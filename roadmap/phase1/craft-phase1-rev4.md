# Craft Phase 1 Rev.4

## Framework Foundations — พื้นฐานสำหรับพัฒนา API framework ด้วย Craft

วันที่: 19 กันยายน 2026  
สถานะ: ผู้ใช้ทดลอง Rev.4 ผ่านแล้วในชุด `craft-test`: check สำเร็จ, tests 7/7 (exit 0), HTTP server เริ่มฟังที่ 127.0.0.1:8080; การตรวจรับเพิ่มเติมยังมีรายการ pending ดู [สถานะ Rev.4](output/phase1-rev4-status.md)  
ฐาน: Craft 0.1.3 / Phase 1 Rev.3  
เป้าหมายรุ่น: 0.1.4 โดยต้องยืนยัน compatibility และ bundle contract ก่อน release

## 1. เป้าหมายและเงื่อนไขสำเร็จ

เพิ่มความสามารถให้ผู้พัฒนาสามารถเขียน API framework เป็นไลบรารี `.craft` และนำไปใช้กับแอปพลิเคชัน `.craft` ได้จริง โดยใช้ Go พัฒนาตัวภาษา runtime และเครื่องมือพื้นฐานที่ Craft ต้องเรียกใช้

**การเพิ่ม route, middleware, validation rule หรือรูปแบบ API error ต้องทำได้ด้วย Craft โดยไม่แก้ Go และไม่ rebuild ตัวภาษา** นี่เป็นเกณฑ์หลักในการตัดสินว่า Craft พร้อมสร้าง framework แล้ว

Rev.4 ส่งมอบทั้งพื้นฐานภาษาและ framework ตัวอย่างขนาดเล็กเพื่อพิสูจน์ว่า API ที่เพิ่มเพียงพอ ไม่ถือว่าการเปิดพอร์ตด้วย Go แล้วเรียก Craft handler ได้อย่างเดียวเป็นการส่งมอบครบ

คำว่า Craft สร้าง framework เองในแผนนี้หมายถึงเขียนและประกอบ framework ด้วยภาษา Craft ไม่ได้หมายถึงให้ภาษาออกแบบ framework อัตโนมัติ หรือย้าย compiler ทั้งหมดจาก Go มาเขียนด้วย Craft

## 2. Baseline และหลักฐานปัจจุบัน

### มีแล้ว

- Static types, named functions, Struct, Map, Array, Optional, Exception และ defer
- JSON values/parse/stringify, conversion, String utilities, filesystem/path และ environment
- Task/Timer, sleep, join/cancel/status และ deep-copy arguments
- CLI check/run/test/fmt/build, source bundles และ VS Code highlighting/themes/snippets

### ผลที่ผู้ใช้รายงานในบทสนทนา

- `craft version` แสดง 0.1.3 / Phase 1 Rev.3 / windows-amd64
- `craft check` สำเร็จ และ `craft run` ของตัวอย่าง Rev.3 จบด้วย `REV3 CHECKS COMPLETED`, exit 0
- ตัวอย่างครอบคลุมสอง task, cleanup, one-shot/repeating timer, pending cancellation, task exception และ zero interval rejection
- ผลนี้เป็น user-reported demo acceptance ไม่ใช่ผล Go tests, stress test, bundle regression หรือการตรวจรับ Rev.3 ทั้งฉบับ

### ช่องว่าง ณ baseline Rev.3 ที่ Rev.4 ต้องแก้

| ช่องว่าง | หลักฐานปัจจุบัน | ผลต่อ framework |
| --- | --- | --- |
| ไม่มี HTTP server primitive | stdlib ยังไม่มี network/server API | Craft เปิดพอร์ตรับ request ไม่ได้ |
| ไม่มี module/import/visibility | project loader รวม declarations จาก source เข้าด้วยกัน | แยก framework กับแอปและป้องกันชื่อชนกันไม่ได้อย่างเป็นระบบ |
| ไม่มี typed function values ทั่วไป | checker ปฏิเสธการเรียกผ่านตัวแปร; task callback เป็นกรณีพิเศษ | เก็บ handler ใน route table และประกอบ middleware ไม่ได้ครบ |
| task ownership จำกัด handles สะสม | จำกัด 256 active operations และ 256 owned handles ต่อ function invocation | server loop ที่ spawn ต่อ request ใน invocation เดิมจะติดเพดานแม้งานเก่าจบแล้ว |
| ไม่มี request context ใน Craft | cancellation ปัจจุบันอยู่ใน interpreter/Go context | ยังจัดการ deadline, client disconnect และ request-scoped cleanup ผ่าน Craft ไม่ได้ |
| deep copy เป็นกฎหลัก | Struct/Map/Array ถูก copy เมื่อส่งต่อ | ต้องกำหนด request state, route table และ native handle semantics ให้ชัด |

