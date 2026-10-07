# Craft Phase 1 Rev.14 — UI Engine แนวตั้งเส้นแรก

วันที่: 6 ตุลาคม 2026  
ฐาน: Craft 0.1.13 / Phase 1 Rev.13  
เป้าหมายรุ่น: Craft 0.1.14  
สถานะ: เริ่ม implement แล้ว; vertical slice พื้นฐานอยู่ใน source แต่ยังไม่ผ่านเกณฑ์รับงานทั้งหมด ดู [สถานะ Rev.14](output/phase1-rev14-status.md)

## 1. เป้าหมาย

ให้โปรแกรม Craft บน Windows สร้าง **UI tree ขนาดเล็กที่วางตำแหน่ง วาด และรับ input ได้จริง** ผ่าน `gui` package ที่เลือกติดตั้งได้ ผลพิสูจน์คือหน้าต่างหนึ่งบานที่มีข้อความและปุ่ม: resize แล้ว layout ปรับ, คลิกปุ่มแล้วข้อความเปลี่ยน, redraw เฉพาะเมื่อ state เปลี่ยน และปิดแล้ว resource ถูกคืนครบ

Rev.14 เป็นก้าวแรกของ Craft UI Engine ไม่ใช่ชุด component ครบถ้วนหรือ Craft Office ภาษา Craft ยังคงเป็นภาษาทั่วไป; โปรเจกต์ CLI/API ที่ไม่ import GUI ไม่ต้องสร้าง UI host ไม่มี WebView, HTML/CSS, JavaScript หรือเครื่องมือออกแบบหน้าจอแบบลากวาง รุ่นนี้ยังเป็น Phase 1 จนกว่าผู้ใช้จะตัดสินใจออก V1.00

```text
Craft application
  → packages/gui (public Craft API)
  → typed UI provider / session
  → UI tree → measure/layout → paint commands → Windows renderer
                         ↘ hit test → typed event → Craft
```

## 2. ฐานที่ต้องยอมรับตามจริง

- Rev.13 เปิดหน้าต่าง รับ click/key และเปลี่ยน title/สีได้ทั้ง source และ bundle ใน Windows smoke test ดู [ผล Rev.13](output/phase1-rev13-status.md)
- Rev.13 ยังไม่มี UI tree, layout, text, hit testing, component, GPU renderer หรือหลักฐาน performance
- งาน Rev.13 ที่ยังไม่ผ่านเกณฑ์เต็ม ได้แก่ DPI/resize, cancel, overflow, invalid handle, installation/uninstallation และการใช้งานจริงด้วยคน; ต้องตรวจและแก้ส่วนที่กระทบ engine ก่อนพึ่ง bridge นั้น
- `craft build` ยังได้ source bundle ที่ต้องใช้ Craft CLI; Rev.14 ไม่เปลี่ยนเป็น executable ของแอปผู้ใช้
- Windows 10+ x64 เป็น platform สำหรับ implementation รอบนี้ API สาธารณะต้องไม่เปิด Win32 handle เพื่อให้เพิ่ม backend อื่นภายหลังได้

## 3. ผลส่งมอบขั้นต่ำและขอบเขต

### 3.1 API สำหรับผู้เขียน Craft

ให้ `packages/gui` มี API แบบประกาศ tree แล้วส่ง state ใหม่อย่างชัดเจน เช่น `gui.mount(window, root)` และ `gui.update(window, root)` ชื่อและ signature สุดท้ายต้องตรึงใน M0 โดยอิง syntax/type ของ Craft ที่มีจริง ไม่มี JSX, hooks หรือ closure เป็นเงื่อนไขของ API หลัก

ชนิด node ขั้นต่ำ: `Column`, `Row`, `Panel`, `Text`, `Button` พร้อม stable ID, ขนาด, padding, gap, alignment, สีพื้นหลัง และข้อความ ปุ่มส่ง `ButtonPressed(id)` กลับสู่ event loop ของ Craft; Craft เป็นเจ้าของ application state และสร้าง tree รุ่นใหม่หลัง event ไม่ให้ native renderer เรียก callback เข้า interpreter จาก OS procedure

API ต้องรองรับการอัปเดตข้อความ/สี/สถานะปุ่มที่มี ID เดิมโดยไม่สร้าง native window ใหม่ กำหนด error เมื่อ ID ซ้ำ, node ผิดชนิด, tree ใหญ่เกินขอบเขต, ใช้ session/handle หลังปิด และค่าขนาดที่ใช้ไม่ได้

### 3.2 Layout และ input

