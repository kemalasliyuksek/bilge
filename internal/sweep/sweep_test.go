package sweep

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
	"github.com/kemalasliyuksek/bilge/internal/engine"
	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

func report(cs ...cleanup.Candidate) cleanup.Report {
	return cleanup.Report{Candidates: cs}
}

func safe(kind cleanup.Kind, id string) cleanup.Candidate {
	return cleanup.Candidate{Kind: kind, ID: id, Name: id, Verdict: cleanup.Safe, Size: 10}
}

func review(kind cleanup.Kind, id string) cleanup.Candidate {
	return cleanup.Candidate{Kind: kind, ID: id, Name: id, Verdict: cleanup.Review, Size: 10}
}

func req(c cleanup.Candidate) Request {
	return Request{Kind: c.Kind, ID: c.ID, Verdict: c.Verdict}
}

func TestPlanRejectsReviewWithOthers(t *testing.T) {
	vol := review(cleanup.KindVolume, "shop_db")
	img := safe(cleanup.KindImage, "sha:1")
	_, _, err := Plan(report(vol, img), []Request{req(img), req(vol)})
	if !errors.Is(err, ErrReviewNotAlone) {
		t.Fatalf("err = %v, want ErrReviewNotAlone", err)
	}

	todo, _, err := Plan(report(vol), []Request{req(vol)})
	if err != nil || len(todo) != 1 {
		t.Fatalf("single review item: todo = %v, err = %v", todo, err)
	}
}

func TestPlanSkipsStaleRequests(t *testing.T) {
	// The UI saw a safe anonymous volume; meanwhile a container mounted it.
	gone := Request{Kind: cleanup.KindVolume, ID: "anon", Verdict: cleanup.Safe}
	// The UI saw a safe container; it now belongs to a Compose project.
	changed := Request{Kind: cleanup.KindContainer, ID: "c1", Verdict: cleanup.Safe}
	ok := safe(cleanup.KindNetwork, "n1")

	todo, skipped, err := Plan(report(review(cleanup.KindContainer, "c1"), ok), []Request{gone, changed, req(ok)})
	if err != nil {
		t.Fatal(err)
	}
	if len(todo) != 1 || todo[0].ID != "n1" {
		t.Errorf("todo = %v, want only n1", todo)
	}
	if len(skipped) != 2 || skipped[0].ID != "anon" || skipped[1].ID != "c1" {
		t.Fatalf("skipped = %v, want anon and c1", skipped)
	}
	for _, s := range skipped {
		if s.Status != Skipped || s.Message == "" {
			t.Errorf("skipped result = %+v, want status skipped with a reason", s)
		}
	}
}

func TestPlanOrdersByDependencyAndDeduplicates(t *testing.T) {
	cache := safe(cleanup.KindBuildCache, "build-cache")
	net := safe(cleanup.KindNetwork, "n1")
	img := safe(cleanup.KindImage, "sha:1")
	ctr := safe(cleanup.KindContainer, "c1")
	vol := safe(cleanup.KindVolume, "anon")

	todo, _, err := Plan(report(cache, net, img, ctr, vol),
		[]Request{req(cache), req(net), req(img), req(img), req(vol), req(ctr)})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, c := range todo {
		got = append(got, c.ID)
	}
	want := []string{"c1", "sha:1", "anon", "n1", "build-cache"}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

type fakeEngine struct {
	inv     inventory.Inventory
	scanErr error
	errs    map[string]error
	removed []string
}

func (f *fakeEngine) Snapshot(context.Context) (inventory.Inventory, error) {
	return f.inv, f.scanErr
}

func (f *fakeEngine) Remove(_ context.Context, kind cleanup.Kind, id string) error {
	f.removed = append(f.removed, string(kind)+":"+id)
	return f.errs[id]
}

func TestRun(t *testing.T) {
	// Three stopped standalone containers: all safe candidates.
	f := &fakeEngine{
		inv: inventory.Inventory{Containers: []inventory.Container{
			{ID: "ok", Name: "ok", State: inventory.StateExited, SizeRw: 7},
			{ID: "busy", Name: "busy", State: inventory.StateExited},
			{ID: "broken", Name: "broken", State: inventory.StateExited},
		}},
		errs: map[string]error{
			"busy":   fmt.Errorf("remove container busy: %w", engine.ErrInUse),
			"broken": errors.New("driver failed"),
		},
	}
	reqs := []Request{
		{Kind: cleanup.KindContainer, ID: "ok", Verdict: cleanup.Safe},
		{Kind: cleanup.KindContainer, ID: "busy", Verdict: cleanup.Safe},
		{Kind: cleanup.KindContainer, ID: "broken", Verdict: cleanup.Safe},
		{Kind: cleanup.KindContainer, ID: "vanished", Verdict: cleanup.Safe},
	}

	var streamed []Result
	results, err := New(f).Run(context.Background(), reqs, func(r Result) { streamed = append(streamed, r) })
	if err != nil {
		t.Fatal(err)
	}

	status := map[string]Status{}
	for _, r := range results {
		status[r.ID] = r.Status
	}
	want := map[string]Status{"ok": Removed, "busy": Skipped, "broken": Failed, "vanished": Skipped}
	for id, s := range want {
		if status[id] != s {
			t.Errorf("%s: status %q, want %q", id, status[id], s)
		}
	}
	if len(streamed) != len(results) {
		t.Errorf("progress reported %d results, want %d", len(streamed), len(results))
	}
	if slices.Contains(f.removed, "container:vanished") {
		t.Error("a stale request reached the engine")
	}
	for _, r := range results {
		if r.ID == "ok" && r.Freed != 7 {
			t.Errorf("freed = %d, want 7", r.Freed)
		}
	}
}

func TestRunNeverRemovesWhatIsInUse(t *testing.T) {
	// The request claims a running container is safe; the fresh scan disagrees.
	f := &fakeEngine{inv: inventory.Inventory{Containers: []inventory.Container{
		{ID: "db", State: inventory.StateRunning},
	}}}
	results, err := New(f).Run(context.Background(),
		[]Request{{Kind: cleanup.KindContainer, ID: "db", Verdict: cleanup.Safe}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.removed) != 0 {
		t.Fatalf("engine removed %v", f.removed)
	}
	if results[0].Status != Skipped {
		t.Errorf("status = %q, want skipped", results[0].Status)
	}
}

func TestRunStopsWhenTheScanFails(t *testing.T) {
	f := &fakeEngine{scanErr: errors.New("daemon down")}
	_, err := New(f).Run(context.Background(),
		[]Request{{Kind: cleanup.KindContainer, ID: "x", Verdict: cleanup.Safe}}, nil)
	if err == nil {
		t.Fatal("want an error when the rescan fails")
	}
	if len(f.removed) != 0 {
		t.Fatalf("engine removed %v without a fresh scan", f.removed)
	}
}