อ้างอิง source: [task runtime](../../internal/interpreter/tasks.go), [task callback checking](../../internal/types/tasks.go), [type checker](../../internal/types/checker.go), [project loader](../../internal/project/project.go), [value semantics](../../internal/runtime/value.go)

## 3. ขอบเขตระหว่าง Go และ Craft

| Go: ภาษา/runtime/เครื่องมือ | Craft: framework/application |
| --- | --- |
| Lexer/parser/type checker สำหรับ modules และ function types | Router, route registration, matching และ route groups |
| โหลด local dependencies และสร้าง bundle ที่เก็บ module identities | Handler/controller/service และ business logic |
| HTTP transport, connection lifecycle และ bridge เข้า interpreter | Middleware pipeline และลำดับการทำงาน |
| Request cancellation/deadline, concurrency limits และ shutdown | Validation และ mapping ของ application errors |
| Body/headers/URL primitives และ opaque resource handles | JSON response helpers, 404/405 และรูปแบบ error response |
| Primitive ด้านเวลา, I/O และ cryptography เมื่อมี use case ที่ยืนยัน | Authentication/authorization policy เมื่อมี primitives เพียงพอ |
| CLI สำหรับตรวจ รัน ทดสอบ และ package โปรแกรม Craft | Template/scaffolding logic ที่เขียนด้วย Craft ได้ในอนาคต |

- Go HTTP bridge ส่งทุก request ที่ transport รับได้ไปยัง Craft entry point เดียว จากนั้น Craft เป็นผู้เลือก route และ handler
- ไม่ใส่ route registry, controller discovery, middleware ordering หรือ validation framework ลงใน Go
- ไม่เพิ่ม Go built-in เฉพาะสำหรับ route หรือ controller ของแอปตัวอย่าง
- ใช้ HTTP implementation มาตรฐานเป็นฐาน เช่น Go `net/http`; ไม่เขียน HTTP wire parser ใหม่ใน Craft เพื่อให้ framework รุ่นแรกทำงาน
- การแก้ framework `.craft` อาจต้อง check/build bundle ใหม่ แต่ต้องไม่ต้อง build `craft.exe` ใหม่

## 4. Local modules และ dependencies

### ต้องส่งมอบ

- Module identity และ explicit import พร้อม alias เพื่อให้สอง module มีชื่อ function/type เหมือนกันได้
- Visibility: public exports กับ private declarations และ diagnostics เมื่อเข้าถึง private symbol
- Type identity อิง module + declaration ไม่เทียบเพียงชื่อ Struct สั้น ๆ
- Local path dependencies ผ่าน manifest มีชื่อและ version ที่ตรวจสอบได้; ยังไม่ต้องมี registry ออนไลน์
- กำหนด canonical path, duplicate imports, dependency cycles, missing dependency และ reserved `std` namespace
- Dependency graph ต้องถูก resolve แบบ deterministic พร้อมข้อจำกัดจำนวนไฟล์/ขนาด source ครอบคลุมทั้ง graph
- กำหนด source roots ของ dependency ชัดเจน ไม่ดึง tests, dist หรือไฟล์นอก package มา compile อัตโนมัติ
- รักษาโปรเจกต์ Rev.3 ที่ไม่มี import ให้ทำงานได้ตามกฎ legacy ที่บันทึกไว้
- Bundle ใหม่ต้องเก็บ module identities, dependency versions และ source ที่จำเป็นครบ รันได้หลังย้ายออกจากโฟลเดอร์ต้นทาง
- ตัดสิน bundle format/version ใหม่อย่างชัดเจน; runtime เก่าต้องปฏิเสธ bundle ที่ไม่รองรับด้วย error ที่อ่านเข้าใจ

