# Craft Phase 1 — สถานะปัจจุบัน

อัปเดต: 17 กันยายน 2026  
เวอร์ชัน: Craft 0.1.0  
สถานะ: พัฒนาฟังก์ชันตามขอบเขต Phase 1 แล้ว และอยู่ระหว่างผู้ใช้ทดลองใช้งานจริง

## ภาพรวม

Craft มีแกนภาษาที่เขียนด้วย Go ตั้งแต่ Lexer, Parser, AST, Static Type Checker ไปจนถึง Interpreter พร้อม CLI และตัวติดตั้ง Windows ผู้ใช้เริ่มทดลองใช้งานแล้ว และแจ้งว่า Phase 1 ดำเนินไปได้อย่างสวยงาม

สถาปัตยกรรมปัจจุบัน:

```text
.craft source
    → Lexer
    → Parser / AST
    → Static Type Checker
    → Interpreter
    → Program output
```

Frontend แยกจาก Execution Layer เพื่อรองรับการเพิ่ม backend ในอนาคต

## สิ่งที่พัฒนาแล้ว

| ส่วน | สถานะและความสามารถ |
| --- | --- |
| CLI | มี `version`, `new`, `init`, `run`, `check`, `build`, `clean` และ `help` |
| Project | มี `craft.toml`, template เริ่มต้น และการอ่านหลายไฟล์ `.craft` ภายใต้ `src/` |
| Lexer | แยก token, literals, operators, comments และตำแหน่ง source |
| Parser / AST | ใช้ Recursive Descent และ Pratt Parser รองรับ statements และ expressions |
| Types | รองรับ `Int`, `Float`, `Bool`, `String`, `Void` |
| Variables | ประกาศด้วย `let` พร้อม type annotation และเปลี่ยนค่าได้ |
| Operators | Arithmetic, comparison, logical, assignment และ compound assignment |
| Functions | Parameters, return values, การเรียกข้ามไฟล์ และ recursion |
| Entry point | ใช้ `func main()` หนึ่งจุด ไม่มี parameters และไม่มี return value |
| Control flow | `if`, `else if`, `else`, `while`, `for`, `break`, `continue`, `return` |
| Output | `print(...)` รองรับหลายค่า คั่นด้วยช่องว่างและขึ้นบรรทัดใหม่ |
| Static checking | ตรวจชื่อ ตัวแปร ชนิดข้อมูล operators arguments return paths และ entry point |
| Diagnostics | แจ้ง syntax/type/runtime errors พร้อมตำแหน่งและ source context |
| Runtime | ตรวจ division by zero, numeric overflow และรองรับหยุดด้วย Ctrl+C |
| Build | สร้าง experimental source bundle นามสกุล `.craftbundle` |
| Packaging | มี Windows x64 CLI, Inno Setup installer และ SHA-256 checksums |
| Tests / docs | เตรียม unit, integration, E2E tests, fuzz targets, ตัวอย่าง และคู่มือทดลองใช้ |

ตารางนี้สรุปสถานะ implementation ไม่ใช่การยืนยันว่าทุกฟังก์ชันผ่านการทดสอบแล้ว

## สิ่งที่ส่งมอบ

- [ตัวติดตั้ง Windows](../../dist/Craft-setup.exe)
- [CLI แบบ portable](../../dist/craft.exe)
- [SHA-256 checksums](../../dist/SHA256SUMS.txt)
- [README](../../README.md)
- [คู่มือทดลองใช้และตรวจรับ](../../docs/TRY-CRAFT.md)
- [รายละเอียดกติกาภาษา](../../docs/LANGUAGE.md)
- [ตัวอย่างโปรแกรม](../../examples/)
- [สคริปต์ build](../../scripts/build.ps1)

คอมไพล์ `craft.exe` และ `Craft-setup.exe` สำเร็จแล้ว ตัว CLI ฝัง compiler, interpreter, built-ins และ template ผู้ใช้ปลายทางไม่ต้องติดตั้ง Go เพื่อใช้ Craft

## ผลการทดลองที่ทราบในปัจจุบัน

