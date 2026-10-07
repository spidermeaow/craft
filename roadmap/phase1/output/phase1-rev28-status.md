# Craft Phase 1 Rev.28 — สถานะ implementation

วันที่: 7 ตุลาคม 2026  
สถานะ: ลง source แล้ว; รอผู้ใช้ทดสอบ CLI/GUI

## ลง source แล้ว

- Runtime 0.1.28: `std.ui.textWrap` แบบ opt-in บน Text; คำนวณบรรทัด/ความสูงจาก allocated width และวาดตามบรรทัดที่ layout ใช้. เพิ่ม `std.ui.grid` ที่คำนวณคอลัมน์ใน Runtime จาก parent width.
- Package 0.2.6: wrapper `textWrap`/`grid`, helper `autoGrid` และ `adaptiveShell`; คง `responsiveGrid` เดิม.
- อัปเดต `examples/gui-package-lab`, `gui-package-dashboard` และ `responsive-tester` พร้อมสำเนา package/lockfile.
- CLI และ Windows Setup ระบุเวอร์ชัน 0.1.28; source package ระบุ 0.2.6.

## หลักฐานและข้อจำกัด

`go build` สำเร็จ. ยังไม่ได้รัน automated tests, `craft check`, `craft run` หรือเปิด GUI เพราะผู้ใช้ทดสอบเอง. การแบ่งบรรทัดภาษาไทยใช้ขอบเขตข้อความพื้นฐานและ fallback ตามตัวอักษร ยังไม่ใช่ Unicode line-breaking เต็มรูปแบบ. Grid ต้องใช้ ScrollView หากหน้าต่างแคบกว่าความกว้างขั้นต่ำของ cell. Canvas พิกัดคงที่, Button และ TextInput ไม่ wrap อัตโนมัติ. ดู [คู่มือ Rev.28](../../../docs/REV28-UI-REFLOW.md).
