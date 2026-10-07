# ทดลอง Craft 0.1.16 UI Engine บน Windows

ใช้ `dist/craft.exe` ของรุ่นนี้ เปิด PowerShell ใหม่ แล้วรัน:

```powershell
cd D:\02-Repository\craft\examples\gui-layout
& D:\02-Repository\craft\dist\craft.exe check
$env:CRAFT_UI_TRACE = '1'
& D:\02-Repository\craft\dist\craft.exe run
```

ตรวจว่า `Renderer:` แสดง `direct2d-hardware` หรือ `gdi`, หน้าต่างไม่ค้าง, คลิกแล้วข้อความเปลี่ยน, กดค้างแล้วลากออกก่อนปล่อยไม่เปลี่ยน, Tab/Shift+Tab/Enter/Space ยังทำงาน ลอง resize ต่อเนื่องและดูว่าภาพไม่กระพริบหรือค้างในพื้นที่เก่า ปิดหน้าต่างเพื่อดู JSON diagnostics บน stderr

ตรวจ bundle:

```powershell
& D:\02-Repository\craft\dist\craft.exe build
& D:\02-Repository\craft\dist\craft.exe run dist\gui-layout.craftbundle
```

เปรียบเทียบกับ 0.1.15 บนเครื่องเดียวกันโดยใช้ workload และจำนวนรอบเท่ากัน บันทึก p50/p95/p99, DPI, refresh rate, renderer, CPU/RAM และความผิดปกติของภาพ แยก `input_to_draw_complete` ออกจากการวัด input ถึงภาพปรากฏจริง ซึ่งต้องใช้อุปกรณ์หรือวิธีวัดภายนอก