- Layout แบบสองขั้น `measure` และ `arrange` บน logical pixels; `Column`/`Row` รองรับ fixed, content และ fill อย่างจำกัด โดยกำหนด min/max constraint, padding, gap และ alignment ชัดเจน
- กำหนดผลของพื้นที่ไม่พอ, ข้อความยาว, node ซ้อนกัน และขนาดศูนย์โดยไม่เกิด negative size, overflow หรือ infinite layout loop
- Resize และ DPI change ต้องคำนวณ layout ใหม่อย่างถูกต้อง; พิกัด pointer ถูกแปลงเป็น logical pixels ก่อน hit test
- Hit testing เลือก node ที่มองเห็นและรับ input ได้ตามลำดับการวาด ปุ่มมี hover/pressed state ขั้นต่ำ และรับ activation จากเมาส์กับคีย์บอร์ดเมื่อมี focus
- Event ที่ไม่เกี่ยวกับ node ยังคงส่งให้ Craft ตามสัญญา Rev.13; กำหนดลำดับ event ระหว่าง resize, pointer, focus และ close

### 3.3 Rendering และ text ขั้นต้น

- แยก scene/paint commands ออกจาก Windows renderer: rectangle, solid fill, clip, border, rounded corner และ glyph run ขั้นต่ำ
- เลือก Windows rendering backend ใน M0 โดยให้แนวทางหลักเป็น Direct2D/DirectWrite เพื่อใช้ hardware acceleration เมื่อมี และกำหนด fallback/diagnostic ที่ชัดเมื่อสร้าง device ไม่ได้; ห้ามอ้างว่า GPU ทำงานจริงจนกว่าจะตรวจ backend และเครื่องที่ใช้
- Static text ต้องวัดและวาดด้วยระบบฟอนต์เดียวกัน รองรับ Unicode พื้นฐาน, DPI scaling, สีและขนาดฟอนต์; ระบุพฤติกรรม fallback font และข้อความล้นพื้นที่
- ทำ invalidation เมื่อ tree, layout หรือ visual state เปลี่ยน; render เฉพาะเมื่อจำเป็นและไม่วน frame ตลอดเวลาโดยไม่มี animation
- มี debug counters สำหรับจำนวน layout pass, paint pass และพื้นที่ที่ invalidated เพื่อใช้วัดจริง ไม่ตั้งเป้า 60/120 FPS จากตัวอย่างนิ่งเพียงอย่างเดียว

## 4. สัญญาและการตัดสินใจที่ต้องตรึงใน M0

ก่อนเพิ่ม primitive ใน runtime ให้เขียน contract เดียวที่ checker, runtime และ package ใช้อ้างอิง ระบุอย่างน้อย:

| ประเด็น | สิ่งที่ต้องตัดสิน |
| --- | --- |
| การสร้าง tree | รูปแบบ Craft source ที่ check ได้จริง, immutable/copy semantics และ stable ID |
| ขนาด | หน่วย logical pixel, fixed/content/fill, min/max, การปัดเศษ และ DPI |
| การ update | replace tree หรือ patch, ผลเมื่อ ID หาย/ซ้ำ, ownership และเวลาที่ commit |
| Event | payload ของ `ButtonPressed`, pointer/focus, ordering, queue limit และ overflow |
| Text | font fallback, การวัด, Unicode, clipping/ellipsis และข้อจำกัดของ Rev.14 |
| Renderer | Direct2D/DirectWrite หรือทางเลือก, device loss/recreate, fallback และ resource cleanup |
| Capability | version gate ของ GUI package/provider, source bundle และ runtime เก่า |

ให้ทดลอง `craft check` กับ API draft ก่อนตรึงชื่อจริง หาก syntax ของ Craft ยังสร้าง tree อย่างอ่านง่ายไม่ได้ ให้ปรับ API ใน package ก่อนพิจารณาเพิ่ม feature ของภาษา

## 5. ลำดับงาน

### M0 — Bridge hardening และ contract

1. เติม headless host ที่ inject ได้เพื่อทดสอบคำสั่ง/event/cleanup โดยไม่เปิดหน้าต่าง และรวม signature/capability version ของ provider ให้ checker กับ runtime ตรงกัน
2. แก้หรือบันทึกข้อจำกัด Rev.13 ที่กระทบ engine: resize/DPI, stale handle, cancel, queue overflow, repeated close และ error ที่ไม่มี provider
3. ตรึง API/tree/layout/event/render contract ตามหัวข้อ 4 พร้อมตัวอย่าง Craft source สั้นที่ `craft check` ได้

**Gate:** ไม่มี crash/resource leak ที่รู้แล้วใน bridge path และ API draft ผ่าน checker ก่อนเริ่ม renderer

### M1 — UI tree และ layout แบบ headless

