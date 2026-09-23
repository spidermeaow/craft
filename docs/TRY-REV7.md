# ทดลอง Craft 0.1.7 / Phase 1 Rev.7

Rev.7 แยก HTTP request timeout ออกจาก DB timeout และตรวจว่า read deadline ของ body ไม่กระทบ handler หลังอ่าน body เสร็จ

เปิด PowerShell ที่ `D:\02-Repository\craft` แล้วรัน:

```powershell
pwsh -File .\scripts\test-rev7.ps1
$LASTEXITCODE
```

สคริปต์โหลด DSN จาก `.local\secrets\rev5-databases.clixml` เข้า process environment ชั่วคราว แล้วคืนค่าเดิมหลังจบ ไม่แสดง DSN หรือรหัสผ่าน

คาดหวัง `TestRev7*` ผ่านครบ MySQL, PostgreSQL และ SQL Server และ exit code เป็น `0` โดยมีการตรวจดังนี้:

- request timeout ตอบ 504 โดย read timeout ยาวกว่า
- body timeout ตอบ 408 และไม่เรียก application handler
- body ที่อ่านเสร็จแล้วสามารถทำงานเกิน read timeout เดิมได้ตราบใดที่ยังไม่เกิน request timeout
- Craft application ทดสอบ DB timeout 500ms และ HTTP timeout 1.5s แยกกัน, รอ `inUse=0`, แล้ว ping ต่อได้

จากนั้นตรวจ regression:

```powershell
pwsh -File .\scripts\test-rev7.ps1 -Regression
$LASTEXITCODE
```

หากเกิด failure ให้ส่งเฉพาะ output ที่ sanitize แล้ว ไม่ส่ง DSN, password หรือไฟล์ secrets
