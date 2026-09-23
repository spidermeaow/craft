# Craft Phase 1 Rev.7 — HTTP + Database evidence and lifecycle hardening

วันที่: 20 กันยายน 2026  
ฐาน: Rev.6 / Craft 0.1.6  
เป้าหมายรุ่น: 0.1.7  
สถานะ: implementation เตรียมแล้ว; targeted tests และ full acceptance ยังรอผู้ใช้ตรวจรับ

## 1. เป้าหมาย

Rev.7 รับช่วงจาก Rev.6 เพื่อทำให้ข้ออ้างเรื่อง HTTP + Database มีหลักฐานที่แยกสาเหตุได้จริง ก่อนตัดสินว่า runtime เสถียรหรือมี defect

ไม่สร้าง API Framework, ORM, migrations, dotenv loader หรือ database driver ใหม่ งานหลักคือทำให้ timeout, client disconnect, pool wait, transaction cleanup และ shutdown มี contract ที่ชัด ตรวจได้ทั้ง MySQL, PostgreSQL และ SQL Server แล้วแก้ runtime เฉพาะข้อที่พิสูจน์ว่าเป็น defect

ผลส่งมอบคือ test matrix ที่ไม่สับสนระหว่าง timeout คนละชนิด, integration tests ที่รายงานเหตุการณ์ครบ, fixes ที่มี regression coverage, status ที่ตรงกับผลจริง และคู่มือที่ระบุข้อจำกัดของ driver ตามหลักฐาน

## 2. สิ่งที่รับมาจาก Rev.6

Rev.6 ยังคงเป็นแผนและ baseline ของ HTTP + Database ตาม [Rev.6](craft-phase1-rev6.md) โดยไม่มีการเปลี่ยนขอบเขตย้อนหลัง Rev.7 อ้างอิงผลทดสอบที่เกิดขึ้นแล้วดังนี้

| หลักฐาน | ผล | ความหมาย |
| --- | --- | --- |
| ชุด `test-rev6.ps1` | FAIL, exit code 1 | มี failures ที่ต้องแยกสาเหตุก่อนนำไปตัดสิน runtime |
| Diagnostics empty DSN/redaction | PASS | เป็นหลักฐานเฉพาะ diagnostics ไม่ใช่ acceptance ของ HTTP + DB ทั้งหมด |
| Pool-wait disconnect/shutdown fixture | PASS ทั้งสาม driver | ยืนยันเส้นทาง fixture นี้ ไม่แทน query ที่เริ่มทำงานบน DB แล้ว |
| Probe Craft: HTTP timeout 1.5s, DB timeout 5s | PASS ทั้งสาม driver | `/slow` ได้ 504 ราว 1.5s, `inUse=0`, `/ping` ใช้ต่อได้ |
| Probe server stop | PASS ทั้งสาม driver | server stop ปกติ exit code 0 ในเส้นทางที่ทดลอง |

ผลที่ยังไม่สรุปเป็น defect:

- HTTP 200 แทน 504/500 ใน pool fixture เดิม เพราะ fixture ตั้ง read timeout 1s ร่วมกับ request timeout 1.5s และไม่ได้บันทึก context cause/response body เพียงพอ
- MySQL activity view ยังเห็น query หลัง client disconnect เพราะ assertion เดิมรวม `SQL active` กับ `pool inUse` และไม่มี timeline ของ cancellation ถึง driver/server
- PostgreSQL/SQL Server เห็น `inUse != 0` หลัง timeout จาก snapshot เดียว จึงยังไม่พิสูจน์ว่าเป็น leak ถาวรหรือ cleanup ที่กำลังจบ

ห้ามแก้ expected result ให้เป็น 200, ลบ assertions, หรือเพิ่ม timeout แบบไม่มีหลักฐานเพื่อให้ suite ผ่าน

## 3. Contract ที่ Rev.7 ต้องตรึง

