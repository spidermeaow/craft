# Craft UI Engine 0.1.18 — ScrollView

`gui.scrollView(id, child)` สร้างพื้นที่เลื่อนสำหรับ UI tree หนึ่งลูก. ใช้กับ Row, Column หรือ Panel ได้โดยไม่ต้องกำหนดขนาดเนื้อหาเอง:

```craft
var rows: UiNode[] = []
rows.append(gui.text("item-1", "First item"))
rows.append(gui.text("item-2", "Second item"))
let list: UiNode = gui.scrollView("list-scroll", gui.column("list", rows))
```

ScrollView รับพื้นที่ว่างจาก parent ที่เป็น Row/Column; เนื้อหาด้านในใช้ preferred size ที่วัดใน logical pixels. Wheel เลื่อนแนวตั้ง, Shift+wheel และ horizontal wheel เลื่อนแนวนอน. เมื่อชี้หรือเลื่อนใน ScrollView นั้น คีย์ลูกศร, PageUp/PageDown และ Home/End จะเลื่อน view ที่ active. เนื้อหาที่อยู่พ้น viewport ถูก clip และคลิกไม่ได้; offset ถูก clamp ที่ขอบ และคงอยู่ตาม ID หลัง `update`/resize.

การ scroll ย้าย geometry ที่วัดไว้โดยไม่วัดข้อความใหม่ทุกครั้ง. Renderer ตรวจ intersection กับ dirty region และ viewport ก่อนวาด subtree. เปิด `CRAFT_UI_TRACE=1` เพื่อดู `scroll_events`, `scroll_to_draw_complete` และ `visible_nodes_last_frame`. ตัวเลข draw completion ไม่ใช่เวลาที่ภาพปรากฏบนจอ.

ScrollView รุ่นนี้ยังสร้าง UI node ทุกตัวตาม tree ของผู้ใช้ ไม่ใช่ virtualization. ตาราง 32×32 ใน Stress Lab มี 1,024 เซลล์จริงและจะถูก clip/วาดเฉพาะส่วนที่มองเห็น แต่ยังมีต้นทุนการสร้างและ layout ของทุกเซลล์. ยังไม่มี scrollbar ที่ลากได้, momentum animation หรือ touch gesture แบบ native.
