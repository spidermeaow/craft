# Craft Phase 1 Rev.5 — Database foundations

วันที่: 19 กันยายน 2026; อัปเดต implementation 20 กันยายน 2026  
สถานะ: พัฒนา `std.db` ใน Craft 0.1.5 แล้ว และผ่าน database acceptance ทั้งสามตัว ดู [สถานะ Rev.5](output/phase1-rev5-status.md) และ [API contract](../../docs/REV5-DATABASE.md) สำหรับผลจริง ข้อจำกัดและรายการ pending  
ฐาน: Craft 0.1.4 / Phase 1 Rev.4

## 1. เป้าหมายที่ผู้ใช้กำหนด

พัฒนาเครื่องมือพื้นฐานของ Craft สำหรับเชื่อมต่อฐานข้อมูล เพื่อให้ผู้พัฒนานำไปสร้างแอปและเครื่องมือระดับสูงของตนเองได้ โดยควบคุมขนาดงาน Rev.5

- เน้นความสามารถพื้นฐานที่ใช้งานได้จาก Craft โดยไม่ผูกกับ API Framework
- รองรับ MySQL, PostgreSQL และ Microsoft SQL Server
- ยังไม่สร้างหรือขยาย framework ใดเป็นผลลัพธ์หลัก
- ตัวอย่าง package ระดับสูงทำได้เฉพาะเมื่อจำเป็นต่อการดำเนินงานหรือใช้ทดสอบความสามารถพื้นฐาน
- ผู้ใช้เปิดให้เพิ่มฐานข้อมูลอื่นได้ แต่แผนรอบนี้จำกัดไว้ที่สามตัวข้างต้นก่อน; ตัวอื่นเป็นงานต่อยอด

## 2. ผลลัพธ์รอบนี้และขอบเขตเสนอ

ผลลัพธ์หลักคือโปรแกรม Craft ที่เชื่อมต่อแต่ละฐานข้อมูล ส่งคำสั่ง SQL พร้อม parameters อ่านผลลัพธ์ และทำ transaction ได้ โดยไม่ต้องมี HTTP server

| ความสามารถ | ขอบเขตขั้นต่ำที่เสนอ |
| --- | --- |
| Connection | เลือก driver, กำหนด connection configuration, ตรวจการเชื่อมต่อ และปิดใช้งาน |
| Pool | มีขีดจำกัด connections และกำหนดอายุ/เวลาว่างของ connections ได้ พร้อม default ที่อธิบายไว้ |
| Execute | INSERT / UPDATE / DELETE และคืนข้อมูลผลการทำงานตามความสามารถของ driver |
| Query | SELECT, ส่ง parameters แยกจาก SQL และอ่านชื่อคอลัมน์/ค่าผลลัพธ์ |
| Values | กำหนดการรับส่ง NULL, Bool, integers, floating point, text, binary, decimal และวันเวลาอย่างชัดเจน |
| Transaction | Begin, commit, rollback; ปิด transaction ที่ยังไม่เสร็จอย่างปลอดภัยตาม ownership contract |
| Cancellation | กำหนด timeout และวิธีเชื่อมกับ context/lifecycle ของ Rev.4 พร้อมข้อจำกัดของแต่ละ driver |
| Errors | Craft exceptions ที่แยกประเภทใช้งานได้ โดยไม่เผย credentials หรือ parameter values ใน diagnostics ตามค่าเริ่มต้น |
| Delivery | Drivers และ runtime support มากับชุดส่งมอบ Craft; มีเอกสาร config และตัวอย่างสำหรับทั้งสามฐานข้อมูล |

ตารางนี้เป็นขอบเขตต้นทาง รายละเอียด API ที่ implement แล้วอยู่ใน [REV5-DATABASE](../../docs/REV5-DATABASE.md); ใช้ `std.db` โดยไม่ต้อง import package

## 3. การแบ่งความรับผิดชอบ

