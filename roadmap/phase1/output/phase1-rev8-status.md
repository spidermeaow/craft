# Phase 1 Rev.8 status

วันที่ตรวจรับ: 20 กันยายน 2026  
รุ่น: Craft 0.1.8 / Phase 1 Rev.8  
สถานะ: **ผ่าน Rev.8 acceptance และปิดขอบเขต local package system**

## ผลส่งมอบ

- รองรับ `[dependencies]` รูปแบบ `alias = { path = "...", version = "..." }` และยังอ่าน dotted keys ของ Rev.4 ได้
- resolve dependency จากตำแหน่ง `craft.toml`, เรียง graph แบบ deterministic และรายงาน missing path, version mismatch, duplicate identity และ cycle พร้อม manifest location
- เพิ่ม `craft package check`, `craft package list` และ `craft package lock`
- เพิ่ม `craft.lock` format 1 ซึ่งเก็บ root graph, relative source path และ SHA-256 ของทุก local package ไม่เก็บ environment หรือ secret
- ทุกคำสั่งที่ load project ตรวจ lockfile ถ้ามี; source หรือ graph เปลี่ยนต้องสร้าง lock ใหม่อย่างตั้งใจ
- bundle เดิมยังเป็น format 3 / language 0.1.7 ตามข้อตกลง ไม่มีการเปลี่ยน native build ใน Rev.8
- เพิ่ม `examples/rev8-packages`, `docs/REV8-PACKAGES.md`, `docs/TRY-REV8.md` และ `scripts/test-rev8.ps1`
- แก้ cleanup context ใน interpreter ให้ทุกเส้นทางคืน cancel function และผ่าน `go vet`

## หลักฐานทดสอบ

- `pwsh -NoProfile -File .\scripts\test-rev8.ps1`: exit 0 หลัง source ชุดสุดท้าย; full regression Rev.1–Rev.8 ผ่าน และไม่มี database driver ถูก skip
- MySQL 8.4.11, PostgreSQL 18.6 และ SQL Server Express 17.0.1135.8 ผ่าน lifecycle, CRUD/concurrency, transactions, disconnect, deadlines, pool cleanup, restart และ shutdown
- lifecycle ผ่านอย่างน้อยสามรอบรวม normal, stress และ race runs
- `pwsh -NoProfile -File .\scripts\test-rev8.ps1 -Race`: exit 0
- `pwsh -NoProfile -File .\scripts\test-rev8.ps1 -Stress`: exit 0; HTTP จริง 10,000 requests, 7 goroutines, heap หลัง GC ประมาณ 822 KiB ในรอบที่บันทึก
- package acceptance ผ่าน manifest ใหม่/เดิม, mixed-form rejection, transitive graph, cycle diagnostics, stale checksum, deterministic lock, deterministic bundle, CLI check/list/lock และ application run
- `go vet ./...`: exit 0
- `pwsh -NoProfile -File .\scripts\build.ps1 -Installer`: exit 0; สร้าง `dist/craft.exe`, `dist/Craft-setup.exe`, VSIX และ `dist/SHA256SUMS.txt`
- ใช้ `Craft-setup.exe` อัปเกรด installation เดิมแบบ silent สำเร็จ; executable ที่ติดตั้งรายงาน 0.1.8 และรัน package example ได้ `Hello Craft`

## Contract ที่เรียนรู้จาก regression

MySQL คืน caller และ pool ภายใน database timeout แต่ delayed statement อาจยังปรากฏฝั่ง server จน timeout ของ operation จบ รอบที่บันทึก cleanup ประมาณ 5 วินาที PostgreSQL และ SQL Server ยกเลิก server-side operation ได้เร็วกว่านั้น Craft ไม่ใช้ privileged `KILL` อัตโนมัติ และข้อจำกัดนี้บันทึกใน `docs/REV5-DATABASE.md`

## ขอบเขตการรับรอง

Rev.8 รับรอง local filesystem packages เท่านั้น ยังไม่มี registry, GitHub download, publish, upgrade หรือ semantic version range

SQL Server ในเครื่องทดสอบใช้ Shared Memory และ `127.0.0.1:1433` ไม่เปิดรับ TCP จึงไม่มีข้ออ้างว่า TCP/TLS ผ่าน Ambiguous commit fault injection ซึ่งต้องตัด transport/server ระหว่าง commit และ OS-handle telemetry แบบ platform-specific ไม่ถูกนำมาใช้เป็น pass claim; runtime ยังคงไม่ auto-retry write/commit ที่ผลไม่แน่นอน งานเหล่านี้เป็น certification ของ deployment environment ไม่ใช่งานค้างของ local package system หรือ full regression ที่ Rev.8 ปิดแล้ว
