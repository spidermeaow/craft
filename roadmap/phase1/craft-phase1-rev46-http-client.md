# Craft Phase 1 Rev.46 (0.1.46) — HTTP Client

**สถานะ:** ผู้ใช้ตรวจรับให้ผ่านแล้ว (10 ตุลาคม 2026)
**ฐาน:** Craft 0.1.45 candidate, `std.http` Server และ `packages/api`
**ขอบเขต:** เพิ่ม HTTP client พื้นฐานใน Standard Library สำหรับแอป CLI, Desktop และ API; ไม่เปลี่ยนสัญญา Server เดิม

## เป้าหมาย

ให้โปรแกรม Craft เรียก HTTP/HTTPS API ภายนอกได้ผ่าน Public API ที่ตรวจชนิดและควบคุมเวลา/ขนาดข้อมูลได้ โดยไม่ต้องเรียก `curl`, PowerShell หรือโปรแกรมอื่น. Transport อยู่ใน `std.http` ของ Runtime; โค้ดแอปเป็นเจ้าของ URL, headers, credentials และการตัดสินใจว่าจะใช้ผลลัพธ์อย่างไร. `packages/api` ยังคงเป็นเครื่องมือสร้าง HTTP Server ไม่ถูกบังคับให้ import เพื่อใช้ client.

## API ที่เสนอ

ชื่อและชนิดสุดท้ายต้องยืนยันกับภาษาและ type checker เมื่อ implement; ตัวอย่างนี้เป็นสัญญาที่ต้องการ ไม่ใช่ API ที่ใช้ได้แล้ว.

| ส่วน | API / ข้อมูล | สัญญา |
| --- | --- | --- |
| Request | `HttpClientRequest { method, url, headers, body }` และ `std.http.clientRequest(method, url, body)` | headers เป็น `Map<String, String[]>`; body เป็น `Bytes`; ผู้เรียกเติม headers ก่อนส่ง. ไม่ผูกกับ `HttpRequest` ของ Server เพราะ `path`, `query` และ `context` มีความหมายต่างกัน |
| Config | `HttpClientConfig` จาก `std.http.clientConfig()` | กำหนด `timeoutMs`, `maxResponseBytes`, `maxResponseHeaderBytes` และขีดจำกัดที่จำเป็น; ค่าเริ่มต้นปลอดภัยและตรวจช่วงก่อนเปิด connection |
| ส่งคำขอ | `std.http.send(request, config): HttpClientResponse` | รองรับ `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`; HTTP status 4xx/5xx เป็น response ปกติ ไม่ throw เพียงเพราะ status |
| Response | `HttpClientResponse { status, headers, body }` | headers เก็บค่าซ้ำได้ เช่น `Set-Cookie`; body เป็น `Bytes` ที่อ่านครบภายใต้ limit; ไม่เปิด streaming cursor ในรอบนี้ |
| Context | การเรียกใช้รับ cancellation จาก execution context ปัจจุบัน | หากเรียกจาก HTTP handler, task หรือช่วง shutdown คำขอต้องหยุดตาม owner/deadline; ไม่เก็บ context หรือ request handle หลัง scope จบ |

ตัวอย่างเป้าหมาย:

```craft
func main() {
    var request: HttpClientRequest = std.http.clientRequest(
        "GET",
        "https://example.com/health",
        std.bytes.fromString(""),
    )
    request.headers.set("accept", ["application/json"])
    var config: HttpClientConfig = std.http.clientConfig()
    config.timeoutMs = 5000
    let response: HttpClientResponse = std.http.send(request, config)
    if response.status == 200 {
        print(response.body.text())
    }
}
```

## กติกา Transport และความปลอดภัย