- ชั้น runtime/driver รับผิดชอบการเชื่อมต่อ ส่งคำสั่ง อ่านค่าผลลัพธ์ pool และ resource lifecycle
- โค้ด Craft รับผิดชอบ SQL, business logic และการประกอบเป็น repository/service/framework ตามความต้องการของแอป
- ตัวอย่างฐานข้อมูลต้องใช้โดยตรงได้โดยไม่ import package ระดับ framework
- ใช้ package reference เดิมเป็น integration fixture ได้เมื่อจำเป็น แต่ไม่เพิ่มระบบ framework ทั่วไปเพื่อให้ database milestone ผ่าน
- API อยู่ใน standard library ชื่อ `std.db`; drivers ฝังใน CLI และไม่ต้องมี manifest dependency
- ชุดส่งมอบ Craft มีตัวเชื่อมต่อ; ไม่รวมการติดตั้ง database server หรือสร้างบริการฐานข้อมูลบนเครื่องผู้ใช้

## 4. ประเด็น contract ต้นทาง

ตัดสินใจแล้วใน [REV5-DATABASE](../../docs/REV5-DATABASE.md): driver DSNs/parameters ตาม vendor, immutable DbValue, bounded materialized rows, function-owned handles, default transaction isolation และ errors ที่ไม่เผยข้อมูลจาก driver

1. **Drivers และ compatibility:** เลือกไลบรารี เวอร์ชัน license และฐานข้อมูลเวอร์ชันที่รองรับจากเอกสารทางการ ตรวจข้อกำหนด build/Windows และบันทึกไว้ก่อนเพิ่ม dependencies
2. **SQL parameters:** ระบุ placeholder syntax และการ bind ของแต่ละฐานข้อมูล ห้ามนำค่ามาต่อ SQL string; ไม่รับรองว่า SQL ชุดเดียวใช้ได้ข้ามทุกฐานข้อมูล และไม่สร้าง SQL translator ในรอบนี้
3. **Type mapping:** กำหนด SQL NULL แยกจากค่าว่าง, overflow, timezone, binary และ decimal ที่ไม่สูญเสีย precision; ค่าหรือชนิดที่ไม่รองรับต้องมี error ชัดเจน
4. **Result lifetime:** เลือกวิธีอ่าน rows และการปิด resources ที่ชัดเจน กำหนดขอบเขตการใช้หน่วยความจำ ไม่โหลดผลลัพธ์ขนาดไม่จำกัดโดยไม่มี contract
5. **Handle ownership:** Connection/pool, transaction และ rows เป็น native resources ต้องกำหนด copy/close/use-after-close/task sharing ให้สอดคล้องกับ value semantics ของ Craft
6. **Transactions:** กำหนด scope, commit failure, rollback หลัง exception/cancellation และ isolation ที่รองรับ ไม่เพิ่ม nested transactions/savepoints ในขั้นต่ำ
7. **Driver differences:** กำหนดผล affected rows และ generated IDs ตามความสามารถจริง ไม่สมมติว่าทุก driver คืนค่าแบบเดียวกัน
8. **Configuration:** รองรับ credentials จาก runtime configuration/environment และทางเลือก TLS ของ driver; ไม่ฝัง secrets ลง source bundle หรือแสดงในข้อความผิดพลาด
9. **Cancellation:** บอกข้อจำกัดเมื่อ driver/server ยกเลิกงานไม่ได้ทันที และวิธีคืน connection/resources หลัง timeout

## 5. สิ่งที่เลื่อนไปก่อน

- ฐานข้อมูลเพิ่มเติม เช่น SQLite และฐานข้อมูลแบบ NoSQL
- ORM, query builder, schema migration CLI, schema introspection และ code generation
- Automatic model mapping, relation loading และระบบ repository สำเร็จรูป
- Database administration UI, database server installer และการ provision cloud database
- Community registry, การดาวน์โหลด package จาก GitHub และระบบจัดการ third-party packages
- การสร้าง framework ใหม่ หรือขยาย router/authentication/DI/OpenAPI โดยไม่มีความจำเป็นต่อ database work
- การรับรอง production throughput, distributed transactions, replication หรือ automatic retries ของคำสั่งที่อาจเขียนข้อมูลซ้ำ

