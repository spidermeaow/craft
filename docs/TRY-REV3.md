# ทดลอง Craft 0.1.3 / Rev.3

Agent ไม่รัน automated tests หรือเปิด CLI/VS Code/installer เพื่อทดสอบตามคำขอของผู้ใช้
การ build เป็นคนละขั้นกับการทดสอบ ด้านล่างเป็นคำสั่งให้ผู้ใช้รันเอง

## Runtime

```powershell
cd D:\02-Repository\craft
go test ./...
$craftRev3 = (Resolve-Path .\dist\craft.exe).Path
& $craftRev3 version
Push-Location .\examples\rev3-tasks
try {
    & $craftRev3 check
    & $craftRev3 test
    & $craftRev3 run
    & $craftRev3 fmt
    & $craftRev3 fmt --check
    & $craftRev3 test
    & $craftRev3 build
    & $craftRev3 run .\dist\rev3-tasks.craftbundle
} finally { Pop-Location }
```

คาดหวัง version 0.1.3 / Rev.3, Craft tests 4 รายการผ่าน, demo มี finished/cleanup
ของทั้งสองงาน, once หนึ่งครั้ง, tick ตามเวลาที่ scheduler ทำได้ และจบด้วย
`done completed completed completed cancelled` โดยไม่กำหนดจำนวน tick หรือลำดับ
ระหว่าง task เป็นค่าตายตัว การรัน bundle ควรมีพฤติกรรมเดียวกัน
ตรวจ `$LASTEXITCODE` หลังคำสั่งแต่ละตัว: 0 เมื่อสำเร็จ

ในโปรเจกต์ทดลองใหม่ ตั้ง repeating timer แล้ว join โดยไม่ cancel จากนั้นกด Ctrl+C:
โปรแกรมต้องออกตาม cancellation convention เดิม (runtime exit 1), งานลูกต้องหยุด
และ defer ที่ลงทะเบียนแล้วต้องได้โอกาส cleanup ดูข้อจำกัด blocking I/O ใน contracts

ทดสอบ `craft fmt check` ใน directory ที่ไม่มี path ชื่อ check:
ต้องแจ้ง error พร้อม hint `craft fmt --check`, exit 4 และไม่แก้ไขไฟล์
`craft fmt path/to/file.craft` และ `craft fmt --check` ยังคงทำงานตามปกติ

## Regression และชุดผู้ใช้เดิม

```powershell
cd D:\02-Repository\craft
$craftRev3 = (Resolve-Path .\dist\craft.exe).Path
$previousRev2Value = $env:CRAFT_REV2_VALUE
Push-Location .\examples\rev2-user-acceptance
try {
    $env:CRAFT_REV2_VALUE = 'ready'
    & $craftRev3 check
    & $craftRev3 test
} finally {
    $env:CRAFT_REV2_VALUE = $previousRev2Value
    Pop-Location
}
```

คาดหวัง 16 passed (15 ที่กู้จาก full.craft + starter 1) ตัวช่วยถูกสร้างกลับตาม
assertions เพราะไม่พบ helper เดิมใน source ปัจจุบัน จึงต้องตรวจรับชุดนี้ใหม่
filesystem test ใช้ `craft-rev2-test-data` ใต้ cwd และจะปฏิเสธถ้ามีไฟล์ชื่อที่ใช้แล้ว
defer ลบไฟล์ที่ทดสอบแต่คง directory ว่างไว้ ลบ directory ว่างนี้เองได้หลังตรวจผล
รัน `craft test` ใน examples/rev2-features, rev2-bank และ rev2-cli เพิ่มด้วย
ผล Rev.2 ที่เคยผ่านก่อน fmt ไม่ใช้แทนผล regression บน 0.1.3

## VS Code

ติดตั้ง `dist/craft-language-support-0.1.0.vsix` ผ่าน Extensions → Install from VSIX
ใน VS Code ไม่ใช้ Visual Studio Installer แล้วทำตาม
[คู่มือ Language Support](../craft-vscode/README.md)

- เปิด fixtures/highlighting.craft: String/comment ไม่ทำให้ keyword ภายในเปลี่ยนสี
  และ unfinished String ไม่ทำให้สีรั่วไปบรรทัดถัดไป
- ตรวจไทย/Unicode, 1..3, exponent, Map ซ้อน, named arguments และ chained calls
- เปรียบเทียบ Craft Dark, Craft Light, theme เดิม และ High Contrast ที่ zoom 100/125/150%
- ทดลอง selection, comment toggle, bracket/quote pairing, indentation/folding และ snippets
- ติดตั้งร่วมกับ icon VSIX แล้วถอนทีละ extension เพื่อตรวจความเป็นอิสระ
- จาก craft-vscode รัน `npm install --ignore-scripts` แล้ว `npm test`

บันทึก VS Code version, theme, zoom, คำสั่งและผลจริงเมื่อส่งผลกลับ
Format Document, Problems และ semantic services เป็นงานระดับ B ที่ยังไม่เปิดใช้

## Build เอง

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 -Installer -Icons -Language
```

คำสั่งนี้สร้าง CLI/VSIX/installer/checksums โดยไม่รัน tests ไม่ติดตั้งโปรแกรม
และไม่เปิด VS Code ผู้ใช้ทดสอบ install/upgrade/PATH/uninstall แยกต่างหาก
