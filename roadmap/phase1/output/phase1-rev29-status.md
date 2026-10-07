# Craft Phase 1 Rev.29 — สถานะ implementation

วันที่: 7 ตุลาคม 2026  
สถานะ: ลง source/binary/Setup แล้ว; รอผู้ใช้ทดสอบจริง; งาน accessibility ยังไม่ครบแผน

Runtime 0.1.29 เพิ่ม focus policy สำหรับ Button/TextInput, focus request, การเลื่อน ScrollView เมื่อ focus, focus ring ของ Button/Canvas target ใน GDI และ Direct2D, keyboard shortcut ระดับหน้าต่าง และ `CommandInvoked`. คำสั่งไม่ยิงซ้ำจาก key repeat; native EDIT เก็บคีย์แก้ไขข้อความและ IME ไว้. Package 0.2.7 เพิ่ม wrapper และ command helper. อัปเดต `gui-package-lab`, `gui-package-dashboard`, `responsive-tester`, package source/lockfile และ Setup.

`std.ui.accessible` เก็บ name/role metadata ใน UiNode แล้ว แต่ยังไม่มี Windows UI Automation provider สำหรับ control ที่ Runtime วาดเอง. Focus scope ของ Dialog/Popup, Canvas target policy และ screen-reader acceptance ยังไม่เสร็จ. ดังนั้น Rev.29 **ยังไม่ผ่านเกณฑ์ครบตามแผน** แม้ตัวเลขเวอร์ชันและ Setup เป็น 0.1.29. ดู [คู่มือ](../../../docs/REV29-UI-FOCUS-COMMANDS.md).

Go build, package lock, static `craft check` ของสามตัวอย่าง และ Inno Setup compile สำเร็จ. ไม่ได้รัน automated tests หรือเปิดแอปทดสอบตามความต้องการของผู้ใช้.
