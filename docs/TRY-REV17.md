# ลอง Craft UI Engine 0.1.17 ด้วยตนเอง

จาก `D:\02-Repository\craft`:

```powershell
go build -o dist/craft.exe ./cmd/craft
cd D:\02-Repository\test-gui
D:\02-Repository\craft\dist\craft.exe run
```

เริ่มที่ 8×8: เซลล์ต้องเห็นข้อความและคลิกตรงตำแหน่ง. สลับ 16×16: เซลล์ต้องไม่หายเพราะความสูงศูนย์. สลับ 32×32: ควรเห็นส่วนที่อยู่ในหน้าต่าง พร้อมข้อความบอกว่าเซลล์นอกหน้าต่างถูก clip; ไม่มี scrollbar. กด `Update selected` และ `Update all` ตรวจว่าข้อความเปลี่ยนครบ ไม่เหลือ glyph เก่า. ลอง resize หน้าต่าง และ DPI 100%/150%/200% ถ้ามีอุปกรณ์.

Trace แบบ opt-in:

```powershell
$env:CRAFT_UI_TRACE = '1'
D:\02-Repository\craft\dist\craft.exe run
Remove-Item Env:CRAFT_UI_TRACE
```

ปิดหน้าต่างแล้วดู JSON ที่ stderr: `layout`, `paint_to_draw_complete`, `input_to_draw_complete`, `overflow_nodes`. เวลาจบคำสั่งวาดยังไม่ใช่เวลาที่ภาพปรากฏจริงบนจอ. ให้เทียบ workload เดียวกับ 0.1.16 ก่อนรับรอง performance.