### ขอบเขตเริ่มต้น

เริ่มจาก local packages กับ explicit imports ไม่เพิ่ม dynamic imports, remote download, top-level side effects หรือ package initialization order ที่ซับซ้อน รุ่นแรกปฏิเสธ import cycles พร้อมแสดงเส้นทางวงจร

Syntax ของ import/export และ manifest fields สรุปใน [Rev.4 contracts](../../docs/REV4-LANGUAGE.md); ตัวอย่างโครงสร้างด้านล่างไม่มีใน 0.1.3 แต่มี implementation สำหรับ0.1.4 แล้ว

```text
packages/
  api-framework/
    craft.toml
    src/
      router.craft
      middleware.craft
      response.craft
      validation.craft
examples/
  api-server/
    craft.toml
    src/
      main.craft
      handlers.craft
```

## 5. Typed function values และ handler composition

- เพิ่ม type ของฟังก์ชันที่ระบุ parameter types และ return type พร้อมตรวจตอน assign, pass, return และ call
- Named functions ต้องแปลงเป็น function values, เก็บในตัวแปร/Struct/Array/Map และเรียกผ่านค่าเหล่านั้นได้
- ฟังก์ชันรับหรือคืน function value ได้ เพื่อให้ไลบรารีเขียนการประกอบ handler ด้วย Craft
- เริ่มจาก non-capturing named functions; lambda และ lexical closures ยังไม่บังคับใน Rev.4
- Function values เป็น immutable references ไปยัง code ที่ตรวจชนิดแล้ว ไม่ deep copy ตัว code และไม่ serialize เป็น JSON
- Argument values ยังคงใช้กฎ copy เดิม ยกเว้น native handles ที่ระบุ contract แยก
- กำหนด shadowing, module-qualified function references, signature mismatch, nullability และการเรียก named arguments ผ่าน function values ให้แน่นอน
- ปรับ task/timer callback ให้ใช้ระบบ callable เดียวกัน โดยรักษา syntax Rev.3 และข้อจำกัด callback คืน Void ของ task/timer เดิม
- Parser, formatter, diagnostics และ highlighting ต้องรองรับ function-type syntax โดยไม่ทำให้โค้ดเก่าที่ถูกต้องเปลี่ยนความหมายโดยเงียบ

### การทำ middleware โดยไม่รอ closures

Framework รุ่นแรกสามารถเก็บรายการ named before/after hooks ใน Struct และวนเรียกด้วย Craft: before hook คืน context ที่ปรับแล้วหรือ response สำหรับหยุด pipeline ส่วน after hook รับ/คืน response ตามกติกาเดียวกันทั้งหมด

ต้องกำหนดลำดับ before/after, short-circuit และ error propagation พร้อมทดสอบด้วย middleware อย่างน้อยสองตัว ไม่ใช้ mutable captures เป็นทางลัด และไม่อ้างว่ารองรับรูปแบบ `next()` แบบ closure จนกว่าจะมี implementation

## 6. Runtime สำหรับ server ที่ทำงานต่อเนื่อง

### Task ownership และ resource reclamation

- แยกขีดจำกัดงานที่กำลังทำจากจำนวนงานที่เคยทำ ไม่ให้ server invocation ติดเพดานสะสมหลังรับ request 256 ครั้ง
- ออกแบบ server scope และ request scope; แต่ละ request มี environment/call stack แยก และงานลูกผูกกับ request เจ้าของ
- เมื่อ request จบ ให้ยกเลิก/รอ cleanup งานลูกและคืนข้อมูลหนักของงานที่จบแล้ว โดยไม่เก็บ request body หรือ stack ตลอดอายุ server
- Handle ที่ผู้ใช้ยังถืออยู่ต้องมี terminal status และผล join เดิม; กำหนดนโยบาย unobserved failure ก่อนถอดงานจาก owner registry
- หากใช้ worker pool ต้องมี queue ที่จำกัดขนาดและ cancellation ของงานที่ยังไม่เริ่ม ห้ามใช้ goroutine/queue แบบไม่จำกัด
- กำหนด ownership ของงานระดับ server เช่น timer housekeeping แยกจากงานระดับ request โดยยังไม่มี detached tasks ที่ไม่มีเจ้าของ
- ไม่เปลี่ยน lifetime ของ task/timer Rev.3 โดยเงียบ หากจำเป็นต้องเปลี่ยนต้องมี migration note และ regression

