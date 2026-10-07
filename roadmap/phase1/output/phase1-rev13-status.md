# Phase 1 Rev.13 status

Date: 6 October 2026. Target: Craft 0.1.13. Core GUI bridge smoke tests passed on Windows; the full acceptance matrix remains open.

## Implemented in source

- Typed `std.ui` signatures and a Craft source wrapper in `packages/gui`; `UiWindow` is session-bound and `UiEvent` carries native input/window data.
- Windows UI host on a locked OS thread with a message pump, one window per run, title/background commands, event queue and cleanup at interpreter exit. Other platforms return `UiUnavailable`.
- `craft gui init` copies the embedded package into an existing project and adds a local dependency/lock without network access.
- GUI source bundles are marked language 0.1.13; non-GUI bundles retain 0.1.11 compatibility. CLI and installer metadata target 0.1.13.
- `examples/gui-bridge` and Rev.13 user/reference guides are included.

## Evidence and limits

- `go build ./cmd/craft` and `go build -o dist/craft.exe ./cmd/craft` completed after code changes. `scripts/build.ps1 -Installer -Icons -Language` also produced the Windows CLI, Setup, VSIX files and SHA-256 manifest. This confirms artifact compilation/package creation only.
- After the user explicitly requested testing, `go test ./...` passed. `craft check` and `craft build` passed for `examples/gui-bridge`.
- Windows smoke runs of both source and bundle opened a real window, verified the initial title, posted click and key messages, observed matching title and background color changes, then posted a close request and observed process exit.
- In a fresh temporary project, `craft gui init`, `craft package check`, and `craft check` passed. A second `craft gui init` was rejected without overwriting the package.
- Manual human interaction, DPI and resize behavior, cancellation during an event wait, queue overflow, invalid handle cases, installation/uninstallation, and performance remain unverified. The smoke run used synthetic Windows messages; it does not establish full GUI acceptance.
- No Craft Office, visual designer, component library, text engine, GPU renderer, Linux/macOS GUI backend or standalone application executable is part of this revision.

See [TRY-REV13](../../../docs/TRY-REV13.md) for the remaining manual acceptance steps. Record actual results here before declaring the GUI bridge accepted.
