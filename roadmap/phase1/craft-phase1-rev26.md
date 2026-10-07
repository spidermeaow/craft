# Craft Phase 1 Rev.26 (0.1.26) — Responsive UI composition

วันที่: 7 ตุลาคม 2026  
สถานะ: ลง source; รอผู้ใช้ตรวจ GUI จริง

## เป้าหมาย

แก้ช่องว่างจากผลทดสอบล่าสุดที่ข้อความปุ่มถูกตัดและ dashboard พิกัดคงที่ไม่พอดีกับหน้าต่างหลายขนาด. ให้ผู้พัฒนา Craft วัดข้อความก่อนกำหนดความกว้าง control และจัดกลุ่มการ์ดเป็นคอลัมน์ตามพื้นที่จริง.

## งาน Runtime

- เปิด native text measurement ของ UI host เป็น `std.ui.measureTextWidth`, ใช้ font size/weight และ logical pixels ที่สอดคล้องกับ renderer.
- ตรวจ handle/argument และคง API เดิมให้ใช้ได้.

## งาน package

- เปิด wrapper การวัดข้อความใน `packages/gui` 0.2.3.
- เพิ่ม `responsiveGrid` เป็น composition บน Row/Column/size/spacing เดิม. แอปเป็นเจ้าของ breakpoints, viewport, state และ overflow.
- ปรับ `examples/gui-package-lab` ให้เห็นการเปลี่ยนจำนวนคอลัมน์ระหว่าง resize.

## เกณฑ์ตรวจรับ

1. `craft version` รายงาน 0.1.26 และ package dependency เป็น 0.2.3.
2. หน้าต่างตัวอย่างสลับจำนวนคอลัมน์โดยไม่ตัดการ์ดเมื่อขนาดพอสำหรับหนึ่งการ์ด.
3. ความกว้างที่วัดของข้อความไทย/อังกฤษใช้กำหนด minimum control width ได้โดยไม่ตัดข้อความที่ DPI 100%/125%.
4. การเปิด GUI, เปลี่ยน DPI, และ regression ของ input/event ตรวจโดยผู้ใช้ตาม `docs/TRY-REV26.md`.

## ข้อจำกัด

Canvas แบบ fixed coordinates ใน `craft-test` ยังต้องแยกเป็นองค์ประกอบ responsive หรือวาดใหม่จาก viewport เอง. Rev.26 ไม่เพิ่ม CSS-like auto layout, text wrapping หรือการ reflow Canvas โดยอัตโนมัติ. ดู [รายละเอียด API](../../docs/REV26-UI-RESPONSIVE.md) และ [สถานะ implementation](output/phase1-rev26-status.md).