### Context, deadline และ shutdown

- เปิด context API ให้ Craft ตรวจ cancellation/deadline และส่งต่อ context ไปยัง I/O ที่รองรับได้
- แยก process shutdown, client disconnect, request timeout และ explicit cancellation เพื่อรายงานเหตุได้ถูกต้อง
- Context เป็น opaque handle ห้าม forge/serialize และการ copy ไม่สร้าง deadline หรือ request ใหม่
- Stop accepting → drain requests ภายใน deadline → cancel งานที่เหลือ → cleanup → close resources เป็นลำดับ shutdown ที่ต้องกำหนด
- Ctrl+C ต้องเข้ากระบวนการ shutdown และคืน exit convention ที่ประกาศไว้; failure ของ request เดียวต้องไม่ทำให้ server ทั้งตัวหยุด
- ต้องกำหนดเวลาสูงสุดของ drain/cleanup และข้อจำกัด I/O ที่ยกเลิกไม่ได้ ไม่รับรอง deadline ว่าฆ่า OS call ได้ทันที
- รักษาการตรวจ cancellation ใน loop ที่ใช้ CPU เพื่อให้ handler ที่วนไม่จบไม่กีดขวาง shutdown ตลอดไป

### State และประสิทธิภาพ

- Route configuration ต้องสร้างก่อนเริ่ม serve และ freeze ระหว่างรับ request; dynamic route mutation ยังไม่อยู่ในรุ่นแรก
- Request-local state ใช้ value semantics: helper/middleware คืนค่าที่เปลี่ยนแล้ว และ caller รับค่าคืนนั้น ไม่สมมติว่าแก้ Struct parameter แล้วผู้เรียกจะเปลี่ยนด้วย
- ห้ามแชร์ mutable Craft Map/Struct ระหว่าง request โดยไม่มี synchronization contract
- การใช้ immutable snapshot หรือ read-only shared representation ภายใน runtime ต้องรักษาพฤติกรรม value semantics และทดสอบ aliasing
- วัดต้นทุน copy ของ route table, headers และ body ก่อนเพิ่ม optimization; ไม่ใช้ผล demo เดิมอ้าง throughput ของ server

## 7. HTTP primitives และ Craft bridge

หัวข้อนี้ระบุข้อกำหนดต้นทาง; ชื่อ type/API ที่ implement แล้วดู [Rev.4 contracts](../../docs/REV4-LANGUAGE.md)

| Primitive | ขอบเขตที่ต้องรองรับ |
| --- | --- |
| Server configuration | bind address, port, timeouts, body/header limits, concurrency/queue limits |
| Request data | method, raw target/path, decoded path ตามกฎที่ระบุ, query และ headers หลายค่า |
| Request body | bounded read, byte data และ explicit UTF-8 decoding สำหรับ text/JSON |
| Response data | status, headers หลายค่า และ body; ตรวจความถูกต้องก่อนเขียน socket |
| Request context | cancellation/deadline และอายุการใช้ resource ที่ผูกกับ request |
| Server lifecycle | serve, stop/drain และการรายงาน bind/configuration errors |

### Contract ที่ต้องสรุปก่อน implementation

