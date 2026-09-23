# ทดลอง Craft 0.1.8 / Phase 1 Rev.8

เปิดตัวอย่าง local package:

```powershell
cd examples/rev8-packages
craft package check
craft package list
craft package lock
craft check
craft run
```

ผลที่คาดหวังจาก `craft run`:

```text
Hello Craft
```

หลังสร้าง lock แล้ว หากแก้ `packages/greeting/src/greeting.craft`, คำสั่ง `craft check` ต้องแจ้งว่า `craft.lock is out of date` ตรวจ source ที่เปลี่ยนแล้วรัน `craft package lock` เพื่อยอมรับ checksum ใหม่

ผู้พัฒนา repository สามารถรัน acceptance ทั้งหมดด้วย:

```powershell
pwsh -NoProfile -File .\scripts\test-rev8.ps1
```

สคริปต์ใช้ encrypted database credentials เดิมใน `.local/secrets/rev5-databases.clixml`, รัน full regression ของ Rev.1–Rev.8 และไม่พิมพ์ DSN ออกมา
