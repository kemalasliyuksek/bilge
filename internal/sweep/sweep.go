// Package sweep removes the candidates a user confirmed. It does not trust the
// request: every item is checked against a fresh scan before anything is
// removed. The rules are in docs/design.md under "What the backend enforces".
package sweep

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
	"github.com/kemalasliyuksek/bilge/internal/engine"
	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

// ErrReviewNotAlone rejects a request that mixes a review item with others.
var ErrReviewNotAlone = errors.New("a review item must be removed on its own")

// Request is one item the user confirmed, with the verdict they saw.
type Request struct {
	Kind    cleanup.Kind    `json:"kind"`
	ID      string          `json:"id"`
	Verdict cleanup.Verdict `json:"verdict"`
}

type Status string

const (
	Removed Status = "removed"
	Skipped Status = "skipped"
	Failed  Status = "failed"
)

type Result struct {
	Kind    cleanup.Kind `json:"kind"`
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Status  Status       `json:"status"`
	Message string       `json:"message"`
	Freed   int64        `json:"freed"` // expected space, from the fresh scan
}

// order removes containers first: they are what pins images, volumes and
// networks. Build cache goes last because its prune is the slowest call.
var order = map[cleanup.Kind]int{
	cleanup.KindContainer:  0,
	cleanup.KindImage:      1,
	cleanup.KindVolume:     2,
	cleanup.KindNetwork:    3,
	cleanup.KindBuildCache: 4,
}

// Plan checks reqs against a fresh report. It returns the candidates to
// remove in dependency order and a skipped result for every request that is
// no longer valid.
func Plan(report cleanup.Report, reqs []Request) ([]cleanup.Candidate, []Result, error) {
	for _, r := range reqs {
		if r.Verdict == cleanup.Review && len(reqs) > 1 {
			return nil, nil, ErrReviewNotAlone
		}
	}

	current := make(map[string]cleanup.Candidate, len(report.Candidates))
	for _, c := range report.Candidates {
		current[key(c.Kind, c.ID)] = c
	}

	var todo []cleanup.Candidate
	var skipped []Result
	seen := map[string]bool{}
	for _, r := range reqs {
		k := key(r.Kind, r.ID)
		if seen[k] {
			continue
		}
		seen[k] = true

		c, ok := current[k]
		switch {
		case !ok:
			skipped = append(skipped, Result{Kind: r.Kind, ID: r.ID, Name: r.ID, Status: Skipped,
				Message: "no longer a cleanup candidate; it is in use or already gone"})
		case c.Verdict != r.Verdict:
			skipped = append(skipped, Result{Kind: r.Kind, ID: r.ID, Name: c.Name, Status: Skipped,
				Message: fmt.Sprintf("changed since the scan: now %s, was %s", c.Verdict, r.Verdict)})
		default:
			todo = append(todo, c)
		}
	}
	sort.SliceStable(todo, func(i, j int) bool { return order[todo[i].Kind] < order[todo[j].Kind] })
	return todo, skipped, nil
}

func key(kind cleanup.Kind, id string) string {
	return string(kind) + ":" + id
}

type remover interface {
	Snapshot(context.Context) (inventory.Inventory, error)
	Remove(context.Context, cleanup.Kind, string) error
}

type Sweeper struct {
	engine remover
}

func New(e remover) *Sweeper {
	return &Sweeper{engine: e}
}

// Run removes the confirmed items and reports each result through progress
// as soon as it is known. It returns an error only when nothing could be
// attempted (scan failed, invalid request); per-item problems are results.
func (s *Sweeper) Run(ctx context.Context, reqs []Request, progress func(Result)) ([]Result, error) {
	if progress == nil {
		progress = func(Result) {}
	}
	inv, err := s.engine.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("rescan before removal: %w", err)
	}
	todo, skipped, err := Plan(cleanup.Classify(inv), reqs)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(skipped)+len(todo))
	for _, r := range skipped {
		progress(r)
		results = append(results, r)
	}
	for _, c := range todo {
		r := Result{Kind: c.Kind, ID: c.ID, Name: c.Name}
		switch err := s.engine.Remove(ctx, c.Kind, c.ID); {
		case err == nil:
			r.Status = Removed
			r.Freed = max(c.Size, 0)
		case errors.Is(err, engine.ErrInUse):
			r.Status = Skipped
			r.Message = "now in use"
		default:
			r.Status = Failed
			r.Message = err.Error()
		}
		progress(r)
		results = append(results, r)
	}
	return results, nil
}
