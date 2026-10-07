# Craft UI Layout Foundation 0.1.23

Rev.23 เพิ่ม primitive ใน `std.ui` สำหรับจัดพื้นที่หน้าจอ โดย API เดิมยังใช้ได้:

| API | ความหมาย |
| --- | --- |
| `size(node, width, height, minWidth, minHeight, weight)` | กำหนดขนาดที่ต้องการและขั้นต่ำเป็น logical pixels; `0` หมายถึง auto. `weight` แบ่งพื้นที่ว่างบนแกนหลักของ Row/Column; `0` หมายถึงไม่รับพื้นที่เพิ่มจากกฎนี้ |
| `maxSize(node, maxWidth, maxHeight)` | จำกัดขนาดสูงสุด; `0` หมายถึงไม่จำกัด |
| `align(node, horizontal, vertical)` | จัดแนวภายในพื้นที่ของ parent: `stretch`, `start`, `center`, `end` |
| `spacing(node, padding, gap)` | กำหนดระยะขอบและช่องว่างให้ Row/Column/Panel; ค่า 0 ใช้ได้ |
| `splitView(id, first, second, ratio, minFirst, minSecond)` | จัดสอง pane แนวนอน; `ratio` คือเปอร์เซ็นต์ความกว้าง pane แรก (0–100); ขั้นต่ำเป็น logical pixels |

API จัดวางทุกตัวคืน `UiNode` ใหม่โดยไม่แก้ node ต้นทาง. ค่าขนาดถูกจำกัด 0–8192, weight 0–1000, padding/gap 0–256. SplitView ใช้ divider gap 4 logical pixels; ยังลากแบ่งไม่ได้. เมื่อหน้าต่างแคบกว่าผลรวมขั้นต่ำ จะลดขนาด pane ให้ยังอยู่ภายใน viewport; เนื้อหาที่ใหญ่กว่าต้องวางใน ScrollView. ขนาด fixed เป็น preferred size และอาจถูกลดลงเมื่อพื้นที่ไม่พอ.

`ScrollView` ยังเก็บ intrinsic size ของเนื้อหา; ถ้าต้องการ scroll ภายใน pane ให้กำหนด `size` ที่เนื้อหาด้วย. เมื่อมีเนื้อหาแนวตั้งล้น viewport จะแสดง scrollbar ที่ลากได้. Wheel บน scroll ชั้นในที่ถึงขอบจะถูกส่งต่อไปชั้นนอก. Layout/scroll state ผูกกับ node ID เดิมเมื่อ `update` และ resize. TextInput เป็น native Windows EDIT ตาม Rev.22.

ตัวอย่างจริงอยู่ที่ `examples/gui-layout-lab/src/main.craft`; วิธีลองด้วยตาอยู่ใน `docs/TRY-REV23.md`. ตัวอย่างใช้ปุ่มแทนแถบ tab/list และช่องข้อความแทน form เต็มรูปแบบ. ยังไม่มี theme, table/grid, popup, drag splitter หรือ multi-window.
