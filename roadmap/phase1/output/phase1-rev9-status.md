# Phase 1 Rev.9 status

วันที่ตรวจ: 21 กันยายน 2026  
รุ่นเป้าหมาย: Craft 0.1.9 / Phase 1 Rev.9  
สถานะ: **พร้อมติดตั้ง Craft 0.1.9 บน Windows x64**

## ผลส่งมอบ

- เพิ่ม `std.path.clean`, `std.path.absolute` และ `std.path.relative` โดยใช้ host path semantics และ relative path ของ process working directory
- เพิ่ม `std.fs.writeTextAtomic`, `std.fs.readBytes` และ `std.fs.writeBytesAtomic`
- atomic write สร้าง temporary file ใน destination directory, ปิดไฟล์ก่อน replace และ cleanup temporary file แบบ best effort เมื่อไม่สำเร็จ
- read text/bytes ตรวจ regular file และกำหนด complete-file limit 16 MiB; bytes ไม่ถูก decode เป็น UTF-8
- filesystem diagnostics ระบุ operation และ requested path โดยไม่แสดง arbitrary OS error text, file contents หรือ environment values
- เพิ่ม `examples/rev9-tools`, `docs/REV9-TOOLING.md`, `docs/TRY-REV9.md` และ `scripts/test-rev9.ps1`
- logging ยังไม่รับเข้า Rev.9 เพราะยังไม่มี contract ของ redaction/structured fields/output ที่ไม่ผูก application policy

## หลักฐานทดสอบ

- `pwsh -NoProfile -File .\scripts\test-rev9.ps1`: **exit 0**
- full regression Rev.8 ผ่าน รวม language, formatter, lexer/parser fuzz seeds, project/package/bundle/CLI, HTTP, database และ Rev.1–Rev.7 test matrix
- `TestRev9AtomicFilesystemAndPaths`: ผ่าน text atomic replacement, destination directory failure ที่ไม่ทำให้ target เดิมหาย, binary bytes round-trip, `clean`, `absolute` และ `relative`
- database lifecycle ผ่านกับ MySQL 8.4.11, PostgreSQL 18.6 และ SQL Server Express 17.0.1135.8 ใน run นี้

## ผลตรวจ release เพิ่มเติม

- `scripts/test-rev9.ps1 -Race -Stress`: exit 0; full regression และ HTTP 10,000 requests ผ่าน (build/rev9-release-tests.log)
- Windows tests ผ่าน locked destination, read-only destination, original preservation, temporary cleanup, UNC lexical paths, mixed separator, drive-relative และ relative path ข้าม volume รวม write limit/invalid UTF-8 rejection
- `go vet ./...`: exit 0
- CLI/packaging เป็น 0.1.9; build installer exit 0 (build/rev9-release-build.log)
- silent installer upgrade exit 0; installed executable รายงาน 0.1.9 / Rev.9, check ตัวอย่างผ่าน และเขียน JSON ได้สอง keys (build/rev9-install.log)
- UNC ตรวจเฉพาะ path semantics ไม่ได้รับรอง I/O ผ่าน network share; read-only failure ไม่ใช่การตรวจ Windows ACL ทุกแบบ

## ขอบเขตที่ยังไม่รับรอง

atomic replacement เป็น best effort ตาม filesystem/OS: ไม่รับรอง multi-file transaction, crash/power-loss durability, filesystem sandbox หรือการป้องกัน race จาก process อื่น ไม่มี recursive deletion, dotenv loader, process execution, watcher หรือ logging API ใน Rev.9
