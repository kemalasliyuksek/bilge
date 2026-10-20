# bilge

Cross-platform desktop app (Go + Wails v2, Svelte 5 + TypeScript frontend) that
monitors, manages and cleans up a Docker Engine. Personal open source
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
- `scripts/make_icon.py`: source of every app icon; `just icon` regenerates
  `build/appicon.png`, `build/windows/icon.ico`, `build/linux/bilge.png` and
  the sidebar icon.
- `build/linux/bilge.desktop`: desktop entry shipped in the Linux archive
  (`Exec=bilge`, so the binary must be on `PATH`); CI validates it.
- `frontend/wailsjs`: generated bindings, regenerate with `just bindings`.
- `.github/workflows/ci.yml`: quality gate plus Linux, macOS (universal) and
  Windows builds, packaged as archives. Also callable by the release workflow.
- `.github/workflows/release.yml`: on a `v*` tag, validates it, reruns CI with
  that version and drafts a GitHub release.
- `CHANGELOG.md`: Keep a Changelog format; the source of release notes.
- `docs/screenshots/`: README images (see README below).

## Commands

`just --list` shows everything. `just check` is the quality gate.
`just test-integration` runs the read-only check against the local engine,
`just icon` regenerates the icons.

## README

`README.md` and `README.tr.md` say the same thing; change both together.
Screenshots in `docs/screenshots/` are taken from an isolated `docker:dind`
engine with neutral sample data, never from the developer's own engine:

1. `docker run -d --privileged --name bilge-demo -p 127.0.0.1:23750:2375
   -e DOCKER_TLS_CERTDIR= docker:29-dind` (loopback only).
2. With `DOCKER_HOST=tcp://127.0.0.1:23750`, create a running Compose project
   (`shop`), a removed one that leaves its volume (`blog`), a stopped one
   (`analytics`), a stopped standalone container, an anonymous volume, an
   empty network, an untagged build and some build cache.
3. `DOCKER_HOST=tcp://127.0.0.1:23750 just dev`, then capture 1280x800 at 2x
   with headless Chrome: overview (light and dark), containers, cleanup and
   the named-volume dialog.
4. Remove the demo: `docker rm -f -v bilge-demo` and `docker rmi docker:29-dind`.

## Releasing

1. Move the `[Unreleased]` entries in `CHANGELOG.md` under the new version
   and commit. Make sure `main` is green in CI.
2. `git tag -a vX.Y.Z -m "vX.Y.Z"` and push the tag (needs Kemal's approval).
3. The Release workflow validates the tag, reruns CI with that version and
   creates a **draft** release with the three archives and `SHA256SUMS`.
4. GitHub's generated notes are empty here (work goes straight to `main`, no
   pull requests), so put the changelog entry on the draft with
   `gh release edit vX.Y.Z --notes-file <file>`.
5. Check the draft (checksums, version string in the binary), then Kemal
   publishes it on GitHub. Never publish from here.

Released: v0.1.0 (2026-10-20).

## Rules specific to this project

- Commit only after Kemal has seen a summary of the change (what, why, how it
  was verified) and asked for it. Pushes need his approval too.
- Add user-facing changes to `CHANGELOG.md` under `[Unreleased]`.

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
