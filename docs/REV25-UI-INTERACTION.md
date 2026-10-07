# Craft UI Runtime 0.1.25 — viewport, style และ Canvas targets

0.1.25 เพิ่ม native primitives สำหรับ `packages/gui` และแอป Craft บน Windows. API เดิมของ 0.1.24 ยังอยู่. ตัวอย่างใน `D:\02-Repository\craft-test\gui-package-dashboard` ใช้ `import gui "gui"` สำหรับ API ที่ package มีอยู่ และเรียก `std.ui` ตรงเฉพาะ primitive 0.1.25 ที่ package 0.2.1 ยังไม่ได้ห่อ.

| API ใหม่ | ความหมาย |
| --- | --- |
| `std.ui.viewportWidth(window): Int` / `viewportHeight(window): Int` | ขนาด client area ปัจจุบันใน logical pixels; อ่านได้หลัง `openWindow` |
| `std.ui.textStyle(node, size, weight, color): UiNode` | กำหนดตัวอักษรให้ `Text`/`Button`; size 8–96, weight 100–900 ทีละ 100, สีจาก `std.ui.color` |
| `std.ui.surface(node, fill, border, radius): UiNode` | พื้นหลัง/ขอบของ `Panel`/`Button`; fill ทึบ, border ทึบหรือโปร่งใส, radius 0–256 |
| `std.ui.drawTextStyled(value, x, y, width, height, color, size, weight, align): UiDraw` | ข้อความ Canvas ที่กำหนดขนาด/น้ำหนัก/แนว `start`, `center`, `end`; wrap ภายในกรอบ |
| `std.ui.hitRegion(id, x, y, width, height): UiDraw` | พื้นที่รับ input ที่ไม่ถูกวาด; target ID ไม่ซ้ำกับ node/target อื่นใน tree; region หลังสุดอยู่บนสุด |

`Resized` และ `DpiChanged` เดิมส่ง `width`/`height` ของ client area เป็น logical pixels. แอปควรอ่าน viewport ครั้งแรกหลังเปิดหน้าต่าง แล้ว rebuild Canvas ตามขนาดเมื่อรับสอง event นี้. Canvas ขนาดคงที่ยังคงขนาดเดิม; หากใหญ่กว่า ScrollView ให้ใช้ scrollbar แนวนอนที่เพิ่มในรุ่นนี้หรือ wheel แนวนอน/Shift+wheel. `size`/`weight` และ min/max ของ layout ยังคงสัญญาเดิม; min ที่รวมกันเกิน viewport ถูก clip และมี trace overflow.

เมื่อ Canvas มี hit region:

- `CanvasPointerDown`/`CanvasPointerUp` ส่ง target ID (หรือ Canvas ID ถ้าไม่อยู่ใน region) กับ local coordinates.
- `CanvasTargetEntered`/`CanvasTargetLeft` ส่ง target ID เมื่อ pointer ข้ามขอบ; `CanvasPointerMoved` ส่งได้สูงสุดประมาณทุก 16 ms และ drop ได้หาก queue เต็ม เพื่อไม่ทำ click/key/close หาย. พิกัดของ move เป็น local Canvas; enter/leave ให้ใช้ ID เป็นหลัก.
- คลิก region ที่ down/up อยู่บน ID เดิมส่ง `CanvasPressed` เดิม แล้ว `FocusChanged` และ `CanvasTargetPressed` ของ target. `CanvasPressed` ยังเป็น event ระดับ Canvas เพื่อรักษา compatibility.
- Region อยู่ในลำดับ Tab; Enter/Space บน region ที่ focus ส่ง `CanvasTargetPressed`. `KeyPressed`/`KeyReleased` มี `nodeId` ของ focus ปัจจุบัน. แอปเปลี่ยนรูปลักษณ์ hover/pressed/focus ผ่าน state แล้ว `update` ได้.
- `PointerDown`, `PointerUp`, `PointerDoubleClicked` ระดับหน้าต่างใช้พิกัด global logical; อย่าสับสนกับพิกัด local ของ Canvas.

Text แบบ node ยังคงบรรทัดเดียว ส่วน `drawTextStyled` รองรับการ wrap ในกรอบที่กำหนด. TextInput ยังใช้ native EDIT และ `textStyle` ไม่รับ TextInput. GDI fallback ไม่ blend alpha ของข้อความและอาจต่างด้าน anti-alias/rounding. ยังไม่มี shadow, gradient, transform, automatic responsive reflow, form controls ขั้นสูง, data grid, animation หรือ accessibility tree สำหรับ region. Hit region เป็น target ข้อมูล ไม่ใช่ native child control; ผู้พัฒนาต้องจัด label/visual state และตรวจการใช้งานด้วย keyboard เอง.

`CRAFT_UI_TRACE=1` ยังรายงาน measure/layout/paint และ input-to-end-draw ตามเดิม; **ยังไม่ใช่ input-to-present latency**.
