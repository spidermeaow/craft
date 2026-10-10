# Craft 0.1.36 — editor inside VirtualGrid

`gui.gridEditor(grid, row, column, editor)` places one enabled `TextInput` over a VirtualGrid cell. `row` and `column` are zero based. The grid continues to receive only single-line `Text` nodes for its loaded data cells; the editor is a separate overlay child and does not count toward the loaded range.

```craft
let grid: UiNode = gui.virtualGrid("orders", 1000, 8, 34, 140, firstRow, firstColumn, loadedColumns, cells)
let selected: UiNode = gui.gridSelection(grid, selectedRow, selectedColumn)
let editor: UiNode = gui.textInput("orders-editor", currentValue, "")
let editable: UiNode = gui.gridEditor(selected, selectedRow, selectedColumn, editor)
```

Handle `GridCellPressed` or `GridSelectionMoved` to update the selected coordinates and value, then call `gui.update`. Keep the editor ID stable and call `gui.setText(window, "orders-editor", currentValue)` after changing cells. `TextChanged` and `TextSubmitted` arrive with node ID `orders-editor`; the app owns storage, validation, commit and cancel rules. `gui.focus(window, "orders-editor")` moves the caret into the cell. The Runtime positions and clips the native input as the grid scrolls.

The UI Showcase at `D:\02-Repository\app` demonstrates the API on its Excel page. See [Rev.36 plan](../roadmap/phase1/craft-phase1-rev36.md) for scope and manual acceptance steps. Range selection, spreadsheet semantics and virtual accessibility are not part of this API.
