# Changelog

All notable changes to bilge are listed here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.1] - 2026-10-21

### Added

- The Linux archive ships a desktop entry and an icon, so bilge can be added
  to the application menu.

### Fixed

- Windows file properties and the macOS bundle now show the real version,
  name and copyright instead of Wails' defaults (1.0.0 and placeholders).

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

[Unreleased]: https://github.com/kemalasliyuksek/bilge/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/kemalasliyuksek/bilge/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/kemalasliyuksek/bilge/releases/tag/v0.1.0
