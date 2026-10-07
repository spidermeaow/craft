# Phase 1 Rev.15 status

Date: 6 October 2026. Target: Craft 0.1.15. **Windows smoke test completed; broader visual and performance acceptance pending.**

## Implemented in source

- Windows hardware Direct2D target with DirectWrite text, retained COM resources, device-loss target recreation and GDI fallback. `gui.renderer` reports the selected backend.
- Pointer release activation/cancellation, mouse leave tracking, focus retention after tree update and DPI boundary adjustments.
- UI host idle polling interval increased from 1 ms to 8 ms after measuring unnecessary idle CPU use.
- Rev.15 package, example, bundle metadata, CLI/installer metadata and guides.

## Evidence and limits

- `go test ./...` passed both before and after the idle polling adjustment. `craft version` reported 0.1.15 and `craft check` passed for `examples/gui-layout`.
- `craft run` and `craft build` followed by `craft run dist/gui-layout.craftbundle` both opened the Windows example and reported `direct2d-hardware`. The bundled example showed the initial text, changed it after a button click, left it unchanged after pressing then dragging outside the button, and activated through Tab then Enter. The window closed cleanly.
- A second GUI run after the polling adjustment still reported `direct2d-hardware` and responded to a click. In two 3-second idle samples, CPU time changed from 0.31 s before to 0.17 s after; these are single observations, not a benchmark. The latter process used about 65 MiB working set. The host still polls and can be optimized further.
- DirectWrite typography quality, device loss, GDI fallback, DPI variants, focus retention after an update, installer installation/uninstallation and frame time remain **unverified by execution**.
- Layout remains equal-share; text measurement, scrolling, animation and virtualization are not implemented. Do not infer modern app performance or 60/120 FPS from this revision.

See [TRY-REV15](../../../docs/TRY-REV15.md) for manual acceptance. Record actual results before marking this revision accepted.
