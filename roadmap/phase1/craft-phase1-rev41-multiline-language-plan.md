# Craft Phase 1 Rev.41 (0.1.41) — Multiline Syntax และ Implicit Line Continuation

สถานะ: **ลง source รอบแรกใน 0.1.41 candidate; รอผลทดสอบและ compatibility**  
ขอบเขต: Craft Language Core ทั้งระบบ; ไม่ใช่การแก้เฉพาะ UI หรือ package ใด package หนึ่ง  
แผนร่วมรุ่น 0.1.41: [โมดูลภายในโปรเจกต์และ import](craft-phase1-rev41-project-modules-plan.md)  
งานต่อเนื่อง 0.2.2: [การแพ็ก Desktop และ Server จาก CraftBundle](../craft-0.2.2-deployment-plan.md)

## เป้าหมาย

ให้โค้ด Craft ขนาดใหญ่แบ่งบรรทัดได้อย่างเป็นธรรมชาติ โดยไม่ต้องใช้ backslash และโดยที่โค้ดบรรทัดเดียวกับหลายบรรทัดมี AST, type, ผลรัน และขอบเขต statement/block เหมือนกัน

## สัญญา Grammar ที่ต้องกำหนดก่อนแก้

1. Newline ภายใน `()` และ `[]` เป็น whitespace ในรายการหรือ expression ที่ยังไม่จบ; newline หลังตัวดำเนินการ infix, comma, dot หรือ token ที่ยังต้องมี operand/member ถัดไปให้ต่อ expression โดยปริยาย
2. Newline ที่ expression สมบูรณ์แล้วและอยู่นอกวงเล็บยังคงเป็น statement boundary เดิม. ห้ามทำให้ statement ถัดไปถูกกลืนเข้า expression โดยบังเอิญ; กำหนดกรณี newline **ก่อน** operator/dot ให้ชัดพร้อมตัวอย่างที่ยอมรับและไม่ยอมรับ
3. Function call/declaration, parameter/argument lists, array literals/indexing, struct initialization, nested expressions, boolean expressions และ chained calls ต้องใช้กฎเดียวกันทุก module และ library
4. รายการที่คั่นด้วย comma รองรับ comma ตัวท้ายก่อนตัวปิดอย่างสม่ำเสมอ รวมทั้งกรณีบรรทัดเดียว; ระบุข้อยกเว้นที่ grammar จำเป็นต้องมีแทนการแก้เฉพาะ library
5. วงเล็บปิด, block scope, precedence, short-circuit, comments และตำแหน่ง error ต้องรักษาความหมายเดิม. ไม่เพิ่ม backslash continuation

ก่อน implement ให้ทำตารางตัวอย่าง valid/invalid ที่เป็น normative grammar รวมกรณี newline หลัง/before operator, nested delimiters, trailing comma, comments, blank lines, `return`, `if`, loops และ chained calls เพื่อป้องกันความกำกวม

## งาน Language Core

| ชั้น | งานที่ต้องทำ |
| --- | --- |
| Lexer | ตรวจการปล่อย newline tokens และการติดตาม delimiter depth; รักษา line/column ของ token และ comments |
| Parser | จัดการ newline ตาม grammar กลางในทุก expression และ comma-delimited list; แยก statement boundary กับ continuation อย่างแน่นอน |
| AST | ให้ multiline/single-line สร้างโครงสร้างความหมายเดียวกัน โดยยังรักษา source span เพื่อ diagnostics และ formatter |
| Type Checker | ตรวจว่าการแบ่งบรรทัดไม่เปลี่ยนการ resolve ชื่อ, type inference/checking, call arity หรือ overload ที่มีอยู่ |
| Interpreter | ตรวจ evaluation order, short-circuit, scope และข้อผิดพลาด runtime ให้ตรงกับรูปแบบบรรทัดเดียว |
| Formatter | `craft fmt` **มีอยู่แล้ว**; ปรับ formatter เดิมและ `craft fmt --check` ให้พิมพ์รูปแบบ multiline ตาม grammar เดียวกับ Parser, คง comments, idempotent และไม่เปลี่ยน AST/ผลรัน |

ห้ามแก้ด้วยการบังคับรวมโค้ดเป็นบรรทัดเดียวหรือเพิ่ม parser workaround เฉพาะ `std.ui`, `gui`, HTTP หรือ Database

## ลำดับดำเนินการ

1. สำรวจ Lexer, Parser, AST, Type Checker, Interpreter และ formatter ปัจจุบัน; บันทึก syntax ที่รองรับจริงและกรณีที่ยังพัง
2. ล็อก grammar และชุดตัวอย่าง valid/invalid; แก้ Language Core ก่อน
3. ปรับ `craft fmt` ให้ใช้ grammar เดียวกันและกำหนดรูปแบบ multiline ที่อ่านง่ายโดยไม่เปลี่ยนความหมาย
4. เพิ่ม regression tests ระดับ lexer/parser/AST, type checking, interpreter และ formatter; เทียบ single-line กับ multiline และตรวจ statement boundary, error span, comments และ trailing comma
5. ทดสอบ compatibility กับ source Craft เดิมทั้งหมด รวม Standard Library, GUI, HTTP, Database, UI Showcase และโปรเจกต์ตัวอย่าง; แยกผล parser/type/runtime/format และบันทึก migration หากจำเป็น
6. ส่งให้ผู้ใช้ทดสอบแอปด้วยตนเองพร้อมตัวอย่าง multiline และขั้นตอนตรวจที่ทำซ้ำได้

## เกณฑ์ตรวจรับ

- ตัวอย่าง syntax ทุกกลุ่มในข้อกำหนดผ่าน `craft check`/`craft run` และรูปแบบบรรทัดเดียวกับหลายบรรทัดให้ผลเท่ากัน
- Trailing comma ใช้ได้ใน comma-delimited lists ตาม grammar ที่ประกาศ; invalid syntax ยังรายงานตำแหน่งและสาเหตุที่อ่านเข้าใจ
- Statement boundary และ block scope เดิมไม่เปลี่ยน; regression tests ครอบคลุมกรณีต่อบรรทัดที่กำกวม
- `craft fmt` รอบที่สองไม่เปลี่ยนไฟล์, `craft fmt --check` สอดคล้องกัน, comments ไม่หาย และ AST ก่อน/หลัง format เท่ากันในเชิงความหมาย
- โค้ด Standard Library, GUI, HTTP, Database, UI Showcase และ examples เดิมยังตรวจและทำงานได้ หรือมีรายการ breaking change และ migration ที่เจ้าของโครงการยอมรับ

มี implementation รอบแรกใน Lexer, Parser และ formatter พร้อม regression cases ที่ยังไม่ได้รันตามความต้องการให้ผู้ใช้ทดสอบเอง. ยังไม่รับรองผล compatibility สำหรับ 0.1.41
