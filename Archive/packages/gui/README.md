# Craft GUI 0.1.20

Windows native GUI package for Craft. This package wraps the built-in `std.ui` provider. Rev.14 adds a retained UI tree with `row`, `column`, `panel`, `text`, `button`, `mount`, and `update`. It supports one window per interpreter run.

Use `craft gui init` in an existing Craft project to copy this package into `packages/gui` and add the local dependency. Import it with `import gui "gui"`. Run the app with Craft 0.1.20 on Windows. See `examples/gui-stress` for ScrollView and Virtualized List usage.

`nextEvent(-1)` waits, `nextEvent(0)` polls, and positive values wait up to 60 seconds. A timeout returns `null`. Events also have `nodeId`. `ButtonPressed` names the button in that field. Key events are physical key codes, not text or IME input.

Window handles are session-bound. Close each window once; the runtime also closes it when the program exits or is cancelled. A long-running Craft computation that does not call `nextEvent` can make the window unresponsive.

Nodes have unique stable IDs within each tree, with a 4096-node and 64-level limit. `mount` installs the first tree; `update` replaces it. Rev.15 requests a hardware Direct2D target and uses DirectWrite for text, falling back to GDI when unavailable. Rev.16 uses event-driven host wake and partial repaint. Rev.17 measures Text/Button content. Rev.18 adds `scrollView(id, child)` for content larger than its viewport. Rev.19 corrects ScrollView sizing. Rev.20 adds `virtualList(id, itemCount, rowHeight, firstIndex, items)` and `VisibleRangeChanged` events so Craft supplies only visible rows. Wheel and page keys move the active viewport; offsets persist across update and resize by node ID. `CRAFT_UI_TRACE=1` reports timings and virtual row counts. `renderer(window)` reports `direct2d-hardware` or `gdi`. Tab/Shift+Tab move basic button focus. Virtualized Grid, momentum animation, TextBox/IME and full accessibility remain future work. `craft build` still produces a source bundle requiring Craft CLI.
