# Craft Phase 1 Rev.19 — ScrollView sizing correctness

วันที่: 7 ตุลาคม 2026  
ฐาน: Craft 0.1.18  
เป้าหมายรุ่น: Craft 0.1.19  
สถานะ: implement ใน source แล้ว; รอผู้ใช้ตรวจรับ GUI จริง

## หลักฐานที่พบจากการทดสอบ 0.1.18

- `go test ./...` และ `craft check` ของ Stress Lab ทั้งใน repo และ `D:\02-Repository\test-gui` ผ่าน แต่ชุดเดิมไม่มีการทดสอบ ScrollView โดยตรง
- ภาพหน้าต่างจริง 8×8 ของ 0.1.18 แสดงปุ่ม control ฝั่งซ้ายกว้างเกือบครึ่งหน้าต่าง จนคอลัมน์สุดท้ายของ grid ถูก clip ทั้งที่พื้นที่รวมพอ
- เมื่อ viewport สูงกว่าเนื้อหา ScrollView ขยาย child ให้สูงเท่า viewport ทำให้แถว/ปุ่มใน grid สูงผิดธรรมชาติ
- Native host รับ PageDown, wheel และ click หลังเลื่อนได้ถูกต้องในกรณีที่ทดสอบ; ปัญหาหลักที่ยืนยันได้อยู่ที่กฎจัดพื้นที่ ไม่ใช่การแปลงพิกัด click
- คำสั่ง `craft version` ของ binary 0.1.19 ยังพิมพ์ `Language: Phase 1 Rev.13` จาก metadata เก่า จึงต้องแสดง milestone ปัจจุบันให้ชัด

## งาน 0.1.19

1. ให้ ScrollView ที่อยู่เป็นลูกตรงของ Row/Column รับพื้นที่ว่างก่อน container พี่น้อง; controls/toolbar คง preferred size ของตน
2. ให้ content ใน ScrollView คง intrinsic width/height แม้ viewport ใหญ่กว่า; offset ยัง clamp ตามส่วนที่ล้นจริง
3. เพิ่ม regression test ของ layout, clipping/hit test, native PageDown, wheel สองแกน, click หลังเลื่อน และ update ที่ทำให้เนื้อหาหด
4. ตรวจภาพจริง 8×8/32×32 และ trace; อัปเดต GUI package, template, CLI/bundle version, installer และเอกสาร
5. แก้ข้อความ `craft version` ให้รายงาน milestone Phase 1 Rev.19 โดยไม่สื่อว่ากฎภาษาเปลี่ยนตามทุก revision ของ UI Engine

## เกณฑ์ตรวจรับ

- 8×8 เห็นครบทุกคอลัมน์ในหน้าต่างอ้างอิง; controls ไม่ขยายกินพื้นที่โดยไม่จำเป็น และปุ่ม/แถวมีขนาดตามเนื้อหา
- 32×32 เลื่อนแนวตั้งและแนวนอนถึงเนื้อหาส่วนท้ายได้; คลิกตรงภาพหลังเลื่อน
- Offset คงอยู่เมื่อ update แล้วเนื้อหายังล้น และ clamp ลงเมื่อเนื้อหาหด/หน้าต่างขยาย
- Direct2D/GDI, DPI และการ resize จริงต้องให้ผู้ใช้ตรวจรับก่อนประกาศผ่าน; `go build` หรือ test ที่ผ่านเพียงอย่างเดียวไม่ยืนยันภาพทุกเครื่อง

## นอกขอบเขต

Virtualized List/Grid, scrollbar ที่ลากได้, momentum animation, TextBox/IME, Craft Office และ native standalone application binary. การแก้ขนาด ScrollView ให้ถูกต้องเป็นฐานก่อนเพิ่ม virtualization.
