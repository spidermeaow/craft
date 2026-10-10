# Craft Phase 1 Rev.43 (0.1.43) — Windows Standard Library Foundation

**สถานะ:** Implemented candidate; รอผู้ใช้ตรวจรับ  
**ฐาน:** Craft 0.1.42; รองรับ Windows amd64 เป็นเป้าหมายของรอบนี้  
**ขอบเขต:** เพิ่ม Standard Library หกหมวด โดยไม่ขยาย UI Engine และไม่ย้ายงาน Craft Sheet เข้า Runtime

## เป้าหมาย

ให้แอป Craft ทำงานพื้นฐานที่ต้องใช้ในการสร้าง CLI, API และ Desktop บน Windows
ผ่าน API ที่ตรวจชนิดได้ มี error และขีดจำกัดชัดเจน. ใช้ Go Standard Library
หรือ Windows API ภายใน Runtime; ไม่มีความจำเป็นต้องให้โค้ด Craft เรียก Win32
โดยตรง. API ที่ไม่ผูก Windows (`encoding`, `crypto`, `url`, `regex`) ออกแบบให้
ความหมายไม่ขึ้นกับ OS แม้รอบนี้ตรวจรับบน Windows เท่านั้น.

## Public API ที่เสนอ

Signature ด้านล่างเป็น Public API ที่ลง source ใน 0.1.43; ดูรายละเอียดการใช้งานใน [คู่มือ Rev.43](../../docs/REV43-WINDOWS-STDLIB.md).

| ลำดับ | หมวด | API ที่เสนอ | ข้อกำหนดสำคัญ |
| --- | --- | --- | --- |
| 1 | `std.encoding` | `base64Encode(Bytes): String`, `base64Decode(String): Bytes`, `hexEncode(Bytes): String`, `hexDecode(String): Bytes` | ใช้ Base64 มาตรฐานและ hex lowercase; decode ปฏิเสธข้อมูลผิดรูปแบบ; round-trip ข้อมูลไบนารีรวมศูนย์ไบต์ |
| 2 | `std.crypto` | `sha256(Bytes): Bytes`, `hmacSha256(key: Bytes, data: Bytes): Bytes`, `secureRandom(length: Int): Bytes`, `uuid(): String`, `secureEqual(left: Bytes, right: Bytes): Bool` | digest/MAC เป็น raw Bytes; `uuid` เป็น UUID v4; `secureRandom` ใช้ CSPRNG, ไม่ใช้ `std.math.random`; `secureEqual` ใช้เปรียบเทียบ MAC โดยไม่เทียบแบบ String ปกติ |
| 3 | `std.url` | `parse(String): UrlParts`, `encode(String): String`, `decode(String): String`, `query(String): Map<String, String[]>` | กำหนดว่า `encode/decode` เป็น **query-component encoding**; `query` เก็บค่าซ้ำและลำดับค่าของแต่ละ key; `parse` ไม่เปิด network และแยก scheme/host/path/rawQuery/fragment |
| 4 | `std.process` | `run(executable: String, args: String[], timeoutMs: Int, maxOutputBytes: Int): ProcessResult`, `start(executable: String, args: String[], maxOutputBytes: Int): Process`, `wait(process: Process, timeoutMs: Int): ProcessResult`, `kill(process: Process): Void`, `exitCode(process: Process): Int?` | ส่ง executable และ args แยกกัน ไม่ผ่าน shell; `ProcessResult` มี exit code, stdout/stderr เป็น Bytes; process handle มี owner, cleanup, timeout และ output limit |
| 5 | `std.regex` | `match(pattern: String, text: String): Bool`, `find(pattern: String, text: String): RegexMatch?`, `replace(pattern: String, text: String, replacement: String): String`, `split(pattern: String, text: String): String[]` | ใช้ไวยากรณ์ Go RE2; กำหนดตำแหน่ง `RegexMatch` เป็น Unicode code-point offsets ให้สอดคล้องกับ Craft String; จำกัด pattern/input/output และจำนวนผลลัพธ์ |
| 6 | `std.os` | `version(): String`, `architecture(): String`, `processId(): Int`, `open(pathOrUrl: String): Void` | `version` ระบุ Windows แบบที่ตรวจสอบได้; `architecture` ระบุสถาปัตยกรรมของ process อย่างชัด; `open` ใช้ default handler ของ Windows พร้อม error เมื่อเปิดไม่ได้ |

