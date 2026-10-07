# Craft Phase 1 Rev.13 — สะพานเชื่อม Craft กับ Native UI Engine

วันที่: 6 ตุลาคม 2026  
ฐาน: Craft 0.1.12 / Phase 1 Rev.12  
เป้าหมายรุ่น: Craft 0.1.13  
สถานะ: implementation ใน source แล้ว; รอผู้ใช้ตรวจ GUI จริง ดู [สถานะ Rev.13](output/phase1-rev13-status.md)

## 1. เป้าหมาย

ทำให้โปรแกรม Craft บน Windows **ส่งคำสั่งไปยัง native UI host และรับ event กลับมาประมวลผลได้จริง** โดยไม่ใช้ WebView และไม่เพิ่ม syntax เฉพาะ GUI ในภาษา งานรอบนี้สร้างขอบเขตเชื่อมระหว่าง interpreter กับ UI engine เพื่อให้รอบถัดไปพัฒนา layout, renderer และ components บนฐานที่ตรวจสอบได้

ผลลัพธ์ขั้นต่ำคือโปรแกรม Craft เปิดหน้าต่างหนึ่งบาน เปลี่ยนชื่อ/พื้นหลังของหน้าต่างตาม event จากเมาส์หรือคีย์บอร์ด และปิดได้อย่างปลอดภัย ทั้งจากผู้ใช้และจากการยกเลิกโปรแกรม ตัวอย่างนี้เป็นเครื่องมือพิสูจน์ bridge เท่านั้น ไม่ใช่ Craft Office หรือ framework GUI ที่พร้อมใช้งานทั่วไป

คำว่า **UI host** ในแผนนี้หมายถึงชั้น native ที่สร้างหน้าต่าง รับข้อความจาก OS และแสดงผลขั้นต่ำ คำว่า **Craft GUI package** หมายถึง Craft source ที่ห่อ host API ให้ผู้พัฒนาเรียกใช้ การมี bridge ใช้งานได้ยังไม่เท่ากับมี UI engine สำหรับ layout, text หรือ GPU rendering เต็มรูปแบบ

Rev.13 ยังเป็น Phase 1 ตามนโยบายโครงการ การเปลี่ยน Phase หรือรุ่น 0.1.x ไม่ได้ประกาศว่า Craft พร้อมเป็น V1.00; ผู้ใช้เป็นผู้ตัดสินใจเรื่องนั้น

การสร้างหน้าจอ GUI ของ Craft ใช้วิธีเขียน Craft source โดยตรง **ไม่มีแผนทำเครื่องมือออกแบบหน้าจอแบบลากวาง** ทั้งใน Rev.13 และในเป้าหมายระยะยาวของสายงาน GUI

## 2. ฐานปัจจุบันและการตัดสินใจด้านขอบเขต

- Craft 0.1.12 มี static checker, AST interpreter, modules/imports, source packages และ `Fn<...>` สำหรับ named functions แต่ยังไม่มี lambda/closures; API ที่ต้องจับ mutable state ใน callback จึงยังไม่เหมาะเป็นรูปแบบหลักของ Rev.13
- `craft build` ยังสร้าง source bundle ที่ต้องใช้ Craft CLI; bundle ไม่บรรจุ OS backend หรือกลายเป็น `.exe` ของแอปผู้ใช้
- Rev.13 ใช้ **native provider ที่ลงทะเบียนกับ Craft runtime** สำหรับความสามารถ UI บน Windows ก่อน ไม่เปิดให้ package ดาวน์โหลดและโหลด DLL ใด ๆ ได้เอง การโหลด native extension ทั่วไปต้องออกแบบ trust, ABI, compatibility และ distribution ต่างหาก
- API ที่ผู้พัฒนาใช้ห่อด้วย Craft source package แยกจาก standard library; Go code ใต้ runtime ทำเฉพาะ host boundary และการเรียก OS ที่ Craft source ทำเองไม่ได้
- Windows 10+ x64 เป็น platform สำหรับ implementation และการตรวจรับรอบนี้ตามฐาน installer ปัจจุบัน บน platform ที่ไม่มี UI provider ต้องได้ error ที่บอกความสามารถที่ขาดอย่างชัดเจน ไม่ panic หรือทำงานครึ่งหนึ่ง

เส้นทางข้อมูลที่ต้องพิสูจน์:

