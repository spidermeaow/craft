# ทดลอง Craft 0.1.13 GUI bridge บน Windows

เริ่มจาก Craft CLI 0.1.13 ที่ build ใน `dist/craft.exe` หรือจาก `Craft-setup.exe` แล้วเปิด terminal ใหม่ การทดลองเปิดหน้าต่างเป็นขั้นตอนที่ผู้ใช้ทำเอง

```powershell
craft version
cd D:\02-Repository\craft\examples\gui-bridge
craft check
craft run
```

คาดหวังหน้าต่าง `Craft GUI bridge` พื้นหลังสีขาว คลิกภายในหน้าต่างแล้ว title เปลี่ยนเป็น `Clicked in Craft` พร้อมพื้นหลังสีม่วง กดคีย์แล้ว title เปลี่ยนเป็น `Key pressed` พร้อมพื้นหลังสีส้ม กดปุ่มปิดแล้วโปรแกรมออกโดยไม่เหลือ process ค้าง

ตรวจ bundle:

```powershell
craft build
craft run dist\gui-bridge.craftbundle
```

ผลด้านหน้าต่างควรเหมือน source; bundle ยังต้องใช้ Craft CLI ตรวจ package ที่แจกมาใน CLI:

```powershell
cd <your-existing-craft-project>
craft gui init
craft package check
```

คาดหวัง `packages/gui` และ dependency/lock ของ GUI รุ่น 0.1.13 สั่ง `craft gui init` ซ้ำต้องปฏิเสธโดยไม่เขียนทับไฟล์เดิม

กรณีที่ควรส่งผลกลับ: เวอร์ชัน Windows และ DPI scale, output ของ `craft check`, ผลการคลิก/กดคีย์/ปรับขนาด/ปิด, พฤติกรรมเมื่อกด Ctrl+C ระหว่างรอ event และข้อความ error หากขั้นใดไม่ตรงคาด อย่าถือว่า build สำเร็จเป็นผลรับรองการแสดงหน้าต่าง
