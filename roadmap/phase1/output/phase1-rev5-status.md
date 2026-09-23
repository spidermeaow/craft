# Phase 1 Rev.5 — Database foundations status

วันที่ 20 กันยายน 2026 — Craft 0.1.5 / Language Support 0.2.1

สถานะ: **implementation และ database acceptance ผ่านกับฐานข้อมูลทั้งสามที่ผู้ใช้เตรียมไว้** ยังแยก regression เดิมทั้งชุด, race checks, installer execution และ editor UI เป็น pending user testing

## สิ่งที่ส่งมอบ

| Milestone | ผลส่งมอบ | หลักฐาน |
| --- | --- | --- |
| M0 | Contract ของ std.db, native handles, types, SQL parameters และ errors | [REV5-DATABASE](../../../docs/REV5-DATABASE.md) |
| M1–M2 | MySQL, PostgreSQL, SQL Server drivers; connect/ping/query/execute/transactions | Craft integration tests ผ่านทั้งสามฐานข้อมูล |
| M3 | Pool limits, execution context/timeout, function ownership, auto rollback/close, bounded results | Tests ของ pool wait/cancellation, scope cleanup, child task, repeated transactions และ result limits ผ่าน |
| M4 | ตัวอย่าง Craft, คู่มือ, scripts/test-rev5.ps1, driver license notices, CLI/editor/installer packaging | [TRY-REV5](../../../docs/TRY-REV5.md); build ledger ด้านล่าง |

ไม่สร้าง API framework เพิ่ม Database primitives อยู่ในตัวภาษาและใช้จากแอป console ได้โดยตรง

## ผลที่รันจริง

ได้รับอนุญาตให้ agent ทดสอบฐานข้อมูล Rev.5 โดยเฉพาะ ใช้ credentials ที่เข้ารหัสด้วย DPAPI โดยไม่พิมพ์ DSNs และไม่ฝังใน source/artifacts

คำสั่ง: `pwsh -NoProfile -File scripts/test-rev5.ps1` → `go test ./tests ./internal/stdlib -run '^TestRev5' -count=1 -v -timeout 180s`, exit 0

| Database | Server | Database / transport | ผล |
| --- | --- | --- | --- |
| MySQL | 8.4.11 | craft_test / localhost TCP | PASS |
| PostgreSQL | 18.6 | craft_test / localhost TCP | PASS |
| SQL Server Express | 17.0.1135.8 | craft_test / SQLEXPRESS Shared Memory | PASS |

Tests ครอบคลุม:

- Type checking/native value construction, NULL/type errors และ exact decimal precision
- CRUD ด้วย bound parameters, Unicode/quotes, binary 0/255, decimal, วันเวลา และ nullable values
- Commit, explicit rollback, auto rollback หลัง exception และ duplicate-key constraint errors
- Child task ใช้ connection ของ parent, handle ที่เจ้าของ return แล้วถูกปิด, pool stats และ transaction 270 รอบไม่สะสม registry
- Query timeout, pool wait deadline/cancellation, reuse หลัง cancelled operation และ owner exit ปิด pool
- maxRows/maxBytes, row cleanup หลัง error และ driver-message redaction
- ตัวอย่าง Craft SELECT/parameters ผ่านทั้งสาม driver พร้อม Craft value tests 3 รายการ
- Portable database source bundle ย้ายออกจาก source tree ผ่าน PostgreSQL; ตรวจว่าไม่บรรจุ runtime DSN

ใช้ fixture tables ชื่อ `craft_rev5_...` ที่สร้างเฉพาะการทดสอบและลบด้วย defer ไม่แก้ server configuration, credentials หรือข้อมูลอื่น

## Build และข้อที่รอตรวจรับ

- `go build ./...` และ release script `scripts/build.ps1 -Installer -Language` ผ่าน สร้าง `dist/craft.exe`, `dist/Craft-setup.exe`, `dist/craft-language-support-0.2.1.vsix` และ `dist/SHA256SUMS.txt` พร้อม license notices
- CLI ที่ build แล้วรายงาน Craft 0.1.5 / Phase 1 Rev.5 / windows-amd64; ตัวอย่าง `rev5-database` ผ่าน `craft check`, `craft test` 3/3 และ `craft run` ทั้งสาม driver โดยแสดง `Connections in use: 0`
- ไม่ได้รัน regression ของ Rev.1–Rev.4 หรือ Go tests ทั้งโปรเจกต์ ตาม working preference ที่ให้ผู้ใช้ทดสอบเองนอกข้อยกเว้น database Rev.5
- Installer upgrade/uninstall, Windows ที่ไม่มี Go และ editor visuals ยังไม่ได้ทดลอง
- SQL Server TCP/TLS, HTTP request-disconnect integration กับ DB, race detector และการวัด memory/resource plateau ระยะยาวยังไม่ได้รับรอง
- ผลทดสอบวันเวลายืนยันชนิดที่อ่านได้ ไม่ใช่การตรวจทุก timezone/SQL date type; unsigned overflow และ invalid UTF-8 มี conversion tests แต่ยังไม่ได้ exhaustively ตรวจทุก vendor type

## ข้อจำกัดที่ตั้งใจไว้

Function-owned handles; factory ที่คืน connection ออกนอก owner จะได้ handle ที่ปิดแล้ว ต้องเปิดใน scope ภายนอกแล้วส่งเข้า helper. Transactions ใช้ default isolation และ timeout ครอบคลุมอายุ transaction ทั้งหมด. ไม่มี ORM/migrations/savepoints/streaming cursor. ผล query materialize ภายใต้ row/byte limits แต่ driver อาจ allocate field ก่อนตรวจ limit จึงไม่รับรอง hard process-memory cap. อ่าน contract ก่อนออกแบบแอปใช้งานจริง
