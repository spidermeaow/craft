# Phase 1 Rev.14 status

Date: 6 October 2026. Target: Craft 0.1.14. **Implementation partial; core Windows smoke tests passed. Full acceptance remains open.**

## Implemented in source

- `UiNode` tree primitives and package wrappers for Row, Column, Panel, Text, Button, mount and update.
- Tree validation for duplicate IDs, depth, node count and leaf children; session copies the tree on commit.
- Basic equal-share Row/Column layout, Panel inset, resize relayout, button hit testing, Tab/Shift+Tab focus, hover color, `ButtonPressed` with `nodeId`, and Windows GDI rectangle/text painting.
- `examples/gui-layout`, 0.1.14 CLI/package/bundle metadata, guides and installer metadata.

## Evidence and remaining work

- After the user explicitly requested testing, `go test ./...` passed. `craft check` and `craft build` passed for `examples/gui-layout`.
- Windows smoke runs of source and bundle opened a real window, observed a painted button, confirmed a click outside it did not change the text area, clicked the button and observed a changed rendered text area, resized the client area, and closed the process. The smoke run uses synthetic Windows messages and a pixel fingerprint, not human interaction.
- In a fresh temporary project, `craft gui init`, `craft package check` and `craft check` passed. A second `craft gui init` was rejected without overwriting the package. Ctrl+C during the GUI event loop exited and left no Craft process.
- Go build of `cmd/craft` and the Windows CLI/Setup packaging script completed. The Setup installation/uninstallation workflow, DPI variants, text quality and resource leak behavior have not been verified.
- Direct2D/DirectWrite or GPU rendering, precise content measurement, min/max/fill sizing, clipping, custom style, complete focus/pressed behavior, headless host, provider version gate, device-loss recovery, robust DPI verification and safe in-place upgrade of a user-modified GUI package remain open from the Rev.14 plan.
- No Craft Office, visual designer, standalone application binary, Linux/macOS GUI backend, animation or performance claim is included.

See [TRY-REV14](../../../docs/TRY-REV14.md) for further manual acceptance. Do not mark Rev.14 accepted until the remaining implementation and distribution tests are recorded.
