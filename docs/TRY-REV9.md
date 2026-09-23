# ทดลอง Rev.9 — filesystem/path tooling primitives

Rev.9 เพิ่ม `std.path.clean`, `absolute`, `relative`, `std.fs.readBytes`, `writeTextAtomic` และ `writeBytesAtomic` โดยยังคงให้ paths ที่เป็น relative อ้างอิง process working directory

## ทดลองตัวอย่าง

สร้าง input JSON ใน directory ของตัวอย่าง แล้วสั่งให้ Craft validate และเขียน output:

```powershell
cd examples/rev9-tools
Set-Content -NoNewline input.json '{"name":"Craft","enabled":true}'
craft run -- input.json output.json
Get-Content output.json
```

คาดหวังข้อความ `Wrote <absolute-path> with 2 keys` และ `output.json` เป็น JSON object ที่ valid

ให้ลองส่ง output เป็น directory เพื่อยืนยันว่า destination เดิมไม่ถูกแทนที่:

```powershell
New-Item -ItemType Directory -Force blocked-output | Out-Null
craft run -- input.json blocked-output
Test-Path blocked-output
```

คำสั่ง `craft run` ต้อง fail พร้อม filesystem operation/path diagnostic และ `Test-Path` ต้องคืน `True` เพราะ directory เดิมยังอยู่

## รัน acceptance

จาก repository root ให้รัน:

```powershell
pwsh -NoProfile -File .\scripts\test-rev9.ps1
```

สคริปต์เรียก full Rev.8 regression ก่อน แล้วรัน acceptance ของ atomic write/bytes/path ใน `internal/stdlib` หากต้องรัน race หรือ stress ที่ Rev.8 รองรับ ให้เพิ่ม `-Race` หรือ `-Stress` ตาม environment ที่เตรียมไว้

ก่อนประกาศ release ให้ทดสอบ Windows cases ตาม [แผน Rev.9](../roadmap/phase1/craft-phase1-rev9.md): drive-rooted, drive-relative, UNC, mixed separator, locked destination, permission failure และ target ที่อยู่คนละ volume สำหรับ `std.path.relative` ผลที่ไม่ได้รันต้องบันทึกไว้ใน `roadmap/phase1/output/phase1-rev9-status.md` ไม่ใช่ถือว่าผ่านโดยอัตโนมัติ
