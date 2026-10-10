# Craft 0.1.33 — VirtualGrid prototype

`std.ui.virtualGrid` และ `gui.virtualGrid` เป็น viewport ตารางสองแกนแบบทั่วไปบน Windows. แอปเป็นเจ้าของข้อมูลและส่งเฉพาะ `Text` nodes ของช่วงที่โหลดมาให้ UI Engine. จำนวน node สูงสุดต่อการอัปเดตคือ 1,024; ไม่ต้องสร้าง node ตามจำนวนแถวคูณคอลัมน์ทั้งหมด.

```craft
gui.virtualGrid(id, rowCount, columnCount, rowHeight, columnWidth,
    firstRow, firstColumn, loadedColumns, cells)
```

`cells` เรียงตามแถวก่อน แล้วตามคอลัมน์. จำนวนต้องหารด้วย `loadedColumns` ลงตัว; เมื่อยังไม่มีข้อมูลให้ใช้ `loadedColumns = 0` และ `cells = []`. ทุก cell ต้องเป็น `Text` node ที่มี ID ไม่ซ้ำ. `firstRow`/`firstColumn` เป็นดัชนีเริ่มต้นแบบศูนย์. ความสูงแถวและความกว้างคอลัมน์เป็น logical pixels เท่ากันทั้งตารางใน prototype นี้.

เมื่อ viewport ต้องการข้อมูลเพิ่ม Runtime ส่ง `UiEvent` ชนิด `VisibleGridRangeChanged`. `firstIndex`/`endIndex` คือช่วงแถว และ `firstColumn`/`endColumn` คือช่วงคอลัมน์แบบ end-exclusive รวม overscan. แอปสร้าง snapshot ของช่วงดังกล่าวแล้วเรียก `gui.update`. เมื่อเลื่อนเร็ว Runtime เก็บคำขอช่วงล่าสุดของแต่ละ Grid และแทนคำขอเก่าที่ยังไม่ได้อ่าน; แอปอาจยังประมวลผลคำขอเก่าที่อ่านไปแล้ว. หลังคลิก Runtime ส่ง `GridCellPressed` โดย `firstIndex` เป็นแถวและ `firstColumn` เป็นคอลัมน์ของ cell เชิงตรรกะ. หน้า Layouts ใน `D:\02-Repository\app` แสดงตัวอย่างข้อมูล mock 100,000 × 1,000.

Runtime รักษาตำแหน่ง scroll เมื่อแอปส่ง snapshot ใหม่ และวาดเฉพาะ cell nodes ที่ได้รับ. ช่องที่ยังไม่มีข้อมูลว่างจนแอปตอบกลับ. ไม่มี callback จาก native UI thread เข้าไปอ่านข้อมูลแอปโดยตรง.

## ขอบเขตของ prototype

- ยังไม่มี cell editor, selection หลายช่อง, frozen rows/columns, row/column size แบบต่างกัน, merged cells หรือ virtual accessibility.
- Event ช่วงข้อมูลใช้ดัชนีปัจจุบัน ยังไม่มี generation token/cancellation สำหรับ data source แบบ asynchronous; แอปที่โหลดข้อมูลช้าต้องกันข้อมูลเก่าทับข้อมูลใหม่เอง.
- ยังไม่มี dirty-cell invalidation หรือ benchmark ที่พิสูจน์ performance ระดับ Final. การ scroll อาจวาด viewport ใหม่ทั้งส่วน.
- ผู้ใช้ต้องตรวจ Direct2D และ GDI fallback, DPI, การเลื่อนสองแกน และ click event บนเครื่องจริงก่อนถือว่าผ่าน.

ดู [แผน 0.1.33](../roadmap/phase1/craft-phase1-rev33.md) และ [สัญญาระยะยาว](../roadmap/phase1/craft-phase1-rev32-virtual-grid-contract.md).
