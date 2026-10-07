# Craft Phase 1 Rev.25 — สถานะ implementation

วันที่: 7 ตุลาคม 2026  
สถานะ: **ลง source แล้ว; Go build ผ่าน; รอผู้ใช้ทดสอบ Craft/GUI ตาม `docs/TRY-REV25.md`**

## หัวข้อพัฒนา package ใน Phase 1 Rev.25

`packages/gui` 0.2.2 ห่อ primitives ใหม่ของ Runtime 0.1.25 และเพิ่ม composition/style/adaptive helpers. ตัวอย่างและสำเนา package ใน `craft-test/gui-package-dashboard` อัปเดต dependency แล้ว. [สถานะ package 0.2.2](gui-0.2.2-status.md) แยกหลักฐาน implementation/ข้อจำกัดไว้ภายใน Phase 1 เดียวกัน. ยังไม่ได้รับผลตรวจ GUI จริงจากผู้ใช้; การสร้าง Setup สำเร็จไม่ได้ยืนยันพฤติกรรม UI. `craft version` ยังคงเป็น 0.1.25.

## ลง source แล้ว

- `viewportWidth`/`viewportHeight` สำหรับ client area ใน logical pixels; `Resized`/`DpiChanged` ยังส่งขนาดเช่นเดิม.
- ScrollView มี scrollbar แนวนอนที่มองเห็นและลาก/คลิก track ได้ โดยใช้ scrollX และ clip เดิม.
- `textStyle` สำหรับ Text/Button; `surface` สำหรับ Panel/Button; `drawTextStyled` สำหรับ Canvas พร้อม font size/weight/align/wrap. DirectWrite และ GDI มีทางวาดของตน.
- `hitRegion` ใน Canvas, event pointer/target/focus/keyboard ขั้นต้น; `CanvasPressed` เดิมยังส่ง. `CanvasPointerMoved` coalesce/drop ได้เพื่อป้องกัน event queue จาก motion จำนวนมาก.
- ตัวอย่างใน `craft-test/gui-package-dashboard` ใช้ viewport เพื่อเลือกหน้า compact/wide และมี target แยก 3 จุด โดยไม่แก้ `packages/gui` 0.2.1.
- CLI/installer version เป็น 0.1.25; เพิ่มคู่มือ API และ manual test.

## ข้อจำกัดและงานตรวจรับ

- ยังไม่ได้รัน Craft checker, automated tests, GDI GUI หรือ Direct2D GUI ในรอบ implement นี้ตามความต้องการให้ผู้ใช้ทดสอบเอง. ยังไม่มี baseline สอง DPI/renderer หรือผล latency.
- Fixed Canvas ยังคงมีขนาดคงที่; แอปต้อง rebuild จาก viewport event เอง. TextInput ยังเป็น native EDIT และไม่รับ `textStyle`.
- `surface` รองรับสีพื้นทึบ/ขอบทึบหรือโปร่งใสเท่านั้น. ไม่มี shadow/gradient/transform/disabled state ระดับ widget หรือ accessibility tree สำหรับ Canvas targets.
- `drawTextStyled` wrap ในกรอบที่แอประบุ; Text/Button ยังเป็นบรรทัดเดียว. GDI alpha text ยังคงทึบ.
- ตาราง 2D, form controls ขั้นสูง, popup, animation, multi-window และ input-to-present instrumentation ยังเป็นงานถัดไป.

ผลตรวจรับและ regression ของ Rev.22–24 จะอัปเดตหลังผู้ใช้ทดลองจริง.
