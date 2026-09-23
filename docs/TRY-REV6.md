# ทดลอง Craft 0.1.6 / Phase 1 Rev.6

รุ่นนี้เน้น HTTP + Database ร่วมกัน ใช้ `std.http` และ `std.db` โดยตรง ไม่ต้อง import `api-framework` ผลตรวจรับยังรอผู้ใช้รัน ห้ามตีความว่าการ build ผ่านเท่ากับผ่าน integration tests

## รันชุดตรวจรับบน repository

เปิด PowerShell ที่ `D:\02-Repository\craft` ใช้ Windows account เดิมที่สร้าง encrypted credentials:

```powershell
pwsh -File .\scripts\test-rev6.ps1
```

สคริปต์โหลด `.local/secrets/rev5-databases.clixml` ชั่วคราว รัน Go tests แล้วคืน environment เดิม ไม่แสดง DSN และไม่บันทึก secrets ลง source ต้องมี Go ตาม go.mod และ DB ทั้งสามเปิดอยู่ สามารถส่ง `-SecretsPath` หากเก็บไฟล์ไว้อีกตำแหน่ง

ผลที่คาดหวัง: `TestRev6ConnectionDiagnostics`, `TestRev6HTTPPoolCancellation` และ `TestRev6HTTPDatabase` ผ่านทุก subtest ของ MySQL/PostgreSQL/SQL Server ไม่มี SKIP เพราะ credentials ขาด และ `$LASTEXITCODE` เป็น 0 ใช้ `craft_test` เท่านั้น สร้าง table `craft_rev6_<unique id>` ต่อ run แล้วลบเมื่อ server owner จบ

จากนั้นตรวจ regression:

```powershell
pwsh -File .\scripts\test-rev6.ps1 -Regression
```

`-Race` เป็นตัวเลือกสำหรับเครื่องที่เตรียม Go race toolchain/CGO C compiler แล้ว ไม่เปลี่ยนเครื่องหรือติดตั้ง compiler อัตโนมัติ เก็บผล exit code, versions และ sanitized output เพื่ออัปเดต status ห้ามส่ง DSN หรือไฟล์ secrets มาเป็น test log

## ทดลอง HTTP ด้วยตัวเอง

Terminal แรก:

```powershell
pwsh -File .\scripts\run-rev6.ps1 -Driver mysql
```

เปลี่ยนเป็น `postgres` หรือ `sqlserver` เพื่อทดลองอีกสองตัว สคริปต์ใช้ `dist/craft.exe` ของ repository โดยตรง ไม่พึ่ง PATH; ตั้ง driver, DSN, unique run ID และ loopback address ก่อนเข้า example แล้วเรียก `craft run` ไม่มี dotenv loader เพิ่มใน CLI