```text
Craft main/event loop
  → Craft GUI source package
  → typed native provider ใน Craft runtime
  → Windows UI host
  → OS event queue
  → typed UiEvent กลับสู่ Craft
```

การออกแบบนี้ไม่ต้องมี native compiler, dynamic library loader หรือ WebView และไม่เปลี่ยนวิธีทำงานของแอป API/CLI ที่ไม่ใช้ GUI

## 3. สัญญาที่ต้องกำหนดก่อน implement (M0)

API ขั้นต่ำที่ใช้ประเมินการออกแบบ (ชื่อจริงและชนิดข้อมูลให้ตรึงใน M0 โดยไม่เพิ่ม syntax ภาษา):

| คำสั่งจาก Craft | ผลที่ต้องได้ |
| --- | --- |
| `openWindow(title, width, height)` | opaque window handle หรือ error |
| `setTitle(window, title)` / `setBackground(window, color)` | ปรับหน้าต่างที่ยังเปิดอยู่ หรือ error ของ handle |
| `nextEvent(timeout)` | event พร้อม window identity หรือผลว่าไม่มี event ภายในเวลา |
| `closeWindow(window)` | ปิด resource ที่เป็นของ session อย่างชัดเจน |

Event ขั้นต่ำมี `CloseRequested`, `Resized`, `PointerClicked` และ `KeyPressed` ส่วนชื่อ API นี้เป็นร่างของ contract ไม่ใช่ syntax ที่ Craft 0.1.12 ใช้ได้แล้ว

กติกาที่ต้องตัดสินและบันทึกเป็นเอกสาร M0 ก่อนเขียน backend:

| เรื่อง | ข้อเสนอสำหรับ Rev.13 |
| --- | --- |
| หน้าต่าง | หนึ่ง active window ต่อ UI session; การเปิดเพิ่มขณะยังไม่ปิดคืน error ที่ระบุชัด |
| ขนาด/พิกัด | ใช้ logical pixels อ้างอิง 96 DPI; แปลงเป็น physical pixels ที่ Windows boundary |
| สีพื้นหลัง | ค่า RGB 0–255; invalid component ถูกปฏิเสธ ไม่ clamp เงียบ ๆ |
| `nextEvent(timeoutMs)` | `0` ไม่รอ, ค่าบวกไม่เกิน 60,000 ms, `-1` รอจนมี event หรือ cancel; timeout คืน `null` |
| การปิดจาก title bar/Alt+F4 | ส่ง `CloseRequested`; Craft ตัดสินใจเรียก `closeWindow`; เมื่อ `main` จบ runtime ปิดให้เสมอ |
| Event payload | typed `UiEvent` มี `kind` และ window identity; field ที่ใช้เฉพาะ event ต้องระบุใน contract และมีค่า neutral เมื่อไม่เกี่ยวข้อง |
| คีย์บอร์ด | `KeyPressed` ส่ง key code กับ modifiers; text composition/IME เป็นงานภายหลัง |
| Error | ใช้ Craft exception ที่แยก unavailable, invalid argument/handle, wrong thread, queue overflow และ closed session |

ตัวอย่าง flow ที่ต้องเกิดได้จริง: Craft เปิดหน้าต่าง → รอ event → คลิกแล้วเปลี่ยน title/สี → กดคีย์แล้วเปลี่ยนกลับ → กดปิดหน้าต่างแล้ว Craft ปิด handle และออกจาก loop ไม่ใช้ callback ที่จับ local state

### 3.1 Runtime/native boundary

