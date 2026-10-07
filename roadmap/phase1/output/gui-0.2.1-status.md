# Craft Phase 1 Rev.24 — สถานะ GUI package 0.2.1

> ผลงาน package ภายใน [Craft Phase 1 Rev.24](phase1-rev24-status.md).

วันที่: 7 ตุลาคม 2026  
สถานะ: ลง source แล้ว; รอผู้ใช้ทดสอบ

## ลง source แล้ว

- เปิด `packages/gui` กลับมาเป็น Craft library version 0.2.1 โดยห่อ public `std.ui` Runtime 0.1.24 ครบตาม signature ที่ประกาศใน `internal/stdlib/library.go`.
- เพิ่ม composition `card`, `toolbar`, `navigation`, `field` โดยใช้ node และ spacing ของ Runtime.
- เพิ่ม `examples/gui-package-lab` ที่ import package จริงและแสดง state update, TextInput, split layout, VirtualList.
- เพิ่มคู่มือ package และ `docs/TRY-REV201.md` สำหรับให้ผู้ใช้ตรวจรับ.

## ข้อจำกัด

- `craft gui init` ยังพักอยู่; ใช้ local dependency และ `craft package lock` ตามระบบ package ปกติ.
- ยังไม่มี theme tokens ที่เปลี่ยนภาพ control ได้จริง, reactive state, animation, popup หรือ widget ขั้นสูง เพราะ Runtime ยังไม่มี primitive ครบ. `card`/`toolbar`/`navigation` เป็น composition ไม่ใช่ control ใหม่.
- Canvas 0.1.24 ยังรอผลตรวจรับ GUI จริง; bundle ไม่ฝัง image assets.
- ยังไม่ได้รัน `craft check`, `craft package check`, automated tests หรือเปิด GUI ในงานนี้ตามความต้องการให้ผู้ใช้ทดสอบเอง. ยังไม่มีหลักฐาน benchmark หรือ renderer parity ของ package 0.2.1.

## ตรวจรับที่รอ

ทำตาม `docs/TRY-REV201.md`; ส่งผล error/check, ภาพหน้าจอ หรืออาการ interaction เพื่อแก้ไขต่อ. การผ่านเกณฑ์ 0.2.1 จะอัปเดตหลังได้ผลทดสอบจริง.
