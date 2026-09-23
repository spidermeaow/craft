# Craft Phase 1 Rev.6 — HTTP + Database stability

วันที่: 20 กันยายน 2026  
ฐาน: Craft 0.1.5 / Phase 1 Rev.5  
เป้าหมายรุ่น: 0.1.6  
สถานะ: ลงมือพัฒนาแล้ว มี runtime fixes, ตัวอย่าง, ชุดตรวจรับ และ build 0.1.6; ยังไม่รันทดสอบตาม preference ของผู้ใช้ ดู [status](output/phase1-rev6-status.md)

## 1. เป้าหมาย

ทำให้ `std.http` และ `std.db` ใช้งานร่วมกันได้อย่างถูกต้องและมีเสถียรภาพ โดยพิสูจน์ผ่านแอป Craft ขนาดเล็กที่รับ HTTP requests และทำงานกับฐานข้อมูลจริง ครอบคลุม requests พร้อมกัน, transactions, timeout, cancellation และการคืน resources

ใช้ primitives ที่มีอยู่ก่อน เพิ่มหรือแก้ API เฉพาะเมื่อพบช่องว่างที่จำเป็นจากกรณีใช้งานจริง ไม่สร้าง API Framework ใหม่เป็นผลส่งมอบหลัก

ผลที่ผู้ใช้ควรได้รับคือแอปตัวอย่างที่เปิด server, ใช้ connection pool ร่วมกัน, ทำ CRUD และ transaction ต่อ request และหยุด server ได้โดยไม่ทิ้ง connection หรือ transaction ที่ยังเปิดอยู่ พร้อมเอกสารข้อจำกัดที่ตรวจสอบได้

## 2. Baseline ที่มีหลักฐาน

- Rev.4 มี HTTP transport, request context, limits, lifecycle และ `api-framework` ตัวอย่าง ผู้ใช้เคยรายงาน tests 7/7 และ server startup ผ่าน
- Rev.5 มี database drivers สำหรับ MySQL, PostgreSQL, SQL Server พร้อม typed parameters/results, pool, timeout, transaction และ function-owned cleanup
- Database acceptance ของ Rev.5 ผ่านทั้งสามตัว ผู้ใช้ยืนยัน Craft 0.1.5 เชื่อมต่อและรัน SELECT/parameter smoke test ผ่านแล้ว
- ยังไม่มีผลตรวจรับครบชุดสำหรับ HTTP requests ที่ใช้ DB ร่วมกัน, client disconnect ระหว่าง query, transaction isolation ระหว่าง requests และ shutdown ขณะมี DB operations
- Regression Rev.1–Rev.4 บน runtime รุ่นล่าสุดยังไม่ได้รันครบ การผ่านแต่ละ revision ในอดีตไม่แทนผล regression บน Rev.6

อ้างอิง: [Rev.4 status](output/phase1-rev4-status.md), [Rev.5 status](output/phase1-rev5-status.md), [HTTP bridge](../../internal/interpreter/http.go), [HTTP transport](../../internal/stdlib/http.go), [DB runtime](../../internal/stdlib/database.go)


## 3. ขอบเขตและลำดับความสำคัญ

