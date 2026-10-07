# Craft Phase 1 Rev.21 status

Date: 7 October 2026  
Source version: 0.1.21  
Status: implemented; Craft Runtime test suite passed in the checked areas, with manual GUI/DB/installer checks remaining

- Archived `gui` and `api-framework` sources in `Archive/packages/`, then removed them from the active `packages/` tree.
- Converted repository GUI examples to direct `std.ui` calls and the API server example to direct `std.http` calls. Preserved the previous API server example in `Archive/examples/api-server`.
- Paused `craft gui init` and removed its embedded template from the current binary source; saved a template copy in `Archive/templates/gui`.
- Removed package distribution from the Windows installer and updated CLI, bundle language, installer version, README, and manual check guide.
- Kept general package installation, lockfiles, HTTP Engine, and UI Engine in the Runtime.

## ผลทดสอบใน `D:\02-Repository\craft-test`

ทดสอบวันที่ 7 ตุลาคม 2026 บน Windows amd64 ด้วย `dist/craft.exe` ที่รายงาน `Craft 0.1.21`, `Phase 1 Rev.21`. รายละเอียดและคำสั่งอยู่ใน `D:\02-Repository\craft-test\guides\results-2026-10-07.md`.

| หัวข้อ | ผลที่ตรวจได้ |
| --- | --- |
| ภาษาและ Standard Library | `craft check` ผ่าน; 12 test blocks ผ่าน; `run` และ source bundle ทำงาน |
| Task, Timer, Context | `craft check` ผ่าน; 5 test blocks ผ่าน; `run` จบด้วย task/timer status ตามคาด |
| Local package | `package lock/check/list`, `check`, `run`, `build` และการรัน bundle ผ่าน |
| File, Path, JSON | อ่าน `input.json` และเขียน `output.json` ได้; ข้อมูลผลลัพธ์ตรงตามคาด |
| Config และ Logging | 1 test block ผ่าน; log ปิดค่า token; `PORT=bad` ได้ `ConfigError` และไม่สร้าง output ใหม่ |
| HTTP Engine | 1 test block ผ่าน; server จริงตอบ 200, 201, 404 และ 422 ตามกรณี |
| Database | 3 test blocks ของ value/error ผ่าน; ยังไม่ได้ต่อ database จริง |
| UI Engine | `craft check` ผ่าน; เปิดหน้าต่างและรายงาน `direct2d-hardware`; ยังไม่ได้ตรวจภาพและ interaction ด้วยตา |
| Diagnostics | type/immutable error มี Craft exit code 3; runtime/unhandled error มี exit code 1 พร้อมตำแหน่งและ stack/defer ตามคาด |
| Console | 1 test block ผ่าน; args, Unicode input และ EOF ทำงาน |
| CLI และ bundle | `version`, `build --release`, run bundle ผ่าน; `craft gui init` แจ้งว่า paused ด้วย exit code 5 |

รวม 23 test blocks ที่ผ่านในชุดที่มี test block. ทดสอบ `craft check` ผ่านทั้ง 9 โปรเจกต์ปกติในชุด. โฟลเดอร์ diagnostics ตั้งใจให้บางคำสั่งล้มเหลวและตรวจ exit code แยก.

## สิ่งที่ยังต้องตรวจ

- การติดตั้ง/อัปเกรดจริงด้วย `Craft-setup.exe` 0.1.21. ตัวติดตั้งถูก build สำเร็จและ version metadata เป็น 0.1.21 แต่ยังไม่ได้ติดตั้งทดสอบ.
- ภาพ, click, scroll, resize และความลื่นของ UI ด้วยตา; การเปิดหน้าต่างยืนยันเพียง startup/backend.
- การเชื่อมต่อฐานข้อมูลจริง เพราะไม่มี DSN สำหรับการทดสอบรอบนี้.
- การติดตั้ง package จาก public GitHub ด้วย tag จริง.

ผลทดสอบนี้เป็นหลักฐานเฉพาะกรณีที่ระบุ ไม่ใช่การรับรองว่า Runtime ทุก API หรือ Craft ทั้งหมดพร้อม production. คู่มือทดสอบด้วยตนเองอยู่ที่ `docs/TRY-REV21.md` และคู่มือแยกหัวข้ออยู่ใน `D:\02-Repository\craft-test\guides\`.
