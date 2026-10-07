# ทดลอง Craft 0.1.15 GPU renderer บน Windows

ใช้ CLI 0.1.15 ใน `dist/craft.exe` หรือ Setup แล้วเปิด terminal ใหม่:

```powershell
craft version
cd D:\02-Repository\craft\examples\gui-layout
craft check
craft run
```

บันทึกบรรทัด `Renderer:` ที่แสดง `direct2d-hardware` หรือ `gdi` ตรวจหน้าต่างและข้อความไทย/Unicode บน DPI ที่มี คลิกปุ่มแล้วข้อความเปลี่ยน กดปุ่มค้างแล้วลากออกก่อนปล่อยต้องไม่เปลี่ยนข้อความ ลอง Tab/Shift+Tab, Enter/Space, resize และปิดหน้าต่าง

ตรวจ bundle:

```powershell
craft build
craft run dist\gui-layout.craftbundle
```

บันทึก Windows version, GPU, DPI, renderer ที่รายงาน, เวลาเปิด, RAM และ CPU/GPU ตอนหน้าต่างอยู่นิ่ง รวมถึง error หรือภาพผิดปกติ เมื่อมีผลแล้วอัปเดต `roadmap/phase1/output/phase1-rev15-status.md`