1. ผู้ใช้เรียก `craft check` ผ่าน PowerShell ในโปรเจกต์ `D:\02-Repository\my-app` ได้แล้ว
2. CLI อ่าน `src/main.craft` และแจ้ง syntax error ที่บรรทัด 106 คอลัมน์ 9 พร้อม source และ caret เมื่อพบ `var transactionNumber: Int = 1`
3. สาเหตุคือ Phase 1 ใช้ `let` สำหรับตัวแปร จึงแนะนำให้เปลี่ยน `var` เป็น `let` และส่งโค้ดตัวอย่าง Bank ที่แก้แล้วให้ผู้ใช้
4. ส่งตัวอย่าง `while` นับเลข 1–10 ให้ผู้ใช้ทดลองเพิ่มเติม
5. ผู้ใช้แจ้งภาพรวมว่า Phase 1 ดำเนินไปได้อย่างสวยงาม

ยังไม่มีผล output หลังแก้โปรแกรม Bank หรือผลรันตัวอย่าง `while` ส่งกลับมาในบทสนทนา จึงยังไม่บันทึกว่าสองกรณีนี้ผ่านการทดสอบแล้ว รวมถึงยังไม่ยืนยันจากหลักฐานดังกล่าวว่าผู้ใช้เรียก CLI ผ่าน installer หรือ portable binary

## ข้อสรุปด้านภาษาและขอบเขต

- ใช้ `func main()` ตามแผน Phase 1 ไม่ใช้ `app { start {} }` จากแนวคิดเดิม
- `let` เปลี่ยนค่าได้ในเวอร์ชันนี้ และยังไม่รองรับ keyword `var`
- ต้องระบุชนิดข้อมูลตัวแปรและ parameters ไม่มี implicit conversion ระหว่าง `Int` กับ `Float`
- `for i in 0..5` วนค่า 0–4 ไม่รวมขอบบน
- รองรับ semicolon แบบ optional; simple statements หลายคำสั่งบนบรรทัดเดียวต้องคั่นด้วย semicolon
- `.craftbundle` ยังต้องรันด้วย `craft run <bundle>` ไม่ใช่ standalone executable
- `craft build --release` ยังใช้ source bundle เช่นเดียวกับ `craft build` ไม่มี native compilation หรือ optimization
- `craft.toml` รองรับเฉพาะ scalar string keys ที่กำหนดสำหรับ Phase 1 ยังไม่ใช่ TOML reader แบบเต็ม

## การทดสอบที่ยังรอผู้ใช้ตรวจรับ

Agent ไม่ได้รัน automated tests หรือเปิดทดลอง CLI/installer ตามความต้องการให้ผู้ใช้ทดสอบเอง การคอมไพล์ไฟล์ส่งมอบสำเร็จจึงแยกจากการผ่าน runtime acceptance

- [ ] บันทึกผล `craft version` → `craft new` → `craft run` → `craft check` ของโปรเจกต์ใหม่
- [ ] ยืนยันผลโปรแกรม Bank หลังแก้ syntax: ยอดสุดท้าย `10762.5`, สถานะ `profit`, รายการต้องสงสัย `5`
- [ ] ยืนยันตัวอย่าง `while` พิมพ์เลข 1–10 คนละบรรทัด
- [ ] ทดลอง control flow, functions, หลายไฟล์ และ error cases ตามคู่มือ
- [ ] ทดลอง build, รัน bundle นอกโปรเจกต์ และ clean โดยไม่กระทบไฟล์อื่น
- [ ] รัน `go test ./...` ใน repository และบันทึกผล
- [ ] ทดลองติดตั้ง อัปเกรด ถอนการติดตั้ง และตรวจ PATH
- [ ] ทดลองบน Windows x64 ที่ไม่ได้ติดตั้ง Go

## สิ่งที่ยังไม่อยู่ใน Phase 1

Standalone executable ของโปรแกรมผู้ใช้, Bytecode VM, Native Backend, optimizer, generics, structs แบบเต็ม, interfaces, package manager, HTTP/API framework, database, migration, OpenAPI, dependency injection, async และ concurrency

คำสั่ง `craft test`, `craft fmt`, `craft add`, `craft upgrade` และคำสั่งอื่นจากเอกสารแนวคิดระยะยาวยังไม่อยู่ใน CLI ของเฟสนี้

## ขั้นตอนต่อไป

ให้ผู้ใช้ทดลองตามคู่มือ เก็บผลและแก้ปัญหาที่พบ จากนั้นยืนยัน Definition of Done ของ Phase 1 ก่อนสรุปขอบเขต Phase ถัดไป

เอกสารอ้างอิง: [แผน Phase 1](craft-phase-1.md) และ [สถานะการส่งมอบทางเทคนิค](../../docs/PHASE1-STATUS.md)