| ระดับ | งาน | ผลส่งมอบ |
| --- | --- | --- |
| P0 | Regression baseline ของ Rev.1–Rev.5 | รายการทดสอบ ผลจริง และ defects ที่ต้องแก้ก่อนส่งมอบ |
| P0 | แอป HTTP + DB ด้วย Craft | ใช้ `std.http`/`std.db` โดยตรงและตรวจรับได้ทั้งสาม driver |
| P0 | Concurrent requests และ pool limits | ไม่ปะปนข้อมูล/transaction, ไม่เกิน connection limit และ pool wait ยกเลิกได้ |
| P0 | Timeout/cancellation/cleanup | Request timeout, disconnect และ exception ไม่ทิ้ง transaction หรือปิด pool ของ requests อื่น |
| P0 | Server shutdown | Stop admission, drain/cancel, จบงานลูก แล้วปิด pool ตามลำดับ |
| P0 | Diagnostics | แยก DSN ว่างจาก connection failure; ข้อความมีประโยชน์โดยไม่เปิดเผย secrets |
| P0 | คู่มือและผลตรวจรับ | ระบุ environment, workload, lifecycle และข้อจำกัดตามผลจริง |
| P1 | Lifecycle race checks และ repeated-load measurements | ตรวจ races และแนวโน้ม resource หลัง workload หาก environment รองรับ |
| P1 | Bundle/deployment smoke และ installer regression | ตรวจตัวอย่างนอก source tree, version/PATH และขั้นตอนติดตั้งที่เปลี่ยน |
| ภายหลัง | Configuration library / dotenv convenience | ไม่เป็นเงื่อนไขหลักของ Rev.6 |

Race หรือ resource leak ที่พบจริงต้องแก้ก่อนสรุปว่ากรณีนั้นผ่าน แม้การวัดขยายผลจะเป็น P1 หากลดขอบเขตต้องระบุรายการที่ยังไม่ได้ทดสอบ ไม่ลดมาตรฐานด้วยการซ่อน failures


## 4. แอปพิสูจน์ความพร้อม

เสนอ `examples/rev6-http-db` เป็นโปรเจกต์แยกจากตัวอย่าง Rev.4/Rev.5 เพื่อไม่ให้ main ใหม่ทำให้ tests เดิมเสีย ใช้ dispatcher Craft ธรรมดา; routing/response helpers มีเท่าที่ fixture ต้องใช้

| Endpoint ที่เสนอ | จุดประสงค์ | ผลที่ต้องตรวจ |
| --- | --- | --- |
| `GET /health` | ตรวจว่า HTTP process ยังตอบได้ โดยไม่เรียก DB | Request อื่นยังใช้ได้เมื่อ DB query ล้มเหลว |
| `GET /ready` | ตรวจ DB ด้วย bounded ping/query | ตอบ unavailable เมื่อ DB ใช้ไม่ได้ โดยไม่เผยรายละเอียดภายใน |
| `POST /items` | สร้าง fixture row ด้วย parameters | ข้อมูลถูกบันทึกตาม response ที่ยืนยัน commit แล้ว |
| `GET /items/:id` | อ่านข้อมูลเฉพาะแถว | ไม่พบให้ 404; ไม่ปะปนข้อมูลของ concurrent requests |
| `PUT /items/:id` | แก้ข้อมูลด้วย parameters | อ่านกลับได้ค่าที่แก้; ไม่ใช้ string concatenation กับ user values |
| `DELETE /items/:id` | ลบ fixture row | ลบเฉพาะข้อมูลเป้าหมาย |

Slow-query, forced-error และ transaction-rollback scenarios อยู่ใน test fixtures เท่านั้น ไม่เปิดเป็น debug endpoints ของแอปปกติ กำหนด status/body ที่คาดหวังในชุดทดสอบก่อน implementation; policy การ map errors เป็น HTTP responses อยู่ในแอป Craft ไม่ย้าย business policy เข้า Go bridge

เลือก driver ตอนเริ่มโปรแกรมด้วย configuration ที่มีอยู่ ไม่เปลี่ยน driver กลาง request SQL และ placeholders แยกตาม driver โดยใช้ interface ของ Craft เดิม ไม่เพิ่ม SQL translator

โครงสร้าง lifetime ที่ต้องพิสูจน์:

```text
main / server owner
  เปิด DbConnection หนึ่ง pool
  เรียก std.http.serve โดยส่ง pool ใน state
    request A → query / transaction A → cleanup
    request B → query / transaction B → cleanup
  server drain/cancel และรอ handlers จบ
  ปิด pool
```

HTTP state แบ่งปันเฉพาะ pool handle และ configuration ที่เหมาะสม ไม่แบ่งปัน `DbTransaction` หรือ mutable request state

