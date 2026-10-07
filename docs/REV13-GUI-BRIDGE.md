# Craft 0.1.13 — Windows GUI bridge

Rev.13 adds the first native UI bridge for Craft. It is a foundation for future GUI libraries, not a complete GUI framework. `packages/gui` is Craft source; the `std.ui` provider is compiled into the Windows Craft runtime. No WebView, HTML/CSS, JavaScript runtime, or arbitrary native DLL loading is involved.

## Start a project

```powershell
craft new my-gui
cd my-gui
craft gui init
```

`craft gui init` copies the bundled `gui` source package into `packages/gui` and adds a local dependency to `craft.toml` and `craft.lock`. It refuses an existing `gui` dependency or `packages/gui` directory. It does not download from GitHub. The CLI contains the same package source whether used through Setup or as a portable executable.

Import it using `import gui "gui"`. See `examples/gui-bridge/src/main.craft` for a small event loop. A program can open one window per interpreter run, set title/background, poll events, and close the window. `craft run` and `craft run <bundle>` still require the Craft CLI; Rev.13 does not build standalone application executables.

## API contract

| Craft package API | Meaning |
| --- | --- |
| `gui.openWindow(String, Int, Int): UiWindow` | Open one native Windows window. Width and height must be 1–8192. |
| `gui.setTitle(UiWindow, String)` | Set Unicode title, up to 1024 UTF-8 bytes, without NUL. |
| `gui.setBackground(UiWindow, Int, Int, Int)` | Set RGB components, each 0–255. |
| `gui.nextEvent(Int): UiEvent?` | Poll at 0 ms, wait up to 60,000 ms, or wait indefinitely with -1. `null` means timeout. |
| `gui.closeWindow(UiWindow)` | Close the window. A second close or use after close is an error. |

`UiEvent` fields: `kind`, `windowId`, `x`, `y`, `width`, `height`, `keyCode`, `modifiers`, and `scale`. Kinds are `CloseRequested`, `Resized`, `PointerClicked`, `KeyPressed`, and `DpiChanged`. The window ID is 1 for this revision. Modifier bits are Shift=1, Ctrl=2, Alt=4. Key codes are Windows virtual key codes; this is not text input or IME. Position and dimensions are logical pixels based on the reported scale. Fields unrelated to the event are zero.

The UI host runs on its own locked Windows OS thread. OS callbacks only queue events; they do not reenter the Craft interpreter. Craft reads events in its loop. The event queue holds at most 1024 events; overflow ends the session with `UiQueueOverflow`. Craft code that does not return to `nextEvent` for a long time may make the window unresponsive. The runtime closes the window when `main` exits, throws, or is cancelled, even if the program did not call `closeWindow`.

`craft build` marks GUI bundles with language `0.1.13` so older runtimes reject them. Bundles with no `std.ui` use retain the previous language marker. Linux/macOS runtimes report `UiUnavailable` on GUI invocation.

## Current limits

No component system, layout, text rendering/editing, IME, accessibility, GPU renderer, animation, virtualized lists, drag and drop, standalone `.exe`, or visual designer. The window background and event bridge prove the Craft ↔ native host path only. Do not infer smooth modern app performance from this revision.
