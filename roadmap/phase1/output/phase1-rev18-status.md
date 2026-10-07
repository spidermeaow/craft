# Phase 1 Rev.18 status

Date: 6 October 2026. Target: Craft 0.1.18. **Implemented in source; awaiting the user's visual and performance testing.**

## Implemented

- `gui.scrollView(id, child)` and native ScrollView node with measured content, clipped viewport and persistent scroll offsets by ID.
- Vertical and horizontal wheel input, Shift+wheel, keyboard scrolling, boundary clamping and viewport-aware hit testing/painting in Direct2D and GDI paths.
- Scroll trace metrics, stress examples, package/template, bundle version and Windows installer metadata.

## Evidence and remaining acceptance

- `go build ./...` and `go build -o dist/craft.exe ./cmd/craft` completed. Inno Setup also produced `dist/Craft-setup.exe`. These confirm compilation and packaging only.
- Automated tests and GUI execution were not run for this implementation task because the user performs testing personally.
- User acceptance is required for scrolling, hit testing, update/resize/DPI behavior and responsiveness. See [TRY-REV18](../../../docs/TRY-REV18.md).
