# Linux builds link against WebKitGTK 4.1; other platforms ignore the tag.
tags := if os() == "linux" { "webkit2_41" } else { "" }

default:
    @just --list

# Run the app with hot reload
dev:
    wails dev -tags "{{tags}}"

# Build a production binary into build/bin
build:
    wails build -tags "{{tags}}" -clean

# Regenerate the TypeScript bindings for Go methods
bindings:
    wails generate module -tags "{{tags}}"

# Regenerate the app icons from scripts/make_icon.py (needs Pillow)
icon:
    python3 scripts/make_icon.py

format:
    golangci-lint fmt ./...
    cd frontend && pnpm format

# go:embed needs at least one file in frontend/dist, even before the first frontend build
_dist:
    mkdir -p frontend/dist
    test -n "$(ls -A frontend/dist)" || touch frontend/dist/.placeholder

format-check:
    test -z "$(golangci-lint fmt --diff ./...)"
    cd frontend && pnpm format:check

lint: _dist
    golangci-lint run ./...
    cd frontend && pnpm lint

typecheck: _dist
    go build -tags "{{tags}}" ./...
    cd frontend && pnpm typecheck

test: _dist
    go test -race ./...
    cd frontend && pnpm test

# Read-only checks against the local Docker Engine
test-integration: _dist
    go test -tags integration -run LocalEngine -v ./internal/engine/

# Everything CI runs
check: format-check lint typecheck test
