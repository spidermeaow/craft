# ทดลอง Craft 0.1.5 / Phase 1 Rev.5

Rev.5 เพิ่ม `std.db` สำหรับ MySQL, PostgreSQL และ SQL Server ไม่ต้องแก้ `craft.toml` หรือ import package เพื่อใช้ DB primitives ไม่รวม database server ในตัวติดตั้ง

## ตัวอย่างที่ไม่แก้ข้อมูล

ติดตั้ง CLI รุ่นใหม่แล้วเปิด PowerShell ใหม่ หรือเรียก `dist/craft.exe` ด้วย absolute path ตรวจ `craft version` ให้แสดง `0.1.5` และ `Phase 1 Rev.5`

ตั้ง environment ด้วยข้อมูลของคุณ (ตัวอย่างเป็น placeholders):

```powershell
$env:CRAFT_DB_DRIVER = 'postgres'
$env:CRAFT_DB_DSN = 'postgres://USER:PASSWORD@127.0.0.1:5432/DATABASE?sslmode=disable&connect_timeout=5'
Set-Location D:\02-Repository\craft\examples\rev5-database
craft check
craft test
craft run
```

คาดหวัง check ผ่าน, tests แบบไม่เชื่อมต่อ DB ผ่าน 3 รายการ และ run พิมพ์ชื่อ driver/database, `Hello from Craft Rev.5`, `Connections in use: 0` ตัวอย่างใช้ SELECT อย่างเดียว

เปลี่ยนเป็น `mysql` พร้อม DSN `USER:PASSWORD@tcp(127.0.0.1:3306)/DATABASE?charset=utf8mb4&parseTime=true&loc=UTC&timeout=5s` หรือ `sqlserver` พร้อม DSN ของคุณ สำหรับ SQL Express ในเครื่องนี้ใช้ host/path `localhost/SQLEXPRESS` และ `protocol=lpc`; อ่าน [contract](REV5-DATABASE.md) สำหรับข้อจำกัด Shared Memory และรูปแบบ URL

## ใช้ credentials ที่เข้ารหัสไว้ในเครื่องพัฒนา

รันจาก root repository ด้วยบัญชี Windows ที่บันทึกไฟล์ไว้:

```powershell
$rev5Local = Import-Clixml -LiteralPath .\.local\secrets\rev5-databases.clixml
try {
    $env:CRAFT_DB_DRIVER = 'postgres'
    $env:CRAFT_DB_DSN = [System.Net.NetworkCredential]::new('', $rev5Local.CRAFT_POSTGRES_DSN).Password
    Push-Location .\examples\rev5-database
    try { & ..\..\dist\craft.exe run } finally { Pop-Location }
} finally {
    Remove-Item Env:CRAFT_DB_DSN -ErrorAction SilentlyContinue
    Remove-Item Env:CRAFT_DB_DRIVER -ErrorAction SilentlyContinue
    Remove-Variable rev5Local -ErrorAction SilentlyContinue
}
```

เปลี่ยน key เป็น `CRAFT_MYSQL_DSN`/`CRAFT_SQLSERVER_DSN` พร้อม driver ที่ตรงกัน ห้ามพิมพ์ credentials ลง terminal หรือใส่ source; `.local` ไม่รวมใน installer/bundle

## Database acceptance suite

ใช้คำสั่งนี้จาก repository root:

```powershell
pwsh -File scripts/test-rev5.ps1
```

สคริปต์โหลด secrets แบบ DPAPI เฉพาะ process แล้วรัน Go tests ชื่อ `TestRev5` เท่านั้น ชุด integration ตรวจ current database ต้องเป็น `craft_test` ก่อนเขียน สร้างและลบเฉพาะ fixture tables ชื่อ `craft_rev5_...` ที่สร้างเอง หากไม่มี Go ใช้ตัวอย่าง Craft ด้านบนสำหรับ smoke test และให้ผู้พัฒนารัน integration suite

ทดสอบ CRUD/parameters, Unicode, decimal, binary, NULL, transactions, cleanup, pool stats, bounds, timeout, typed errors และ portable bundle ผลที่ agent รันจริงอยู่ใน [status](../roadmap/phase1/output/phase1-rev5-status.md)

สำหรับ regression ทั้งโปรเจกต์ ผู้ใช้รันเองตาม working preferences:

```powershell
go test ./... -timeout 180s
```

ตรวจ installer upgrade และ VS Code highlighting ของ `DbConnection`, `DbValue`, `DbRows` ด้วยตนเอง ชุด database ไม่ใช่การยืนยันว่า UI/installer หรือ regression ทั้งหมดผ่าน