- กำหนดทะเบียน native provider ที่มีชื่อ capability, version และ typed signatures ให้ checker และ interpreter ใช้สัญญาเดียวกัน ห้ามให้ checker ยอมรับ API ที่ runtime เรียกไม่ได้
- เลือกชื่อ namespace ภายในของ provider และวิธีประกาศ signature ให้มีแหล่งข้อมูลเดียว; Craft GUI package เป็น public API ส่วน primitive ใน provider เป็น implementation detail ที่ไม่สัญญาว่าจะคงชื่อข้ามรุ่น
- ไม่ให้ Craft package ระบุ arbitrary DLL path, symbol name หรือ pointer/address ได้; ขอบเขตนี้ใช้ provider ที่มาพร้อม runtime เท่านั้น
- กำหนดการแปลงค่า Craft ↔ Go สำหรับ `Int`, `Float`, `Bool`, `String`, optional, opaque handle และ event โดยระบุ range, UTF-8/Unicode, copy/ownership และ error ที่ตรวจได้
- ตรวจ `craft check` จาก signature/target metadata โดยไม่เปิดหน้าต่าง; `craft run` ตรวจ provider จริงก่อนเริ่ม `main`; `craft build` บันทึกความต้องการ provider ใน bundle โดยไม่เริ่ม GUI
- Bundle ใหม่ที่ใช้ GUI ต้องระบุ language/capability contract ของ 0.1.13 ให้ runtime เก่าปฏิเสธอย่างชัด; bundle รุ่นก่อนยังอ่านได้ การไม่มี provider หรือ version ไม่ตรงต้องเป็น diagnostic ก่อนใช้ native handle
- รักษา behavior ของโปรแกรมที่ไม่ import GUI และของ source bundle รุ่นก่อน; ไม่บังคับ GUI provider ให้ทำงานระหว่าง `craft check` หรือ `craft test` ที่ใช้ headless backend

### 3.2 Session, handle และ thread

- หนึ่ง interpreter run มี UI session ของตนเอง; ไม่ใช้ mutable global state ของหน้าต่างข้ามโปรเจกต์หรือ test run
- Window handle เป็นค่า opaque ที่มี session ID และ identity/generation โดยไม่เปิดเผย raw OS pointer; การ copy ค่า handle ยังอ้างหน้าต่างเดียวกัน การใช้ handle ผิด session, หลังปิด หรือปิดซ้ำต้องมีผลที่กำหนดชัด
- Windows window creation, message pump และ destruction อยู่บน OS thread เดียวกันตลอดอายุ UI session; ล็อก thread ตั้งแต่เริ่ม session ถึง cleanup และปล่อยเมื่อ native resources ปิดแล้ว
- Rev.13 อนุญาตคำสั่ง UI จาก Craft event loop บน owning thread เท่านั้น การเรียกจาก task/timer หรือ thread อื่นคืน `UiThreadError` แทนการ queue แบบไม่มีกติกา; ออกแบบจุดขยายสำหรับ cross-thread dispatch ในรุ่นหลัง
- ทุก window, callback registration, event queue และ native resource ต้องถูกปิดเมื่อ loop จบ, `main` คืนค่า, เกิด exception หรือ context ถูกยกเลิก; cleanup ทำงานได้แม้ Craft ไม่เรียก close เอง
- กำหนดลำดับการปิด: หยุดรับ event ใหม่, จัดการ event ที่ค้างตาม policy, ทำลาย resource, แล้วปล่อย session โดยไม่มี callback เรียกกลับหลังปิด
- `closeWindow` สำเร็จครั้งแรก; การปิดซ้ำและการใช้ handle หลัง close คืน `UiInvalidHandle` อย่างคาดเดาได้ การปิด session จะ invalidate handle ทุกตัวก่อนคืนจาก interpreter

### 3.3 Event และการตอบสนอง

- Rev.13 ให้ Craft ควบคุม state ผ่าน **event loop แบบดึง event** เช่น `nextEvent(timeout)` ซึ่ง pump ข้อความของ OS แล้วคืน event ที่มีชนิดและข้อมูลชัดเจน; ชื่อและ signature จริงต้องยืนยันใน M0 ก่อนเขียนตัวอย่างถาวร
- รองรับอย่างน้อย close request, resize, DPI change, pointer click และ key input; event ต้องระบุ window ที่มาและค่าพิกัดใน logical pixels โดยกำหนด rounding เมื่อแปลงจาก physical pixels
- กำหนด queue สูงสุด 1024 events ต่อ session; coalesce resize/DPI change ที่ยังไม่ถูกอ่านได้ แต่ close/key/click ต้องรักษาลำดับ หาก non-coalescible event ทำให้ queue เต็ม ให้คืน `UiQueueOverflow` และปิด session อย่างปลอดภัย ห้ามทิ้ง event โดยเงียบ ๆ
- การรอ event ต้องเคารพ cancellation และไม่ทำให้ window ค้างเมื่อ Craft มีงานยาว; ระบุข้อจำกัดว่า Craft code ที่ block UI thread นานยังทำให้ UI ไม่ตอบสนอง
- ไม่มี callback เข้า interpreter จาก OS window procedure โดยตรง; native side แปลงเป็น event ขนาดจำกัด แล้ว `nextEvent` ส่งให้ Craft หลังออกจาก OS callback เท่านั้น
- callback แบบ `Fn<...>` สำหรับ API ระดับสูงให้พัฒนาหลังมีสัญญาเรื่อง state, reentrancy และ exception ที่พิสูจน์แล้ว หรือเพิ่ม adapter แบบ opt-in ที่ใช้ event queue เดียวกัน