| เหตุการณ์ | HTTP transport | App Craft ตัวอย่าง | หลักฐานผ่าน |
| --- | --- | --- | --- |
| Body อ่านเกินเวลา | 408 หากส่ง response ได้ และไม่ dispatch | ไม่มี DB operation | handler call count เป็น 0 |
| Request deadline หมดหลังรับ body | 504 หาก client ยังรับ response | app ไม่สามารถเปลี่ยนเป็น success | response/body/cause/elapsed ตรงกัน |
| DB deadline หมดก่อน HTTP deadline | dispatch error fallback 500 หาก app ไม่จับ | app map `DbTimeoutError` เป็น 504 ตาม policy ตัวอย่าง | deadlines อื่นยาวกว่าและ log error type |
| รอ DB pool แล้วหมดเวลา | operation คืน `DbTimeoutError`/`DbCancelledError` | request response ตาม context ที่เกิดจริง | WaitCount, cause, response และ subsequent request |
| Client disconnect | ไม่ต้องมี HTTP status | rollback/cleanup ตาม owner scope | transaction outcome และ pool quiescence |
| SQL started แล้ว cancel | ไม่ถือว่า client returned แปลว่า SQL หยุด | ไม่ทำ automatic write retry | timeline ของ DB return และ server activity แยกกัน |
| Shutdown เกิน drain deadline | ปิด admission, cancel handlers แล้วรายงาน failure | owner ปิด pool หลัง handlers/tasks จบ | listener reuse, pool closed, child task complete |

การทดสอบแต่ละกรณีต้องมี deadline ที่เป็นตัวหลักเพียงหนึ่งตัว: body timeout, request timeout หรือ DB timeout ส่วน timeout อื่นต้องยาวพอและบันทึกค่าไว้ใน test output

## 4. Work packages

### R7.1 — กู้ความน่าเชื่อถือของ tests

- แยก CRUD, pool wait, in-flight query, transaction rollback, client disconnect และ shutdown เป็น independent subtests ต่อ driver
- สร้าง `RunTrace` ที่บันทึก run ID, request ID, start/end monotonic time, HTTP status/body, context cause, DB error type, pool `open/inUse/idle/waitCount`, transaction outcome และ cleanup result
- failure ของ scenario หนึ่งต้องไม่หยุด execution ของ scenario อื่น; รายงานเป็น PASS, FAIL, SKIP หรือ NOT RUN ให้ถูกต้อง
- cleanup ต้องบันทึก table/port/pool outcome โดยไม่กลบ failure ต้นเหตุ และห้ามลบข้อมูลนอก prefix fixture

เกณฑ์ผ่าน: ผลหนึ่งบรรทัดตอบได้ว่า scenario ใด, driver ใด, context ใด และจบอย่างไร โดยไม่ต้องตีความจาก log ที่ขาดข้อมูล

### R7.2 — Deadline isolation

- สร้าง minimal raw transport fixture และ Craft application fixture แยกกัน
- request-timeout: HTTP deadline สั้นกว่า DB, write, read และ client budgets
- db-timeout: DB deadline สั้นกว่า HTTP, write, read และ client budgets
- body-timeout: ส่ง request body ช้าและยืนยันว่า dispatch ไม่ถูกเรียก
- ตรวจว่า read deadline หลังอ่าน body ครบไม่ยกเลิก handler ระหว่าง query โดยไม่ตั้งใจ

เกณฑ์ผ่าน: ได้ 408/504/500/504 ตาม contract ในตาราง โดยบันทึก cause ที่ยืนยันว่า timeout ตัวถูกต้องเป็นต้นเหตุ

### R7.3 — Pool and request cleanup

- ใช้ physical connection hold เป็น barrier สำหรับ pool wait และยืนยัน `WaitCount` ก่อน cancel
- รอ request completion signal ก่อนตรวจ pool แล้ววัดซ้ำภายใน cleanup budget ที่กำหนดล่วงหน้า
- แยก `inUse=0` (ไม่มีงานใช้ pool), connection returned/discarded, and `open=0` after owner close
- ตรวจ subsequent ping/query หลัง timeout และ start/stop ซ้ำบน port เดิม

เกณฑ์ผ่าน: ไม่มี operation ค้าง, pool ไม่เกิน maxOpen, request ถัดไปทำงานได้ และ owner close ทำให้ open connections เป็น 0 ภายใน budget

### R7.4 — In-flight SQL cancellation

- ใช้ SQL delay ที่มี marker เฉพาะ run และ activity view ของแต่ละ driver โดยไม่นับ observer connection
- บันทึก event แยก: SQL observed active, client disconnected, Craft DB call returned, scope cleanup finished, SQL no longer active
- เปรียบเทียบ direct driver probe กับ Craft path เมื่อพฤติกรรมต่างกัน เพื่อหา context propagation ที่หายไป
- ไม่เพิ่ม `KILL`, privileges, server configuration หรือ driver-specific cancel connection โดยอัตโนมัติ

เกณฑ์ผ่าน: หาก driver รองรับ cancellation ตาม transport ปัจจุบัน SQL ต้องจบภายใน budget ที่ประกาศ; หากไม่รองรับ ต้องมี direct-driver evidence, documented limitation และนำเสนอ contract change ให้ผู้ใช้ตัดสิน ไม่ถือ Rev.7 ผ่านเอง

