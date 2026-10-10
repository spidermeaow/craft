# Craft 0.1.48: Enterprise sidebar

Rev.48 refines the shared `packages/gui-template` sidebar. The package remains a composition over native `packages/gui` and `std.ui`; it adds no web runtime or external library.

## API

- `navItem(id, label, selected)` and `navigation(id, brand, caption, items, footer)` remain available.
- `navItemIcon(id, label, icon, selected)` adds a leading icon.
- `navItemState(id, label, icon, selected, hovered)` lets the application supply hover state. Supported icons are `layout-dashboard`, `package`, `arrow-down-to-line`, `arrow-up-from-line`, and `history`.
- `navigationWithStatus(id, brand, caption, items, status, runtime)` puts a connection/status line and runtime label in a fixed footer below a separately scrolling navigation list.
- `navigationFooter(id, status, runtime)` is available for custom sidebar compositions.

Navigation item Canvas IDs are unchanged. Each item also has a full-row target ID `<id>-target`. Handle `CanvasTargetEntered` and `CanvasTargetLeft`, update the hovered ID, and call `gui.update` only when hover changes. Click routing can continue to use `CanvasPressed` and the original ID.

The engine offers Canvas lines, rectangles, hit regions, pointer target events, scroll views, and DPI-aware logical pixels. It does not provide named icon assets or automatic Canvas hover painting, so the five icons are native strokes and hover is application state. It also does not prevent a disabled Canvas from emitting `CanvasPressed`; this release does not expose a disabled navigation variant. The app should update its width on `Resized` and `DpiChanged` and rebuild the shell when the compact breakpoint changes.

InventoryManagement demonstrates the five icons, hover routing, and MySQL status footer. The requested footer text says `Craft Runtime 0.1.47` even when using the 0.1.48 candidate toolchain.
