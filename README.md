# bilge

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

> Status: early development. Nothing is removed yet.

Docker is a trademark of Docker, Inc. bilge is an independent project and is
not affiliated with or endorsed by Docker, Inc.

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
