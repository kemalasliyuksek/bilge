// Package engine is the only part of bilge that talks to the Docker Engine.
// It turns API responses into an inventory.Inventory and carries out the
// actions the user confirmed in the UI.
package engine

import (
	"context"
	"errors"
	"fmt"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/client"

	"github.com/kemalasliyuksek/bilge/internal/cleanup"
	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

// ErrInUse means the engine refused a removal because the resource became
// used after the last scan. It is a skip, not a failure.
var ErrInUse = errors.New("in use")

// api is the subset of *client.Client that bilge uses. Keeping it narrow lets
// tests replace the engine with a small fake.
type api interface {
	ContainerList(context.Context, client.ContainerListOptions) (client.ContainerListResult, error)
	DiskUsage(context.Context, client.DiskUsageOptions) (client.DiskUsageResult, error)
	NetworkList(context.Context, client.NetworkListOptions) (client.NetworkListResult, error)
	ContainerStart(context.Context, string, client.ContainerStartOptions) (client.ContainerStartResult, error)
	ContainerStop(context.Context, string, client.ContainerStopOptions) (client.ContainerStopResult, error)
	ContainerRestart(context.Context, string, client.ContainerRestartOptions) (client.ContainerRestartResult, error)
	ContainerRemove(context.Context, string, client.ContainerRemoveOptions) (client.ContainerRemoveResult, error)
	ImageInspect(context.Context, string, ...client.ImageInspectOption) (client.ImageInspectResult, error)
	ImageRemove(context.Context, string, client.ImageRemoveOptions) (client.ImageRemoveResult, error)
	VolumeRemove(context.Context, string, client.VolumeRemoveOptions) (client.VolumeRemoveResult, error)
	NetworkRemove(context.Context, string, client.NetworkRemoveOptions) (client.NetworkRemoveResult, error)
	BuildCachePrune(context.Context, client.BuildCachePruneOptions) (client.BuildCachePruneResult, error)
	Info(context.Context, client.InfoOptions) (client.SystemInfoResult, error)
	ClientVersion() string
	Close() error
}

type Engine struct {
	api api
}

// New connects to the engine named by DOCKER_HOST, or the platform default
// (the Unix socket on Linux and macOS, the named pipe on Windows). The API
// version is negotiated on the first request.
func New() (*Engine, error) {
	c, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &Engine{api: c}, nil
}

func (e *Engine) Close() error {
	return e.api.Close()
}

// Snapshot reads containers, images, volumes, networks and build cache.
// Containers include stopped ones: they still pin images, volumes and networks.
func (e *Engine) Snapshot(ctx context.Context) (inventory.Inventory, error) {
	cs, err := e.api.ContainerList(ctx, client.ContainerListOptions{All: true, Size: true})
	if err != nil {
		return inventory.Inventory{}, fmt.Errorf("list containers: %w", err)
	}
	// Verbose is required for the per-item lists; without it the engine
	// returns only totals.
	du, err := e.api.DiskUsage(ctx, client.DiskUsageOptions{
		Images: true, Volumes: true, BuildCache: true, Verbose: true,
	})
	if err != nil {
		return inventory.Inventory{}, fmt.Errorf("read disk usage: %w", err)
	}
	ns, err := e.api.NetworkList(ctx, client.NetworkListOptions{})
	if err != nil {
		return inventory.Inventory{}, fmt.Errorf("list networks: %w", err)
	}
	info, err := e.api.Info(ctx, client.InfoOptions{})
	if err != nil {
		return inventory.Inventory{}, fmt.Errorf("read engine info: %w", err)
	}
	return inventory.Inventory{
		// ClientVersion is the negotiated API version once a request has run.
		Engine:     toEngineInfo(info.Info, e.api.ClientVersion()),
		Containers: mapSlice(cs.Items, toContainer),
		Images:     mapSlice(du.Images.Items, toImage),
		Volumes:    mapSlice(du.Volumes.Items, toVolume),
		Networks:   mapSlice(ns.Items, toNetwork),
		BuildCache: mapSlice(du.BuildCache.Items, toBuildCacheRecord),
	}, nil
}

func (e *Engine) StartContainer(ctx context.Context, id string) error {
	_, err := e.api.ContainerStart(ctx, id, client.ContainerStartOptions{})
	return wrap("start container", id, err)
}

// StopContainer uses the container's own stop timeout, so services that need
// time to shut down cleanly get it.
func (e *Engine) StopContainer(ctx context.Context, id string) error {
	_, err := e.api.ContainerStop(ctx, id, client.ContainerStopOptions{})
	return wrap("stop container", id, err)
}

func (e *Engine) RestartContainer(ctx context.Context, id string) error {
	_, err := e.api.ContainerRestart(ctx, id, client.ContainerRestartOptions{})
	return wrap("restart container", id, err)
}

// Remove deletes one resource. It never forces: a resource that became used
// since the scan is left alone and reported as ErrInUse. A resource that is
// already gone counts as removed.
func (e *Engine) Remove(ctx context.Context, kind cleanup.Kind, id string) error {
	var err error
	switch kind {
	case cleanup.KindContainer:
		_, err = e.api.ContainerRemove(ctx, id, client.ContainerRemoveOptions{})
	case cleanup.KindImage:
		err = e.removeImage(ctx, id)
	case cleanup.KindVolume:
		_, err = e.api.VolumeRemove(ctx, id, client.VolumeRemoveOptions{})
	case cleanup.KindNetwork:
		_, err = e.api.NetworkRemove(ctx, id, client.NetworkRemoveOptions{})
	case cleanup.KindBuildCache:
		// All selects every record that is not in use, which is exactly what
		// cleanup counted; records in use are never pruned by the engine.
		_, err = e.api.BuildCachePrune(ctx, client.BuildCachePruneOptions{All: true})
	default:
		return fmt.Errorf("remove %s: unknown kind %q", id, kind)
	}
	if cerrdefs.IsNotFound(err) {
		return nil
	}
	return wrap("remove "+string(kind), id, err)
}

// removeImage removes a tagged image tag by tag. Removing it by ID would need
// Force when it has more than one tag, and Force would also untag an image a
// container started using in the meantime. Untagging the last tag deletes the
// image itself.
func (e *Engine) removeImage(ctx context.Context, id string) error {
	img, err := e.api.ImageInspect(ctx, id)
	if err != nil {
		return err
	}
	refs := img.RepoTags
	if len(refs) == 0 {
		refs = []string{id}
	}
	for _, ref := range refs {
		opts := client.ImageRemoveOptions{PruneChildren: true}
		if _, err := e.api.ImageRemove(ctx, ref, opts); err != nil && !cerrdefs.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func wrap(op, id string, err error) error {
	if err == nil {
		return nil
	}
	if cerrdefs.IsConflict(err) {
		return fmt.Errorf("%s %s: %w: %w", op, id, ErrInUse, err)
	}
	return fmt.Errorf("%s %s: %w", op, id, err)
}

func mapSlice[T, U any](in []T, f func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}
