# ตรวจ Craft UI 0.1.26 ด้วยตนเอง

ใช้ PowerShell จาก repository นี้. รอบ implementation นี้ไม่ได้รัน automated tests หรือเปิด GUI เพื่อทดสอบ.

```powershell
go build -o craft.exe ./cmd/craft
.\craft.exe version
Set-Location examples/gui-package-lab
..\..\craft.exe package check
..\..\craft.exe check
..\..\craft.exe run
```

คาดว่า version เป็น `0.1.26`, package check และ source check สำเร็จ. ในหน้าต่าง GUI ให้ปรับจากกว้างประมาณ 1000 px ไป 700 px และ 500 px: การ์ดสถิติเปลี่ยนจากหลายคอลัมน์เป็นคอลัมน์เดียวตามพื้นที่; sidebar ย้ายลงเมื่อแคบกว่า 760 px. กดปุ่มและพิมพ์ไทยในช่อง Search เพื่อตรวจว่า input/event เดิมยังทำงาน. ย้ายหน้าต่างระหว่างจอ DPI 100%/125% ถ้ามี และตรวจข้อความไม่ถูกตัดผิดปกติ.

สำหรับ dashboard ใน `craft-test` ที่ใช้ Canvas พิกัดตายตัว ให้แยกผลจากตัวอย่าง grid นี้: Canvas เดิมอาจยังเลื่อนแนวนอนหรือมีส่วนถูก clip เมื่อหน้าต่างแคบ. บันทึกขนาดหน้าต่าง, DPI, renderer และภาพเมื่อพบปัญหา.
