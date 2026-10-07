# ทดลอง Craft GUI package 0.2.2

ใช้ Craft Runtime 0.1.25 ใน `D:\02-Repository\craft\dist\craft.exe` บน Windows. ผลต่อไปนี้เป็นขั้นตอนให้ผู้ใช้ตรวจ ไม่ใช่ผลที่ผู้ช่วยรันแล้ว.

## 1. Package lab

```powershell
Set-Location 'D:\02-Repository\craft\examples\gui-package-lab'
& 'D:\02-Repository\craft\dist\craft.exe' package check
& 'D:\02-Repository\craft\dist\craft.exe' check
& 'D:\02-Repository\craft\dist\craft.exe' run
```

คาดว่า app เปิดได้, list เลื่อน, ปุ่มเปลี่ยน status, TextInput รับภาษาไทย, resize หน้าต่างให้แคบกว่า 760 logical pixels แล้ว navigation ย้ายขึ้นด้านบน. ลอง DPI 100% และ 125% แล้วตรวจ content ที่เลื่อนไปถึงได้.

## 2. Dashboard ใน craft-test

```powershell
Set-Location 'D:\02-Repository\craft-test\gui-package-dashboard'
& 'D:\02-Repository\craft\dist\craft.exe' package check
& 'D:\02-Repository\craft\dist\craft.exe' check
& 'D:\02-Repository\craft\dist\craft.exe' run
```

ลองหน้าต่างกว้างและแคบกว่า 1500 logical pixels. ในหน้า compact คลิก target เอกสาร, ข้อมูล และ Refresh; ใช้ Tab + Enter/Space และดู ID ที่ status. ตรวจ styled native Button, TextInput/IME, การเลื่อนทั้งสองแนว และ resize. ตาราง, chart, tree และ control หลายตัวใน Canvas เป็นภาพจำลองตาม README ของ test app.

## 3. Renderer

ใน terminal ใหม่ที่โฟลเดอร์ dashboard:

```powershell
$env:CRAFT_UI_RENDERER = 'gdi'
& 'D:\02-Repository\craft\dist\craft.exe' run
Remove-Item Env:CRAFT_UI_RENDERER
```

เปรียบเทียบข้อความ, สี, clipping และ hit area กับ Direct2D. หากพบปัญหา ส่ง error output, ภาพหน้าจอ, ขนาดหน้าต่าง, DPI, renderer, การกดที่ทำให้เกิดปัญหา และบอกว่าเกิดกับ package lab หรือ dashboard. ยังไม่มีตัวเลข performance baseline ของรุ่นนี้.
