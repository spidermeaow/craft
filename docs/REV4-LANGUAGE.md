# Craft 0.1.4 — Phase 1 Rev.4 contracts

สถานะ: implementation และ build สำหรับให้ผู้ใช้ตรวจรับ ไม่ใช่ผลรับรอง runtime/production ดู [สถานะ](../roadmap/phase1/output/phase1-rev4-status.md) และ [คำสั่งทดสอบ](TRY-REV4.md)

## Modules และ library authors

หนึ่ง `craft.toml` คือหนึ่ง module ใช้ `name@version` เป็น identity; Struct ต่าง module เป็นคนละ type แม้ชื่อและ fields ตรงกัน ตัวอย่าง manifest แอป:

```toml
name = "api-server"
version = "0.1.0"
edition = "2026"
entry = "main"
dependency.api = "../../packages/api-framework"
dependency-version.api = "0.1.0"
```

Package ที่ไม่มี main ใช้ `entry = "library"`; `craft check/test/build` รองรับ ส่วน `craft run` ปฏิเสธ library. Manifest รองรับ scalar strings และ dotted dependency keys ตามตัวอย่าง ยังไม่รองรับ TOML tables, version ranges หรือ remote registry. Path อิง directory ของ manifest ต้องชี้ package root และ version ตรงทุกตัวอักษร

```craft
import api "api"

func hello(request: api.Request): HttpResponse {
    return api.text(200, "hello")
}

func main() {
    let app: api.App = api.add(api.create(), "GET", "/hello", hello)
    api.serve(std.http.config("127.0.0.1:8080"), app)
}
```

`import alias "dependency-key"` มี scope ต่อไฟล์ ต้อง import ในแต่ละไฟล์ที่ใช้ alias; declarations ภายใน module มองเห็นกันทุก source file. `export func` และ `export struct` เปิดให้ module อื่นใช้; ที่ไม่ export เป็น private. Fields ของ exported struct ใช้ mutability ตามประกาศ ไม่มี field-private. Exported function คืน private type ได้เพื่อส่งต่อค่า แต่ client ตั้งชื่อ/สร้าง private type โดยตรงไม่ได้

ห้าม alias `std`, alias ซ้ำ, import module เดิมซ้ำในไฟล์เดียว, alias ชน declaration หรือ local binding/parameter. Aliases ใช้ ASCII identifier; declaration identifiers ยังคงรองรับ Unicode. Builtin names สงวนไว้เหมือนกันในทุก module. ไม่มี re-export, wildcard import, top-level initialization หรือ dynamic loading

โหลดเฉพาะ `.craft` ใต้ `src/`; `check/test` เพิ่ม `tests/` ของ root เท่านั้น ไม่ดึง tests ของ dependencies. ไม่ตาม source symlinks และข้าม nested dist/build/node_modules/.git. Resolve canonical paths, ปฏิเสธ identity เดียวหลาย paths และ cycles พร้อม dependency chain. Graph จำกัด 64 modules, 1024 source files, รวม 16 MiB, ไฟล์ละ 2 MiB; traversal เรียง aliases และ modules

## Function values

`Fn<Return, Parameter1, Parameter2, ...>` ระบุ return type ก่อน เช่น `Fn<Void>` และ `Fn<HttpResponse, HttpRequest, Int>`. เริ่มจาก named functions เท่านั้น ไม่มี closures/lambda/captures. Builtins ไม่แปลงเป็น function values โดยตรง ให้สร้าง named Craft wrapper

```craft
func increment(n: Int): Int { return n + 1 }
func choose(): Fn<Int, Int> { return increment }
func main() {
    let callback: Fn<Int, Int> = choose()
    let handlers: Fn<Int, Int>[] = [callback]
    print(handlers[0](41))
}
```

ส่ง/คืน/เก็บใน Struct, Array, Map และ Optional ได้ ต้อง unwrap Optional ก่อน call. Signature ตรงทุก type ไม่มี variance หรือ implicit coercion. เรียกผ่านค่าใช้ positional arguments; direct named function call ยังใช้ named arguments เดิมได้. Local variable ที่ชื่อชน function จะบัง function; ถ้า local ไม่ callable จะเป็น type error. Function references immutable และไม่ serialize เป็น JSON; arguments ยัง deep copy ตาม value semantics

