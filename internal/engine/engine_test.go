package engine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

// fakeAPI embeds the interface so that any method a test does not expect
// panics with a nil dereference instead of silently returning zero values.
type fakeAPI struct {
	api

	containers []container.Summary
	du         client.DiskUsageResult
	networks   []network.Summary
	imageTags  map[string][]string

	removeErr error    // returned by every *Remove call
	calls     []string // "<method> <id>" in call order
	duOpts    client.DiskUsageOptions
	listOpts  client.ContainerListOptions
	pruneOpts client.BuildCachePruneOptions
	rmOpts    []any
}

func (f *fakeAPI) ContainerList(_ context.Context, o client.ContainerListOptions) (client.ContainerListResult, error) {
	f.listOpts = o
	return client.ContainerListResult{Items: f.containers}, nil
}

func (f *fakeAPI) DiskUsage(_ context.Context, o client.DiskUsageOptions) (client.DiskUsageResult, error) {
	f.duOpts = o
	return f.du, nil
}

func (f *fakeAPI) NetworkList(context.Context, client.NetworkListOptions) (client.NetworkListResult, error) {
	return client.NetworkListResult{Items: f.networks}, nil
}

func (f *fakeAPI) Info(context.Context, client.InfoOptions) (client.SystemInfoResult, error) {
	return client.SystemInfoResult{Info: system.Info{ServerVersion: "29.9.1", Driver: "overlayfs", NCPU: 32}}, nil
}

func (f *fakeAPI) ClientVersion() string { return "1.56" }

func (f *fakeAPI) ContainerRemove(_ context.Context, id string, o client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
	f.calls = append(f.calls, "ContainerRemove "+id)
	f.rmOpts = append(f.rmOpts, o)
	return client.ContainerRemoveResult{}, f.removeErr
}

func (f *fakeAPI) ImageInspect(_ context.Context, id string, _ ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	tags, ok := f.imageTags[id]
	if !ok {
		return client.ImageInspectResult{}, fmt.Errorf("no such image %s: %w", id, cerrdefs.ErrNotFound)
	}
	var r client.ImageInspectResult
	r.ID = id
	r.RepoTags = tags
	return r, nil
}

func (f *fakeAPI) ImageRemove(_ context.Context, ref string, o client.ImageRemoveOptions) (client.ImageRemoveResult, error) {
	f.calls = append(f.calls, "ImageRemove "+ref)
	f.rmOpts = append(f.rmOpts, o)
	return client.ImageRemoveResult{}, f.removeErr
}

func (f *fakeAPI) VolumeRemove(_ context.Context, id string, o client.VolumeRemoveOptions) (client.VolumeRemoveResult, error) {
	f.calls = append(f.calls, "VolumeRemove "+id)
	f.rmOpts = append(f.rmOpts, o)
	return client.VolumeRemoveResult{}, f.removeErr
}

func (f *fakeAPI) NetworkRemove(_ context.Context, id string, _ client.NetworkRemoveOptions) (client.NetworkRemoveResult, error) {
	f.calls = append(f.calls, "NetworkRemove "+id)
	return client.NetworkRemoveResult{}, f.removeErr
}

func (f *fakeAPI) BuildCachePrune(_ context.Context, o client.BuildCachePruneOptions) (client.BuildCachePruneResult, error) {
	f.calls = append(f.calls, "BuildCachePrune")
	f.pruneOpts = o
	return client.BuildCachePruneResult{}, f.removeErr
}

