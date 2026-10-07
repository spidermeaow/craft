# วิธีให้ผู้ใช้ตรวจรับ Craft 0.1.25

ใช้ Craft CLI ที่ build จาก source 0.1.25 และเปิด PowerShell ใน `D:\02-Repository\craft-test\gui-package-dashboard`. โปรเจกต์ทดสอบนี้ใช้สำเนา `packages/gui` 0.2.1 เดิม ไม่แก้ source package ระหว่างตรวจ Runtime.

```powershell
& 'D:\02-Repository\craft\dist\craft.exe' version
& 'D:\02-Repository\craft\dist\craft.exe' package check
& 'D:\02-Repository\craft\dist\craft.exe' check
& 'D:\02-Repository\craft\dist\craft.exe' run
```

ตรวจ 1: ที่ DPI 100%/125% และหน้าต่างขนาดเล็ก/ใหญ่ หน้า compact ต้องแสดงการ์ดภายในความกว้าง viewport. ขยายหน้าต่างจนกว้างมากเพื่อดู dashboard แบบเดิม; ส่วนที่ใหญ่เกิน viewport ต้องเลื่อนแนวนอนด้วย scrollbar ได้. ขยับขนาดหลายครั้งแล้วภาพ/ตำแหน่ง input ไม่ค้าง.

ตรวจ 2: คลิกพื้นที่ `เอกสาร 24`, `ข้อมูล 1,248` และ `Refresh` ในหน้า compact; status ต้องแสดง target ID ที่ต่างกัน. ใช้ Tab โฟกัสแต่ละ target และ Enter/Space; status ต้องเปลี่ยนเช่นกัน. คลิก sidebar, พิมพ์ไทย/อังกฤษใน TextInput และเลื่อน ScrollView เพื่อดูว่า interaction เดิมยังทำงาน.

ตรวจ 3: ดูขนาด/น้ำหนักข้อความหัวข้อ, ขอบปุ่ม native และ Canvas text. ตั้ง `$env:CRAFT_UI_RENDERER='gdi'` ก่อนรันอีกครั้งเพื่อเทียบ fallback; หลังตรวจให้ลบด้วย `Remove-Item Env:CRAFT_UI_RENDERER`. บันทึก renderer, DPI, ขนาดจอ/หน้าต่าง และภาพที่ต่างกัน.

หากต้องวัด trace ให้ตั้ง `$env:CRAFT_UI_TRACE='1'` ก่อนรัน แล้วบันทึก startup, layout/paint และการตอบสนองขณะ resize/scroll. ผลนี้วัดถึง draw completion เท่านั้น. ไม่มีเกณฑ์ผ่านด้าน input-to-present จนกว่าจะมีเครื่องมือวัดภาพบนจอ.

ผู้ช่วยยังไม่ได้รัน `craft check`, automated tests หรือเปิด GUI ในงาน implement นี้. ผล Go build เพียงยืนยันว่า source ฝั่ง Go คอมไพล์ได้ ไม่ยืนยันภาพและ interaction.