Task/timer ใช้ callable เดียวกัน รับ function value ที่คืน Void และ arguments ตรง signature. ไม่เพิ่ม detached tasks. Limit เป็น 256 active operations ต่อ interpreter run และ 256 active/unobserved failed children ต่อ function frame; terminal successful/cancelled/observed children ถูกถอดจาก owner registry เมื่อจะชน limit. Handle ที่ยังถืออยู่คง status/join result; unobserved failure ไม่ถูกทิ้ง. Request/task จบจะ cancel/join งานลูกก่อนคืน scope ตาม Rev.3

## Bytes และ context

| API | Contract |
| --- | --- |
| `std.bytes.fromString(String): Bytes` | immutable bytes จาก String สูงสุด 16 MiB |
| `std.bytes.fromArray(Int[]): Bytes` | แต่ละค่าต้อง 0..255 สูงสุด 16 MiB |
| `bytes.length(): Int` | จำนวน bytes |
| `bytes.text(): String` | ตรวจ UTF-8; throw เมื่อ invalid |
| `bytes.toArray(): Int[]` | คืน copy ของ byte values |
| `std.context.current(): Context` | execution context ปัจจุบัน รวม child-task cancellation |
| `context.cancelled(): Bool` | มี cancellation หรือไม่ |
| `context.reason(): String` | ว่างเมื่อยังทำงาน; otherwise cause text |
| `context.remainingMilliseconds(): Int?` | deadline ที่เหลือ clamp ที่ 0; null ถ้าไม่มี |

Bytes share immutable backing memory ได้. Context เป็น opaque reference ไม่มี constructor และ JSON conversion; copy อ้าง context เดิม. Request context และ current-context handles ที่ derive จาก request ใช้หลัง handler/cleanup จบไม่ได้ (throw expired). Synthetic `std.http.request` ใช้ context ของ execution ที่เรียก ไม่ถือ server resource

Request timeout cause `request timeout`; forced shutdown `server shutdown`; task cancellation/client disconnect `context canceled`. Root process Ctrl+C ส่ง cancellation เข้า server lifecycle. Cancellation ถูกตรวจตาม statement/loop/native wait; context API จึงมีประโยชน์ใน cleanup หรือส่ง handle ต่อ แต่ไม่ใช้หลบ cancellation เพื่อให้ loop ทำต่อไม่จบ

## HTTP native bridge

`std.http.serve(config, dispatch, state)` ต้องรับ `Fn<HttpResponse, HttpRequest, State>` และ state type ตรงกัน. Serve block จนหยุด; bind/configuration errors เป็น Craft exceptions. ห้ามเริ่ม server จาก request/task ของ request หรือ defer. Go ส่ง request ไป entry เดียวและไม่รู้จัก routes/middleware

State freeze โดย snapshot เมื่อเริ่ม serve แล้ว copy แยกต่อ request. ห้าม Task/Timer/Context ใน state แม้ซ้อน collection; ไม่มี mutable state ร่วมข้าม request. Request machine/call stack แยกกัน; AST/function declarations share แบบอ่านอย่างเดียว. Console input ใน request ไม่มี interactive reader. งานที่ spawn ภายใน request ผูกกับ frame/request; งานใน outer main frame ก่อน serve อยู่ใน server lifetime

| Type | Fields (mutable ตาม value semantics) |
| --- | --- |
| `HttpRequest` | `method`, `rawTarget`, `path`: String; `query`, `headers`: Map<String,String[]>; `body`: Bytes; `context`: Context |
| `HttpResponse` | `status`: Int; `headers`: Map<String,String[]>; `body`: Bytes |
| `HttpConfig` | `address`: String; ตัวเลือกด้านล่างทั้งหมด Int |

`std.http.config(address)` สร้าง config. Defaults/ranges:

| Option | Default | Allowed |
| --- | ---: | --- |
| maxConcurrent | 64 | 1..256 |
| maxConnections | 128 | 1..1024 |
| maxBodyBytes | 1048576 | 1..16777216 |
| maxHeaderBytes | 32768 | 1024..1048576 |
| readTimeoutMs | 10000 | 1..300000 |
| writeTimeoutMs | 15000 | 1..300000, >= requestTimeoutMs |
| idleTimeoutMs | 30000 | 1..300000 |
| requestTimeoutMs | 10000 | 1..300000 |
| shutdownTimeoutMs | 5000 | 1..300000 |

Transport เป็น plain HTTP ผ่าน Go net/http; ไม่มี TLS configuration, WebSocket, SSE, streaming หรือ multipart helper. ตัวอย่าง bind loopback; deployment ใช้ reverse proxy สำหรับ TLS และไม่ trust forwarded headers อัตโนมัติ. Header reading ใช้ net/http MaxHeaderBytes semantics (รวม internal read allowance ของ transport), ไม่อ้างเป็น byte-exact wire cap