## 6. Milestones ที่ตรวจรับแยกได้

| Milestone | ผลส่งมอบ | เกณฑ์ก่อนเดินต่อ |
| --- | --- | --- |
| M0 — Contract | API design, driver/version matrix, type mapping, resource/error/cancellation rules | ข้อกำหนดหลักชัดเจนและมีตัวอย่าง syntax เสนอ โดยระบุว่ายังไม่ใช่ API ที่มีแล้ว |
| M1 — Core + first driver | Connection/query/execute/transaction สำหรับหนึ่งฐานข้อมูล | เตรียมตัวอย่างและ acceptance fixtures ครบวงจร; เลือกฐานข้อมูลแรกตาม environment ที่ผู้ใช้ตรวจรับได้ |
| M2 — Three drivers | MySQL, PostgreSQL และ SQL Server ภายใต้ contract เดียว พร้อมความต่างที่บันทึกไว้ | มี acceptance matrix แยกทั้งสามตัว ไม่ใช้ผล driver เดียวแทนตัวอื่น |
| M3 — Lifecycle integration | Pool bounds, timeout/cancellation, cleanup และการใช้งานร่วมกับ tasks | เตรียมกรณี failure/cleanup และเพิ่ม HTTP integration เฉพาะที่จำเป็น |
| M4 — Handoff | CLI/installer artifacts, เอกสาร, ตัวอย่าง Craft และผลตรวจรับที่ผู้ใช้รายงาน | บันทึกสิ่งที่ built/source-reviewed/user-tested/pending แยกกัน |

การผ่าน driver แรกเป็น milestone ย่อย ไม่ถือว่า Rev.5 รองรับครบสามฐานข้อมูลแล้ว

## 7. เกณฑ์ตรวจรับที่เสนอ

Working preferences เดิมให้ผู้ใช้เป็นผู้รันทดสอบ แต่เมื่อ 19 กันยายน 2026 ผู้ใช้เตรียมฐานข้อมูลเฉพาะและอนุญาตให้ agent ใช้ทดสอบ Rev.5 ได้ตามรายละเอียดหัวข้อ 8 ข้อยกเว้นนี้ใช้กับการทดสอบ database ของ Rev.5 ในฐานข้อมูลที่ระบุ; งานอื่นยังใช้ working preferences เดิม

- [x] โปรแกรม Craft เชื่อมต่อ MySQL, PostgreSQL และ SQL Server ได้ตาม version matrix ที่ประกาศ
- [x] แต่ละ driver ผ่าน parameterized INSERT / SELECT / UPDATE / DELETE และจัดการค่าที่มี quotes เป็นข้อมูลได้ถูกต้อง
- [ ] NULL, binary, Unicode, decimal, วันเวลา และ numeric boundaries เป็นไปตาม type mapping
- [x] Commit บันทึกข้อมูล และ rollback/exception ไม่ทิ้งการเปลี่ยนแปลงของ transaction ที่ยังไม่ commit
- [x] Connection failure, invalid SQL และ constraint errors ให้ข้อผิดพลาดที่ใช้งานได้โดยไม่เผย secrets (sanitized errors; ดู tests/status)
- [ ] Timeout/cancellation, early return และการอ่าน rows ไม่ครบ คืน resources ตาม contract
- [ ] Pool มีขอบเขต และไม่สะสม connections/rows/transactions หลังทำงานซ้ำ
- [ ] ตัวอย่างใช้งานจาก source และ bundle ได้ พร้อม dependencies/runtime support ที่ต้องใช้
- [ ] ชุดส่งมอบ Windows ใช้งานได้โดยไม่ต้องติดตั้ง Go; database server และเงื่อนไขภายนอกระบุในคู่มือ
- [ ] ความสามารถ Rev.4 ที่ได้รับผลกระทบมี regression results จากผู้ใช้