1. เพิ่มโครงสร้าง node ที่มี stable ID และตรวจ tree ก่อน commit โดยไม่ส่ง pointer Go/Win32 ออกให้ Craft
2. Implement measure/arrange ของ `Row`, `Column`, `Panel`, `Text`, `Button` ภายใต้ constraints ที่กำหนด
3. เพิ่ม scene diff/invalidation และ hit testing โดยแยกจาก Windows API
4. ทำ fixtures สำหรับ nested layout, resize, overflow, ID ซ้ำ, hidden/disabled node และ DPI 100%/150%/200%

**Gate:** layout และ hit test ได้ผล deterministic ใน headless tests; update tree ไม่ทำลาย window/session

### M2 — Windows paint และ text

1. สร้าง renderer ตาม backend ที่เลือกใน M0 และผูกกับ window/session เดิม
2. วาด panel, text, button และ state hover/pressed; จัดการ clip, redraw, device loss และ DPI change
3. ตรวจขนาดข้อความกับพื้นที่วาดจริง และ cleanup ของ font, brush, target และ callback

**Gate:** probe Windows เปิดได้, resize แล้วไม่เหลือภาพเก่า, ข้อความอ่านได้บน DPI ที่ทดสอบ และปิดได้โดยไม่มี process ค้าง

### M3 — Craft package, ตัวอย่าง และ distribution

1. เปิด API ผ่าน `packages/gui` และ `craft gui init` รุ่น 0.1.14 พร้อม migration/compatibility จาก package 0.1.13; ห้ามเขียนทับ package ที่ผู้ใช้แก้เอง
2. เพิ่ม `examples/gui-layout` เป็นหน้าต่างเล็กที่มี `Column`, `Text`, `Button`; กดปุ่มแล้ว Craft เปลี่ยน state และ update tree; resize แล้วจัดวางใหม่
3. ตรวจ source/bundle, CLI/Setup/portable package, version/capability gate และ checksum ของ artifacts

**Gate:** ผู้ใช้ทำตามคู่มือจาก project ใหม่จนเห็น layout และกดปุ่มได้โดยไม่ต้องติดตั้ง Node.js/WebView

## 6. เกณฑ์รับงานและหลักฐาน

| กรณี | ผลที่ต้องเห็น | หลักฐาน |
| --- | --- | --- |
| Craft source | `craft check` ผ่านและ `craft run` แสดง text/button; กดแล้วข้อความเปลี่ยน | CLI output + Windows manual run |
| Layout/resize/DPI | ตำแหน่งและขนาดตรง contract ที่ 100%, 150%, 200%; ไม่มีภาพค้าง | Headless layout fixtures + ภาพ/บันทึกจาก Windows |
| Hit test/focus | คลิกเฉพาะในปุ่มจึงได้ `ButtonPressed`; Tab/Enter ใช้ได้ตาม contract | Headless event cases + Windows manual run |
| Tree update | เปลี่ยนข้อความ/สีโดย window เดิม; ID ซ้ำหรือ invalid node ได้ error | Unit/contract tests + sample |
| Renderer/text | Unicode พื้นฐานแสดงครบ, clipping ถูก, device loss/fallback มีผลที่ระบุ | Backend tests + Windows manual run |
| Lifecycle | close, exception, cancel และ repeated close คืน resource; ไม่มี process ค้าง | Headless tests + Windows run |
| Distribution | source/bundle ทำงานตรงกัน; CLI/Setup มี package/provider รุ่นเดียวกัน | Integration + installation trial |
| Regression | โปรแกรม CLI/API และ bundle เก่าไม่ถดถอย | `go test ./...` และ compatibility fixtures |

บันทึกผลเป็น `roadmap/phase1/output/phase1-rev14-status.md` โดยแยก **ผ่าน / ไม่ผ่าน / ยังไม่ได้ทดสอบ** และบอก Windows version, DPI, rendering backend ที่ใช้จริง การ build สำเร็จไม่ใช่หลักฐานว่า GPU, layout หรือ interaction ผ่าน

## 7. ขอบเขตที่ยังไม่ทำ

Craft Office, visual designer แบบลากวาง, TextBox/TextArea และ IME, complex text editing, ScrollView/virtualization, animation, blur/shadow, accessibility เต็มรูปแบบ, drag and drop, Linux/macOS backend, arbitrary FFI, standalone application executable และการรับประกัน 60/120 FPS

Rev.14 ต้องพิสูจน์ vertical slice ที่ใช้งานได้ก่อนขยาย component library ถ้า bridge หรือ text/render backend ยังไม่เสถียร ให้บันทึกข้อจำกัดและไม่ประกาศว่า UI Engine พร้อมใช้งานทั่วไป
