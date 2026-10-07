# Craft Phase 1 Rev.24 — สถานะ implementation

วันที่: 7 ตุลาคม 2026  
สถานะ: **ลง source แล้ว; Go build ผ่าน; ยังไม่ได้ `craft check`, รัน GUI หรือรันทดสอบในงาน implement นี้; รอผู้ใช้ทดสอบ**

## หัวข้อพัฒนา package ใน Phase 1 Rev.24

บนฐาน UI Runtime 0.1.24 มีการนำ `packages/gui` กลับมาเป็น Craft library รุ่น 0.2.1 พร้อม wrapper/composition และ `examples/gui-package-lab`. สถานะ package อยู่ที่ [package 0.2.1 status](gui-0.2.1-status.md) ภายใต้ Phase 1 Rev.24; Craft version ยังคงเป็น 0.1.24.

## ลง source แล้ว

- ชนิด `UiDraw`, API สี/รูปทรง/เส้น/ข้อความ/ภาพ และ retained Canvas node ผ่าน `std.ui`.
- ตรวจขอบเขตจำนวนคำสั่งและข้อมูลภาพ; decode PNG/JPEG เป็น premultiplied BGRA พร้อม cache ต่อ session.
- วาด Canvas ด้วย Direct2D/DirectWrite, GDI fallback, clip ใน viewport, GPU bitmap cache และ release เมื่อภาพออกจาก UI tree.
- `CanvasPressed` ส่ง node ID กับ local coordinates; update คำสั่งที่ไม่เปลี่ยน geometry สามารถ invalidate เฉพาะ Canvas.
- `CRAFT_UI_RENDERER=gdi` สำหรับเลือก fallback; trace เพิ่มจำนวน commands กับ image bytes ของ tree ล่าสุด.
- Paint Lab, คู่มือ API/การทดสอบ และเลข CLI/installer 0.1.24.

## ข้อจำกัดและสิ่งที่ต้องรับผลทดสอบ

- `go build -o dist\craft.exe .\cmd\craft` ผ่าน; **ยังไม่ได้ยืนยัน Craft type-check หรือ GUI interaction** ในงานนี้. อย่าอ้างว่าเกณฑ์เหล่านั้นผ่านจนกว่าจะมีผลจริง.
- ภาพอ้างไฟล์จาก working directory; source bundle ยังไม่ฝัง asset ภาพ. ต้องส่งไฟล์คู่กับ bundle หรือออกแบบ asset manifest ภายหลัง.
- GDI fallback สำหรับข้อความที่มี alpha ยังวาดแบบทึบ; anti-alias/rounding อาจต่างจาก Direct2D. ต้องรับภาพเปรียบเทียบจากผู้ใช้ก่อนปรับ.
- ยังไม่มี path/gradient/shadow/transform, per-shape hit testing, reactive state, multi-window หรือ application-specific editor.
- ยังไม่ได้วัด input-to-present; trace หยุดที่ draw completion.

วิธีทดลอง: `docs/TRY-REV24.md`.