ก่อนตรวจรับจริง ต้องระบุฐานข้อมูลเวอร์ชันและ environment ที่ใช้ทดสอบ ใช้ฐานข้อมูลทดสอบแยกและข้อมูล fixture สำหรับคำสั่งเขียน/ลบ ไม่ใช้ฐานข้อมูล production

## 8. ฐานข้อมูลทดสอบที่ผู้ใช้เตรียมไว้

บันทึกเมื่อ 19 กันยายน 2026: ผู้ใช้เตรียมฐานข้อมูลทั้งสามสำหรับให้ agent ทดสอบโดยเฉพาะ อนุญาตให้เก็บ credentials อย่างปลอดภัย และให้ปรับข้อมูลเชื่อมต่อหลังการทดสอบครั้งแรกไม่ผ่าน ผลล่าสุดเชื่อมต่อและอ่าน database identity/server version ผ่านครบสามตัว โดยใช้ฐานข้อมูล `craft_test` ที่มีอยู่แล้วทุกตัว ไม่ได้สร้างฐานข้อมูลหรือแก้ข้อมูลภายใน

| Database | Endpoint | Database name | Authentication | Environment variable สำหรับชุดทดสอบ |
| --- | --- | --- | --- | --- |
| SQL Server | `.\SQLEXPRESS` ผ่าน Shared Memory (`protocol=lpc`) | `craft_test` | SQL login `ai_tester` | `CRAFT_SQLSERVER_DSN` |
| PostgreSQL | `127.0.0.1:5432` | `craft_test` | User `postgres` | `CRAFT_POSTGRES_DSN` |
| MySQL | `127.0.0.1:3306` | `craft_test` | User `root` | `CRAFT_MYSQL_DSN` |

### ผลการตรวจการเชื่อมต่อล่าสุด

| Database | Server version | ผล |
| --- | --- | --- |
| MySQL | 8.4.11 | Ping และ `SELECT DATABASE(), VERSION()` ผ่าน |
| PostgreSQL | 18.6 / Windows x64 | Ping และ `SELECT current_database(), version()` ผ่าน |
| SQL Server | 17.0.1135.8 / Express Edition (64-bit) | Ping และอ่าน `DB_NAME()`/`SERVERPROPERTY` ผ่านด้วย SQL login |

- ใช้โปรแกรม Go ชั่วคราวนอก repository: `go-sql-driver/mysql` v1.10.1, `pgx/v5` v5.11.0 และ `microsoft/go-mssqldb` v1.11.0; ยังไม่ได้เพิ่ม dependencies เหล่านี้ในโปรเจกต์ Craft
- PostgreSQL แก้ชื่อ user จาก `postgre` เป็น `postgres`; MySQL เปลี่ยน database จาก `craft_rev5_test` เป็น `craft_test`; SQL Server เปลี่ยนจาก TCP port 1433/database `craft__test` เป็น instance `SQLEXPRESS`/database `craft_test`
- SQL Server instance นี้ปิด TCP/IP และ SQL Browser หยุดอยู่ จึงใช้ driver extension `github.com/microsoft/go-mssqldb/sharedmemory` กับ URL host/path `localhost/SQLEXPRESS` และ `protocol=lpc`; ต้องรวม extension นี้หาก Rev.5 ใช้ config ชุดนี้ทดสอบบน Windows ไม่มีการแก้ server configuration หรือ restart service
- Shared Memory ใช้สำหรับการเชื่อมต่อภายในเครื่อง Windows เท่านั้น ผลนี้ยังไม่ยืนยัน SQL Server TCP/TLS transport; หากต้องทดสอบ TCP ภายหลังต้องตั้ง listener และ config เพิ่ม
- ผลนี้ยืนยัน credentials, การเข้าถึง database และ read-only query เท่านั้น ยังไม่ยืนยันสิทธิ์ CRUD/DDL, transaction, lifecycle หรือ database API ของ Craft

### การเก็บและโหลด credentials