- เพิ่ม byte representation แบบมีขอบเขตและ immutable/value semantics ที่ชัดเจน เพื่อไม่บังคับ binary HTTP body เป็น UTF-8 String
- Headers ใช้แนวคิด `Map<String, String[]>` หรือ type ที่เก็บหลายค่าเทียบเท่า; lookup ไม่ไวต่อ case และต้องไม่รวม `Set-Cookie` ผิดกฎ
- Query parameters ต้องรักษาค่าซ้ำ/ค่าว่าง กำหนด percent decoding และ malformed encoding อย่างชัดเจน
- แยก raw กับ decoded path ป้องกันการ decode ซ้ำ; framework ใช้ contract นี้ในการ route matching
- ใช้ response แบบ buffered ก่อน: Craft handler คืน response value ให้ Go bridge เขียนหนึ่งครั้ง ไม่เปิด raw response writer ให้ส่งข้าม task โดยไม่มี ownership
- กำหนดอายุ request body/context/native handles; การใช้หลัง request จบต้องได้ error ไม่เข้าถึง resource ที่ปิดไปแล้ว
- จำกัด body/header size ก่อนจัดสรรไม่จำกัด และกำหนด read/write/header/idle timeouts
- ระบุวิธีจัดการ invalid status/header, header injection, malformed request, body ใหญ่เกิน และ handler exception
- Application error mapping อยู่ใน Craft; bridge มี fallback ที่จำเป็นเมื่อ Craft dispatch ล้มเหลวจนสร้าง response ไม่ได้ และไม่ส่ง internal stack trace ให้ client โดยอัตโนมัติ
- เมื่อ response ส่งแล้วหรือ client disconnect ห้ามพยายามเขียน error response ซ้ำ; log เหตุโดยแยกจาก response
- Client cancellation ต้องส่งถึง Craft invocation และ task ลูก ไม่ใช้ context ของ process แทน context ของทุก request
- เริ่มจาก HTTP/1.1 และ buffered text/JSON endpoints; กำหนดขอบเขต TLS ว่าใช้ transport option หรือ reverse proxy ก่อนใช้งานภายนอกเครื่อง
- Authentication, cookie/session policy, CORS และ trusted proxy configuration ต้องระบุว่าอะไรยังไม่รองรับ ไม่เปิดใช้โดยเดาค่าที่ผู้ใช้ต้องการ

ไม่บังคับ streaming, multipart uploads, WebSocket, SSE, HTTP/2 หรือ HTTP/3 ในรอบนี้

## 8. Framework ตัวอย่างที่ต้องเขียนด้วย Craft

### Minimum framework

- Route registration และ matching ตาม method/path พร้อม static route และ path parameter
- Route precedence และ duplicate/ambiguous route diagnostics ที่กำหนดชัดเจน
- Query/path parameter helpers, JSON request parsing และ JSON response construction
- Middleware before/after อย่างน้อยสองตัว รองรับ short-circuit และลำดับ error handling
- Validation helpers สำหรับ required fields, type และช่วงค่า พร้อม field-level errors
- Error response format สม่ำเสมอ และ 404/405 รวม Allow header ตาม method ที่ลงทะเบียน
- Framework dispatch เรียกได้ด้วย request value จำลองเพื่อเขียน unit tests โดยไม่เปิดพอร์ต
- เพิ่ม routes และ middleware ในโปรเจกต์แอปผ่าน imports ไม่แก้ source ของ framework หรือ Go runtime

### แอปพิสูจน์ความพร้อม

- `GET /health`: แสดงว่า server ทำงานและคืน JSON
- `GET /items/:id`: ทดสอบ path/query parsing และ not-found response
- `POST /items`: รับ JSON, validate และคืน response ที่กำหนด โดยไม่บังคับ database
- Endpoint สำหรับทดสอบ handler error และ slow/cancelled request อยู่เฉพาะชุดทดสอบ
- middleware ตัวอย่างสำหรับ request metadata/logging และ validation/short-circuit
- ตัวอย่างเป็น stateless หรือใช้ fixture คงที่ก่อน ไม่สร้าง in-memory shared mutable database เพื่อหลีกเลี่ยง state contract ที่ยังไม่พร้อม

เกณฑ์สำคัญ: ผู้ใช้เพิ่ม endpoint และ middleware ใหม่ด้วย Craft แล้ว check/run/build ได้ด้วย CLI เดิม ไม่ต้องเพิ่ม built-in เฉพาะใน Go

## 9. เครื่องมือ เอกสาร และ compatibility

