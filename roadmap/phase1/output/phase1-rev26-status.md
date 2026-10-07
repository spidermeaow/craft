# Craft Phase 1 Rev.26 — สถานะ implementation

วันที่: 7 ตุลาคม 2026  
สถานะ: ลง source แล้ว; รอผู้ใช้ทดสอบ GUI

## เปลี่ยนแล้ว

- Runtime เพิ่ม `std.ui.measureTextWidth` บน UI host ของ Windows โดยใช้ renderer และ DPI ของหน้าต่าง.
- `packages/gui` 0.2.3 เพิ่ม wrapper และ `responsiveGrid` สำหรับการ์ดหลายคอลัมน์ตามพื้นที่จริง.
- ตัวอย่าง `examples/gui-package-lab` ใช้ grid ใหม่; CLI และ Windows Setup ระบุ 0.1.26.
- เพิ่มคู่มือ API และวิธีตรวจด้วยตนเองใน `docs/REV26-UI-RESPONSIVE.md` และ `docs/TRY-REV26.md`.

## หลักฐานและสิ่งที่ยังไม่ตรวจ

`go build -o craft.exe ./cmd/craft` สำเร็จ และสร้าง package lock ของตัวอย่างใหม่แล้ว. ยังไม่ได้รัน automated tests, Craft checker หรือ GUI ตามความต้องการให้ผู้ใช้ทดสอบเอง. Dashboard Canvas เดิมใน `craft-test` ยังใช้พิกัดคงที่; การจัด grid ใน package ไม่เปลี่ยน Canvas นั้นโดยอัตโนมัติ.

สร้างไฟล์แจกใหม่ด้วย `scripts/build.ps1 -Installer` แล้ว: `dist/craft.exe` รายงาน 0.1.26 และ metadata ของ `dist/Craft-setup.exe` ระบุ ProductVersion 0.1.26. การ build นี้ไม่ใช่ผลทดสอบการติดตั้งหรือการใช้งาน GUI.