- เก็บ DSNs แบบเข้ารหัสด้วย Windows DPAPI ผ่าน PowerShell SecureString/Export-Clixml ที่ `.local/secrets/rev5-databases.clixml` ภายใต้ project root
- ถอดรหัสได้ด้วยบัญชี Windows และเครื่องเดิมที่บันทึกเท่านั้น พร้อมจำกัด ACL ของโฟลเดอร์ให้บัญชีปัจจุบันและ SYSTEM
- `.local/` ถูก exclude ใน `.gitignore`; ไม่ใส่ credentials ใน roadmap, `craft.toml`, source, bundle, installer หรือ logs และไม่คัดลอกไฟล์ private นี้ไปใน artifacts
- หากย้ายเครื่องหรือเปลี่ยนบัญชี ให้ provision credentials ใหม่; ห้ามถือว่าไฟล์ CLIXML นี้ใช้ข้ามเครื่องได้
- SQL Server/PostgreSQL ใช้ URL ที่เติม scheme ให้ครบ และ percent-encode อักขระพิเศษในส่วนรหัสผ่าน; MySQL คง DSN syntax ของ driver ไว้
- คง DSN options ตามที่ผู้ใช้ให้; SQL Server ใช้ Shared Memory จึงยังไม่ใช่หลักฐานตรวจรับ TCP/TLS แม้ DSN ระบุ encryption/trust certificate; PostgreSQL ใช้ `sslmode=disable`; MySQL ใช้ `utf8mb4`, `parseTime=true`, `loc=UTC`

โหลดเข้ากระบวนการ PowerShell สำหรับรันชุดทดสอบเมื่อเริ่ม implementation โดยรันจาก project root (ห้ามพิมพ์ค่าตัวแปรหรือ DSNs ลง output):

```powershell
$rev5Secrets = Import-Clixml -LiteralPath .\.local\secrets\rev5-databases.clixml
try {
    foreach ($rev5Key in $rev5Secrets.Keys) {
        $rev5Credential = [System.Net.NetworkCredential]::new('', $rev5Secrets[$rev5Key])
        [Environment]::SetEnvironmentVariable($rev5Key, $rev5Credential.Password, 'Process')
    }
} finally {
    Remove-Variable rev5Credential, rev5Secrets, rev5Key -ErrorAction SilentlyContinue
}
```

ค่าจะเป็น plaintext ใน environment ของ process ระหว่างทดสอบ จึงโหลดเฉพาะ terminal/process ที่ใช้ทดสอบและปิดหลังใช้ หรือเคลียร์ด้วย:

```powershell
'CRAFT_SQLSERVER_DSN', 'CRAFT_POSTGRES_DSN', 'CRAFT_MYSQL_DSN' | ForEach-Object {
    [Environment]::SetEnvironmentVariable($_, $null, 'Process')
}
```

### ขอบเขตการทดสอบที่ได้รับอนุญาต

- ใช้เฉพาะ endpoints และ databases ในตาราง แม้บัญชีที่ให้จะมีสิทธิ์กว้างกว่านั้น
- สร้าง fixtures ที่ตั้งชื่อขึ้นต้น `craft_rev5_` สำหรับ CRUD, type mapping, transactions, timeout และ cleanup; ลบเฉพาะ objects ที่ชุดทดสอบสร้างเอง
- ไม่ drop database, เปลี่ยนบัญชี/สิทธิ์, แก้ server configuration หรือแตะข้อมูลนอก fixtures
- ก่อนรันทดสอบจริง อ่าน server version และตรวจ current database ให้ตรงกับเป้าหมาย; บันทึกผลแยกทั้งสาม driver โดยปกปิด credentials
- ผลตรวจการเชื่อมต่อวันที่ 19 กันยายนเป็น baseline; ผู้ใช้อนุญาตให้ดำเนิน implementation เมื่อ 20 กันยายนแล้ว ผล database API ที่รันจริงและข้อที่ยังรอตรวจรับอยู่ใน [สถานะ Rev.5](output/phase1-rev5-status.md)

อ้างอิง: [Rev.4 roadmap](craft-phase1-rev4.md), [Rev.4 status](output/phase1-rev4-status.md)
