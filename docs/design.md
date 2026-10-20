# bilge design

## Goal

A small, cross-platform desktop app to monitor, manage and clean up a Docker
Engine: it shows what the engine holds (containers, images, volumes, networks,
build cache) and what it costs on disk, starts and stops containers, and
removes what is truly waste without ever surprising the user by deleting data.

The built-in `docker system prune` and the "Prune" buttons in existing GUIs use
Docker's own definition of reclaimable: anything not referenced right now. That
definition treats a stopped Compose stack's database volume the same as a
leftover anonymous volume. bilge separates the two.

## Non-goals (for now)

- Switching between hosts or contexts in the UI, Swarm, Kubernetes. The engine
  named by `DOCKER_HOST` is used when set.
- Building, pulling or running new containers.
- Log viewers, exec shells, live charts.

## Architecture

```
main.go / app.go        Wails entry point and the methods bound to the UI
internal/engine         talks to the Docker Engine API, produces an Inventory,
                        performs actions (start, stop, remove, prune)
internal/inventory      plain data model of one snapshot of the engine
internal/cleanup        pure classification: Inventory -> Report
frontend/               Svelte + TypeScript UI
```

`cleanup` has no I/O and no dependency on the Docker SDK. Every rule below is a
unit test against a hand-built `Inventory`. `engine` is the only package that
imports `github.com/moby/moby/client`.

Usage is derived from the container list, not from counters returned by the
API. `ImageList` reports `Containers: -1` unless asked for disk usage, and
`NetworkList` does not report attached containers at all. Deriving both from
the containers (running and stopped) keeps a single source of truth.

## Cleanup rules

Each candidate gets a verdict:

- **safe**: no data is lost and nothing the user set up on purpose breaks.
  Can be removed in bulk after one confirmation.
- **review**: probably waste, but removing it can lose data or undo a
  deliberate setup. Shown separately, selected one by one, never in bulk.

Anything not listed as a candidate is in use and is never offered for removal.

| Resource | safe | review | never |
|---|---|---|---|
| Image | dangling (no tag) and used by no container | tagged and used by no container | used by any container, running or stopped |
| Container | stopped (`exited`, `created`, `dead`) and not part of a Compose project | stopped and part of a Compose project | `running`, `paused`, `restarting`, `removing` |
| Volume | anonymous and mounted by no container | named and mounted by no container | mounted by any container |
| Network | user-defined and attached to no container | | `bridge`, `host`, `none`, swarm-scoped, ingress |
| Build cache | records not in use | | records in use |

Any resource labelled `bilge.keep=true` is never a candidate.

A volume is anonymous when it carries the `com.docker.volume.anonymous` label
(set by Docker 23+) or, for older engines, when its name is a 64 character hex
string.

Reclaimable size for an image is `Size - SharedSize`: layers shared with other
images are not freed by removing one of them.

Reclaimable size for the build cache counts only records that are neither in
use nor shared. Shared records hold layers that images also use; pruning drops
the cache record but the data stays on disk. This matches the engine's own
`Reclaimable` figure in `/system/df`.

## Removal

- Removal always starts with a preview: the exact list and the expected space.
- Items are removed one by one through the API, not with the engine's prune
  endpoints, so the engine never decides on its own what goes. Build cache is
  the exception: it is removed through the build cache prune endpoint, limited
  to records that are not in use.
- The engine rejects removing an image or volume that became in use after the
  scan. That error is shown as "skipped, now in use", not as a failure.

### What the backend enforces

The UI is not trusted to have current data or to follow the rules. Before
removing anything the backend:

1. Takes a fresh snapshot and classifies it again.
2. Skips every requested item that is no longer a candidate, or whose verdict
   changed since the user saw it ("skipped, changed since the scan").
3. Rejects a request that contains a review item together with anything else.
   Review items are removed one per request.
4. Removes in dependency order: containers, then images, volumes, networks,
   and the build cache last. Removing a container never makes bilge remove
   something it did not list.

Each item ends as `removed`, `skipped` (with the reason) or `failed` (with the
engine's error). Progress is reported per item while the removal runs.

### What the UI asks

- Safe items: a checkbox per item, one confirmation dialog that lists every
  selected item and the expected space. Everything is selected by default
  except the build cache: removing it loses no data, but the next builds that
  relied on it run from scratch, so it is an explicit choice.
- Review items: no bulk selection. Each has its own "Remove" button and its own
  confirmation dialog showing the reason. A named volume additionally requires
  typing its name, because it is the one kind that can hold irreplaceable data.
