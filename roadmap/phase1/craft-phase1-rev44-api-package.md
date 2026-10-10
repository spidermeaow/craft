# Craft Phase 1 Rev.44 (0.1.44) — `packages/api`

**สถานะ:** Implemented candidate; รอผู้ใช้ตรวจรับ  
**ฐาน:** Craft 0.1.43 candidate และ HTTP Engine (`std.http`) ปัจจุบัน  
**แผนรายละเอียด:** [packages/api](../packages-api-plan.md)

## เป้าหมาย

ให้ผู้พัฒนาสร้าง HTTP API ด้วย Craft ได้สะดวกขึ้นผ่าน first-party `packages/api` ที่มี router, middleware, response helpers และ validation. ตัว package เขียนด้วย `.craft` และทำงานบน Public API ของ `std.http`; งานเฉพาะแอปไม่ถูกย้ายเข้า HTTP Engine หรือ private Runtime.

## งานใน 0.1.44

1. สร้าง `packages/api` เป็น Craft library ที่มีเวอร์ชันตรงกับ Craft toolchain และใช้ `import package api "api"`.
2. ลงทะเบียน route พร้อม path parameter และ route group; กำหนดการชนกันของ route, 404/405 และ `Allow` ให้แน่นอน.
3. เพิ่ม before/after middleware, JSON/text/error response helpers, query/JSON request helpers และ validation พื้นฐาน.
4. ส่งต่อ `HttpConfig`, request context, timeout และ graceful shutdown ให้ `std.http` เดิมจัดการ.
5. ทำตัวอย่างและ tests ของ package; ตรวจ `craft check`, `craft test`, `craft fmt --check`, `craft run`, `craft build` รวมถึง HTTP request จริงบน localhost และ regression ของ `std.http` โดยตรง.
6. อัปเดตคู่มือ, lockfile และตัวติดตั้งให้มี package ที่ตรงกับ 0.1.44 หลัง implementation ผ่านเกณฑ์.

## เกณฑ์จบรุ่น

แอปตัวอย่างสร้าง REST API พื้นฐานได้ผ่าน Public API ของ package โดยไม่แก้ private Runtime เพื่อรองรับ route เฉพาะแอป. พฤติกรรม router, middleware, error, JSON, validation, timeout และ shutdown ต้องมีผลทดสอบที่ตรวจซ้ำได้. ผู้ใช้ตรวจรับก่อนเปลี่ยนสถานะจาก candidate.

HTTP client, TLS configuration ใน Craft Server, WebSocket, SSE, streaming, multipart, authentication framework, ORM และ deployment อยู่นอกขอบเขต 0.1.44; รายการและเหตุผลอยู่ใน [แผนรายละเอียด](../packages-api-plan.md).

## ผลการลงงานและการทดสอบ candidate

- สร้าง `packages/api` ด้วย Craft source ทั้งหมด โดยปรับ framework ต้นแบบจาก Archive ให้ใช้ package import ปัจจุบัน. เพิ่มการตรวจชื่อ path parameter และปฏิเสธ `%`/backslash ใน route pattern. `std.http` ยังคงเป็น transport เดิม.
- ผูก package `api` กับ Craft toolchain version ใน manifest parser แบบเดียวกับ `gui`; ตัวอย่างใช้ path dependency โดยไม่ระบุเวอร์ชัน. เพิ่ม regression test ของการอนุมานเวอร์ชันและการปฏิเสธ version field ที่ระบุเอง.
- `craft check`, `craft test` (9 cases) และ `craft fmt --check` ของ package ผ่าน. `examples/api-server-package` ผ่าน `craft check`, `craft test` (2 cases), `craft fmt --check` และ `craft build`.
- HTTP Server จริงบน localhost ตอบ 200, 201, 400, 404, 405, 415 และ 422 ตามกรณี; ตรวจ `Allow`, `Content-Type` และ header จาก middleware. Test fixture เรียก `std.http.stop` ผ่าน handler แล้วตอบ 200 และ Server จบด้วย exit code 0. แอปที่รันจาก `.craftbundle` ตอบ `/health` 200 พร้อม middleware header.
- `go test ./...` ผ่านทั้ง repository. `craft check` ของ `examples/gui-package-lab`, `D:\02-Repository\app` และ `D:\02-Repository\craft-sheet` ผ่านหลังซิงก์ GUI lockfile เป็น 0.1.44.
- ยังต้องให้ผู้ใช้ทดลอง API package และตัวติดตั้ง 0.1.44 ก่อนตรวจรับ. ผลทดสอบนี้ไม่ใช่การรับรอง production throughput หรือ HTTPS.
