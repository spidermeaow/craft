# Craft Phase 1 Rev.10 — Logging, Configuration และ Typed Errors

วันที่: 22 กันยายน 2026  
ฐาน: Rev.9 / Craft 0.1.9  
เป้าหมายรุ่น: 0.1.10  
สถานะ: ผ่าน Windows x64 acceptance วันที่ 23 กันยายน 2026; ดูหลักฐานและข้อจำกัดใน [สถานะ Rev.10](output/phase1-rev10-status.md)

## 1. เป้าหมาย

ช่วยให้ผู้ใช้ Craft ตั้งค่าโปรแกรมและตรวจหาสาเหตุของปัญหาได้ง่ายขึ้น ผ่าน logging, configuration validation และ filesystem errors ที่แยกประเภทได้ พร้อมเก็บรายละเอียดด้าน cancellation, cleanup และ bundle compatibility ของ Rev.9

ปิดรับรองเฉพาะ API ที่กำหนดในรอบนี้ ไม่ถือว่า standard library ทั้งหมดเสร็จสมบูรณ์ เครื่องมือและตัวอย่างต้องใช้ได้โดยไม่พึ่ง API Framework

## 2. ขอบเขตที่รับเข้า

### 2.1 Logging ขั้นพื้นฐาน

- รองรับระดับ debug, info, warn และ error พร้อมการกำหนดระดับขั้นต่ำที่ต้องการแสดง
- รับ message และ structured fields โดยกำหนดชนิดข้อมูลที่ยอมรับ ขนาดสูงสุด และวิธีจัดการข้อมูลที่ไม่รองรับก่อน implement
- ส่ง log ไป stderr เพื่อให้ stdout ใช้ส่งผลลัพธ์ของ CLI ได้; รองรับ writer ที่ inject ได้สำหรับ tests และ embedding
- กำหนดรูปแบบ log ที่แน่นอน เช่น JSON หนึ่ง record ต่อบรรทัด พร้อม escaping newline/control characters, timestamp และลำดับ field ที่คาดเดาได้
- ให้ผู้เรียกระบุ field ที่ sensitive อย่างชัดเจน และ redact ก่อน serialize/ส่งออก; ไม่รับรองการเดา secret จากข้อความทั่วไปหรือ path โดยอัตโนมัติ
- ระบุ behavior เมื่อเขียน log ล้มเหลว และกำหนดการเขียนพร้อมกันจาก tasks/HTTP handlers ไม่ให้ record ปะปนกัน
- ใช้ configuration ที่ผูกกับ execution/session หลีกเลี่ยง mutable global state ที่ทำให้โปรแกรมหรือ tests กระทบกัน
- รอบแรกไม่เพิ่ม background queue, file rotation หรือ remote log transport

### 2.2 Configuration validation

- ต่อยอด `std.env.args`, `std.env.get`, `std.env.lookup` และ conversion APIs เดิม โดยไม่เปลี่ยน semantics ของ unset กับ present-empty
- เพิ่ม helper ขนาดเล็กสำหรับ required value, default value, conversion และ range validation ตามชนิดที่เลือกใน M0
- กำหนด behavior ของ empty/whitespace, invalid number, overflow, Bool และค่าที่อยู่นอกช่วงอย่างชัดเจน
- error ระบุชื่อ setting/argument และเงื่อนไขที่ผิด โดยไม่แสดงค่าที่รับเข้าหรือค่าลับ
- ให้ application เลือกลำดับความสำคัญของ arguments/environment/default เอง พร้อมตัวอย่างลำดับที่ชัดเจน; core ไม่โหลด config source เพิ่มโดยปริยาย
- ทำ helper ด้วย Craft package เมื่อ primitives เดิมเพียงพอ เพิ่ม native API เฉพาะส่วนที่จำเป็น
- ไม่เพิ่ม `.env` auto-loading, secret store, configuration framework หรือ command-line parser เต็มรูปแบบ

### 2.3 Filesystem errors ที่แยกประเภทได้

- ให้โปรแกรม catch และแยก missing path, permission denied, invalid path, unsupported file kind และ size limit ได้โดยไม่ parse message
- ระบุ mapping ไปยัง Exception/type contract เดิมก่อน implement; ชื่อ error ขั้นสุดท้ายและ field ที่เปิดเผยต้องบันทึกในคู่มือ
- แยก invalid UTF-8 และ generic I/O failure; mapping ข้อผิดพลาด Windows ที่ไม่ชัดต้องใช้ fallback โดยไม่เดาสาเหตุ
- ครอบคลุม APIs ของ filesystem/path ที่เกี่ยวข้อง รวม atomic replace; ข้อความระบุ operation และ path โดยไม่ส่ง arbitrary OS error text หรือ file contents
- ระบุว่า path ที่ผู้เรียกส่งเข้ามาอาจมีข้อมูล sensitive และไม่ได้ถูก redact อัตโนมัติทุกกรณี
- รักษา error propagation และ catch เดิม; หาก error type เปลี่ยน ต้องมี compatibility tests และ migration note

### 2.4 เก็บรายละเอียด Rev.9

