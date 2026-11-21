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

func namedVol(name string) cleanup.Candidate {
	c := review(cleanup.KindVolume, name)
	c.ConfirmByName = true
	return c
}

func TestPlanAllowsReviewItemsInBulk(t *testing.T) {
	img := review(cleanup.KindImage, "sha:old")
	ctr := review(cleanup.KindContainer, "shop-web-1")
	net := safe(cleanup.KindNetwork, "n1")

	todo, skipped, err := Plan(report(img, ctr, net), []Request{req(img), req(ctr), req(net)})
	if err != nil {
		t.Fatal(err)
	}
	if len(todo) != 3 || len(skipped) != 0 {
		t.Errorf("todo = %v, skipped = %v, want all three removed", todo, skipped)
	}
}

func TestPlanRejectsNamedVolumeWithOthers(t *testing.T) {
	vol := namedVol("shop_db")
	img := review(cleanup.KindImage, "sha:old")
	if _, _, err := Plan(report(vol, img), []Request{req(img), req(vol)}); !errors.Is(err, ErrNamedVolumeNotAlone) {
		t.Fatalf("err = %v, want ErrNamedVolumeNotAlone", err)
	}

	// The request is rejected as a whole even when the volume is no longer a
	// candidate: a UI that bundles a named volume is breaking the rule.
	gone := Request{Kind: cleanup.KindVolume, ID: "gone_db", Verdict: cleanup.Review}
	if _, _, err := Plan(report(img), []Request{req(img), gone}); !errors.Is(err, ErrNamedVolumeNotAlone) {
		t.Fatalf("stale named volume: err = %v, want ErrNamedVolumeNotAlone", err)
	}

	// A request that hides a named volume behind a safe verdict is caught by
	// the fresh report, not by what the request claims.
	disguised := Request{Kind: cleanup.KindVolume, ID: "shop_db", Verdict: cleanup.Safe}
	if _, _, err := Plan(report(vol, img), []Request{req(img), disguised}); !errors.Is(err, ErrNamedVolumeNotAlone) {
		t.Fatalf("disguised named volume: err = %v, want ErrNamedVolumeNotAlone", err)
	}

	todo, _, err := Plan(report(vol), []Request{req(vol)})
	if err != nil || len(todo) != 1 {
		t.Fatalf("named volume alone: todo = %v, err = %v", todo, err)
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
