# ทดลอง Craft GUI package 0.2.1

ผู้ใช้เป็นผู้ทดสอบเองตาม workflow ปัจจุบัน. ใช้ Craft CLI ที่สร้างจาก source ปัจจุบันหรือรุ่นที่มี Runtime 0.1.24 และเปิด PowerShell ใน `D:\02-Repository\craft\examples\gui-package-lab`.

```powershell
craft package lock
craft package check
craft check
craft run
```

สิ่งที่ควรเห็น: หน้าต่างเดียวที่มี header/sidebar, ช่อง Search, รายการ 10,000 แถวแบบ virtual และ status ด้านล่าง. คลิกปุ่มแล้ว status/selection เปลี่ยน; พิมพ์ภาษาไทยใน Search แล้ว status เปลี่ยน; เลื่อนรายการและ resize แล้ว UI ยังตอบสนอง. ปิดหน้าต่างแล้วโปรแกรมจบ.

ตรวจว่า package ถูกเลือกใช้จริง: `craft package list` ควรมี `gui@0.2.1`. ตัวอย่าง import `gui "gui"` และไม่ได้เรียก `std.ui` ในแอปโดยตรง. หากใช้ `CRAFT_UI_RENDERER=gdi` ให้เทียบภาพกับค่าเริ่มต้น Direct2D; ความต่างของข้อความโปร่งใส/ขอบมนใน GDI ยังเป็นข้อจำกัด Runtime.

หาก `craft check` แจ้ง lock mismatch ให้ตรวจว่า `craft.toml` อ้าง package ถูก path/version แล้วรัน `craft package lock` ใหม่. หากทดสอบจากโปรเจกต์อื่นให้คัดลอก `packages/gui` ไปไว้ภายในโปรเจกต์และใช้ path `packages/gui`.

บันทึกผล: เวอร์ชัน CLI, renderer, DPI, ผล check/run และอาการผิดปกติพร้อมขั้นตอนทำซ้ำ. รอบ implement นี้ผู้ช่วยยังไม่ได้รัน check หรือเปิดหน้าต่าง.
