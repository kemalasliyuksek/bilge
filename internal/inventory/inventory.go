// Package inventory is the plain data model of one snapshot of a Docker
// Engine. It is deliberately independent of the Docker SDK so that the rest of
// the app, and its tests, never depend on SDK types.
package inventory

import "time"

// Label keys that carry meaning for bilge.
const (
	LabelComposeProject  = "com.docker.compose.project"
	LabelAnonymousVolume = "com.docker.volume.anonymous"
	LabelKeep            = "bilge.keep"
)

// ContainerState mirrors the engine's container states.
type ContainerState string

const (
	StateCreated    ContainerState = "created"
	StateRunning    ContainerState = "running"
	StatePaused     ContainerState = "paused"
	StateRestarting ContainerState = "restarting"
	StateRemoving   ContainerState = "removing"
	StateExited     ContainerState = "exited"
	StateDead       ContainerState = "dead"
)

// Stopped reports whether the container is not running and not in transition.
func (s ContainerState) Stopped() bool {
	return s == StateCreated || s == StateExited || s == StateDead
}

type Container struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	ImageID    string            `json:"imageId"`
	State      ContainerState    `json:"state"`
	Status     string            `json:"status"`
	Created    time.Time         `json:"created"`
	SizeRw     int64             `json:"sizeRw"`
	Labels     map[string]string `json:"labels"`
	Volumes    []string          `json:"volumes"`    // names of mounted volumes
	NetworkIDs []string          `json:"networkIds"` // IDs of attached networks
}

type Image struct {
	ID         string            `json:"id"`
	Tags       []string          `json:"tags"`
	Created    time.Time         `json:"created"`
	Size       int64             `json:"size"`
	SharedSize int64             `json:"sharedSize"` // -1 when the engine did not compute it
	Labels     map[string]string `json:"labels"`
}

type Volume struct {
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Created time.Time         `json:"created"`
	Size    int64             `json:"size"` // -1 when the engine did not compute it
	Labels  map[string]string `json:"labels"`
}

type Network struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Driver   string            `json:"driver"`
	Scope    string            `json:"scope"`
	Ingress  bool              `json:"ingress"`
	Created  time.Time         `json:"created"`
	Labels   map[string]string `json:"labels"`
	Internal bool              `json:"internal"`
}

type BuildCacheRecord struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	InUse      bool       `json:"inUse"`
	Shared     bool       `json:"shared"`
	Size       int64      `json:"size"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

// EngineInfo describes the daemon itself, for the overview screen.
type EngineInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	OS            string `json:"os"`
	OSType        string `json:"osType"`
	Architecture  string `json:"architecture"`
	KernelVersion string `json:"kernelVersion"`
	StorageDriver string `json:"storageDriver"`
	RootDir       string `json:"rootDir"`
	CPUs          int    `json:"cpus"`
	MemoryBytes   int64  `json:"memoryBytes"`
}

// Inventory is everything bilge knows about the engine at one point in time.
type Inventory struct {
	TakenAt    time.Time          `json:"takenAt"`
	Engine     EngineInfo         `json:"engine"`
	Containers []Container        `json:"containers"`
	Images     []Image            `json:"images"`
	Volumes    []Volume           `json:"volumes"`
	Networks   []Network          `json:"networks"`
	BuildCache []BuildCacheRecord `json:"buildCache"`
}
