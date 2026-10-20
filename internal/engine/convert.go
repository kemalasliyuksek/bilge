package engine

import (
	"strings"
	"time"

	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/api/types/volume"

	"github.com/kemalasliyuksek/bilge/internal/inventory"
)

func toContainer(c container.Summary) inventory.Container {
	out := inventory.Container{
		ID:      c.ID,
		Name:    containerName(c.Names),
		Image:   c.Image,
		ImageID: c.ImageID,
		State:   inventory.ContainerState(c.State),
		Status:  c.Status,
		Created: time.Unix(c.Created, 0),
		SizeRw:  c.SizeRw,
		Labels:  c.Labels,
	}
	for _, m := range c.Mounts {
		// Bind mounts and tmpfs are host paths or memory, not engine volumes.
		if m.Type == mount.TypeVolume && m.Name != "" {
			out.Volumes = append(out.Volumes, m.Name)
		}
	}
	if c.NetworkSettings != nil {
		for _, ep := range c.NetworkSettings.Networks {
			if ep != nil && ep.NetworkID != "" {
				out.NetworkIDs = append(out.NetworkIDs, ep.NetworkID)
			}
		}
	}
	return out
}

// containerName drops the leading slash the API keeps for historical reasons.
func containerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

func toImage(i image.Summary) inventory.Image {
	return inventory.Image{
		ID:         i.ID,
		Tags:       i.RepoTags,
		Created:    time.Unix(i.Created, 0),
		Size:       i.Size,
		SharedSize: i.SharedSize,
		Labels:     i.Labels,
	}
}

func toVolume(v volume.Volume) inventory.Volume {
	out := inventory.Volume{
		Name:   v.Name,
		Driver: v.Driver,
		Size:   -1,
		Labels: v.Labels,
	}
	if v.UsageData != nil {
		out.Size = v.UsageData.Size
	}
	if t, err := time.Parse(time.RFC3339, v.CreatedAt); err == nil {
		out.Created = t
	}
	return out
}

func toNetwork(n network.Summary) inventory.Network {
	return inventory.Network{
		ID:       n.ID,
		Name:     n.Name,
		Driver:   n.Driver,
		Scope:    n.Scope,
		Ingress:  n.Ingress,
		Internal: n.Internal,
		Created:  n.Created,
		Labels:   n.Labels,
	}
}

func toBuildCacheRecord(r build.CacheRecord) inventory.BuildCacheRecord {
	return inventory.BuildCacheRecord{
		ID:         r.ID,
		Type:       r.Type,
		InUse:      r.InUse,
		Shared:     r.Shared,
		Size:       r.Size,
		LastUsedAt: r.LastUsedAt,
	}
}

func toEngineInfo(i system.Info, apiVersion string) inventory.EngineInfo {
	return inventory.EngineInfo{
		Name:          i.Name,
		Version:       i.ServerVersion,
		APIVersion:    apiVersion,
		OS:            i.OperatingSystem,
		OSType:        i.OSType,
		Architecture:  i.Architecture,
		KernelVersion: i.KernelVersion,
		StorageDriver: i.Driver,
		RootDir:       i.DockerRootDir,
		CPUs:          i.NCPU,
		MemoryBytes:   i.MemTotal,
	}
}
