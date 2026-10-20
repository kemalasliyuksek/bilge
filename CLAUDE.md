# bilge

Cross-platform desktop app (Go + Wails v2, Svelte 5 + TypeScript frontend) that
lists Docker Engine resources and removes waste safely. Personal open source
project, MIT. Design and cleanup rules: `docs/design.md`.

## Overrides of the global rules

- Everything in the repository is English: code, comments, docs, commit
  messages. Commits use `/commit-en` (Conventional Commits), not `/commit-tr`.
- Commit identity is the personal global one (`kemal@kemalasliyuksek.com`).
  This is not a brand product: no brand name or logo anywhere.
- Go tests live next to the code (`foo_test.go`), frontend tests next to the
  module (`foo.test.ts`). There is no `src/` or `tests/` directory.

## Layout

- `main.go`, `app.go`: Wails entry point and the methods bound to the UI.
- `internal/inventory`: SDK-independent data model of one engine snapshot.
- `internal/cleanup`: pure classification, Inventory -> Report. No I/O.
- `internal/sweep`: validates a removal request against a fresh scan and
  removes in dependency order.
- `internal/engine`: the only package that imports
  `github.com/moby/moby/client`. Snapshot and actions; never forces a removal.
- `frontend/wailsjs`: generated bindings, regenerate with `just bindings`.

## Commands

`just --list` shows everything. `just check` is the quality gate.

## Rules specific to this project

- A cleanup rule change starts in `docs/design.md`, then a test in
  `internal/cleanup`, then the code.
- Never call the engine's prune endpoints for containers, images, volumes or
  networks; remove item by item so bilge, not the engine, decides what goes.
- Linux builds need `-tags webkit2_41` (Ubuntu ships only WebKitGTK 4.1); the
  justfile adds it. `wails doctor` wrongly reports libwebkit as missing.
- Unit tests never touch a real Docker Engine; `internal/engine` tests use a
  fake of its narrow `api` interface. The only exception is
  `just test-integration` (build tag `integration`), which is read-only.
- Removal goes through `internal/sweep`, which rescans and re-classifies
  before removing; never call `engine.Remove` from the app layer directly.
- Manual removal tests use resources created for the test (`bilge-test-*`),
  never the developer's own containers, volumes or build cache.