1. รองรับเฉพาะ absolute `http://` และ `https://` URL; ปฏิเสธ scheme อื่น, URL ที่มี userinfo และ URL ที่ parse ไม่ได้. TLS ตรวจ certificate และ hostname ตามระบบ; ไม่มีตัวเลือกปิดการตรวจ TLS ใน Public API รอบนี้.
2. **ไม่ตาม redirect อัตโนมัติ** ในรอบแรก. คืน 3xx และ `Location` ให้แอปตัดสินใจเอง เพื่อไม่ส่ง `Authorization` หรือข้อมูลลับข้าม host โดยไม่ตั้งใจ. ไม่มี automatic retry โดยเฉพาะคำสั่งเขียนที่อาจสำเร็จไปแล้ว.
3. ตรวจชื่อ/ค่า header ก่อนส่ง; ปฏิเสธ CR/LF และ control characters. Runtime จัดการ hop-by-hop/transport-owned headers และไม่ให้แอปกำหนดค่าที่ขัดกับ body/connection โดยตรง. ไม่พิมพ์ URL query, credentials, request/response body หรือ secret headers ใน exception/log โดยปริยาย.
4. จำกัด request body, response headers และ response body; ตรวจ `Content-Length` ล่วงหน้าเมื่อมี และอ่านแบบ bounded เพื่อรับมือกรณีไม่ประกาศขนาดหรือประกาศผิด. ปิด response body และคืน connection ให้ pool เมื่อสำเร็จ, เกิน limit, timeout หรือ cancellation.
5. กำหนด timeout รวมของคำขอที่ครอบคลุม DNS, connect, TLS, response headers และ body read; ใช้ context cancellation ของ Craft ร่วมด้วย. Error ต้องแยกอย่างน้อย config/URL, network, TLS, timeout, cancellation และ limit โดยไม่กลบสาเหตุที่สำคัญ.
6. ระบุให้ชัดว่ารอบแรกใช้ proxy จาก environment หรือไม่ก่อน implement; ห้ามให้ค่าระบบที่ไม่ระบุในเอกสารเปลี่ยนปลายทางเงียบ ๆ. หากใช้ environment proxy ต้องนิยาม `NO_PROXY`, TLS และนโยบาย credentials. แอป Server ที่เรียก URL จากผู้ใช้ต้องตรวจปลายทางเอง; HTTP client ไม่ใช่ SSRF guard หรือ sandbox.

## ลำดับงาน

1. ตรวจรูปแบบ `Bytes`, `Map<String,String[]>`, native struct, context และ exception ที่มีอยู่. สรุป API/ค่า default/range ก่อนเพิ่ม signature ใน lexer/type checker/interpreter/stdlib.
2. Implement transport บน Go `net/http` โดยใช้ client/transport ที่ reuse connection ได้ภายในขอบเขต Runtime; หลีกเลี่ยงการสร้าง transport ใหม่ต่อคำขอและกำหนดการปิด idle connections เมื่อจบโปรแกรม.
3. เพิ่มตัวอย่าง CLI ที่อ่าน URL จาก config ที่กำหนดเองและตัวอย่างเรียกจาก HTTP handler เพื่อพิสูจน์ cancellation โดยไม่ใส่ secret ใน source. ตัวอย่างต้องใช้ Public API เท่านั้น.
4. เพิ่มเอกสาร `std.http` client แยกจาก Server, อัปเดต CLI/README และแผน release. ไม่เพิ่ม wrapper ใน `packages/api` เว้นแต่มีกรณีใช้ซ้ำที่ชัดและทดสอบแยก.

## เกณฑ์ทดสอบและตรวจรับ

