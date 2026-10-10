# แผน `packages/api` — เครื่องมือสร้าง HTTP API ด้วย Craft

**สถานะ:** ลง source ใน Craft 0.1.44 candidate; รอผู้ใช้ตรวจรับ  
**ฐาน:** Craft 0.1.43 candidate และ `std.http` ที่มีอยู่  
**รุ่นเป้าหมาย:** Craft 0.1.44; ดู [ผลการลงงาน](phase1/craft-phase1-rev44-api-package.md)

## เป้าหมายและขอบเขต

สร้าง `packages/api` เป็น **first-party Craft package** สำหรับประกอบ HTTP API โดยไม่ต้องเขียนการจับคู่ route, รูปแบบ response และ validation ซ้ำในทุกแอป. Package อยู่เหนือ `std.http`: Runtime ยังคงดูแล listener, HTTP transport, timeout, cancellation, concurrency และขีดจำกัดทรัพยากร; package เขียนด้วย `.craft` และไม่เพิ่ม built-in API ใน `craft.exe` เพื่อรองรับ framework เฉพาะทาง.

แอปที่ต้องการควบคุมทุกอย่างเองยังเรียก `std.http` โดยตรงได้. `packages/api` ไม่ผูกกับ `std.db`; handler เลือกเรียกฐานข้อมูลหรือบริการอื่นเอง.

## รูปแบบ package

- วาง source ที่ `packages/api/` โดย `entry = "library"`, มี Public API, ตัวอย่าง และ tests ของ package เอง. ใช้ `import package api "api"` ในไฟล์ที่เรียก API จาก package.
- ในฐานะ package ที่มากับ Craft ให้ใช้ **เวอร์ชันตรงกับ toolchain** ตามแนวทาง `packages/gui`; ผู้ใช้ระบุ path dependency และรัน `craft package lock` โดยไม่ต้องเลือกเวอร์ชัน API แยก. การเผยแพร่ package อิสระในอนาคตให้ตัดสินใจเมื่อมีนโยบาย compatibility แล้ว.
- ใช้ [framework ต้นแบบที่เก็บใน Archive](../Archive/packages/api-framework/README.md) เป็นข้อมูลออกแบบและชุดพฤติกรรมอ้างอิง แต่ตรวจ API กับภาษา Craft ปัจจุบันก่อนนำโค้ดกลับมา. ไม่คืน `api-framework` เก่าเป็น dependency โดยตรง.

## ความสามารถรอบแรก

| กลุ่ม | สิ่งที่จะทำใน `packages/api` | สัญญาที่ต้องกำหนด |
| --- | --- | --- |
| App และ router | สร้าง App; ลงทะเบียน GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS; path parameter เช่น `/items/:id`; route group/prefix | ตรวจ method/path ตั้งแต่ลงทะเบียน; route ซ้ำหรือกำกวมต้อง error; กำหนดลำดับ static กับ parameter, trailing slash และ case sensitivity |
| Dispatcher | แปลง `HttpRequest` เป็น request ของ package, เลือก handler แล้วคืน `HttpResponse` | 404 เมื่อไม่มี path, 405 พร้อม `Allow` เมื่อ method ไม่ตรง; ไม่มี route พิเศษติดตั้งอัตโนมัติ |
| Middleware | before/after hooks ที่รับและคืนค่าอย่างชัดเจน | ลำดับการทำงาน, short circuit, error และ after-hook เมื่อ 404/405 ต้องแน่นอน; ไม่มี shared mutable state ข้าม request |
| Response helpers | JSON, text และ error response | ตั้ง Content-Type ถูกต้อง; JSON error มีรูปแบบคงที่; ไม่กลบ validation หรือ transport error |
| Request helpers | อ่าน query ซ้ำ, path params, JSON body | แยก parse error กับ validation error; ไม่ parse JSON หรือแปลง body โดยปริยายทุก request |
| Validation ขั้นต้น | required string, integer range และ response สำหรับ field errors | คืนข้อผิดพลาดที่ระบุ field ได้; แอปเป็นผู้ตัดสินใจว่าจะตอบ 400 หรือ 422 |
| Lifecycle | `api.serve(config, app)` ส่งต่อ `std.http.serve` | เคารพ config, request context และ graceful shutdown ของ Runtime; ไม่มี listener อีกชุดใน package |

ตัวอย่าง Public API ที่ลง source และตรวจด้วย `craft check`:

```craft
import package api "api"

func health(request: api.Request): HttpResponse {
    return api.json(200, std.json.fromObject({"status": std.json.fromString("ok")}))
}

func main() {
    var app: api.App = api.create()
    app = api.add(app, "GET", "/health", health)
    api.serve(std.http.config("127.0.0.1:8080"), app)
}
```