## 4. ขอบเขต implementation

### M0 — ตรึง contract ก่อนเขียน backend

- ตัดสินชื่อ API/type/event, native capability/version, กติกา thread, queue/overflow, DPI, error codes และ bundle compatibility ตามหัวข้อ 3 แล้วเขียนเป็นเอกสารที่ checker/runtime/package ใช้อ้างอิงร่วมกัน
- ร่างตัวอย่าง Craft source แบบสั้นและตรวจ syntax/type ด้วย local stub package เพื่อพิสูจน์ว่าการจัดการ state ด้วย polling เขียนได้ด้วยภาษา Craft ปัจจุบันโดยไม่ต้องเพิ่ม closure หรือ syntax ใหม่; stub นี้ยังไม่เปิดหน้าต่าง
- ระบุ Windows API ที่จะใช้, dependency/license ที่ต้องเพิ่ม (ถ้ามี), และ failure behavior เมื่อ OS สร้างหน้าต่างไม่สำเร็จ
- ผลส่งมอบ M0: contract ฉบับตรึงพร้อมตาราง signature/event/error และรายการ test cases จาก acceptance matrix; ถ้า API ต้องเปลี่ยนแกนภาษาเกินขอบเขตนี้ให้ทบทวนแผนก่อนเริ่ม M1

### M1 — Headless bridge

- แยก interface ของ UI host ออกจาก Windows implementation และเพิ่ม fake/headless host สำหรับตรวจลำดับคำสั่ง, event, cancellation และ cleanup โดยไม่เปิดหน้าต่าง
- เพิ่ม typed native provider registration และ session lifecycle ใน runtime/checker เท่าที่ต้องใช้กับ UI host; หลีกเลี่ยงการกระจายชื่อ built-in GUI ทั่ว lexer/parser/type checker การลงทะเบียนต้องเป็น per-run หรือ immutable descriptor ไม่เปิด global mutable window registry
- เพิ่ม diagnostics สำหรับ provider ไม่มี, version ไม่ตรง, handle ไม่ถูกต้อง, invalid size/title, ใช้จาก thread ที่ไม่อนุญาต และการปิด session
- ผลส่งมอบ M1: unit/contract tests ที่สร้าง session, ส่ง command, รับ event, ทำ timeout/cancel, ทดสอบ wrong-session/stale handle และยืนยัน cleanup แบบไม่เปิด UI

### M2 — Windows host ที่เล็กที่สุด

- สร้าง/แสดง/ปิดหน้าต่างหนึ่งบาน, เปลี่ยน title และสีพื้นหลัง, pump OS messages และแปลง input เป็น typed events
- ใช้ Windows APIs เพื่อทำ host โดยตรง; ไม่ใช้ browser, HTML/CSS renderer หรือ WebView และไม่ผูก public Craft API กับ Win32 handles
- รองรับ DPI ของหน้าต่างและพิกัดอย่างสอดคล้อง; resize/DPI change ส่งขนาดและ scale ล่าสุด ไม่ส่งค่า physical pixel ให้ Craft โดยไม่มีการแปลง
- ให้ OS/window resources ถูก cleanup ทั้ง normal close, exception, Ctrl+C และ repeated open/close
- ไม่มี text engine, layout engine, component tree หรือ animation ในรอบนี้ การเปลี่ยนสีพื้นหลังเป็นเพียงการพิสูจน์คำสั่ง Craft → host → display
- ผลส่งมอบ M2: โปรแกรม probe ที่เปิดหน้าต่างจริงได้บน Windows พร้อมบันทึก OS version, DPI/scale และผล close/cancel โดยยังไม่อ้างว่า renderer หรือ performance ตามวิสัยทัศน์เสร็จแล้ว

### M3 — Craft package และตัวอย่างพิสูจน์