### ข้อตัดสินก่อน implement

1. `std.process.start` ต้องมี `wait` เพื่อเก็บ exit status และคืน resource; `kill`
   ไม่เท่ากับ `wait`. stdout/stderr ต้องถูกอ่านพร้อมกันระหว่างทำงานและมี
   ขีดจำกัดร่วมกัน เพื่อไม่ให้ pipe ค้างหรือหน่วยความจำโตไม่จำกัด. กำหนด
   พฤติกรรมเมื่อ handle หลุด scope, ถูกเรียกซ้ำ หรือ
   Craft ถูกยกเลิก. `run` ต้องจบด้วยผลลัพธ์ที่ตรวจ exit code ได้ แม้โปรแกรมลูก
   คืน nonzero. แยก launch/timeout/output-limit errors จาก nonzero exit.
2. บน Windows ให้กำหนด semantics ของ executable path, working directory,
   argument quoting, `.bat`/`.cmd` และการซ่อนหน้าต่าง helper อย่างชัดเจน.
   ห้ามประกอบ command line ด้วยการต่อ String แล้วส่งเข้า `cmd.exe /c` โดยปริยาย.
3. `std.os.open` เป็นการเปิดเอกสาร โฟลเดอร์ หรือ URL ผ่าน file association;
   ไม่ใช่ API สำหรับรับ exit code ของโปรแกรม. จำกัด scheme ที่อนุญาตสำหรับ URL
   และปฏิเสธ path/URL ที่มี control characters. การเปิดอาจทำให้แอปภายนอกทำงาน
   จึงต้องเป็นการเรียกอย่างชัดเจนจากแอป.
4. จำกัด Bytes ตามสัญญาเดิมที่ 16 MiB; กำหนดขนาดสูงสุดของ random output,
   URL/regex input และ process output ก่อนลง source. ปฏิเสธค่าที่เกินขีดจำกัด
   ก่อนจัดสรรหน่วยความจำจำนวนมาก. ไม่พิมพ์ key, token หรือ output ที่อาจมี
   secrets ลง diagnostic/log.
5. `std.crypto.sha256` และ `hmacSha256` ไม่ใช่การเข้ารหัสข้อมูลหรือการ hash
   รหัสผ่าน; ตัวอย่างต้องไม่ชวนให้ใช้ SHA-256 แบบตรง ๆ เก็บ password.

## ลำดับลงงาน

1. **Encoding:** เพิ่ม signatures, interpreter dispatch, error types,
   เอกสาร และตัวอย่าง binary round-trip. ใช้ `encoding/base64` และ `encoding/hex`.
2. **Crypto:** เพิ่ม SHA-256, HMAC, secure random, UUID v4 และ constant-time
   comparison. ใช้ `crypto/sha256`, `crypto/hmac`, `crypto/rand`; ไม่สร้าง
   algorithm เอง.
3. **URL:** เพิ่ม `UrlParts` และ query utilities ด้วย `net/url`; ตรวจการแปลง
   UTF-8, `%` ผิดรูปแบบ, `+` กับ space, key ซ้ำ และ URL ที่เป็น relative.
4. **Process:** ออกแบบ `Process` handle และ lifecycle ก่อนลง API; ใช้
   `os/exec` กับ context/timeout และ output limit. ทดสอบ CLI child, nonzero
   exit, timeout, kill/wait และ path ที่มีช่องว่างบน Windows.
5. **Regex:** ใช้ `regexp` ของ Go; ระบุ RE2 syntax และข้อจำกัดที่ต่างจาก
   PCRE เช่น backreference/lookbehind. แปลง byte offsets เป็น code-point offsets.
6. **OS:** ใช้ Go `runtime`/`os` สำหรับข้อมูลที่มีอยู่ และ Windows API เฉพาะ
   ส่วนที่จำเป็น เช่น version/default-app open; อย่าใช้ค่าที่ตีความผิดว่าเป็น
   Windows version. จัด Windows-only implementation ให้แยกไฟล์ชัดเจน.

## เกณฑ์ตรวจรับ

