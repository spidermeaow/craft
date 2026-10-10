# Craft 0.1.37 — VirtualGrid ranges and clipboard text

`gui.gridRangeSelection(grid, anchorRow, anchorColumn, focusRow, focusColumn)` marks a rectangular range on a VirtualGrid. Coordinates are zero based. The focus cell remains the active cell for keyboard navigation and the optional `gridEditor`; the anchor stays fixed while Shift+Arrow moves the focus. A simple `gridSelection` call clears the range.

`gui.gridRangeStyle(grid, fill, border)` sets opaque colors for the selected range. It can wrap the result of `gridRangeSelection` before `gridEditor`. Without it, the Runtime uses a light blue fill and blue border.

The Runtime emits `GridRangeDragged` when a pointer drag reaches another cell. `firstIndex`/`firstColumn` are the drag anchor, and `endIndex`/`endColumn` are the current cell. Existing `GridCellPressed` and `GridSelectionMoved` events still report the active cell; `modifiers` identifies Shift for a range extension. The application owns selection policy and data changes.

`gui.clipboardWriteText(window, value)` and `gui.clipboardReadText(window)` transfer Unicode text through the Windows clipboard. Errors are reported as `UiClipboardError`. Craft does not interpret rows, columns, formulas, or workbook formats. The UI Showcase at `D:\02-Repository\app` uses tab/newline text for its Copy and Paste buttons and Ctrl+C/Ctrl+V shortcuts.

## Manual check in UI Showcase

1. Open Excel. Click B2, then Shift+click D4, or hold Shift and press Arrow keys. The highlighted rectangle should use B2 as its anchor; the active cell border should follow the focus.
2. Drag from an ordinary cell across several cells and past the visible grid edge. The selected range should follow the pointer while the grid scrolls. Drag from the active editor cell as a separate case.
3. Select a 2×3 range, choose Copy, paste into a text editor and check tab/newline separation. Choose Paste in a different cell and confirm all values land in the expected cells. Repeat with Thai text and with a range near the sheet edge.
4. Edit one cell, then use Enter, Escape, Tab and Apply. Confirm the Rev.36 editor still updates only its active cell.
5. Inspect at DPI 100%/125%, Direct2D/GDI, after resize and rapid scroll. No visual artifact or incorrect hit region should return.

This revision does not yet implement virtualized UI Automation Grid/ItemContainer patterns. The currently loaded Text nodes remain available in the accessibility tree, but that is not complete screen-reader access to an arbitrarily large grid. This work remains open in the Rev.37 plan and must not be counted as passed.
