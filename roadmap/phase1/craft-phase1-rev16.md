# Craft Phase 1 Rev.16 — UI responsiveness และ frame pipeline

วันที่: 6 ตุลาคม 2026  
ฐาน: Craft 0.1.15  
เป้าหมายรุ่น: Craft 0.1.16  
สถานะ: **implementation ใน source แล้ว; รอทดสอบและวัดผลจริง** ดู [สถานะ Rev.16](output/phase1-rev16-status.md)

## เป้าหมายของรุ่น

ปรับ Craft UI Engine ตาม [เป้าหมายใหม่](../craft-gui-direction.md): **Performance first. Simplicity always.** ให้การคลิก การเปลี่ยน state และการวาด frame ใหม่ตอบสนองเร็วและสม่ำเสมอขึ้น โดย API ที่ผู้เขียนแอปใช้ยังเรียบง่าย Craft source ยังคงเป็นจุดเขียน UI และ Windows native host เป็นตัวจัดการ input, layout และ rendering

รุ่นนี้เน้นเส้นทาง **input → event → Craft handler → update → layout/paint → Direct2D EndDraw** และการ resize หน้าต่าง ไม่ใช้ตัวเลข RAM หรือ CPU idle เป็นตัวแทนความไว ไม่ประกาศว่าเร็วเท่า React หรือรองรับ 60/120 FPS ทุก workload จนกว่าจะวัดบนงานและเครื่องที่เทียบกันได้

## ปัญหาที่พบในฐาน 0.1.15

1. UI host ตรวจ Go command และ Win32 message ด้วย `PeekMessage` แล้วนอน 8 ms ในทุก loop จึงอาจเพิ่มเวลารอก่อนรับ input หรือ update แม้เครื่องยังมีทรัพยากรเหลือ
2. `gui.update` สร้างและ clone tree ใหม่ จากนั้น relayout ทั้ง tree; การเปลี่ยนข้อความเพียงจุดเดียวมีงานมากกว่าที่จำเป็น
3. Direct2D path ใช้ `Clear` และเดินวาดทั้ง tree ทุกครั้ง แม้มีเพียง hover/focus หรือ node เดียวเปลี่ยน; การ invalidate เฉพาะ rect ยังไม่เท่ากับการวาดเฉพาะ rect
4. ยังไม่มีตัวเลขแยกเวลารอ event, เวลาใน Craft handler, layout, paint และเวลาถึง `EndDraw` จึงยังบอกไม่ได้ว่าจุดใดทำให้หน่วง
5. layout แบบแบ่งพื้นที่เท่ากันทำให้ปุ่มและข้อความกินพื้นที่เกินเนื้อหา; ยังไม่มีงานวัดข้อความเพื่อสร้าง layout ที่ใช้งานจริงอย่างเหมาะสม แต่รุ่นนี้ให้ความสำคัญกับเส้นทางตอบสนองก่อน

## ขอบเขต implementation

### A. วัดก่อนปรับ

- เพิ่มโหมด diagnostics สำหรับการพัฒนาเท่านั้น เช่น flag/environment ที่ปิดเป็นค่าเริ่มต้น บันทึก timestamp ด้วย clock แบบ monotonic และตัวนับ โดยไม่พิมพ์ทุก frame ในเส้นทางปกติ
- แยกช่วงเวลา: Win32 input เข้า host, ส่ง event ให้ Craft, เริ่ม/จบ `gui.update`, layout, paint, `EndDraw`; ใช้ trace ID ภายในเพื่อจับคู่ input กับ update/frame ที่เกิดจาก input นั้นโดยไม่เปลี่ยน public event API และระบุชัดว่า `EndDraw` **ไม่ใช่** เวลาที่พิกเซลปรากฏจริงบนจอ
- ทำตัวอย่าง workload ขนาดเล็กที่รันซ้ำได้: คลิกปุ่มเปลี่ยนข้อความ, hover/focus, resize ต่อเนื่อง และ update tree จำนวนหลายระดับ/หลาย node ไม่ใช้ Craft Office
- เก็บ baseline 0.1.15 บนเครื่องอ้างอิงเดียวกันก่อนแก้ และรายงาน p50/p95/p99, frame time, CPU/RAM พร้อมจำนวน node, DPI, renderer, refresh rate และจำนวนรอบวัด

### B. ปลุก host ตามเหตุการณ์

- แทนการนอน 8 ms แบบ polling ด้วยกลไก wake ที่ผูกกับ Win32 message queue และคำสั่งจาก Go โดยไม่ให้ channel command ค้างเมื่อ UI thread รอ message
- กำหนด ownership ของ wake handle/message และการปิด session ให้ชัด; cancellation, close, resize และ device loss ต้องปลุก host ได้ ไม่เกิด deadlock หรือ busy loop
- ตรวจลำดับ input กับ command และรักษา event สำคัญ เช่น click, key, close โดยไม่ทิ้งหรือจัดลำดับผิด ส่วน mouse move/resize ที่ถี่มากอาจ coalesce ได้เมื่อไม่ทำให้ state สุดท้ายผิด
- ต้องวัด latency ก่อนและหลังแทนการสรุปจาก CPU idle; หากวิธีใหม่ลด CPU แต่ทำให้ p95/p99 แย่ลง ให้แก้ scheduler ก่อนรับงาน

### C. จำกัดงาน update และ frame