Terminal อีกหน้าต่าง:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/health
Invoke-RestMethod http://127.0.0.1:8080/ready
Invoke-RestMethod -Method Post 'http://127.0.0.1:8080/items?id=demo' -Body 'hello' -ContentType 'text/plain; charset=utf-8'
Invoke-RestMethod http://127.0.0.1:8080/items/demo
Invoke-RestMethod -Method Put http://127.0.0.1:8080/items/demo -Body 'updated' -ContentType 'text/plain; charset=utf-8'
Invoke-RestMethod http://127.0.0.1:8080/items/demo
Invoke-RestMethod -Method Delete http://127.0.0.1:8080/items/demo
```

คาดหวัง health=`ok`, ready=`ready`, POST=201, GET=`hello`, PUT=200, GET=`updated`, DELETE=204 และ GET หลังลบ=404; POST id ซ้ำ=409 ตัวอย่างรับ body เป็น UTF-8 text ไม่ใช่ JSON และ id เป็น lowercase/digit/underscore 1..60 ตัวอักษร

กด Ctrl+C ใน Terminal server เพื่อหยุดและคืน resources ขณะ process ถูก cancel CLI อาจรายงาน `context canceled` ตาม convention เดิม ถ้า drain เกินเวลาเพิ่มข้อความ `HTTP shutdown drain deadline exceeded` และรอ cleanup ก่อนคืน control การปิดหน้าต่างหรือ kill process อาจทำให้ table เหลือ ใช้ชื่อ fixture ที่พิมพ์ตอนเริ่มเพื่อตรวจ/ลบเฉพาะ table ของ run นั้น ห้ามลบด้วย wildcard

## Ownership และ timeout

- `main` เปิด pool และยังไม่ return ระหว่าง serve; request ยืม pool ผ่าน immutable state และเป็นเจ้าของ transaction ของตนเอง
- helper ที่สร้างแล้ว return `DbConnection` จะคืน handle ที่ปิดแล้วเมื่อ helper จบ กติกา Rev.5 ไม่เปลี่ยน: เปิดใน owner แล้วส่งเป็น argument
- Request exception/cancellation ก่อน commit จะ rollback เมื่อ scope จบ; หลัง commit สำเร็จ disconnect ไม่ย้อนข้อมูล
- Failed commit บางกรณีอาจไม่ทราบผลบน server ไม่ auto-retry writes ให้ตรวจผลด้วย application identifier
- HTTP deadline ของตัวอย่าง 3s, DB operation/transaction lifetime 5s, drain 2s, HTTP concurrency 8, DB maxOpen 2; transaction timeout ครอบคลุมทั้งอายุ ไม่ต่อเวลาใหม่ทุก statement
- Shutdown หยุดรับงาน, drain, cancel งานที่เกิน deadline, รอ child scopes/transactions จบ, จากนั้น owner ลบ fixture และปิด pool Cooperative native calls อาจคืนช้ากว่า deadline จึงไม่รับรอง hard wall-clock shutdown bound
- `.env` ไม่ถูกโหลดอัตโนมัติ; `std.env` อ่าน process environment และ `craft.toml` ไม่ใช่ที่เก็บ password

## สิ่งที่ชุดตรวจรับครอบคลุม

Craft example ถูก compile และเปิด HTTP จริงทุก driver; CRUD ใช้ parameters และมี independent DB read-back, UTF-8/quote values, concurrent IDs, pool stats, exception rollback, disconnect ก่อน/หลัง commit, DB/HTTP deadlines, request-owned pool, graceful stop และ restart/forced shutdown ที่มี transaction

HTTP transport integration ใช้ connection ที่ยึดไว้เป็น barrier และรอ `WaitCount` ก่อน cancel จึงไม่อาศัย sleep เดาสถานะ ทดสอบ overload ไม่แตะ DB, pool wait deadline สองลำดับ, client disconnect, forced shutdown และ subsequent query แยกกับแอป Craft

In-flight disconnect รอให้ activity view ของ DB แสดง SQL ที่มี unique fixture marker ก่อนตัด client connection แล้วรอให้ SQL หายไป ใช้ MySQL PROCESSLIST, PostgreSQL pg_stat_activity และ SQL Server sys.dm_exec_requests/sys.dm_exec_sql_text ผู้ใช้ DB ต้องอ่านกิจกรรมของ connections ที่ใช้ login เดียวกันได้ หากสิทธิ์ไม่พอ test จะ FAIL พร้อมแจ้งขั้นตอนที่ค้าง ไม่ grant permissions อัตโนมัติ

ข้อจำกัดของหลักฐาน: ambiguous commit fault injection, OS handle measurements และ bundle/installer execution ยังต้องตรวจเพิ่มเติมตาม status; ไม่มี automatic retry writes ผล commit ที่ไม่แน่นอนต้องตรวจจากแอป SQL Server Shared Memory ไม่แทน TCP/TLS ไม่ปรับ service configuration ในชุดนี้ Child-task DB work ระหว่าง forced shutdown มี fixture แล้วแต่ยังรอผลรันจริง

ค่าที่วัด heap/goroutines จาก warmup + 3 batches เป็น observations ไม่ใช่ production guarantee; idle connections ไม่ใช่ leak ต้องดู `inUse=0`, pool limits และ owner close แยกกัน