- ตรวจ `craft check`, `craft test`, `craft fmt --check`, `craft build` ของตัวอย่าง และ `go test ./...` รวม regression ของ Server/`packages/api` เดิม.
- ใช้ HTTP/HTTPS test server ในเครื่องทดสอบทั้งหก method, header/body แบบ Unicode และ binary, header ซ้ำ, 2xx/3xx/4xx/5xx, response ว่าง และ `HEAD`. ยืนยันว่า 3xx ไม่ตามต่อและ 4xx/5xx ไม่กลายเป็น transport exception.
- ทดสอบ URL/scheme/userinfo และ header ที่ผิด, certificate/hostname ที่ไม่ถูกต้อง, timeout ในแต่ละช่วงที่จำลองได้, cancellation จาก task และจาก HTTP handler/client disconnect, response ที่เกิน limit ทั้งแบบมีและไม่มี `Content-Length`, รวมถึง connection cleanup หลัง failure.
- ทดสอบคำขอพร้อมกันและ reuse connection โดยไม่เกิด resource leak; ตรวจว่า error/output ไม่มี token, URL query หรือ body. ทดสอบการรันจาก `.craftbundle` ด้วย.
- ให้ผู้ใช้ทดสอบการเรียก API จริงด้วยตนเองก่อนเปลี่ยนสถานะจาก candidate เป็นผ่าน; ผลจาก localhost เพียงอย่างเดียวไม่อ้างว่ารองรับเครือข่ายทุกสภาพแวดล้อม.

## ไม่รวมใน 0.1.46

WebSocket, SSE, streaming request/response, multipart upload, cookie jar, redirect อัตโนมัติ, retry/backoff, OAuth/authentication framework, cache, custom TLS trust/insecure mode, client certificate, proxy configuration API, HTTP/3 และการปรับ `packages/api` ให้เป็น HTTP client. งานเหล่านี้ต้องมีสัญญาและกรณีทดสอบของตนเองก่อนเพิ่ม.

## ผลการลงงาน

- เพิ่ม native types และ API `std.http.clientRequest`, `clientConfig`, `send` โดยไม่เปลี่ยน Server API; `packages/api` ยังเป็น Server framework.
- ใช้ Go `net/http` transport ที่ reuse connection; ปิด redirect และ environment proxy; ตรวจ TLS ตาม system trust. กำหนด request/response limits, total timeout และ execution-context cancellation.
- เพิ่มเอกสาร [HTTP Client](../../docs/REV46-HTTP-CLIENT.md), ตัวอย่าง Craft และ Go/Craft integration tests.
- `go test ./... -count=1 -timeout 600s` ผ่าน. Tests ของ client ครอบคลุมหก method, 3xx ไม่ตาม redirect, 4xx เป็น response, header ซ้ำ, TLS certificate ผิด, config/header/URL ผิด, body/header limit, timeout, cancellation และ connection reuse. Craft test เรียก Public API ผ่าน interpreter และรันตัวอย่างจาก source/bundle บน localhost ผ่าน.
- ตัวอย่าง Rev.46 ผ่าน `craft check`, `craft fmt --check`, `craft build`; package `api` ผ่าน `craft test` 9/9 และ package `db` ผ่าน 5/5 หลังขยับเวอร์ชัน. ตัวอย่าง package ทั้งสามผ่าน `craft check` หลังสร้าง lockfile ของ 0.1.46 ใหม่.
- ทดสอบ HTTPS ภายนอกกับ `GET https://covid-api.com/api/regions` โดยใช้ Craft 0.1.46 candidate: `craft run` จาก source และ `.craftbundle` ได้ HTTP 200 และ body 7,179 ไบต์ทั้งคู่ (10 ตุลาคม 2026). การตรวจนี้ยืนยันการส่งคำขอและรับข้อมูลผ่าน TLS; ไม่ยืนยันความสดใหม่หรือความถูกต้องเชิงระบาดวิทยาของข้อมูลจากบริการภายนอก.
- ผู้ใช้ตรวจรับ Rev.46 จากผลการทดสอบ HTTPS จริงแล้ว. สร้าง Windows Setup 0.1.46 สำเร็จจาก source นี้; `craft.exe version` และ metadata ของ Setup รายงาน 0.1.46, build log ระบุว่ารวม source ของ `gui`, `api` และ `db` แล้ว. ยังไม่รันตัวติดตั้งจริงหรือทดสอบการเรียกจาก UI/Desktop และ cancellation ที่เกิดจาก client disconnect ของ HTTP handler. Proxy ถูกปิดตามสัญญารอบนี้. ไม่มีการเปลี่ยนสถานะ UI Engine Final จากงานนี้.