- ใช้ node ID และชนิด node เพื่อตรวจ subtree ที่เปลี่ยน; เก็บ layout/paint data ของ subtree ที่ไม่เปลี่ยน โดยกำหนดกฎ invalidate เมื่อข้อความ ขนาด parent, DPI หรือ style กระทบ geometry
- แยก dirty flags อย่างน้อยเป็น layout, paint และ interaction; hover/focus ควรทำ paint เท่านั้น ส่วนการเปลี่ยนข้อความให้ relayout เฉพาะบรรพบุรุษที่ขนาดขึ้นกับเนื้อหา
- รวม dirty rect ที่ทับกัน และตั้ง clip ให้ตรงกับพื้นที่ที่วาดจริงบน Direct2D; หลีกเลี่ยง `Clear` ทั้ง target ใน frame ที่เปลี่ยนเฉพาะจุด โดยรักษาความถูกต้องของพื้นหลังและลำดับการซ้อน
- กำหนด threshold สำหรับ fallback เป็น full layout/paint เมื่อ dirty rect มากหรือเปลี่ยนทั้งหน้าต่าง เพื่อไม่ให้การจัดการ dirty region แพงกว่าการวาดใหม่
- ให้ GDI fallback วาดผลถูกต้องตาม dirty region หรือ full repaint ที่กำหนดไว้ ไม่ปล่อยภาพค้างต่างจาก Direct2D

### D. Developer experience และ compatibility

- รักษา `gui.openWindow`, `mount`, `update`, `nextEvent`, `renderer` และ source bundle เดิมให้ทำงานได้; รุ่นนี้ไม่บังคับเปลี่ยน API เพื่อรับประโยชน์จาก scheduler และ renderer ใหม่
- แอปตัวอย่างยังเขียนด้วย Craft เท่านั้น; ผู้สร้างแอปไม่ต้องจัดการ dirty rect, Direct2D resource หรือ scheduler เอง
- อัปเดต `packages/gui`, template, docs, CLI/version metadata และ installer เฉพาะเมื่อ implementation ผ่านเกณฑ์ ไม่เปลี่ยนแผนให้กลายเป็นการปล่อย V1.00

## ลำดับทำงาน

1. สร้าง instrumentation และเก็บ baseline 0.1.15 พร้อม workload ที่ทำซ้ำได้
2. เปลี่ยน host เป็น event-driven wake; ทดสอบ close/cancel และวัด input latency ซ้ำ
3. เพิ่ม dirty flags, subtree reuse และ clip/partial paint; ตรวจภาพหลัง hover, update, resize และ DPI change
4. ตรวจ compatibility, package/bundle, build/installer แล้วบันทึกผลจริงและข้อจำกัดใน status Rev.16

## เกณฑ์ตรวจรับ

| ด้าน | เกณฑ์ |
| --- | --- |
| การตอบสนอง | บนเครื่องอ้างอิงเดียวกัน p95/p99 ของเส้นทาง input ถึง `EndDraw` สำหรับ workload คลิก/เปลี่ยนข้อความต้องไม่แย่กว่า 0.1.15 และต้องรายงานค่าจริง; ตั้งเป้า p95 ไม่เกินหนึ่ง refresh interval ที่ 60 Hz และ p99 ไม่เกินสอง interval สำหรับ workload เล็ก หากไม่ถึงต้องระบุสาเหตุและยังไม่อ้างว่าผ่านเป้าหมาย |
| ความลื่น | ระหว่าง resize/update ต่อเนื่องรายงาน frame time และจำนวน frame ที่เกินงบ 16.7 ms ที่ 60 Hz; ไม่มีภาพค้าง เส้นรอยวาด หรือการกระพริบจาก partial paint |
| งานที่หลีกเลี่ยง | Hover/focus ไม่กระตุ้น relayout; update ข้อความหนึ่ง node ไม่จัด layout/วาดทุก node โดยไม่จำเป็น; เมื่อหน้าต่างนิ่งไม่มีการ paint วน |
| ความถูกต้อง | Click, drag-out cancellation, Tab/Shift+Tab, Enter/Space, close, DPI change และ device-loss/fallback ยังทำงาน; event สำคัญไม่หายและไม่มี deadlock |
| ข้อความ/layout | การเปลี่ยนข้อความและ resize ไม่ทำให้ตำแหน่งวาดกับ hit test ไม่ตรงกันบน DPI ที่รองรับ; behavior เดิมยังใช้ได้ |
| Compatibility | ตัวอย่าง GUI Rev.13–15, `craft check/run/build`, source bundle, `go test ./...`, Windows x64 build และ installer ไม่ถดถอย |

การทดสอบ input ถึง `EndDraw` วัดเวลาภายใน engine ส่วนการยืนยัน **input ถึงภาพปรากฏจริง** ต้องใช้วิธีวัดภายนอกที่เหมาะสมและรายงานแยก ห้ามนำสองค่านี้ไปเรียกชื่อเดียวกัน ผลเทียบ React จะทำเฉพาะเมื่อมีแอปตัวอย่างและวิธีวัดเดียวกัน

## งานนอกขอบเขต 0.1.16

- Craft Office และเครื่องมือออกแบบหน้าจอแบบลากวาง
- Text measurement และ layout แบบ content/min-max/fill, animation engine เต็มรูปแบบ, ScrollView/virtualization, TextBox/IME/editing, blur/shadow ครบชุด และ accessibility เต็มรูปแบบ
- Linux/macOS GUI backend, binary แอปปลายทางที่รันโดยไม่ใช้ Craft CLI และการประกาศ V1.00

รุ่นนี้เป็น **ฐานความไวของ UI Engine** ไม่ใช่การประกาศว่า Framework สำหรับแอป GUI ทุกรูปแบบเสร็จแล้ว
