// Package cleanup decides which resources in an inventory are waste. It is
// pure: no I/O, no clock, no Docker SDK. The rules are documented in
// docs/design.md and each one is covered by a test.
package cleanup

import (
	"fmt"
	"regexp"

	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

type Kind string

const (
	KindContainer  Kind = "container"
	KindImage      Kind = "image"
	KindVolume     Kind = "volume"
	KindNetwork    Kind = "network"
	KindBuildCache Kind = "buildCache"
)

type Verdict string

const (
	// Safe candidates lose no data and undo no deliberate setup; they may be
	// removed in bulk after one confirmation.
	Safe Verdict = "safe"
	// Review candidates are probably waste but can hold data or belong to a
	// stack stopped on purpose; they are removed one by one.
	Review Verdict = "review"
)

type Candidate struct {
	Kind    Kind    `json:"kind"`
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Verdict Verdict `json:"verdict"`
	Reason  string  `json:"reason"`
	Project string  `json:"project"` // Compose project, empty when none
	Size    int64   `json:"size"`    // bytes expected to be freed, -1 when unknown
}

type Report struct {
	Candidates []Candidate `json:"candidates"`
}

// SafeSize is the space freed by removing every safe candidate with a known size.
func (r Report) SafeSize() int64 {
	var total int64
	for _, c := range r.Candidates {
		if c.Verdict == Safe && c.Size > 0 {
			total += c.Size
		}
	}
	return total
}

// Classify returns the removal candidates in inv. Resources that are in use,
// predefined by the engine or labelled bilge.keep=true never appear.
func Classify(inv inventory.Inventory) Report {
	u := collectUsage(inv.Containers)
	var out []Candidate
	out = append(out, containers(inv.Containers)...)
	out = append(out, images(inv.Images, u)...)
	out = append(out, volumes(inv.Volumes, u)...)
	out = append(out, networks(inv.Networks, u)...)
	if c, ok := buildCache(inv.BuildCache); ok {
		out = append(out, c)
	}
	return Report{Candidates: out}
}

// usage is derived from all containers, running and stopped, because a
// stopped container still pins its image, volumes and networks.
type usage struct {
	images   map[string]bool
	volumes  map[string]bool
	networks map[string]bool
}

func collectUsage(cs []inventory.Container) usage {
	u := usage{images: map[string]bool{}, volumes: map[string]bool{}, networks: map[string]bool{}}
	for _, c := range cs {
		u.images[c.ImageID] = true
		for _, v := range c.Volumes {
			u.volumes[v] = true
		}
		for _, n := range c.NetworkIDs {
			u.networks[n] = true
		}
	}
	return u
}

func kept(labels map[string]string) bool {
	return labels[inventory.LabelKeep] == "true"
}

func containers(cs []inventory.Container) []Candidate {
	var out []Candidate
	for _, c := range cs {
		if !c.State.Stopped() || kept(c.Labels) {
			continue
		}
		cand := Candidate{
			Kind:    KindContainer,
			ID:      c.ID,
			Name:    c.Name,
			Project: c.Labels[inventory.LabelComposeProject],
			Size:    c.SizeRw,
			Verdict: Safe,
			Reason:  "stopped container",
		}
		if cand.Project != "" {
			cand.Verdict = Review
			cand.Reason = "stopped container of Compose project " + cand.Project + "; the stack may be stopped on purpose"
		}
		out = append(out, cand)
	}
	return out
}

func images(is []inventory.Image, u usage) []Candidate {
	var out []Candidate
	for _, i := range is {
		if u.images[i.ID] || kept(i.Labels) {
			continue
		}
		cand := Candidate{
			Kind:    KindImage,
			ID:      i.ID,
			Name:    imageName(i),
			Project: i.Labels[inventory.LabelComposeProject],
			Size:    imageReclaim(i),
		}
		if dangling(i) {
			cand.Verdict = Safe
			cand.Reason = "dangling image, no tag and no container"
		} else {
			cand.Verdict = Review
			cand.Reason = "tagged image used by no container; it can be pulled or built again"
		}
		out = append(out, cand)
	}
	return out
}

func dangling(i inventory.Image) bool {
	for _, t := range i.Tags {
		if t != "<none>:<none>" {
			return false
		}
	}
	return true
}

func imageName(i inventory.Image) string {
	if dangling(i) {
		return "<none>"
	}
	return i.Tags[0]
}

// imageReclaim excludes layers shared with other images, which stay on disk.
func imageReclaim(i inventory.Image) int64 {
	if i.SharedSize < 0 {
		return i.Size
	}
	return i.Size - i.SharedSize
}

var hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

func anonymous(v inventory.Volume) bool {
	if _, ok := v.Labels[inventory.LabelAnonymousVolume]; ok {
		return true
	}
	// Engines older than Docker 23 do not set the label; their anonymous
	// volumes are recognisable only by the generated name.
	return hex64.MatchString(v.Name)
}

func volumes(vs []inventory.Volume, u usage) []Candidate {
	var out []Candidate
	for _, v := range vs {
		if u.volumes[v.Name] || kept(v.Labels) {
			continue
		}
		cand := Candidate{
			Kind:    KindVolume,
			ID:      v.Name,
			Name:    v.Name,
			Project: v.Labels[inventory.LabelComposeProject],
			Size:    v.Size,
		}
		if anonymous(v) {
			cand.Verdict = Safe
			cand.Reason = "anonymous volume mounted by no container"
		} else {
			cand.Verdict = Review
			cand.Reason = "named volume mounted by no container; it may hold data"
		}
		out = append(out, cand)
	}
	return out
}

var predefinedNetworks = map[string]bool{"bridge": true, "host": true, "none": true}

func networks(ns []inventory.Network, u usage) []Candidate {
	var out []Candidate
	for _, n := range ns {
		if predefinedNetworks[n.Name] || n.Scope == "swarm" || n.Ingress {
			continue
		}
		if u.networks[n.ID] || kept(n.Labels) {
			continue
		}
		out = append(out, Candidate{
			Kind:    KindNetwork,
			ID:      n.ID,
			Name:    n.Name,
			Project: n.Labels[inventory.LabelComposeProject],
			Verdict: Safe,
			Reason:  "user-defined network attached to no container",
		})
	}
	return out
}

// buildCache folds all unused records into one candidate: individual records
// mean nothing to a user, and they are removed together through one prune call.
func buildCache(rs []inventory.BuildCacheRecord) (Candidate, bool) {
	var size int64
	var n int
	for _, r := range rs {
		if r.InUse {
			continue
		}
		n++
		size += r.Size
	}
	if n == 0 {
		return Candidate{}, false
	}
	return Candidate{
		Kind:    KindBuildCache,
		ID:      "build-cache",
		Name:    "build cache",
		Size:    size,
		Verdict: Safe,
		Reason:  fmt.Sprintf("%d build cache records not in use", n),
	}, true
}
