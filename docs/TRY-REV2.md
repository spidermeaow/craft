# ทดลอง Craft 0.1.2 — Phase 1 Rev.2

สถานะ: Agent build artifacts แล้ว แต่ **ไม่ได้รัน tests หรือทดลอง CLI/installer/VS Code** ตาม preference ของผู้ใช้
คำสั่งด้านล่างเป็นขั้นตอนให้ผู้ใช้ตรวจรับ ผลลัพธ์ที่ระบุเป็นสิ่งที่คาดหวัง ไม่ใช่ผลทดสอบที่เกิดขึ้นแล้ว

## 1. เลือก executable รุ่นใหม่

ติดตั้ง `dist/Craft-setup.exe` แล้วเปิด PowerShell ใหม่ หรือใช้ portable path โดยตรง:

```powershell
$craftExe = 'D:\02-Repository\craft\dist\craft.exe'
& $craftExe version
Get-Command craft -All
```

คาดหวัง Craft 0.1.2 / Phase 1 Rev.2; เครื่องผู้ใช้ไม่ต้องมี Go
Installer ไม่ติดตั้ง VSIX ให้อัตโนมัติ ไฟล์จะอยู่ใน `editor/` ของตำแหน่งติดตั้งเมื่อ build มาพร้อม icon artifact

## 2. Feature demo และ tests

```powershell
cd D:\02-Repository\craft\examples\rev2-features
& $craftExe check
& $craftExe run -- hello --flag
& $craftExe test
& $craftExe fmt
& $craftExe fmt --check
& $craftExe test
& $craftExe build
& $craftExe run dist/rev2-features.craftbundle -- hello --flag
$LASTEXITCODE
```

คาดหวัง feature tests 12 รายการผ่าน, arguments `[hello, --flag]`, Original math 90 / Copy 95 และ Exact Int `9007199254740993`
formatter ไม่เปลี่ยนพฤติกรรม, fmt --check หลัง fmt และคำสั่งสำเร็จทั้งหมด exit 0
ผลที่อ่านจาก Random และ DateTime ไม่ควรอาศัยเวลาหรือ seed ของเครื่อง

## 3. Interactive Bank

แนะนำคัดลอกโฟลเดอร์ `examples/rev2-bank` ไปพื้นที่ทดลองที่เขียนไฟล์ได้ ก่อนรัน

```powershell
cd D:\02-Repository\craft\examples\rev2-bank
& $craftExe check
& $craftExe test
& $craftExe run -- bank-trial.json
```

1. เลือก 2 ฝาก `250000` minor units แล้วเลือก 3 ถอน `120000` → ยอด `130000`
2. เลือก 4 → ประวัติฝาก/ถอน; เลือก 5 → บันทึก `bank-trial.json`; เลือก 0 ออก
3. เปิดรอบใหม่ด้วยชื่อไฟล์เดิม แล้วเลือก 1 → ยอด `130000` และประวัติเดิม
4. ลองถอนเกินยอด, ใส่ `abc`, `0`, ค่าติดลบ, เลขเกิน Int64 → ไม่เปลี่ยนยอด
5. Ctrl+Z แล้ว Enter ขณะรอ input → EOF จบโปรแกรม; Ctrl+C → ยกเลิก ไม่ค้างรอ input
6. ทดลองกับไฟล์ JSON ที่ผิดรูปแบบโดยตั้งชื่อไฟล์ทดลองแยก → โปรแกรมแจ้ง error และออกก่อนเขียนทับ

จำนวนเงินใช้ Int minor units: `100 = 1 บาท` เพื่อไม่เก็บยอดใน Float
เมนู 5 เขียนทับไฟล์ที่เลือกอย่างชัดเจน; ออก/EOF ไม่ autosave
writeText ไม่ atomic: ตัวอย่างนี้สำหรับทดลองภาษา ไม่ใช่ระบบบัญชี production

ทดสอบ piped input (ใช้ path ใหม่เพื่อให้เริ่มจากศูนย์):

```powershell
@('2', '250000', '3', '120000', '1', '0') | & $craftExe run -- bank-pipe-trial.json
```

## 4. CLI utilities

```powershell
cd D:\02-Repository\craft\examples\rev2-cli
& $craftExe test
& $craftExe run -- time '2024-02-29T00:00:00Z'
& $craftExe run -- words 'D:\path\to\sample.txt'
& $craftExe run -- config 'D:\path\to\config.json'
```

คาดหวังวันเวลา UTC/UTC+07:00, จำนวนคำที่คั่นด้วย space/tab/newline และ config keys/values ตามลำดับชื่อ

## 5. Regression และ Go tests

```powershell
cd D:\02-Repository\craft
go test ./...
cd examples\rev1
& $craftExe check
& $craftExe test
```

Go tests ใช้ temporary directories สำหรับ filesystem/CLI; ครอบคลุม invalid types, immutable fields, input/EOF/cancellation,
project/bundle arguments, formatter round-trip, JSON duplicate/limits/precision, filesystem overwrite protection และ date/conversion boundaries
ชุด 39 tests ในโปรเจกต์ส่วนตัวที่เคยผ่าน Rev.1 ควรรันซ้ำด้วย 0.1.2 ด้วย

## 6. Craft Forge File Icons

ติดตั้ง `dist/craft-forge-file-icons-0.1.0.vsix` ผ่าน VS Code → Extensions → Install from VSIX
จากนั้นเลือก Preferences: File Icon Theme → Craft Forge File Icons
ดู [คู่มือและตาราง fixtures](../craft-file-icon-theme/README.md) เพื่อเช็ก mappings, fallback, Light/Dark/High Contrast และการถอน extension
การติดตั้ง VSIX ไม่ต้องมี Craft CLI/Go; version icon theme 0.1.0 แยกจากภาษา 0.1.2

## 7. Installer lifecycle

บน Windows x64 ที่ไม่มี Go: ติดตั้ง, เปิด Terminal ใหม่, ตรวจ version/PATH, สร้างโปรเจกต์, run/test, upgrade จากรุ่นเดิม และ uninstall
ตรวจว่าไม่ลบ project ของผู้ใช้หรือ PATH entries ที่ไม่เกี่ยวข้อง
บันทึกผลจริงแยกจากผล build พร้อม Windows/VS Code version และ exit codes ใน [status](../roadmap/phase1/output/phase1-rev2-status.md)
