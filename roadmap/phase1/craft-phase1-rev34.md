# Craft Phase 1 Rev.34 (0.1.34) — VirtualGrid selection

วันที่: 8 ตุลาคม 2026  
สถานะ: implement ใน source; รอผู้ใช้ทดสอบ UI จริง  
ฐาน: 0.1.33 ซึ่งผู้ใช้ให้ผ่าน พร้อมข้อจำกัดว่าบางช่วงเรนเดอร์ตามไม่ทันเมื่อเลื่อนเร็วมาก

## เป้าหมาย

เพิ่มการโต้ตอบพื้นฐานของตารางทั่วไป โดยไม่ย้ายข้อมูลเซลล์หรือกฎของ Spreadsheet เข้า Craft. รอบนี้เน้นการเลือกเซลล์เดียวที่มองเห็นได้, keyboard navigation และการเลื่อน viewport ให้ตามเซลล์ที่เลือก. งดปรับแต่ง fast scroll ต่อในรอบนี้ตามผลตรวจรับของผู้ใช้.

## งานใน 0.1.34

- `std.ui.gridSelection(grid, row, column)` / `gui.gridSelection` รับดัชนีฐานศูนย์ ตรวจช่วง และวาดกรอบเซลล์ที่เลือกบน Direct2D กับ GDI fallback. แอปถือสถานะ selection และส่งกลับผ่าน `update`.
- VirtualGrid รับ focus จาก click และ Tab. Arrow, Page Up/Down, Home/End และ Ctrl+Home/End ส่ง `GridSelectionMoved` พร้อมแถว/คอลัมน์เป้าหมาย; Runtime เลื่อน viewport ให้เซลล์เป้าหมายอยู่ในภาพ. Event `GridCellPressed` เดิมยังใช้ได้.
- UI Showcase หน้า Layouts แสดงการเลือกเซลล์และอัปเดตสถานะจาก event เพื่อให้ผู้ใช้ตรวจจริง.

## เกณฑ์ตรวจรับโดยผู้ใช้

1. เปิดหน้า Layouts คลิกเซลล์ แล้วกรอบสีน้ำเงินตรงกับเซลล์และสถานะ Row/Column.
2. กด Arrow, Page Up/Down, Home/End, Ctrl+Home/End; กรอบและ viewport เดินตามโดยไม่ออกนอกขอบตาราง. Tab ย้าย focus ออกได้.
3. ลองที่ DPI/ขนาดหน้าต่างต่างกัน และตรวจว่าเส้นกรอบไม่เบี่ยงจากเซลล์. หาก Direct2D ใช้ไม่ได้ ให้ตรวจ GDI fallback.

## งานที่ยังไม่รวม

Cell editor ที่วางทับเซลล์, selection หลายช่อง, clipboard ช่วงข้อมูล, frozen panes, dirty-cell rendering, metric latency, virtual accessibility และ backend macOS/Linux ยังต้องออกแบบและทำในรอบถัดไป. การยอมรับ 0.1.33 ไม่ได้ยืนยันว่า component/cursor ผ่านทุกสภาพแวดล้อม.

ผลทดสอบเพิ่มเติมจาก UI Showcase หน้า Excel พบการวาด native TextInput ผิดตำแหน่ง/ซ้อนกันเมื่อ scroll. บันทึกการวิเคราะห์และเกณฑ์แก้ไว้ใน [แผน Rev.35](craft-phase1-rev35.md); ยังไม่ถือว่าประเด็นนี้ได้รับการแก้ใน 0.1.34.
