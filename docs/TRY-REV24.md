# ให้ผู้ใช้ทดลอง UI Paint Lab 0.1.24

เปิด PowerShell ใน repository และ build Craft ล่าสุด:

```powershell
cd D:\02-Repository\craft
go build -o dist\craft.exe .\cmd\craft
cd examples\gui-paint-lab
..\..\dist\craft.exe check
..\..\dist\craft.exe run
```

ตรวจรูปทรงมุมมน, เส้น, ขอบ, ข้อความไทย/อังกฤษ, ภาพ PNG และความโปร่งใส. กด **Change color** แล้วดูว่า Canvas แรกเปลี่ยนสีโดย Canvas ที่สองยังอยู่; คลิก Canvas เพื่อดู local coordinates ใน terminal; scroll แล้วคลิกอีกครั้ง. Resize ให้แคบและกว้างเพื่อดู clipping/scroll.

เปรียบเทียบ GDI fallback ใน PowerShell ใหม่:

```powershell
$env:CRAFT_UI_RENDERER = 'gdi'
..\..\dist\craft.exe run
Remove-Item Env:\CRAFT_UI_RENDERER
```

ควรเห็น `Renderer: gdi` ใน terminal. บันทึกภาพหรือข้อความผิดพลาดพร้อม renderer, DPI และขั้นตอนทำซ้ำ. สำหรับ trace ตั้ง `$env:CRAFT_UI_TRACE = '1'` ก่อนรัน; ค่าที่รายงานเป็นเวลา **จน draw complete** ไม่ใช่เวลาภาพแสดงจริง. ยังไม่ต้องทดสอบ Craft Office หรือ HTTP ในรอบนี้.