- ตรวจ cancellation ก่อนเริ่ม I/O และก่อน commit การแทนที่ไฟล์ พร้อมระบุจุดที่ยกเลิกได้และข้อจำกัดของ blocking OS calls
- ตรวจ cleanup ของ temporary file เมื่อ write, close หรือ replace ล้มเหลว รวมวิธีรายงาน cleanup failure โดยไม่กลบสาเหตุหลัก
- เพิ่ม fault-injection tests ที่พิสูจน์ว่า destination เดิมไม่ถูกทำลายก่อน replace สำเร็จ
- ระบุ metadata/permission behavior, symlink policy และข้อจำกัดด้าน durability โดยไม่อ้างว่า rename รับรอง power-loss recovery
- ตรวจ bundle language metadata ของโปรแกรมที่เรียก API ใหม่: runtime เก่าต้องปฏิเสธอย่างชัดเจน และ runtime ใหม่ต้องอ่าน bundle เดิมตาม compatibility policy
- ทบทวนเอกสารและ version references ให้ตรงกับ source/release ปัจจุบัน

### 2.5 ตัวอย่างและการส่งมอบ

- เพิ่มตัวอย่าง CLI ที่รับ config จาก environment/arguments, validate, เขียนไฟล์แบบ atomic, บันทึก log และจัดการ typed errors
- แสดง failure cases ที่ตรวจได้ เช่น missing setting, invalid value, missing file และ write failure โดยไม่แสดง secret
- เพิ่มคู่มือ API/ข้อจำกัด, คู่มือทดลอง, acceptance script และ status report แยก implementation ออกจากผลที่ทดสอบจริง
- เตรียม CLI และ Windows Setup 0.1.10, checksum และ release notes หลังผ่านเกณฑ์รับงาน

## 3. สิ่งที่ไม่ทำใน Rev.10

- `craft upgrade`, remote registry, GitHub dependencies, publish หรือ dependency update
- process execution, signals, watcher, REPL, native backend หรือ VM
- API Framework, ORM, migrations, authentication หรือ application policy ใหม่
- logging backend ภายนอก, log rotation, automatic secret discovery หรือ telemetry ที่ส่งออกเครือข่าย
- filesystem sandbox, recursive deletion, multi-file transactions หรือการรับรอง network filesystem/ACL ทุกแบบ

## 4. แผนดำเนินงาน

1. **M0 — API และ compatibility contract:** inventory ของเดิม กำหนด signatures/error mapping, logging format/redaction/concurrency และ config validation rules ให้ชัดก่อน implement
2. **M1 — Typed errors และ Rev.9 fixes:** ปรับ error mapping, cancellation/cleanup และ bundle compatibility พร้อม regression fixtures
3. **M2 — Logging:** implement session-scoped logging, stderr writer, filtering, structured fields และ explicit redaction
4. **M3 — Configuration:** เพิ่ม helpers เท่าที่จำเป็น พร้อม tests ของ unset/empty/conversion/range และตัวอย่าง precedence
5. **M4 — Example และเอกสาร:** เพิ่ม CLI example, acceptance script, migration note และคู่มือทดลองที่รันได้ด้วย Windows PowerShell โดยไม่บังคับ PowerShell 7
6. **M5 — Acceptance และ Setup:** ตรวจ regression, race สำหรับ concurrent logging, package/bundle compatibility และ installer upgrade; บันทึกผลก่อนประกาศ release

## 5. เกณฑ์ผ่าน Rev.10

- API และ error types ที่รับเข้าขอบเขตมี contract และ limits ชัดเจน พร้อม tests ทั้ง success/failure
- logging แยก stderr จาก stdout, filter ตาม level และ redact fields ที่ระบุได้; concurrent records ไม่ปะปนและ writer failure มี behavior ที่ทดสอบได้
- config validation แยก unset/empty และรายงานเงื่อนไขที่ผิดโดยไม่แสดงค่ารับเข้า
- Craft program แยก filesystem error types ได้โดยไม่ parse message และ generic fallback ทำงานได้
- fault injection ยืนยัน atomic-write preservation/cleanup ตาม contract และไม่มี claim เกินหลักฐาน
- bundle/runtime compatibility ของ APIs ใหม่และเก่าถูกตรวจจริง
- ตัวอย่าง CLI ผ่าน check/test/run/build และรัน bundle ได้
- regression Rev.1–Rev.9 ผ่าน; race และการทดสอบที่เกี่ยวข้องผ่านตาม scope โดยรายงาน skip/ข้อจำกัด
- Setup 0.1.10 ติดตั้งหรืออัปเกรดได้ executable ที่ติดตั้งรายงานเวอร์ชันถูกต้องและรันตัวอย่างได้

## 6. การทดสอบและสถานะ

ค่าเริ่มต้นให้ผู้ใช้รันทดสอบตาม working preference; หากผู้ใช้ขอให้ agent ทดสอบจึงรันภายใน scope ที่อนุญาต ไม่ถือว่าการเขียนแผนนี้เป็นคำสั่งให้รันทดสอบหรือเริ่ม implementation

ใช้ temporary fixtures เฉพาะงานและ synthetic secrets สำหรับ redaction tests ห้ามนำ credential จริงมาแสดงใน output เก็บผล command/exit code, environment ที่เกี่ยวข้อง และรายการที่ไม่ได้รันใน `output/phase1-rev10-status.md` เมื่อเริ่มส่งมอบ

ดู [แผน Rev.9](craft-phase1-rev9.md), [สถานะ Rev.9](output/phase1-rev9-status.md) และ [ทิศทางภาษา](../craft-language-direction.md)
