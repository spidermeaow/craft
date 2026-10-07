# ลอง Craft UI Engine 0.1.18 ด้วยตนเอง

ใช้ binary ที่ build ไว้จากโฟลเดอร์ `D:\02-Repository\test-gui`:

```powershell
cd D:\02-Repository\test-gui
D:\02-Repository\craft\dist\craft.exe run
```

1. เริ่มที่ 8×8 แล้วเลื่อน wheel ในพื้นที่ grid. สลับ 16×16 และ 32×32; ต้องเลื่อนไปถึงแถว/คอลัมน์ท้าย ๆ ได้โดยไม่มีเซลล์วาดทับ control panel.
2. ลอง Shift+wheel และ horizontal wheel เพื่อเลื่อนแนวนอน. ลองลูกศร, PageUp/PageDown, Home/End หลังชี้ grid. เลื่อนถึงขอบแล้วตำแหน่งต้องไม่เกินเนื้อหา.
3. คลิกเซลล์หลังเลื่อน; label และ revision ของเซลล์นั้นต้องเปลี่ยนตรงกับตำแหน่งที่คลิก. กด Update selected/Update all แล้วตำแหน่ง scroll ไม่ควรกระโดดโดยไม่จำเป็น.
4. ลอง resize และเปลี่ยน DPI เมื่อมีอุปกรณ์; ตรวจภาพ, clipping และตำแหน่งคลิก. ลองทั้ง Direct2D และ GDI fallback เมื่อมีเครื่องที่เข้าเงื่อนไข.

Trace แบบ opt-in:

```powershell
$env:CRAFT_UI_TRACE = '1'
D:\02-Repository\craft\dist\craft.exe run
Remove-Item Env:CRAFT_UI_TRACE
```

ปิดหน้าต่างแล้วดู JSON ที่ stderr. `scroll_to_draw_complete` คือเวลาถึงจบคำสั่งวาด ไม่ใช่เวลาที่พิกเซลปรากฏจริง. เปรียบเทียบกับ 0.1.17 บนหน้าต่างและเครื่องเดียวกันก่อนรับรอง responsiveness.
