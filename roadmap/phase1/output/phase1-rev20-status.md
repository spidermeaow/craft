# Phase 1 Rev.20 status

Date: 7 October 2026. Target: Craft 0.1.20. **Implemented in source; awaiting the user's GUI acceptance.**

## Implemented

- Fixed-height vertical `VirtualList` with total logical extent, clipped viewport, retained scroll offset, bounded visible-range requests and Craft-supplied row nodes.
- `VisibleRangeChanged` event adds `firstIndex`/`endIndex` fields. The GUI package exposes `virtualList`; existing event fields remain available.
- Range math, item count/height limits, count-shrink clamping and a source-level regression test. `test-gui` and `examples/gui-stress` include a 100,000-row scenario.
- CLI/package/template/installer version metadata, source-bundle compatibility, documentation and trace row/range counters updated.

## Build and acceptance

- `go build ./...` and the release build script compiled `dist/craft.exe`; Inno Setup compiled `dist/Craft-setup.exe`. The 0.1.20 source bundle for `examples/gui-stress` was also built. These confirm compilation and packaging only.
- The user prefers to perform GUI testing. Automated tests, `craft check` for `test-gui`, and GUI execution were not run for this implementation request.
- The user should follow `docs/TRY-REV20.md`, including click-after-scroll, resize/DPI and the GDI fallback when available. Responsiveness and p50/p95/p99 comparisons remain unverified.
- The implementation keeps only requested rows in the UI tree; it does not yet include direct jump-to-index, draggable scrollbar, variable-height rows or Virtualized Grid.
