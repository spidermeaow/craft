# Craft Phase 1 Rev.41 (0.1.41) — Language Core และ Project Modules

**สถานะ:** ลง source รอบแรกเป็น 0.1.41 candidate; รอผู้ใช้ทดสอบและตรวจ compatibility  
**ฐาน:** Craft 0.1.40 candidate; GUI package ยังคงผูกเวอร์ชันกับ toolchain จึงขยับเป็น 0.1.41 ด้วย

## ขอบเขต

1. [Multiline Syntax และ Implicit Line Continuation](craft-phase1-rev41-multiline-language-plan.md) สำหรับ Craft Language Core ทั้งระบบ รวม trailing comma และการปรับ `craft fmt` ให้ใช้ grammar เดียวกัน
2. [โมดูลภายในโปรเจกต์และ import](craft-phase1-rev41-project-modules-plan.md) เพื่อแยก source ในโปรเจกต์จาก package ภายนอกอย่างชัดเจน โดยรักษาพฤติกรรมโปรเจกต์เดิมที่ไม่ประกาศโมดูล

ทั้งสองงานต้องตรวจความเข้ากันได้กับ Standard Library, GUI, HTTP, Database, ตัวอย่าง CLI/API และ UI Showcase ก่อนรับรุ่น. รายละเอียด grammar, migration, regression cases และเกณฑ์ตรวจรับอยู่ในแผนย่อยข้างต้น.

## สิ่งที่ลง source รอบนี้

- Lexer ต่อบรรทัดหลัง operator, comma, dot และ colon; newline ภายใน `()`/`[]` ยังเป็น whitespace. Newline ก่อน operator นอกวงเล็บยังเป็น statement boundary.
- Parser รองรับ trailing comma ใน parameter, argument, array, generic function type และ `Map` type. Map literal รองรับอยู่ก่อนแล้ว.
- `craft fmt` เก็บรูปแบบหลายบรรทัดและ indentation ของรายการในวงเล็บ พร้อมตรวจ AST ก่อน/หลัง format. เพิ่มกรณี regression สำหรับ syntax นี้ แต่ยังไม่ได้รัน tests.
- เพิ่ม `module pages.excel`, `import module alias "pages.excel"` และ `import package alias "gui"`; `import alias "gui"` เดิมยังหมายถึง package. Module identity มาจาก declaration ไม่ได้อนุมานจากโฟลเดอร์; หลายไฟล์ใช้ module เดียวกันได้, private/export แยกขอบเขต, import alias เป็น file-local และตรวจวงจร import.
- เพิ่ม `examples/rev41-language` และให้ UI Showcase ใช้ project module จริงหนึ่งจุด (`showcase.labels`). ยังรอผู้ใช้ตรวจ `craft check`, `craft fmt --check`, `craft run` และแอปจริง.

ยังไม่มีผลยืนยัน Standard Library, GUI, HTTP, Database, API/CLI, bundle round-trip หรือ UI Showcase จากการรันในรอบนี้ จึงยังไม่ถือว่าผ่าน 0.1.41.

งานสร้าง binary และแจกจ่ายแอปจาก CraftBundle ยังคงอยู่ใน [Craft 0.2.2](../craft-0.2.2-deployment-plan.md) หลังรูปแบบ source/module ของ Rev.41 นิ่งแล้ว. เลข `0.2.1` ที่ปรากฏในเอกสารย้อนหลังของ `packages/gui` เป็นเวอร์ชัน package ในอดีต ไม่ใช่แผน Craft Language Core ปัจจุบัน.
