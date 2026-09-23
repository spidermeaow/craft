# Craft Phase 1 Rev.1 — สถานะการส่งมอบ

เวอร์ชัน: **0.1.1**  
สถานะ: implementation พร้อมสำหรับผู้ใช้ทดสอบรับงาน; ยังไม่ยืนยันว่า Definition of Done ด้าน tests ผ่านครบ

## สิ่งที่ทำแล้ว

| กลุ่ม | การเปลี่ยนแปลง |
| --- | --- |
| Language Safety | `let` immutable, `var` mutable, type คงเดิม, scope และ return-path analysis |
| Error Handling | Exception, try/catch/throw, runtime faults ที่จับได้, propagation และ stack trace ข้ามไฟล์ |
| Core Data | Array ซ้อนกันได้, index/length/append/remove, foreach และ String operations |
| Value Safety | Array deep-copy ที่ binding/assignment/argument/return ป้องกันการแก้ let ผ่าน alias |
| Runtime | Function frames, defer แบบ LIFO และ unwinding เมื่อ return/throw/จบฟังก์ชัน |
| Functions | Named arguments โดยรักษาลำดับ evaluation ตาม source |
| Developer Tools | craft fmt, fmt --check, craft test, test blocks, assert และ reporter |
| Diagnostics | Error codes E1001–E5001 และ exit codes แยกประเภท |
| Library Foundation | แยก built-ins จาก Interpreter; String/Array methods, math/fs/time/json/env ขั้นต้น |
| Compatibility | Migration Note, ตัวอย่างเดิมใช้ var เมื่อต้องเปลี่ยนค่า, bundle format 2 ระบุ language 0.1.1 |
| Delivery | Portable CLI และ installer ปรับ version เป็น 0.1.1 พร้อมสคริปต์ build/checksums |

## ส่วนที่เป็น foundation หรือเลื่อนไป Rev.2

Struct, Map และ Module/import มีเอกสารข้อกำหนดเบื้องต้น ยังไม่มี syntax/runtime ให้ใช้จริง โดยอิงลำดับท้าย roadmap ที่วาง Module, Map, Optional, Lambda, Callback และ Event Library ไว้ Rev.2

ยังไม่รองรับ Enum, Optional, default parameters, lambda, callback, Event Bus, REPL, full JSON structured parsing, native compilation หรือ VM `defer` และ named arguments ทำใน Rev.1 แล้ว ไม่ได้เลื่อน

รายละเอียด: [REV1-FOUNDATIONS.md](../../../docs/REV1-FOUNDATIONS.md)

## Validation

- ตรวจทาน source และจัดรูปแบบ Go source
- คอมไพล์ release ด้วย `scripts/build.ps1 -Installer` สำเร็จแล้ว: มี `dist/craft.exe`, `dist/Craft-setup.exe` และ `dist/SHA256SUMS.txt` รุ่น 0.1.1
- เพิ่ม tests สำหรับ immutability, deep copy, named arguments, exceptions, trace, defer, CLI exit codes, formatter และ test discovery
- **Agent ไม่ได้รัน automated tests, Craft programs หรือ installer ตามความต้องการของผู้ใช้**
- การคอมไพล์สำเร็จไม่ใช่การยืนยันว่า runtime/installer acceptance ผ่านแล้ว

## ผู้ใช้ตรวจรับต่อ

1. ติดตั้งรุ่นใหม่และตรวจ `craft version` ให้เป็น 0.1.1
2. สร้างโปรเจกต์ใหม่ ลอง run/check/test/fmt
3. ทดลอง `examples/rev1`: main แสดงตัวอย่าง และ craft test คาดหวัง 6 tests ผ่าน
4. ทดลอง let mutation, syntax/type error และ unhandled exception พร้อม `$LASTEXITCODE`
5. ทดลอง `craft build`, bundle execution และ clean
6. รัน `go test ./...` จาก repository
7. ตรวจ installer upgrade/uninstall/PATH และเครื่อง Windows x64 ที่ไม่มี Go

คู่มือ: [TRY-CRAFT.md](../../../docs/TRY-CRAFT.md) · [Migration Note](../../../docs/REV1-MIGRATION.md)