- CLI check ต้องตรวจ module graph และ function signatures โดยไม่เปิด server หรือรัน initializers
- CLI test แยก framework unit tests ออกจาก network integration tests และไม่เปิด listener โดยแฝง
- CLI build เก็บ local dependencies ใน bundle ตาม format ใหม่ที่ตกลงแล้ว พร้อม source positions สำหรับ diagnostics
- Diagnostics ของ native bridge ต้องชี้กลับตำแหน่ง Craft ที่เรียกได้ ไม่แสดงเฉพาะ Go stack
- เพิ่ม fixture/snippets/highlighting สำหรับ syntax ใหม่เมื่อ parser/type checker รองรับจริง ระบุ version ขั้นต่ำ
- จัดทำคู่มือ library author, framework user, migration Rev.3 → Rev.4 และตัวอย่าง deployment ที่ตรงกับ transport ที่ส่งมอบ
- บันทึกสถานะ source-reviewed, built, user-tested, pending และ deferred แยกกันใน `output/phase1-rev4-status.md` เมื่อเริ่ม implementation
- Scaffold ขั้นต่ำเริ่มจากตัวอย่างโปรเจกต์ที่ copy ได้; generator ที่เขียนด้วย Craft เป็นงานต่อยอดหลัง module และ filesystem APIs เพียงพอ

## 10. ลำดับส่งมอบ

| Milestone | สิ่งที่ส่งมอบ | ประตูผ่านไปขั้นถัดไป |
| --- | --- | --- |
| M0 — Contracts/baseline | Module/function syntax, HTTP value model, context/lifecycle, compatibility decisions | ไม่มีข้อกำหนดหลักขัดกัน; บันทึกผล Rev.3 ที่มีจริงและรายการ pending |
| M1 — Language libraries | Local modules/imports/visibility และ typed function values | Craft package เก็บและเรียก handlers จากอีก package ได้ โดยยังไม่ต้องเปิด HTTP |
| M2 — Server runtime | Request scopes, resource reclamation, context/deadline, bounded concurrency | งานต่อเนื่องไม่ติด cumulative task limit; cancellation และ cleanup มี tests เตรียมครบ |
| M3 — HTTP bridge | Transport และ request/response primitives | Named Craft dispatch รับ HTTP และคืน response ได้ พร้อม limits/shutdown |
| M4 — Craft framework | Router/middleware/validation/errors เขียนด้วย Craft | แอปเพิ่ม route/middleware โดยไม่แก้ Go หรือ framework core |
| M5 — Handoff/acceptance | Bundles, examples, docs, test commands และ build artifacts | ผู้ใช้รันตรวจรับและบันทึกหลักฐานตามหัวข้อ 11 |

M1 สามารถเริ่มโดยไม่ต้องรอ LSP หรือ visual acceptance ทุกข้อของ editor Rev.3 แต่ก่อน release Rev.4 ต้องมี regression ของความสามารถ Rev.3 ที่ runtime ใหม่กระทบ

## 11. แผนทดสอบและเกณฑ์พร้อมสร้าง framework

ผู้ใช้เป็นผู้รันทดสอบเองตาม working preferences: Agent เตรียม tests/fixtures/คำสั่ง ตรวจ source และ build ได้ แต่ไม่รัน automated tests หรือเปิด app/browser ทดสอบโดยไม่ได้รับคำสั่งเพิ่มเติม ผลที่คาดหวังต้องแยกจากผลที่เกิดขึ้นจริงเสมอ

### Language/package acceptance

ผลทดลอง Rev.4 ที่ผู้ใช้ยืนยันเมื่อ 19 กันยายน 2026 ใน `D:\02-Repository\craft-test`:

- [x] `craft version` แสดง Craft 0.1.4 / Phase 1 Rev.4 / windows-amd64
- [x] `craft check` ครั้งล่าสุดสำเร็จ
- [x] `craft test` ผ่าน 7/7, ไม่ผ่าน 0, exit code 0 ครอบคลุม function values/collections, Bytes/UTF-8, imported framework/query values, router errors, task ownership เกิน 256 completed tasks, timer function value และ basic arithmetic
- [x] `craft run` แสดง `HTTP listening on 127.0.0.1:8080`

