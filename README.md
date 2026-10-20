# bilge

[![CI](https://github.com/kemalasliyuksek/bilge/actions/workflows/ci.yml/badge.svg)](https://github.com/kemalasliyuksek/bilge/actions/workflows/ci.yml)

A careful cleaner for Docker.

bilge is a small desktop app that shows what your Docker Engine holds
(containers, images, volumes, networks and build cache) and helps you remove
what is actually waste.

`docker system prune` and most GUI "prune" buttons treat everything that is not
referenced right now as reclaimable. That includes the database volume of a
Compose stack you stopped for the weekend. bilge sorts candidates into two
groups instead:

- **safe**: dangling images, anonymous volumes, stopped standalone containers,
  empty networks, unused build cache. Removable in bulk after one confirmation.
- **review**: named volumes, tagged images, stopped containers that belong to a
  Compose project. Shown with the reason, removed one at a time.

Anything in use, or labelled `bilge.keep=true`, is never offered for removal.
The full rules are in [docs/design.md](docs/design.md).

Before anything is removed, bilge scans the engine again and skips every item
that changed since you looked: a volume a container just mounted is left alone.
A named volume must be confirmed by typing its name.

> Status: early development. Try it on an engine you can afford to lose things
> on before trusting it with one you cannot.

Docker is a trademark of Docker, Inc. bilge is an independent project and is
not affiliated with or endorsed by Docker, Inc.

## Download

Builds for Linux, macOS (universal) and Windows are attached to each
[release](https://github.com/kemalasliyuksek/bilge/releases), together with a
`SHA256SUMS` file. The binaries are not signed yet: macOS asks you to confirm
the first launch (right-click, Open) and Windows SmartScreen may warn as well.
On Linux bilge needs WebKitGTK 4.1 (`libwebkit2gtk-4.1-0` on Debian and Ubuntu).

## Requirements

- Go 1.25+
- Node.js 24 and pnpm
- [Wails v2](https://wails.io) CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [just](https://just.systems) and [golangci-lint](https://golangci-lint.run) v2
- Linux: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`

If you use [mise](https://mise.jdx.dev), `mise install` sets up the pinned tool
versions from `mise.toml`.

## Development

```sh
cd frontend && pnpm install && cd ..
just dev      # run with hot reload
just check    # format check, lint, typecheck, tests
just build    # production binary in build/bin
```

## License

[MIT](LICENSE)
