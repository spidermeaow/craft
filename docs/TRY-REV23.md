# ทดลอง Craft UI Layout Lab 0.1.23

เปิด PowerShell แล้ว build Craft จาก source ล่าสุดก่อน (คำสั่งนี้ให้ผู้ใช้เป็นผู้รัน):

```powershell
cd D:\02-Repository\craft
go build -o dist\craft.exe .\cmd\craft
cd D:\02-Repository\craft\examples\gui-layout-lab
..\..\dist\craft.exe check
..\..\dist\craft.exe run
```

ตัวอย่างนี้ใช้ `std.ui` โดยตรงและไม่มี Craft Office. การเรียก `dist\craft.exe` แบบระบุ path ช่วยไม่ให้หยิบตัวติดตั้งรุ่นเก่าจาก PATH.

ตรวจด้วยตาและแจ้งผลกลับ:

1. ขยาย/ย่อหน้าต่าง: top/status bar คงความสูง; sidebar กับพื้นที่งานและ list/detail เปลี่ยนสัดส่วนโดยไม่ทับกัน.
2. Scroll เหนือ list และ detail แล้วลองลาก scrollbar ด้านขวาของแต่ละ pane: แต่ละ pane เลื่อนแยกกัน; คลิก item หลัง scroll ต้องตรงตัวที่เห็น.
3. พิมพ์ไทย/อังกฤษใน Name และ Code แล้ว resize หรือกดปุ่ม: ข้อความและ focus ไม่ควรหาย.
4. ทดลอง DPI 100% และ 150% หากสะดวก; บันทึก renderer ที่พิมพ์ใน terminal และภาพที่ผิดปกติ.
5. ลองหน้าต่างแคบมาก: pane จะย่อจนต่ำกว่า minimum เพื่อให้ยังอยู่ใน viewport; รายงานหากข้อความทับหรือ click ผิดตำแหน่ง.

ปุ่ม Tab A/B เป็นเพียงตัวอย่างการจัดวาง **ยังไม่ใช่ Tabs component**. รายการเป็นปุ่มธรรมดา ไม่ใช่ Table/Grid. SplitView รุ่นนี้กำหนดสัดส่วนด้วย API; ยังลากเส้นแบ่งเพื่อปรับขนาดไม่ได้.
