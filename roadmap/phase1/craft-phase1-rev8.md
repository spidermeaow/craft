# Craft Phase 1 Rev.8 — ปิด full regression และทำ package system ให้เสถียร

วันที่: 20 กันยายน 2026  
ฐาน: Rev.7 / Craft 0.1.7  
เป้าหมายรุ่น: 0.1.8  
สถานะ: ดำเนินการและผ่าน acceptance แล้ว; ดู [สถานะ Rev.8](output/phase1-rev8-status.md)

## 1. เป้าหมาย

Rev.8 มีเป้าหมายให้ Rev.7 ผ่าน full regression ตามหลักฐานจริง และทำให้ package system ใช้งานในโปรเจ็คได้ง่าย คาดเดาได้ และตรวจปัญหาได้ชัดเจน งานทั้งสองส่วนต้องจบในรุ่นเดียวกันก่อนเปิดงานขนาดใหญ่รุ่นถัดไป

Craft ยังคงเป็นภาษาสำหรับสร้างโปรแกรมและเครื่องมือทั่วไป ผู้ใช้สามารถสร้าง API Framework, migration หรือเครื่องมือเฉพาะทางเป็น package ของตนเองได้ แต่สิ่งเหล่านี้ไม่ใช่ release gate ของ Rev.8

## 2. ขอบเขตที่รับเข้า

### 2.1 ปิด Rev.7 full regression

- รัน `pwsh -NoProfile -File .\scripts\test-rev7.ps1 -Regression` โดยต้องไม่ข้าม driver หรือ fixture ใดโดยเงียบ ๆ
- ครอบคลุม regression ของภาษา, parser, checker, formatter, import/module, bundle, CLI, HTTP และ database ตั้งแต่ Rev.1 ถึง Rev.7
- ยืนยัน HTTP body deadline, request deadline, database deadline, pool wait cancellation, disconnect, transaction cleanup, shutdown และ restart แยกเป็นคนละ contract
- ตรวจ MySQL, PostgreSQL และ SQL Server ด้วยฐานข้อมูลทดสอบจริง พร้อมรายงาน version, elapsed time, pool quiescence และสาเหตุของ error โดยไม่เปิดเผย DSN หรือ secret
- ทำ repeated lifecycle, ambiguous-commit fault injection, race check, OS-handle check, installer execution และ SQL Server TCP/TLS เป็นหลักฐานก่อนรับรอง stability
- แก้เฉพาะ defect ที่พิสูจน์จาก failure และเพิ่ม regression test ให้ defect ทุกตัว

### 2.2 ทำ package system ให้ใช้งานง่ายและเสถียร

Rev.8 รองรับ package แบบ local workspace เป็นแกนหลัก ใช้ package เพื่อแบ่งเครื่องมือและ library ภายในโปรเจ็คได้ทันที โดยยังไม่เพิ่ม registry หรือการดาวน์โหลดจาก GitHub

- กำหนดรูปแบบ manifest ที่เป็นทางการสำหรับชื่อ, version และ dependencies โดยยังอ่านรูปแบบ dotted keys เดิมได้เพื่อไม่ทำให้โปรเจ็คเก่าเสีย
- Resolve path ของ dependency จากตำแหน่ง `craft.toml` เสมอ ทำ normalization และแสดง path ที่แก้แล้วใน diagnostic
- ตรวจ package name/version ซ้ำ, dependency หาย, path ไม่ถูกต้อง, import ที่ไม่ตรง manifest และ dependency cycle ก่อน run/build
- ทำ dependency graph ให้ deterministic: ลำดับการ resolve, bundle และ diagnostic ต้องเหมือนเดิมเมื่ออินพุตเหมือนเดิม
- เพิ่ม `craft package check`, `craft package list` และ `craft package lock` ให้ผู้ใช้ตรวจ graph, ดู package ที่ถูกเลือก และบันทึกผลการ resolve ได้โดยไม่ต้องอ่านโค้ดภายใน
- เพิ่ม `craft.lock` สำหรับบันทึก package local ที่ resolve แล้ว พร้อม version, normalized source และ checksum ของแหล่งที่ใช้สร้าง bundle โดยห้ามเก็บ secret
- bundle ต้องฝัง source ของ dependency ที่ผ่านการตรวจแล้ว และไม่ผูกกับ current working directory หรือ path ชั่วคราว
- ทำข้อความ error ให้มี package/import ที่เป็นต้นเหตุ, ไฟล์และบรรทัด, วิธีแก้ และระบุ cycle เป็นเส้นทางเต็ม
- อัปเดตตัวอย่างและคู่มือให้สร้าง package local แล้ว import ได้ภายในขั้นตอนสั้น ๆ โดยไม่ต้องใช้ `api-framework`

## 3. สิ่งที่ไม่ทำใน Rev.8

- remote registry, GitHub dependency, authentication, publish, upgrade หรือ semantic version range
- API Framework, OpenAPI, ORM, migration, CORS หรือ dotenv loader เป็น core feature
- เปลี่ยน syntax ภาษาโดยไม่มี regression และ migration note
- เพิ่ม database driver หรือเปลี่ยน contract timeout ของ Rev.7 เพียงเพื่อให้ผลทดสอบผ่าน

## 4. แผนดำเนินงาน

1. **M0 — Baseline:** ตรึงผล targeted ของ Rev.7, รายการ pending และ manifest/package fixtures ปัจจุบัน
2. **M1 — Regression:** รัน full matrix, แยก failure ตาม contract, แก้ defect และบันทึกหลักฐานซ้ำ
3. **M2 — Resolver:** ทำ manifest compatibility, path normalization, validation, cycle/duplicate diagnostics และ deterministic graph
4. **M3 — Lock และ bundle:** ทำ lockfile, checksum, reproducible bundle และตรวจว่า bundle ไม่รั่ว secret
5. **M4 — CLI และเอกสาร:** เพิ่มคำสั่ง package, ตัวอย่าง package local, migration note และคู่มือ troubleshooting
6. **M5 — Acceptance:** รัน full regression และ package acceptance ซ้ำ, build/installer, อัปเดต status และประกาศ Rev.8 เฉพาะเมื่อหลักฐานครบ

## 5. เกณฑ์ผ่าน Rev.8

- `scripts/test-rev7.ps1 -Regression` exit code 0 และไม่มี driver/fixture ถูก skip โดยไม่รายงาน
- regression เดิมของ Rev.1–Rev.7 และ package acceptance ผ่านบน Windows target ที่รองรับ
- package ที่มี dependency หลายชั้น resolve, check, list, lock, build และ run ได้จาก directory ใดก็ได้ในโปรเจ็ค
- dependency หาย, cycle, duplicate และ checksum เปลี่ยนต้อง fail ด้วย error ที่ระบุต้นเหตุและแนวทางแก้
- สองการ build จาก source และ lock เดียวกันได้ bundle ที่เทียบเท่ากัน และไม่มี secret ใน manifest, lock หรือ bundle
- เอกสารระบุชัดว่า local package คือสิ่งที่ Rev.8 รับรอง ส่วน remote package เป็นงานรุ่นถัดไป

ดู baseline ได้ที่ [Rev.7](craft-phase1-rev7.md), [สถานะ Rev.7](output/phase1-rev7-status.md) และ [ทิศทางภาษา Craft](../craft-language-direction.md)
