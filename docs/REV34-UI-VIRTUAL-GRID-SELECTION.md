# Craft 0.1.34 — VirtualGrid single-cell selection

```craft
let grid: UiNode = gui.virtualGrid("data", rows, columns, 34, 150, firstRow, firstColumn, loadedColumns, cells)
let selected: UiNode = gui.gridSelection(grid, selectedRow, selectedColumn)
```

ดัชนีแถวและคอลัมน์เริ่มจาก 0 และต้องอยู่ในขอบเขตของตาราง. หากยังไม่มี selection ให้ส่ง `grid` ตรง ๆ. แอปเป็นเจ้าของค่า `selectedRow`/`selectedColumn`; เมื่อได้ `GridCellPressed` หรือ `GridSelectionMoved` ให้อ่าน `event.firstIndex` และ `event.firstColumn`, เก็บค่า แล้วเรียก `gui.update` ด้วย `gui.gridSelection` ใหม่.

เมื่อ VirtualGrid มี focus, Arrow เปลี่ยนเซลล์ทีละช่อง, Page Up/Down เปลี่ยนตามจำนวนแถวที่เห็น, Home/End ไปต้น/ท้ายแถว และ Ctrl+Home/End ไปมุมต้น/ท้ายของตาราง. Runtime เลื่อน viewport ให้เห็นตำแหน่งใหม่และส่ง event. ขอบเขต API ยังเป็นตาราง cell ขนาดคงที่; ไม่มี editor, range selection หรือ clipboard ในรุ่นนี้.

ดู [แผน 0.1.34](../roadmap/phase1/craft-phase1-rev34.md) และ [VirtualGrid 0.1.33](REV33-UI-VIRTUAL-GRID.md).
