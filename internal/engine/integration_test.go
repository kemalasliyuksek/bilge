//go:build integration

package engine

import (
	"context"
	"testing"
	"time"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
)

// TestSnapshotLocalEngine reads the local engine and classifies it. It is
// read-only: it never starts, stops or removes anything.
func TestSnapshotLocalEngine(t *testing.T) {
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

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
