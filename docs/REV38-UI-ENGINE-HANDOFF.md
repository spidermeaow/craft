# Rev.38 UI Engine — สถานะและวิธีตรวจด้วยตนเอง

Rev.38 ยังเป็นงานระหว่างดำเนินการ เอกสารนี้ระบุการเปลี่ยนแปลงที่ลง source แล้วและวิธีตรวจ UI รอบปัจจุบัน ไม่ใช่ใบรับรองว่า UI Engine Final ผ่านแล้ว

## สิ่งที่ลง source ในรอบนี้

| หัวข้อ | สถานะ |
| --- | --- |
| `std.ui.virtualGrid`, selection/range/editor และ Unicode clipboard | มี public API จาก Rev.33–37; ยังรอผลตรวจ interaction/DPI/IME/GDI ของผู้ใช้ |
| UIA tree ของ VirtualGrid | ระบุ DataGrid และเซลล์ที่โหลดเป็น DataItem; ชื่อเซลล์ประกอบด้วยแถว คอลัมน์ และค่า |
| UIA focus | `IsKeyboardFocusable` และการสั่ง focus ใช้ข้อกำหนด focus เดียวกับ UI host |
| Timing trace | `CRAFT_UI_TRACE=1` เก็บตัวอย่างล่าสุดสูงสุด 8192 ค่า/metric แล้วสรุปเมื่อปิด session |
| UIA Grid/Table/Selection/VirtualizedItem patterns | ยังไม่มี |
| Variable cell size, pinned regions, dirty-cell rendering และ generic popup/dialog | ยังไม่มี |

## ขั้นตอนให้ผู้ใช้ตรวจ

1. เปิด UI Showcase หน้า Excel ด้วย Craft build ที่รวม source รอบนี้ แล้วลองเลือกช่วง, เลื่อนเร็ว, กลับมาแก้เซลล์ และ copy/paste ภาษาไทย ค่าที่แก้ต้องอยู่ในเซลล์เดิมและข้อความไม่ซ้อน
2. ใช้ Accessibility Insights หรือ screen reader ตรวจ grid และเซลล์ที่มองเห็น ชื่อเซลล์ควรบอกแถว/คอลัมน์ตามพิกัดจริงและอ่านข้อความได้ คาดว่ายังเข้าถึงเซลล์นอก viewport แบบ virtualized ไม่ได้
3. สลับ Direct2D กับ `CRAFT_UI_RENDERER=gdi` และทดลอง DPI 100%, 125%, 150%, ย้ายหน้าต่างข้ามจอ และ IME ไทย แล้วบันทึกกรณีภาพหรือ focus ไม่ตรง
4. หากต้องการเก็บ timing ตั้ง `CRAFT_UI_TRACE=1` ก่อนเปิดแอป เลื่อน/แก้ข้อมูลตามปกติแล้วปิดแอป ดู JSON `craft-ui-engine-timing` ที่ stderr และบันทึกสเปกเครื่อง, DPI, renderer, จำนวนแถว/คอลัมน์ควบคู่กัน ตัวเลขเป็นเวลาถึงจบการวาด ไม่ใช่เวลาที่ภาพขึ้นจอจริง

ยังไม่มีการรันทดสอบโดยผู้พัฒนารอบนี้ตามข้อตกลงให้ผู้ใช้เป็นผู้ทดสอบ ผลทดสอบที่ได้จะใช้ปิดรายการใน [แผน Rev.38](../roadmap/phase1/craft-phase1-rev38.md) ทีละข้อ