- `craft check`, `craft run`, `craft test` และ `craft fmt --check` ใช้ API ใหม่ได้
  โดย signature และ Runtime ให้ผลตรงกัน; source เดิมยังตรวจและรันได้.
- Regression cases ครอบคลุมค่า valid/invalid, Unicode, binary zero, ขีดจำกัด,
  exception, cancellation, resource cleanup และ round-trip. มีตัวอย่าง Craft
  สำหรับทั้งหกหมวด โดยไม่ต้องใช้ `packages/gui`.
- ตรวจจริงบน Windows amd64: path ที่มีช่องว่าง/อักษรไทย, child process ที่
  คืน nonzero, timeout, output ใหญ่, file association และ URL handler.
- Security review ของ process/OS open และ crypto โดยเฉพาะ shell injection,
  argument quoting, secret leakage และการเปรียบเทียบ MAC.
- ผู้ใช้ตรวจผลด้วยตนเองก่อนรับ Rev.43.

## ผลการลงงานและการทดสอบของ candidate

- ลง API ครบหกหมวดใน `internal/stdlib`; เพิ่ม `UrlParts`, `RegexMatch`, `ProcessResult` และ opaque `Process` ในระบบชนิดข้อมูล. `std.process` ใช้ function execution scope เก็บและปิด child process; ปฏิเสธ `.bat`/`.cmd` และจำกัด stdout/stderr รวมกัน.
- ตัวอย่าง `examples/rev43-stdlib` เรียกทั้งหกหมวดโดยใช้ Public API เท่านั้น. `craft check`, `craft test` (2 cases), `craft fmt --check` และ `craft run` ผ่านบน Windows amd64. `craft run` ใช้ `hostname.exe` ทดสอบทั้ง `run` และ `start`/`wait`.
- `go test ./...` ผ่าน รวมกรณี encoding ผิดรูปแบบ, Unicode offset, process nonzero exit, timeout, output limit, kill/wait และ cleanup เมื่อ scope จบ. ตรวจ compatibility ด้วย `craft check` กับ `examples/gui-package-lab`, `D:\02-Repository\app` และ `D:\02-Repository\craft-sheet` หลังอัปเดต lockfile.
- ระหว่าง regression test พบ Lexer กลืน newline หลัง `Map<...>` ใน struct ทำให้ `craft fmt` ขัดกับ Parser; แก้การแยก generic closing token และทำให้ regression test เดิมกลับมาผ่าน. CLI end-to-end test เปลี่ยนจากการเทียบเวอร์ชัน 0.1.25 แบบตายตัวไปใช้ `toolchain.Version`.
- ยังไม่ได้เปิด default handler จริงด้วย `std.os.open` ใน automated test เพราะการทดสอบจะเปิดโปรแกรมภายนอก. ผู้ใช้สามารถทดสอบการเปิดเอกสารและ URL บนเครื่องของตนก่อนตรวจรับ.

## ไม่รวมใน Rev.43

`std.net` raw TCP/UDP, `std.sync`, `std.archive`, `std.compress`, `std.win32`,
HTTP client เต็มรูปแบบ, TLS server configuration, GPU/USB/Serial/Bluetooth,
การรองรับ GUI บน OS อื่น และการแก้ UI Engine Final gate ที่พักไว้. การแจกจ่าย
CraftBundle แบบ standalone ยังอยู่ใน [แผน 0.2.2](../craft-0.2.2-deployment-plan.md).
ข้อบกพร่อง Lexer ที่พบกับ `Map<...>` ใน Craft Sheet ยังอยู่ใน
[บันทึก Rev.42](craft-phase1-rev42.md); ไม่อ้างว่าถูกแก้โดยแผน Standard Library นี้.

## แหล่ง implementation ที่ใช้ประกอบการออกแบบ

- Go [`crypto/rand`](https://pkg.go.dev/crypto/rand) และ
  [`crypto/hmac`](https://pkg.go.dev/crypto/hmac)
- Go [`net/url`](https://pkg.go.dev/net/url), [`regexp`](https://pkg.go.dev/regexp)
  และ [`os/exec`](https://pkg.go.dev/os/exec)
- Microsoft [`ShellExecuteW`](https://learn.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shellexecutew)