- สร้าง `packages/gui` ฝั่ง Craft ที่ห่อ native boundary เป็น API เรียบง่าย พร้อม manifest/version, exported symbols และเอกสาร ownership; package นี้ต้อง check/test/bundle ตามระบบ package ที่มีอยู่
- ฝัง source package รุ่นที่เข้ากันใน CLI เพื่อให้ทั้ง `Craft-setup.exe` และ portable `craft.exe` มี package ชุดเดียวกัน โดยเพิ่ม workflow `craft gui init` สำหรับคัดลอก package เข้าโปรเจกต์และเพิ่ม local dependency/lock โดยไม่ใช้ network; command ต้องไม่เขียนทับ `packages/gui` หรือ manifest/lock ที่ผู้ใช้แก้โดยเงียบ ๆ
- ผู้ใช้ที่สร้าง project หลังติดตั้ง Setup ต้องเริ่มตัวอย่าง GUI ได้โดยไม่ต้องดาวน์โหลด package เพิ่ม; package ไม่ถูก import อัตโนมัติในโปรเจกต์ CLI/API ที่ไม่ใช้ GUI
- ตัวอย่างขนาดเล็กต้องใช้ Craft source จัดการ event loop และ mutable state: เปิดหน้าต่าง, รับ click/key, เปลี่ยน title/สี, ตอบสนอง close request และปิดเองเมื่อโปรแกรมถูกยกเลิก
- ตัวอย่างต้องทำงานจาก source และ bundle ด้วย Craft 0.1.13 บน Windows โดยไม่ต้องติดตั้ง Node.js, browser runtime หรือ WebView
- ผลส่งมอบ M3: `examples/gui-bridge` ที่มี manifest/lock/source/test และคำสั่งจาก project ใหม่จนเห็นหน้าต่าง โดยไม่มี Craft Office หรือ visual designer

### M4 — การตรวจรับและเอกสาร

- จัดทำคู่มือทดลอง Windows, failure matrix, version/compatibility contract และบันทึกผลใน `output/phase1-rev13-status.md`
- ทดสอบ headless contract และ regression ของภาษา/package/CLI; ตรวจ window จริงแยกจากผล headless ไม่ถือว่าการ build หรือ test ผ่านเป็นหลักฐานว่า UI ใช้งานได้
- สร้าง Windows CLI/Setup 0.1.13 และตรวจ installation workflow ตามหลักฐานจริง รวมถึง `craft gui init` จากเครื่องที่ติดตั้งแล้ว, portable CLI, PATH, uninstall และ package files ก่อนระบุว่าการแจกจ่ายผ่านแล้ว
- ผลส่งมอบ M4: เอกสาร API/ข้อจำกัด, คู่มือที่ผู้ใช้ทำตามได้, status ที่ระบุ pass/fail/pending รายกรณี และ Windows artifacts/checksums เมื่อ build จริง

## 5. เกณฑ์รับงาน

- โปรแกรม Craft สร้างหน้าต่าง Windows ผ่าน package, ส่งคำสั่งเปลี่ยน title/สี, รับ click/key/resize/close event กลับใน Craft และปิดได้โดยไม่ค้างหรือทิ้ง process ค้าง
- `craft check` ตรวจ signature และความพร้อมของ provider อย่างสอดคล้องกับ runtime; mismatched types, unsupported provider และ stale handles ได้ diagnostics ที่ชัด ไม่ panic
- การ cancel ระหว่างรอ event, exception ใน Craft, การปิดหน้าต่างซ้ำ และการใช้ handle หลัง close ไม่ทำให้ process ค้างหรือ resource หลุด; headless tests ตรวจ lifecycle และ Windows trial ตรวจพฤติกรรมจริง
- การรัน source และ bundle มีพฤติกรรมเดียวกันบน Craft 0.1.13; bundle ที่ต้องใช้ UI provider ปฏิเสธ runtime ที่ไม่มี provider อย่างชัด และ bundle เก่ากับโปรแกรม non-GUI ยังทำงานตาม compatibility contract
- มีผลตรวจ event ordering, queue bounds/coalescing, Unicode title, key code/modifiers และ DPI/resize ตามกรณีที่ระบุใน M0; บันทึก hardware/Windows version สำหรับผล GUI จริง
- บันทึกสิ่งที่ยังไม่ได้ทดสอบและข้อจำกัดใน status โดยตรง ค่าเริ่มต้นให้ผู้ใช้ทดลอง GUI จริง เว้นแต่ผู้ใช้อนุญาตให้ agent ทดสอบสำหรับงาน implementation นั้น

