# Changelog

All notable changes to bilge are listed here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-10-20

First release.

### Added

- Overview of Docker's disk usage split into kept, needs review and safe to
  remove, with engine details, Compose projects (including leftovers) and the
  largest images and volumes.
- Lists of containers, images, volumes and networks with search, status
  filters and sortable, resizable columns.
- Start, stop and restart containers.
- Careful cleanup: safe items are removed together after one confirmation,
  review items one at a time, named volumes only after typing their name.
  Every item is checked against a fresh scan right before removal.
- Builds for Linux, macOS (universal) and Windows with `SHA256SUMS`.

[Unreleased]: https://github.com/kemalasliyuksek/bilge/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/kemalasliyuksek/bilge/releases/tag/v0.1.0
