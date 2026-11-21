//go:build integration

package engine

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/client"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

// TestSnapshotLocalEngine reads the local engine and classifies it. It is
// read-only: it never starts, stops or removes anything.
func TestSnapshotLocalEngine(t *testing.T) {
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	inv, err := e.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// bridge, host and none exist on every Linux engine.
	if len(inv.Networks) < 3 {
		t.Errorf("networks = %d, want at least the 3 predefined ones", len(inv.Networks))
	}
	t.Logf("containers=%d images=%d volumes=%d networks=%d build-cache=%d",
		len(inv.Containers), len(inv.Images), len(inv.Volumes), len(inv.Networks), len(inv.BuildCache))

	r := cleanup.Classify(inv)
	for _, c := range r.Candidates {
		t.Logf("%-6s %-10s %-40.40s %12d  %s", c.Verdict, c.Kind, c.Name, c.Size, c.Reason)
	}
	t.Logf("safe total: %d bytes", r.SafeSize())
}

// TestDanglingLocalEngine checks that the images cleanup calls safe are exactly
// the ones the engine's own dangling=true filter reports, minus those a
// container uses or bilge.keep protects. cleanup infers dangling from the
// tags alone; this catches an engine or image store where that inference and
// Docker's definition drift apart. The filter is a read, not a prune.
func TestDanglingLocalEngine(t *testing.T) {
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	// The production api interface has no ImageList; bilge never needs it.
	c, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	inv, err := e.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	list, err := c.ImageList(ctx, client.ImageListOptions{Filters: make(client.Filters).Add("dangling", "true")})
	if err != nil {
		t.Fatal(err)
	}

	used := map[string]bool{}
	for _, ct := range inv.Containers {
		used[ct.ImageID] = true
	}
	want := map[string]bool{}
	for _, i := range list.Items {
		if !used[i.ID] && i.Labels[inventory.LabelKeep] != "true" {
			want[i.ID] = true
		}
	}
	got := map[string]bool{}
	for _, cand := range cleanup.Classify(inv).Candidates {
		if cand.Kind == cleanup.KindImage && cand.Verdict == cleanup.Safe {
			got[cand.ID] = true
		}
	}

	for id := range want {
		if !got[id] {
			t.Errorf("image %s: engine says dangling and unused, cleanup does not call it safe", id)
		}
	}
	for id := range got {
		if !want[id] {
			t.Errorf("image %s: cleanup calls it safe, engine does not say dangling", id)
		}
	}
	t.Logf("dangling unused images: engine=%d cleanup=%d", len(want), len(got))
}
