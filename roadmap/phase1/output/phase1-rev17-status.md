# Phase 1 Rev.17 status

Date: 6 October 2026. Target: Craft 0.1.17. **Implemented in source; awaiting the user's visual and performance testing.**

## Implemented

- Content measurement pass for Text/Button, preferred-size Row/Column/Panel layout and clipping for overflow. DirectWrite metrics are used with Direct2D, GDI metrics with fallback.
- Geometry-changing text updates trigger full layout and repaint; same-size text changes retain localized repaint. Text drawing and descendants are clipped to their node/parent bounds.
- Trace includes `overflow_nodes`; package, template, bundle version and installer metadata updated. Stress examples start in visible 8×8 mode and explain 32×32 clipping.

## Evidence and remaining acceptance

- `go build ./...` and `go build -o dist/craft.exe ./cmd/craft` completed after the engine changes. The Inno Setup compiler also produced `dist/Craft-setup.exe`. These confirm compilation and packaging only.
- Automated tests and GUI execution were not run for this task because the user performs testing personally.
- Visual correctness on Direct2D and GDI, dense-grid hit testing, resize/DPI behavior, actual installer installation, and p50/p95/p99 comparison with 0.1.16 remain unverified. See [TRY-REV17](../../../docs/TRY-REV17.md).