## งานที่ต้องตรวจและตัดสินใจก่อนลง API

1. ทดสอบ `Fn<...>` ของ handler/middleware, การส่ง function ผ่าน package, module imports และการ clone App/state ต่อ request กับ Craft 0.1.43 จริง. ถ้าโครงสร้างเดิมจาก Archive ใช้ไม่ได้ ให้ปรับ **Public API ของ package** ก่อนพิจารณาแก้ Language Core; บันทึกเหตุผลและ migration.
2. กำหนดว่า App เป็นค่าที่คืนใหม่จาก `add`/`useBefore`/`useAfter` ทุกครั้ง. ต้องไม่แอบแก้ App เดิม. Route table ถูก freeze เมื่อ `serve` เริ่ม และการลงทะเบียนหลังเริ่ม serve ต้องไม่เกิดขึ้น.
3. กำหนด policy ของ URL/path parameter: ใช้ path ที่ transport decode แล้วครั้งเดียว, ปฏิเสธ parameter ชื่อซ้ำ/ว่าง, ไม่ decode อีกชั้น, ไม่ทำ case folding โดยปริยาย. ให้พฤติกรรมสอดคล้องกับข้อห้าม encoded slash/backslash ของ `std.http`.
4. กำหนด error boundary: ข้อผิดพลาดใน handler/middleware ไม่คืน stack trace หรือข้อมูลลับให้ client; transport errors ที่เกิดก่อน dispatcher อยู่ภายใต้ `std.http` และไม่ผ่าน middleware ของ package.
5. แยกความหมายของ `api.stop` (ถ้าจำเป็น) จาก `std.http.stop(request.context)`. รอบแรกใช้ `std.http.stop` โดยตรงก่อนเพื่อไม่ซ่อน lifecycle.

## ลำดับพัฒนา

1. ย้ายแนวคิด App/Request/Route จาก Archive มาเป็น `packages/api` ใหม่; ลด API ให้เหลือ router, dispatcher และ `serve` ก่อน.
2. เพิ่ม response/query/JSON helpers พร้อม tests ที่ไม่เปิดพอร์ต.
3. เพิ่ม before/after middleware และ validation โดยรักษา value semantics ของ Craft.
4. ทำตัวอย่าง API จริงที่ `examples/api-server` โดยให้มีทั้ง endpoint ผ่าน `api` และตัวอย่าง `std.http` ตรงสำหรับเทียบพฤติกรรม; อย่าทำให้ตัวอย่างเดิมหาย.
5. ตรวจ `craft check`, `craft test`, `craft fmt --check`, `craft run` และ `craft build` ของ package/ตัวอย่าง; ตรวจ request จริงบน localhost, concurrent requests, timeout/cancellation และ graceful shutdown. อัปเดต `craft.lock`, README และ installer หลัง API ผ่านเกณฑ์.

## เกณฑ์รับงาน

- แอปสร้าง REST API พื้นฐานได้โดย import `api` เพียง package เดียวสำหรับ routing/response helpers; ไม่มีการแก้ private Runtime เพื่อให้ route เฉพาะแอปทำงาน.
- ทดสอบ route static/parameter, route ซ้ำ/กำกวม, 404/405, middleware order/short circuit, JSON ที่ผิดรูปแบบ, validation และ header ที่ซ้ำ; ผลเหมือนกันระหว่าง dispatcher test กับ server จริงในส่วนที่ transport ส่งต่อให้ Craft.
- ยืนยันว่า body/response limits, request timeout, cancellation, concurrency และ shutdown ยังเป็นสัญญา `std.http` เดิม; package ไม่อ้างว่าขยายความสามารถ transport.
- ตัวอย่างที่ติดตั้งไปพร้อม Craft ตรวจผ่านด้วย toolchain และ lockfile รุ่นเดียวกัน; ผู้ใช้ทดลองใช้งานจริงก่อนประกาศว่า package พร้อมใช้.

## ไม่รวมในรอบแรก

HTTP client, TLS ใน Craft Server, WebSocket, SSE/streaming, multipart upload, automatic authentication/authorization, ORM, dependency injection container, OpenAPI generator และ application deployment. งานเหล่านี้ต้องมีเกณฑ์ใช้งานและข้อจำกัดของ Runtime ชัดก่อนเพิ่ม. HTTPS สำหรับ deployment ปัจจุบันใช้ reverse proxy ตาม [คู่มือ HTTP Engine](../docs/REV4-LANGUAGE.md).