### R7.5 — Transactions and shutdown

- ก่อน commit: exception, request timeout, client disconnect และ shutdown ต้องยืนยันด้วย fresh observer connection ว่าไม่มี write
- หลัง commit: disconnect ไม่ย้อน write ที่ commit สำเร็จ
- ambiguous commit: fault injection เป็น scenario แยก; ผล unknown ต้องไม่ auto-retry และต้องมี application identifier สำหรับตรวจซ้ำ
- shutdown ขณะ handler และ child task ถือ transaction ต้องตรวจ drain, cancellation, rollback, port reuse และ pool close

เกณฑ์ผ่าน: transaction ของ request ไม่ปะปนกัน, cleanup ไม่ deadlock และ outcome ที่ไม่แน่นอนไม่ถูกประกาศว่า rollback สำเร็จ

## 5. Test environment

ใช้ `craft_test` เท่านั้นและตรวจ current database ก่อนสร้าง fixture

| Driver | Baseline | Transport |
| --- | --- | --- |
| MySQL | 8.4.11 | localhost TCP |
| PostgreSQL | 18.6 | localhost TCP |
| SQL Server Express | 17.0.1135.8 | `.\SQLEXPRESS` Shared Memory |

fixtures ใช้ table/row prefix `craft_rev7_` พร้อม run ID เฉพาะ, loopback ephemeral port และ bounded timeouts ทุก operation ไม่ drop database, ไม่ใช้ wildcard cleanup และไม่บันทึก DSN/password ลง source, bundle, test output หรือ status

SQL Server Shared Memory ไม่แทน TCP/TLS. TCP/TLS, race detector, OS handles, installer execution และ full performance measurements เป็น P1 หลัง P0 ผ่าน

## 6. ลำดับส่งมอบและเกณฑ์รับ

| ขั้น | ผลส่งมอบ | เกณฑ์ผ่าน |
| --- | --- | --- |
| M0 | บันทึก baseline FAIL ของ Rev.6 และ RunTrace contract | ไม่มี assertion ที่ตีความหลายเหตุการณ์รวมกัน |
| M1 | Deadline isolation + pool-wait tests | ระบุ source ของทุก timeout ได้ทั้งสาม driver |
| M2 | In-flight cancellation + transaction/shutdown tests | cleanup and server-side effects มีหลักฐานแยก |
| M3 | Runtime/test fixes พร้อม targeted regressions | failures เดิมแก้โดยไม่ลด contract |
| M4 | Rev.7 integration matrix รันซ้ำ 3 รอบต่อ lifecycle scenario | P0 ผ่านทุก driver ไม่มี unreported cleanup failure |
| M5 | Full regression, docs, status, build | Status ตรงกับ commands, exit codes และ pending work |

สถานะ implementation ปัจจุบัน: ปรับ HTTP transport ให้ล้าง read deadline หลังรับ body, เพิ่ม raw HTTP deadline tests, เพิ่ม Craft integration test ที่แยก DB timeout และ request timeout ต่อสาม drivers, และเพิ่ม `test-rev7.ps1`/คู่มือแล้ว ส่วน R7.4, R7.5, repeated lifecycle, full regression และ release artifacts ยังต้องใช้ผลตรวจรับจริงก่อนปิด milestone

- [ ] Results จาก Rev.6 ถูกบันทึกเป็น baseline ไม่ถูกเขียนทับ
- [ ] Timeout ทุกชนิดแยกสาเหตุได้
- [ ] Client disconnect, pool wait, in-flight SQL และ shutdown มี trace ครบ
- [ ] CRUD/concurrency/transaction ผ่านทั้งสาม driver พร้อม independent read-back
- [ ] `inUse=0` หลัง scope completion และ pool ปิดหลัง owner exit ตาม contract
- [ ] Full regression Rev.1–Rev.6 ผ่านหลัง targeted tests ผ่าน
- [ ] ข้อจำกัด driver ที่ยังทำไม่ได้มีหลักฐานและ user-approved contract change

## 7. งานที่เลื่อน

Rev.7 ไม่ทำ dotenv auto-loading, secret store, API Framework ใหม่, ORM/query builder/migrations, database เพิ่มเติม, ownership transfer, distributed transactions หรือ automatic write retries

Rev.7 จะไม่รายงานว่า “ผ่าน” จาก build, compile-only, manual happy-path หรือ subset ที่ผ่านเพียงส่วนเดียว ต้องใช้ผล P0 ที่รันจริงครบตาม matrix และ status ที่อัปเดตจาก output นั้น
