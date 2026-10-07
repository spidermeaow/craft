# ตรวจรับ Craft UI Engine 0.1.19

จาก `D:\02-Repository\test-gui`:

```powershell
D:\02-Repository\craft\dist\craft.exe run
```

1. ใน 8×8 ต้องเห็นทุกคอลัมน์ในหน้าต่างขนาดอ้างอิง และปุ่ม controls ไม่กินพื้นที่ครึ่งหน้าต่าง.
2. สลับ 16×16 และ 32×32. ใช้ wheel, Shift+wheel/horizontal wheel และ PageDown เพื่อดูแถว/คอลัมน์ท้าย; คลิกเซลล์หลังเลื่อนแล้ว status/label ต้องตรง.
3. กด Update selected และ Update all หลังเลื่อน; scroll offset ควรคงอยู่เมื่อเนื้อหายังล้น. กด Reset แล้ว offset ต้องกลับเข้าเขตของเนื้อหา 8×8.
4. ลอง resize และเปลี่ยน DPI ถ้ามีอุปกรณ์; ตรวจทั้ง Direct2D และ GDI fallback เมื่อมีเครื่องที่เข้าเงื่อนไข.

Trace แบบ opt-in:

```powershell
$env:CRAFT_UI_TRACE = '1'
D:\02-Repository\craft\dist\craft.exe run
Remove-Item Env:CRAFT_UI_TRACE
```

ค่า `scroll_to_draw_complete` วัดถึงจบคำสั่งวาด ไม่ใช่เวลาที่ภาพปรากฏจริงบนจอ. การเปรียบเทียบ performance ต้องใช้ workload, เครื่อง และสภาพหน้าต่างเดียวกัน.
