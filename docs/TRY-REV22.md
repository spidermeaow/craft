# Craft 0.1.22 — คู่มือทดสอบ UI Input

ใช้ Windows และ Craft binary ที่รายงาน `0.1.22`. สร้าง binary จาก source หาก `dist/craft.exe` ยังเป็นรุ่นก่อน:

```powershell
cd D:\02-Repository\craft
go build -o dist/craft.exe ./cmd/craft
dist/craft.exe version
```

เปิด UI Input Lab จาก repository (หรือตัวอย่าง `examples/gui-input` ที่มากับตัวติดตั้ง):

```powershell
cd D:\02-Repository\craft\examples\gui-input
D:\02-Repository\craft\dist\craft.exe check
D:\02-Repository\craft\dist\craft.exe run
```

ใช้ `D:\02-Repository\craft-test\runtime\11-ui-input` แทนได้เมื่อทดสอบใน `craft-test`. ตรวจภาษาอังกฤษ, ภาษาไทย, IME, emoji, Enter, Tab/Shift+Tab, selection, Ctrl+A/C/X/V, Backspace/Delete, ปุ่ม Clear และ resize ตามคู่มือ `D:\02-Repository\craft-test\guides\12-ui-input.md`. เมื่อ status เปลี่ยนหลัง `TextChanged`, ข้อความที่พิมพ์ในช่องต้องไม่หาย. ตรวจ `D:\02-Repository\craft-test\runtime\08-ui` เพิ่มเพื่อดู regression ของ Button, ScrollView และ Virtual List.

บันทึก renderer, Windows/IME, DPI, ขั้นตอนทำซ้ำ, ผลจริง และภาพหน้าจอในรายงานทดสอบ. ยังไม่ต้องทดสอบ HTTP ร่วมกับ UI ใน revision นี้.