func TestSnapshot(t *testing.T) {
	lastUsed := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	f := &fakeAPI{
		containers: []container.Summary{{
			ID:      "c1",
			Names:   []string{"/shop-db-1"},
			ImageID: "sha:pg",
			State:   container.StateExited,
			Created: 1_700_000_000,
			SizeRw:  42,
			Labels:  map[string]string{inventory.LabelComposeProject: "shop"},
			Mounts: []container.MountPoint{
				{Type: mount.TypeVolume, Name: "shop_db"},
				{Type: mount.TypeBind, Source: "/srv/conf"},
				{Type: mount.TypeTmpfs},
			},
			NetworkSettings: &container.NetworkSettingsSummary{Networks: map[string]*network.EndpointSettings{
				"shop_default": {NetworkID: "n1"},
				"broken":       nil,
			}},
		}},
		du: client.DiskUsageResult{
			Images: client.ImagesDiskUsage{Items: []image.Summary{
				{ID: "sha:pg", RepoTags: []string{"postgres:18"}, Size: 100, SharedSize: 10},
			}},
			Volumes: client.VolumesDiskUsage{Items: []volume.Volume{
				{Name: "shop_db", CreatedAt: "2026-09-01T10:00:00Z", UsageData: &volume.UsageData{Size: 2048, RefCount: 1}},
				{Name: "nosize", CreatedAt: "garbage"},
			}},
			BuildCache: client.BuildCacheDiskUsage{Items: []build.CacheRecord{
				{ID: "b1", Size: 7, LastUsedAt: &lastUsed},
			}},
		},
		networks: []network.Summary{{Network: network.Network{ID: "n1", Name: "shop_default", Scope: "local"}}},
	}

	inv, err := (&Engine{api: f}).Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !f.listOpts.All || !f.listOpts.Size {
		t.Errorf("container list options = %+v, want All and Size", f.listOpts)
	}
	if !f.duOpts.Verbose || !f.duOpts.Images || !f.duOpts.Volumes || !f.duOpts.BuildCache {
		t.Errorf("disk usage options = %+v, want verbose images, volumes and build cache", f.duOpts)
	}

	c := inv.Containers[0]
	if c.Name != "shop-db-1" {
		t.Errorf("container name = %q, want leading slash removed", c.Name)
	}
	if !slices.Equal(c.Volumes, []string{"shop_db"}) {
		t.Errorf("container volumes = %v, want only the named volume mount", c.Volumes)
	}
	if !slices.Equal(c.NetworkIDs, []string{"n1"}) {
		t.Errorf("container networks = %v, want [n1]", c.NetworkIDs)
	}
	if c.State != inventory.StateExited || c.SizeRw != 42 {
		t.Errorf("container = %+v", c)
	}

	if inv.Images[0].SharedSize != 10 {
		t.Errorf("image shared size = %d, want 10", inv.Images[0].SharedSize)
	}
	if v := inv.Volumes[0]; v.Size != 2048 || v.Created.IsZero() {
		t.Errorf("volume = %+v, want size 2048 and a parsed creation time", v)
	}
	if v := inv.Volumes[1]; v.Size != -1 || !v.Created.IsZero() {
		t.Errorf("volume without usage data = %+v, want size -1 and zero time", v)
	}
	if b := inv.BuildCache[0]; b.Size != 7 || !b.LastUsedAt.Equal(lastUsed) {
		t.Errorf("build cache = %+v", b)
	}
	if eng := inv.Engine; eng.Version != "29.9.1" || eng.APIVersion != "1.56" || eng.StorageDriver != "overlayfs" || eng.CPUs != 32 {
		t.Errorf("engine info = %+v", eng)
	}
	if inv.Networks[0].ID != "n1" {
		t.Errorf("networks = %+v", inv.Networks)
	}
}

func TestRemoveNeverForces(t *testing.T) {
	f := &fakeAPI{imageTags: map[string][]string{"sha:x": nil}}
	e := &Engine{api: f}
	ctx := context.Background()

	for _, k := range []cleanup.Kind{cleanup.KindContainer, cleanup.KindImage, cleanup.KindVolume} {
		id := "sha:x"
		if err := e.Remove(ctx, k, id); err != nil {
			t.Fatalf("remove %s: %v", k, err)
		}
	}
	for _, o := range f.rmOpts {
		switch o := o.(type) {
		case client.ContainerRemoveOptions:
			if o.Force || o.RemoveVolumes {
				t.Errorf("container remove options = %+v, want no force and volumes kept", o)
			}
		case client.ImageRemoveOptions:
			if o.Force {
				t.Errorf("image remove options = %+v, want no force", o)
			}
		case client.VolumeRemoveOptions:
			if o.Force {
				t.Errorf("volume remove options = %+v, want no force", o)
			}
		}
	}
}

func TestRemoveImageByTag(t *testing.T) {
	f := &fakeAPI{imageTags: map[string][]string{
		"sha:tagged":   {"app:1", "app:latest"},
		"sha:dangling": nil,
	}}
	e := &Engine{api: f}

	if err := e.Remove(context.Background(), cleanup.KindImage, "sha:tagged"); err != nil {
		t.Fatal(err)
	}
	if err := e.Remove(context.Background(), cleanup.KindImage, "sha:dangling"); err != nil {
		t.Fatal(err)
	}
	want := []string{"ImageRemove app:1", "ImageRemove app:latest", "ImageRemove sha:dangling"}
	if !slices.Equal(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
}

func TestRemoveConflictIsInUse(t *testing.T) {
	f := &fakeAPI{removeErr: fmt.Errorf("volume is in use: %w", cerrdefs.ErrConflict)}
	err := (&Engine{api: f}).Remove(context.Background(), cleanup.KindVolume, "shop_db")
	if !errors.Is(err, ErrInUse) {
		t.Fatalf("err = %v, want ErrInUse", err)
	}
}

func TestRemoveAlreadyGoneIsSuccess(t *testing.T) {
	f := &fakeAPI{removeErr: fmt.Errorf("no such network: %w", cerrdefs.ErrNotFound)}
	if err := (&Engine{api: f}).Remove(context.Background(), cleanup.KindNetwork, "n1"); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	// An image that vanished before inspect is also already removed.
	if err := (&Engine{api: f}).Remove(context.Background(), cleanup.KindImage, "sha:gone"); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestRemoveBuildCachePrunesAllUnused(t *testing.T) {
	f := &fakeAPI{}
	if err := (&Engine{api: f}).Remove(context.Background(), cleanup.KindBuildCache, "build-cache"); err != nil {
		t.Fatal(err)
	}
	if !f.pruneOpts.All {
		t.Errorf("prune options = %+v, want All", f.pruneOpts)
	}
}

func TestRemoveUnknownKind(t *testing.T) {
	if err := (&Engine{api: &fakeAPI{}}).Remove(context.Background(), "secret", "x"); err == nil {
		t.Fatal("want error for unknown kind")
	}
}