บันทึกผลและข้อจำกัดของหลักฐานใน [สถานะ Rev.4](output/phase1-rev4-status.md#user-reported-rev4-results--19-september-2026) รายการด้านล่างเป็นเกณฑ์ตรวจรับที่กว้างกว่าชุดทดลองนี้ จึงยังรอผลเฉพาะด้านก่อนทำเครื่องหมายครบ:

- [ ] แยก framework กับแอปเป็น local packages ที่ import กันได้
- [ ] Alias/name collisions/private access/import cycle/missing dependency มี diagnostics
- [ ] Handler signatures ตรวจตอน compile; เก็บและเรียก function values ใน collections/Struct ได้
- [ ] Function/module syntax ผ่าน formatter และ source bundle รักษาความหมาย
- [ ] Rev.3 source และ bundles เดิมรองรับตาม compatibility contract

### Runtime/HTTP acceptance

- [ ] รับอย่างน้อย 10,000 requests ต่อเนื่องใน server process เดียวโดยไม่ติดเพดานสะสม 256 handles
- [ ] ทดสอบ active concurrency limit และ overload แยกจากจำนวน request สะสม; queue และหน่วยความจำไม่เติบโตตามจำนวน request ทั้งหมดหลังจบงาน
- [ ] ระบุเครื่อง/config/payload และวิธีวัด memory/active tasks/open resources ก่อน-หลัง warmup และหลัง drain; ไม่กำหนดตัวเลข throughput โดยไม่มีผลวัด
- [ ] Request หนึ่ง timeout/disconnect แล้ว request อื่นยังทำงาน และ task ลูกของ request ที่จบไม่ค้าง
- [ ] Ctrl+C/shutdown หยุดรับงานใหม่ drain/cancel ตาม deadline และไม่ทิ้ง listener/request tasks
- [ ] Handler exception ไม่ทำให้ server หยุด และไม่มี stack trace ภายในรั่วไป response โดยไม่ตั้งใจ
- [ ] Headers/query หลายค่า, malformed encoding, binary/UTF-8, body limits และ invalid response ถูกจัดการตาม contract
- [ ] ตัวแปร/collections ระหว่าง request ไม่ปะปน และ native handles ใช้หลังหมดอายุไม่ได้
- [ ] Network tests ใช้ local ephemeral ports, มี bounded timeouts และคืน resources แม้ test ล้มเหลว
- [ ] เตรียม fake clock/barriers สำหรับ lifecycle tests และ race checks ที่ผู้ใช้รันบน environment ที่รองรับ

### Framework acceptance

- [ ] Router, middleware และ validation มี implementation และ unit tests เป็น Craft
- [ ] Middleware order, short-circuit, 404/405, JSON errors และ route precedence ตรง contract
- [ ] เพิ่ม route/middleware ใหม่ผ่าน application `.craft` โดยไม่แก้ Go และไม่ rebuild CLI
- [ ] Build bundle แล้วนำไปรันนอก source tree ได้ครบ dependencies
- [ ] ผู้ใช้ตรวจรับ API example และ regression Rev.2/Rev.3 พร้อม exit codes จริง

Rev.4 พร้อมเป็นฐานพัฒนา framework เมื่อผ่าน language/runtime/framework acceptance ที่เกี่ยวข้องทั้งหมด การ build สำเร็จหรือ endpoint `/health` ตอบได้เพียงอย่างเดียวยังไม่พอ และไม่เท่ากับการรับรอง production ทุก workload

## 12. งานที่เลื่อนได้

- Online package registry, remote dependency downloads และ package publishing
- Lambda/closures, generics, interfaces, reflection, annotations/decorators และ automatic Struct JSON mapping
- ORM/database migrations, DI container, OpenAPI generation และ full authentication/session stack
- Shared mutable caches, dynamic route updates และ general channel/mutex API หาก minimum framework ยังไม่ต้องใช้
- Streaming/WebSocket/SSE/multipart และ protocol extensions
- LSP/debugger, VM/native backend, hot reload และการย้าย compiler ไปเขียนด้วย Craft

สิ่งเหล่านี้ไม่ใช่เงื่อนไขเริ่ม API framework รุ่นแรก หากภายหลังมี use case บังคับใช้ต้องเพิ่ม contract และขอบเขตตรวจรับก่อนขยายงาน

อ้างอิง: [Rev.3 roadmap](craft-phase1-rev3.md), [Rev.3 status](output/phase1-rev3-status.md), [Rev.3 runtime contracts](../../docs/REV3-LANGUAGE.md), [Rev.2 language contracts](../../docs/REV2-LANGUAGE.md)