Concurrency เต็มตอบ 503 โดยไม่มี application queue. Connection cap บังคับก่อนรับ connection เพิ่ม; OS backlog อยู่ภายนอก runtime. Body อ่านแบบ bounded ก่อนเข้า Craft; เกิน limit 413, malformed body400, read timeout408. Request deadline รวม body read; handler timeout504 เมื่อยังเขียน socket ได้. Write timeout/disconnect อาจทำให้ client ไม่ได้รับ error response

Path decode ครั้งเดียว; ปฏิเสธ encoded slash/backslash, literal backslash และ NUL. ไม่ normalize case/trailing slash/dot segments. Query decoding แยกจาก path รักษาค่าซ้ำ/ค่าว่าง และ malformed query400. Incoming headers เก็บ lowercase keys; lookup ใช้ lowercase, ค่าซ้ำเป็น array. `Host` เป็น transport metadata ของ net/http และไม่ได้อยู่ใน headers map; ไม่ใช้ headers map เป็น complete raw wire representation. `rawTarget` เก็บ RequestURI

`std.http.response(status, bytes)` สร้าง response headers ว่าง. ยอม status200..599; body<=16MiB; header names ต้องเป็น HTTP tokens, values ห้าม CR/LF/control, รวม header names/values<=64KiB. Content-Length, Transfer-Encoding, Connection, Trailer, Upgrade เป็น transport-owned ห้ามตั้งเอง. Preserve repeated headers เช่น Set-Cookie. HEAD/204/304 ไม่ส่ง body. Invalid response หรือ uncaught handler exception ตอบ500 generic และ log detail เฉพาะ server output

`std.http.request(method,target,bytes)` สร้าง synthetic request สำหรับ pure Craft tests ไม่เปิด listener; query/path decoding เหมือน bridge, headers ว่างแก้ผ่าน var ได้. ไม่จำลอง HTTP wire parser/limits

`std.http.stop(request.context)` ขอหยุด server แบบ nonblocking; handle ต้องยังไม่ expired และมาจาก server. ไม่มี stop HTTP route ติดตั้งอัตโนมัติ. Lifecycle: close admission/listener → drain ภายใน shutdownTimeoutMs → cancel contexts/close connections ที่เหลือ → รอ structured task cleanup. Explicit stop คืนปกติ exit0; Ctrl+C คืน cancellation ตาม CLI เดิม exit1. Blocking OS filesystem calls อาจทำให้ cleanup เกิน drain deadline; ไม่รับรอง hard kill. Rev.3 defer cleanup ยังมี statement/deadline budget เดิม

## Craft framework และ compatibility

ดู [API framework](../packages/api-framework/README.md). Router/middleware/validation อยู่ `.craft` ทั้งหมดและเปลี่ยนได้โดยไม่ build CLI. Route table ถูกกำหนดก่อน serve; callback คืนค่าที่ปรับแล้วเสมอ. การ copy route table ต่อ request เป็นต้นทุนที่ยอมรับในรุ่นนี้ ยังไม่มี throughput/memory benchmark ที่รันจริง

Rev.3 project ที่ไม่มี imports ใช้ entry main เดิมและ declarations ภายใน root ไม่ต้อง export. ชื่อ type ใหม่ `Fn`, `Bytes`, `Context`, `HttpRequest`, `HttpResponse`, `HttpConfig` ถูกสงวน; project เดิมที่ใช้ชื่อเหล่านี้ต้อง rename. `import/export` เป็น contextual top-level syntax, local identifiers ชื่อเดิมยังใช้ได้. Resource reclamation เปลี่ยน cumulative owner limit ให้ใช้ต่อเนื่องได้ แต่ active limit/ownership/error propagation เดิมยังอยู่

Build สร้าง source bundle format3/language0.1.4 พร้อม module identities, graph และ dependency sources; runtime ไม่ resolve local paths ตอนรัน bundle. Root sources มีสำเนาใน compatibility field แต่ module list เป็น authoritative. Runtime0.1.4 รับ format2/language0.1.1–0.1.3 ด้วย; format1 ต้อง migrate/rebuild. Runtime เก่าปฏิเสธ format3. Bundle เป็น source ไม่ใช่ compiled native executable และไม่เก็บ credentials/config จาก environment. `craft fmt` รักษา AST/import/export/function types และ comments