## 5. Lifecycle contract ที่ต้องตรวจและปรับเมื่อพบ defect

### Connection ownership

- Pool ถูกสร้างในฟังก์ชันเจ้าของ server ที่ยังไม่ return ระหว่าง `serve`; helper รับ pool เป็น argument
- คงกติกา Rev.5: connection ที่สร้างใน helper และคืนออกไปจะปิดเมื่อ helper จบ ต้องมีตัวอย่างและ diagnostic อธิบายข้อจำกัดนี้
- Request จบหรือถูกยกเลิกต้องไม่ปิด borrowed pool ที่ main เป็นเจ้าของ
- Pool ที่สร้างเฉพาะ request ต้องปิดเมื่อ owner จบ รวมถึง exception/cancellation
- Native handle aliases ยังคงอ้าง resource เดียวกัน การ close ผ่าน alias ต้องให้ผลสอดคล้องกับ contract และไม่ panic
- ยังไม่เพิ่ม ownership transfer หรือย้าย owner ของ returned handle อัตโนมัติ หาก use case บังคับให้เปลี่ยน ต้องออกแบบ compatibility/migration แยกก่อน

### Transactions และความถูกต้องของข้อมูล

- แต่ละ request สร้าง transaction ของตนเอง; ไม่มี transaction ร่วมกันข้าม requests
- ก่อน commit: failure/exception/cancellation ต้องจบด้วย rollback เมื่อ owner scope จบ และไม่ทิ้ง partial writes
- หลัง commit สำเร็จแล้ว การ disconnect ไม่สามารถย้อนข้อมูลที่ commit ไปแล้ว ต้องตรวจกรณีนี้แยกจาก rollback
- หาก commit ล้มเหลวเพราะ connection ขาด ผลบน server อาจไม่แน่นอน ห้าม auto-retry writes หรืออ้างว่า rollback สำเร็จเสมอ
- Statement error ไม่ควรถูกตีความว่า transaction ทั้งก้อน rollback ทันทีโดยอัตโนมัติ ให้แอปหยุดใช้งานและ rollback ตาม contract
- ยืนยันข้อมูลหลังจบ request ด้วย connection/transaction ใหม่ ไม่ใช้เพียง HTTP status เป็นหลักฐาน

### Timeout และ cancellation

- Operation ต้องได้รับ execution context ของ request/task ที่เรียกจริง ไม่ใช้ background context จน query ทำงานต่อหลัง request จบโดยไม่ตั้งใจ
- ตรวจทั้ง HTTP deadline สั้นกว่า DB timeout และ DB timeout สั้นกว่า HTTP deadline
- ทดสอบ cancellation ขณะรอ pool, ขณะ query ทำงาน และขณะ transaction มี uncommitted changes
- ใช้ driver/server-specific synchronization หรือ barriers เพื่อยืนยันว่า query อยู่ในขั้นตอนเป้าหมายก่อน cancel; ไม่พึ่ง arbitrary sleep เป็นหลักฐานเพียงอย่างเดียว
- หลัง cancellation ต้องคืน connection ที่ยังใช้ได้หรือ discard connection ที่เสีย และ request ถัดไปทำงานได้
- Timeout เป็น cooperative contract: วัดเวลาคืน control/cleanup และบันทึกข้อจำกัดของ driver โดยเฉพาะ Shared Memory ไม่รับรองว่าหยุด SQL ทุกกรณีทันที
- คง Rev.5 ที่ `timeoutMs` ครอบคลุม transaction lifetime ทั้งหมดก่อน แยก timeout fields เพิ่มเฉพาะเมื่อมีผลทดสอบชี้ว่าจำเป็น


### Concurrency และ overload

