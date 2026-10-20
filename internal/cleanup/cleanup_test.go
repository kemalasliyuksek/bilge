package cleanup

import (
	"testing"

	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

func find(t *testing.T, r Report, kind Kind, id string) (Candidate, bool) {
	t.Helper()
	for _, c := range r.Candidates {
		if c.Kind == kind && c.ID == id {
			return c, true
		}
	}
	return Candidate{}, false
}

func expectVerdict(t *testing.T, r Report, kind Kind, id string, want Verdict) {
	t.Helper()
	c, ok := find(t, r, kind, id)
	if !ok {
		t.Fatalf("%s %q: not a candidate, want %s", kind, id, want)
	}
	if c.Verdict != want {
		t.Fatalf("%s %q: verdict %s, want %s (%s)", kind, id, c.Verdict, want, c.Reason)
	}
}

func expectAbsent(t *testing.T, r Report, kind Kind, id string) {
	t.Helper()
	if c, ok := find(t, r, kind, id); ok {
		t.Fatalf("%s %q: unexpected candidate (%s: %s)", kind, id, c.Verdict, c.Reason)
	}
}

func compose(project string) map[string]string {
	return map[string]string{inventory.LabelComposeProject: project}
}

func TestContainers(t *testing.T) {
	inv := inventory.Inventory{Containers: []inventory.Container{
		{ID: "run", State: inventory.StateRunning},
		{ID: "paused", State: inventory.StatePaused},
		{ID: "exited", State: inventory.StateExited},
		{ID: "dead", State: inventory.StateDead},
		{ID: "stack", State: inventory.StateExited, Labels: compose("shop")},
		{ID: "keep", State: inventory.StateExited, Labels: map[string]string{inventory.LabelKeep: "true"}},
	}}
	r := Classify(inv)

	expectAbsent(t, r, KindContainer, "run")
	expectAbsent(t, r, KindContainer, "paused")
	expectAbsent(t, r, KindContainer, "keep")
	expectVerdict(t, r, KindContainer, "exited", Safe)
	expectVerdict(t, r, KindContainer, "dead", Safe)
	expectVerdict(t, r, KindContainer, "stack", Review)
}

func TestImages(t *testing.T) {
	inv := inventory.Inventory{
		Containers: []inventory.Container{
			{ID: "c1", State: inventory.StateRunning, ImageID: "sha:running"},
			{ID: "c2", State: inventory.StateExited, ImageID: "sha:stopped"},
		},
		Images: []inventory.Image{
			{ID: "sha:running", Tags: []string{"app:1"}},
			{ID: "sha:stopped", Tags: []string{"app:0"}},
			{ID: "sha:dangling", Size: 100, SharedSize: 30},
			{ID: "sha:none-tag", Tags: []string{"<none>:<none>"}, Size: 50, SharedSize: -1},
			{ID: "sha:tagged", Tags: []string{"postgres:18"}},
		},
	}
	r := Classify(inv)

	// A stopped container still pins its image: removing the image would
	// make the container impossible to start again.
	expectAbsent(t, r, KindImage, "sha:running")
	expectAbsent(t, r, KindImage, "sha:stopped")
	expectVerdict(t, r, KindImage, "sha:dangling", Safe)
	expectVerdict(t, r, KindImage, "sha:none-tag", Safe)
	expectVerdict(t, r, KindImage, "sha:tagged", Review)

	if c, _ := find(t, r, KindImage, "sha:dangling"); c.Size != 70 {
		t.Errorf("dangling image size = %d, want 70 (shared layers excluded)", c.Size)
	}
	if c, _ := find(t, r, KindImage, "sha:none-tag"); c.Size != 50 {
		t.Errorf("image with unknown shared size = %d, want full size 50", c.Size)
	}
}

func TestVolumes(t *testing.T) {
	anon := "3f1c5d0e8a9b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d"
	inv := inventory.Inventory{
		Containers: []inventory.Container{
			{ID: "db", State: inventory.StateRunning, Volumes: []string{"shop_db"}},
		},
		Volumes: []inventory.Volume{
			{Name: "shop_db", Labels: compose("shop")},
			{Name: "old_db", Labels: compose("old")},
			{Name: "labelled-anon", Labels: map[string]string{inventory.LabelAnonymousVolume: ""}},
			{Name: anon},
		},
	}
	r := Classify(inv)

	expectAbsent(t, r, KindVolume, "shop_db")
	expectVerdict(t, r, KindVolume, "old_db", Review)
	expectVerdict(t, r, KindVolume, "labelled-anon", Safe)
	expectVerdict(t, r, KindVolume, anon, Safe)
}

func TestNetworks(t *testing.T) {
	inv := inventory.Inventory{
		Containers: []inventory.Container{
			{ID: "c", State: inventory.StateExited, NetworkIDs: []string{"n-used"}},
		},
		Networks: []inventory.Network{
			{ID: "n-bridge", Name: "bridge"},
			{ID: "n-host", Name: "host"},
			{ID: "n-none", Name: "none"},
			{ID: "n-swarm", Name: "overlay", Scope: "swarm"},
			{ID: "n-ingress", Name: "ingress", Ingress: true},
			{ID: "n-used", Name: "shop_default"},
			{ID: "n-free", Name: "old_default"},
		},
	}
	r := Classify(inv)

	for _, id := range []string{"n-bridge", "n-host", "n-none", "n-swarm", "n-ingress", "n-used"} {
		expectAbsent(t, r, KindNetwork, id)
	}
	expectVerdict(t, r, KindNetwork, "n-free", Safe)
}

func TestBuildCache(t *testing.T) {
	r := Classify(inventory.Inventory{BuildCache: []inventory.BuildCacheRecord{
		{ID: "a", Size: 10},
		{ID: "b", Size: 20, Shared: true},
		{ID: "c", Size: 40, InUse: true},
	}})
	c, ok := find(t, r, KindBuildCache, "build-cache")
	if !ok {
		t.Fatal("build cache: no candidate")
	}
	if c.Size != 30 {
		t.Errorf("build cache size = %d, want 30 (in-use records excluded)", c.Size)
	}

	r = Classify(inventory.Inventory{BuildCache: []inventory.BuildCacheRecord{{ID: "c", InUse: true}}})
	expectAbsent(t, r, KindBuildCache, "build-cache")
}

func TestSafeSizeIgnoresReviewAndUnknown(t *testing.T) {
	r := Report{Candidates: []Candidate{
		{Verdict: Safe, Size: 100},
		{Verdict: Safe, Size: -1},
		{Verdict: Review, Size: 1000},
	}}
	if got := r.SafeSize(); got != 100 {
		t.Errorf("SafeSize = %d, want 100", got)
	}
}
