# Craft Phase 1 Rev.25 — สถานะ GUI package 0.2.2

> ผลงาน package ภายใน [Craft Phase 1 Rev.25](phase1-rev25-status.md).

วันที่: 7 ตุลาคม 2026  
สถานะ: ลง source แล้ว; รอผู้ใช้ทดสอบ

## สิ่งที่ลงแล้ว

- `packages/gui` รุ่น 0.2.2 เพิ่ม wrapper ของ Runtime 0.1.25: `viewportWidth/Height`, `textStyle`, `surface`, `drawTextStyled`, `hitRegion` โดยคง API 0.2.1.
- เพิ่ม token สีและระยะห่าง, `heading`, `primaryButton`, `surfaceCard`, `appShell`, `targetButton`, `compact`, `viewportChanged` และ `targetPressed`. แอปยังต้องถือ state, เลือก breakpoint และเรียก `update` เอง.
- `examples/gui-package-lab` ใช้ package 0.2.2 และจัด navigation ไว้บน content เมื่อ viewport แคบ.
- `craft-test/gui-package-dashboard` ใช้ `gui.*` แทนการเรียก `std.ui.*` ทั้งห้ากลุ่มใน app source; อัปเดตสำเนา package, manifest, lockfile และคู่มือ.
- `craft package lock` สำเร็จสำหรับสองตัวอย่าง. คำสั่งนี้สร้าง metadata ของ dependency ไม่ใช่ผลทดสอบ syntax หรือ GUI.

## ตรวจ source แบบอ่านอย่างเดียว

เทียบ wrapper กับ signatures ของ `std.ui` ใน `internal/stdlib/library.go`; public runtime functions ทุกตัวมี wrapper ชื่อเดียวกันใน package. ไม่มี `std.ui.` ใน app source ของ dashboard. Runtime-only API ที่ไม่ได้ห่อในรายการ public signatures ปัจจุบัน: ไม่มี. ชนิด `UiNode`, `UiDraw`, `UiWindow`, `UiEvent` และ event fields เป็นชนิด Runtime ที่ package ส่งผ่านโดยตรง.

## รอผลตรวจรับ

ยังไม่ได้รัน `craft package check`, `craft check`, automated tests หรือเปิด GUI ในงานนี้ตามการให้ผู้ใช้ทดสอบเอง. จึงยังไม่ยืนยัน syntax/type ของ helper ใหม่, interaction, DPI, GDI/Direct2D parity หรือ performance. หากพบ error ให้ส่งข้อความ error, renderer, DPI และขั้นตอนทำซ้ำ. ขั้นตอนอยู่ใน `docs/TRY-REV202.md`.

## ขอบเขตที่ยังไม่พร้อม

Canvas target ไม่ใช่ native accessibility control. ยังไม่มี automatic hover/focus/pressed styling, reactive state, dialog, table/grid จริง, animation, shadow, gradient หรือ standalone executable. ตัวเลข performance ต้องวัดหลัง GUI ผ่านการตรวจรับ; ไม่มี benchmark สำหรับ 0.2.2 ตอนนี้.