Acceptance matrix ขั้นต่ำที่ต้องเติมหลักฐานจริงใน status:

| กรณี | ผลที่คาดหวัง | วิธีตรวจ |
| --- | --- | --- |
| Provider ไม่มีหรือ version ไม่ตรง | `check/run` แจ้ง capability ที่ต้องใช้; ไม่เปิดหน้าต่างครึ่งหนึ่ง | Headless fixture + runtime เก่า |
| Invalid size/color/handle หรือ wrong session/thread | Craft exception ที่ระบุสาเหตุ; ไม่มี panic หรือ native resource ค้าง | Headless contract tests |
| `nextEvent(0)`, timeout และ cancel ระหว่างรอ | คืนทันที/คืน `null`/หยุดตาม context ตามลำดับ | Fake clock/event host + Windows trial |
| Click/key/resize/DPI/close | Event ถูกต้อง เรียงลำดับตาม contract และทำให้ตัวอย่างเปลี่ยน title/สี | Headless event tests + manual Windows trial |
| Queue เต็ม | Coalesce เฉพาะ event ที่อนุญาต หรือคืน overflow แล้ว cleanup; ไม่ทิ้ง click/key/close เงียบ ๆ | Flood fixture |
| Exception, `main` คืนค่า, Ctrl+C, close ซ้ำ | Window/session ถูกปิด; process ไม่ค้างและ handle ใช้ต่อไม่ได้ | Headless lifecycle tests + Windows trial |
| `craft gui init` ซ้ำหรือเจอไฟล์ของผู้ใช้ | ไม่เขียนทับ; แสดงวิธีแก้ที่ไม่ทำข้อมูลหาย | CLI fixture |
| Source, bundle, older bundle, non-GUI project | GUI source/bundle ทำงานตรงกัน; ของเดิมไม่ถดถอย | Integration และ regression |
| Portable CLI/Setup หลังติดตั้ง | package source, provider, PATH, uninstall และตัวอย่าง GUI ใช้งานได้ตามเอกสาร | Windows distribution acceptance |

การรัน automated tests ไม่แทนการทดลอง window/input จริง และการเห็นหน้าต่างหนึ่งบานไม่แทนการผ่าน failure matrix ทั้งชุด

## 6. ความเสี่ยงและเงื่อนไขตัดสินใจ

| ความเสี่ยง | วิธีรับมือใน Rev.13 |
| --- | --- |
| Go goroutine ย้าย OS thread แล้ว Win32 message pump ผิด thread | ผูก UI session กับ locked OS thread และมี test/diagnostic ของ wrong-thread calls |
| Craft loop ทำงานหนักจนหน้าต่างไม่ตอบสนอง | เอกสารระบุว่าต้องคืน control ให้ `nextEvent`; วัดการตอบสนองของ probe แต่ยังไม่อ้าง 60/120 FPS |
| Event มาเร็วกว่า Craft อ่าน | จำกัด queue, coalesce เฉพาะรายการที่กำหนด และ fail พร้อม cleanup เมื่อ overflow |
| Package กับ runtime คนละรุ่น | capability/version gate ก่อนรัน และ bundle metadata ที่ runtime เก่าปฏิเสธได้ |
| GUI API ผูกกับ Win32 มากเกินไป | public package ใช้ logical pixels, typed events และ opaque handles; Win32 อยู่หลัง host interface |

หาก M0 พบว่าการเชื่อม UI ต้องเพิ่ม closure, arbitrary FFI หรือเปลี่ยน execution model ของ Craft อย่างกว้าง ให้หยุดขยาย implementation และปรับ contract/ขอบเขตก่อน ไม่ซ่อนงานเหล่านั้นใน Rev.13

## 7. นอกขอบเขต

Craft Office, ชุด UI components, layout, text rendering/editing, IME, accessibility, GPU renderer เต็มรูปแบบ, animation, virtualization, drag & drop ภายในแอป, เครื่องมือออกแบบหน้าจอแบบลากวาง, arbitrary native DLL/FFI, native compiler, standalone application `.exe`, Linux/macOS GUI backend และคำกล่าวอ้าง 60/120 FPS

Rev.13 พิสูจน์ **ความสามารถในการสื่อสารกับ UI host** ก่อน ประสบการณ์ GUI ที่สวยและลื่นตามวิสัยทัศน์ต้องมีแผนและการตรวจรับใน revision ถัดไป