- กำหนด `HttpConfig.maxConcurrent` และ `DbConfig.maxOpen` ให้ต่างกันเพื่อพิสูจน์ทั้ง HTTP admission limit และ DB pool waiting
- Request ที่ admission ปฏิเสธต้องไม่เริ่ม DB operation; ใช้ behavior ของ HTTP contract เดิมเป็น baseline
- Requests ที่รอ pool ต้องออกได้ด้วย deadline/cancellation โดยไม่ทำให้ server ค้าง
- ใช้ unique row/request IDs และตรวจจำนวน/ค่าข้อมูลที่ถูกต้อง ไม่ใช้เพียงจำนวน HTTP 200
- ไม่รับรองว่าไม่มี lost update จาก read-modify-write ของแอปโดยไม่มี transaction/SQL concurrency policy ที่เหมาะสม

### Shutdown

- หยุดรับงานใหม่ → drain handlers ภายใน deadline → cancel งานที่เหลือ → รอ child tasks/DB cleanup → ปิด pool และ listener
- ทดสอบทั้ง explicit stop และ process cancellation/Ctrl+C ตาม convention ที่ประกาศ
- ต้องไม่มี deadlock จากการปิด pool ขณะที่ handler ยังใช้ transaction หรือถือ resource ที่ shutdown รออยู่
- ทดสอบ start/stop ซ้ำและนำ port กลับมาใช้ได้; failures ระหว่าง cleanup ต้องไม่กลบข้อผิดพลาดต้นเหตุ
- ถ้า cleanup เกิน deadline ให้รายงานตามจริงและกำหนดข้อจำกัด ห้ามรายงานว่า shutdown จบทั้งที่งานยังค้าง

## 6. Diagnostics ที่ปรับในขอบเขตนี้

- `std.db.connect` ปฏิเสธ DSN ว่าง/มีแต่ whitespace ก่อนเรียก driver พร้อมข้อความว่าข้อมูลเชื่อมต่อยังไม่ได้กำหนด
- แยก connection refused, authentication failure, unknown database และ timeout เมื่อมี typed error/vendor code ที่เชื่อถือได้
- รักษา exception types ที่ผู้ใช้จับอยู่แล้วเท่าที่ทำได้; หากต้องเปลี่ยนให้มี compatibility note
- Unknown failure ใช้ sanitized fallback ไม่เดาสาเหตุจากข้อความ ไม่เผย DSN/password/SQL parameters/raw driver messages
- HTTP error responses ไม่เผย connection settings, stack trace หรือข้อความภายใน; server diagnostics ควรชี้ตำแหน่ง Craft ที่เกิดปัญหา
- เอกสารอธิบายว่า `std.env` อ่าน process environment; configuration package/framework สามารถจัดการไฟล์เพิ่มเองได้ โดยไม่เพิ่ม dotenv loader ใน CLI รอบนี้

## 7. Test matrix และ environment

ใช้ฐานข้อมูลเฉพาะที่ได้รับอนุญาตจาก Rev.5 และตรวจ current database ก่อนเขียนเสมอ:

| Driver | Server baseline | เป้าหมาย | Transport ที่มีแล้ว |
| --- | --- | --- | --- |
| MySQL | 8.4.11 | `craft_test` | localhost TCP |
| PostgreSQL | 18.6 | `craft_test` | localhost TCP |
| SQL Server | Express 17.0.1135.8 | `craft_test` | `.\SQLEXPRESS` Shared Memory |

บันทึก server version ใหม่ตอนทดสอบ ไม่สมมติว่าเครื่องคงเวอร์ชันเดิม Credentials โหลดด้วยกลไก local ที่มีอยู่ ไม่คัดลอกค่าจริงลงเอกสาร/source/bundle

