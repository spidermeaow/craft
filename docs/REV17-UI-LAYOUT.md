# Craft UI Engine 0.1.17 — sizing and overflow

`Text` ใช้ขนาดข้อความที่วัดได้พร้อมขอบเล็กน้อย; `Button` เพิ่ม padding และมีความสูงขั้นต่ำ 32 logical pixels. Renderer ใช้ DirectWrite metrics เมื่อ Direct2D พร้อมทำงาน และใช้ GDI metrics เมื่อ fallback ไป GDI. ค่า font ปัจจุบันคือ Segoe UI 18 DIP.

`Row` วางลูกจากซ้ายไปขวาตามขนาดที่ต้องการ; `Column` วางจากบนลงล่าง. `Panel` วางลูกทับกันในพื้นที่ด้านใน. แต่ละ container มี padding 6 และ gap 4 logical pixels. พื้นที่เหลือบนแกนหลักจะแจกให้ลูกที่เป็น Row, Column หรือ Panel; Text และ Button ไม่ขยายตามพื้นที่เหลือ. Engine ไม่มี sizing property ที่ Craft ต้องตั้งเองในรุ่นนี้.

เมื่อขนาดที่ต้องการรวมกันเกินพื้นที่หน้าต่าง engine จะรักษาขนาดของลูกและ clip ภาพที่ขอบ parent/window. ไม่มี scroll view ใน 0.1.17 ดังนั้นบางส่วนจะมองไม่เห็นและคลิกไม่ได้. เปิด `CRAFT_UI_TRACE=1` เพื่อดู `overflow_nodes` ณ layout ล่าสุด. ตัวเลขนี้เป็นจำนวน node ที่พื้นที่ได้รับน้อยกว่าขนาดที่ต้องการ ไม่ใช่จำนวนเซลล์ที่มองไม่เห็น.

เมื่อข้อความใน tree รูปร่างเดิมเปลี่ยนแต่ขนาดที่วัดได้เท่าเดิม engine จะ repaint เฉพาะกล่องที่เปลี่ยน. ถ้าขนาดเปลี่ยน engine จะ layout ใหม่ทั้ง tree และ repaint ทั้งหน้าต่างเพื่อให้ sibling, hit test และพื้นที่พื้นหลังตรงกัน. Resize, DPI change และโครงสร้าง tree ที่เปลี่ยนใช้ full layout เช่นกัน.

## Migration จาก 0.1.16

แอปที่อาศัยการแบ่งพื้นที่เท่ากันอัตโนมัติจะเห็นตำแหน่งและขนาดใหม่ โดยเฉพาะ Button ที่เคยขยายเต็มช่อง. ใช้ Row/Column/Panel จัดกลุ่มเนื้อหาโดยตั้งใจ. ตารางที่ใหญ่กว่าหน้าต่างถูก clip; ลดจำนวนเซลล์ที่แสดงพร้อมกันจนกว่าจะมี ScrollView/virtualization ในรุ่นถัดไป. ชื่อและชนิดของ `row`, `column`, `panel`, `text`, `button`, `mount` และ `update` ยังเหมือนเดิม.
