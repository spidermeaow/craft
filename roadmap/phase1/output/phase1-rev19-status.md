# Phase 1 Rev.19 status

Date: 7 October 2026. Target: Craft 0.1.19. **Implemented in source; awaiting the user's GUI acceptance.**

## Findings from Rev.18

- Existing `go test ./...` and both Stress Lab `craft check` runs passed. A new targeted regression test failed on Rev.18: the controls column stretched from preferred width 112 to 478 logical pixels when a sibling ScrollView needed the spare space.
- A real 8×8 Direct2D screenshot confirmed that the controls consumed nearly half the window and clipped the last grid column. ScrollView content also stretched its rows and buttons vertically.
- Native host PageDown, wheel and click-after-scroll worked in the exercised path.

## Rev.19 changes

- A direct ScrollView child receives Row/Column spare space before other flexible containers; its content retains intrinsic size.
- Added scene regression tests and a Windows native host integration test for vertical/horizontal wheel, click-after-scroll and offset clamp after update.
- Package, template, CLI/bundle version, installer metadata, Stress Lab guidance and acceptance docs updated.
- `craft version` now reports the current milestone instead of a stale `Language: Phase 1 Rev.13` line.

## Validation on the development machine

- `go build ./...` and `go build -o dist/craft.exe ./cmd/craft` completed.
- `go test ./...` passed. The native-window integration test is opt-in and passed separately with `CRAFT_UI_INTEGRATION=1`; it exercised PageDown, vertical and horizontal wheel, click-after-scroll and offset clamp after content shrink.
- `craft check` passed for both `examples/gui-stress` and `D:\02-Repository\test-gui` with the 0.1.19 binary.
- `craft package check` passed for `D:\02-Repository\test-gui`; `craft build --release` produced a 0.1.19 source bundle from `examples/gui-stress`. `craft version` reports `Milestone: Phase 1 Rev.19`.
- Real-window Direct2D screenshots showed all 64 cells of 8×8 in view. A 32×32 window after three PageDown inputs showed cells in the 608–900 range while controls stayed in place. The images are local diagnostic files under `build/` and are not packaged.
- A short 32×32 trace recorded four frames and two scroll events (`scroll_to_draw_complete` p50 2,749 µs, p95 6,257 µs). This is a smoke observation, not a performance benchmark; the 0.1.18 comparison run had no painted frames while the window was hidden.
- Inno Setup compiled `dist/Craft-setup.exe`; installation itself was not exercised.

User acceptance remains necessary for other DPI configurations, GDI fallback, installer installation and subjective responsiveness.