| กรณี | MySQL | PostgreSQL | SQL Server |
| --- | --- | --- | --- |
| HTTP CRUD + parameter binding | ต้องตรวจ | ต้องตรวจ | ต้องตรวจ |
| Concurrent requests / pool bounds / state isolation | ต้องตรวจ | ต้องตรวจ | ต้องตรวจ |
| Commit / explicit rollback / exception rollback | ต้องตรวจ | ต้องตรวจ | ต้องตรวจ |
| Request timeout / client disconnect / pool wait cancellation | ต้องตรวจ | ต้องตรวจ | ต้องตรวจ |
| Health หลัง DB error / request ถัดไปหลัง cancellation | ต้องตรวจ | ต้องตรวจ | ต้องตรวจ |
| Drain/shutdown/restart และ resource cleanup | ต้องตรวจ | ต้องตรวจ | ต้องตรวจ |

SQL Server Shared Memory ผ่านไม่เท่ากับ TCP/TLS ผ่าน; TCP/TLS เป็น P1 แยกโดยต้องเตรียม environment และได้รับอนุญาตก่อนเปลี่ยน service configuration ไม่ใช้ปัญหาของ TCP ที่ยังไม่เปิดเป็นเหตุให้แก้เครื่องเอง

ข้อกำหนดของ test fixtures:

- HTTP tests ใช้ loopback และ ephemeral port พร้อม client/startup/operation/shutdown timeouts
- ใช้ชื่อ tables/rows ที่มี prefix `craft_rev6_` และ unique run ID สร้างและลบเฉพาะ fixtures ของ run นั้น ไม่ drop database หรือแตะข้อมูลแอปอื่น
- ลงทะเบียน cleanup ทันทีหลังสร้าง resources; เมื่อ failure ต้องระบุชื่อ fixture ที่อาจเหลือเพื่อให้ตรวจสอบได้ โดยไม่ลบข้อมูลด้วย wildcard กว้าง
- ใช้แอป Craft จริงผ่าน interpreter/CLI อย่างน้อยหนึ่ง end-to-end flow ต่อ driver ไม่ใช้ Go driver probes แทนผลของ Craft
- Regression ต้องรวม parser/checker/module/bundle ที่เกี่ยวข้อง, Rev.3 task/defer, Rev.4 HTTP และ Rev.5 database tests
- เก็บ command, exit code, stdout/stderr ที่ sanitize, driver/server/OS version และ configuration ที่ไม่ลับ


## 8. การวัด resources และเกณฑ์ความเสถียร

แยก acceptance ที่ตรวจได้แน่นอนจาก performance exploration:

- หลัง workload สงบ `inUse` ต้องกลับเป็น 0, open connections ไม่เกิน maxOpen; idle connections ที่อยู่ใน pool ตาม config ไม่ถือว่า leak
- หลัง owner/server จบ pool ต้องปิด, listener ต้องคืน port และไม่มี tasks/transactions ที่เป็นของ workload ค้าง
- Workload ทำซ้ำหลายรอบใน process เดียว ต้องไม่ชน cumulative handle limit ทั้งที่ resources เก่าจบแล้ว
- แผน measured workload เริ่มจาก warmup และอย่างน้อยสาม batch ขนาดเท่ากัน บันทึกจำนวน requests, concurrency, payload, elapsed time, DB pool stats, goroutines, heap และ OS handles ก่อน/หลัง batch และหลัง shutdown
- เปรียบเทียบช่วง quiescent ที่เหมือนกัน; heap/OS memory ไม่จำเป็นต้องกลับค่าเดิมเป๊ะเพราะ GC/pool/allocator ต้องดู retained objects และแนวโน้มร่วมกัน
- กำหนดตัวเลข workload/timeout/tolerance ใน test configuration ก่อนรัน ไม่ปรับเกณฑ์ย้อนหลังเพื่อให้ผลผ่าน และไม่อ้าง production throughput จาก smoke test
- Race checks รันบน environment ที่รองรับและตามสิทธิ์ทดสอบที่ได้รับ หากไม่ได้รันต้องแสดง pending ชัดเจน

## 9. Milestones และเกณฑ์ตรวจรับ

