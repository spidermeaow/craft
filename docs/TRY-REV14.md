# ทดลอง Craft 0.1.14 UI tree บน Windows

ผู้ใช้เป็นผู้รันการทดสอบจริงตามความต้องการของโปรเจกต์ เริ่มจาก CLI 0.1.14 ใน `dist/craft.exe` หรือจาก Setup แล้วเปิด terminal ใหม่

```powershell
craft version
cd D:\02-Repository\craft\examples\gui-layout
craft check
craft run
```

คาดหวังหน้าต่าง `Craft UI layout` มีข้อความ `Hello from Craft` และปุ่ม `Click me` คลิกปุ่มแล้วข้อความเปลี่ยนเป็น `Button pressed in Craft` ปรับขนาดหน้าต่างแล้วทั้งสองส่วนจัดวางใหม่ ปิดหน้าต่างแล้ว process จบ

ทดลอง bundle:

```powershell
craft build
craft run dist\gui-layout.craftbundle
```

คาดหวังผลเหมือน source การทดสอบเพิ่มเติมที่ควรส่งผลกลับ: Windows version, DPI 100%/150%/200%, ข้อความไทย/Unicode, click นอกปุ่ม, Enter/Space, resize, close/Ctrl+C, ความคมชัดของข้อความ และ error ที่พบ