| Milestone | ผลส่งมอบ | เกณฑ์ผ่าน |
| --- | --- | --- |
| M0 — Baseline/contracts | Regression inventory, lifecycle scenarios, test configuration และ HTTP response policy ของ fixture | แยกสิ่งที่มีหลักฐาน/ยังค้าง ไม่เปลี่ยน semantics โดยไม่มีเหตุผล |
| M1 — Vertical integration | แอป Craft HTTP + DB และ isolated fixtures | HTTP CRUD ผ่านทั้งสาม driver พร้อมยืนยันข้อมูลจริง |
| M2 — Failure/lifecycle | Concurrent load, cancellation, rollback และ shutdown tests พร้อม fixes | กรณี P0 ผ่าน ไม่ทิ้ง resources หรือกระทบ requests อื่น |
| M3 — Diagnostics/regression | ข้อความผิดพลาดที่ปลอดภัยและผล regression หลังแก้ | ไม่ทำให้ความสามารถเดิมที่เกี่ยวข้องเสีย และบันทึก suite ที่ยังไม่ได้รัน |
| M4 — Handoff | เอกสาร, ตัวอย่าง, build artifacts และ status ledger | ผู้ใช้ทำตามได้; version/ผลทดสอบ/pending ตรงกับสิ่งที่ส่งมอบ |

- [ ] HTTP CRUD ด้วย Craft ผ่านทั้งสามฐานข้อมูลตาม test matrix
- [ ] Requests พร้อมกันไม่ปะปน state หรือ transaction และไม่เกิน limits
- [ ] Timeout/disconnect ก่อน commit ไม่ทิ้ง uncommitted writes; ผลหลัง commit/ambiguous commit แยกชัดเจน
- [ ] Pool wait และ DB operation ยกเลิกได้ตาม driver contract; subsequent requests ยังใช้ได้
- [ ] Shutdown คืน listener/pool/child resources และ start ใหม่ได้
- [ ] Diagnostics ของ config/DB/HTTP ใช้งานได้โดยไม่เผย secrets
- [ ] Regression P0 ของความสามารถเดิมผ่านบนรุ่นที่ส่งมอบ; tests ที่ไม่ได้รันคง pending ไม่ถือเป็น PASS
- [ ] ตัวอย่างและ docs ระบุ function ownership, transaction deadline และ transport limitations ตรงกับ behavior
- [ ] Build CLI/installer และ checksum สำเร็จเมื่อ implementation พร้อม; build ไม่แทนผล installer execution

ก่อนเริ่ม implementation ให้กำหนดรายชื่อ regression P0 ที่ต้องรันไว้ใน status; ไม่ใช้คำว่า “เสถียรครบแล้ว” หากยังไม่ได้ตรวจ P0 ครบ เมื่อพบปัญหาต้องแนบ reproducible case และผลหลังแก้

## 10. สิทธิ์ทดสอบและงานที่เลื่อน

แผนครั้งนี้ยังไม่ใช่การรันทดสอบ Working preferences ยังคงให้ผู้ใช้ทดสอบเองโดยค่าเริ่มต้น การอนุญาตที่มีอยู่ใช้กับฐานข้อมูลทดสอบและ fixtures ตามขอบเขตเดิม ไม่ขยายเป็นสิทธิ์เปิด installer/editor, เปลี่ยน server config หรือรันทุก suite โดยอัตโนมัติ เมื่อพัฒนาให้เตรียมคำสั่งตรวจรับและใช้ผลที่รันจริงเท่านั้นในการอัปเดต status

เลื่อนไปก่อน: dotenv auto-loading ใน CLI, secret store/`craft secret`, configuration package, API Framework ใหม่, ORM/query builder/migrations, database เพิ่มเติม, ownership transfer, distributed transactions, automatic write retries และ performance guarantees ระดับ production

ผลส่งมอบของ Rev.6 คือพื้นฐาน HTTP + Database ที่พิสูจน์การทำงานร่วมกันแล้ว พร้อมการแก้ defects ที่พบและข้อจำกัดที่ชัดเจน เพื่อใช้เป็นฐานพัฒนาแอปหรือ framework ต่อไป
