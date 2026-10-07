# Craft Phase 1 Rev.24 — Paint Primitives และ Custom Drawing Foundation

วันที่: 7 ตุลาคม 2026  
เป้าหมายรุ่น: Craft 0.1.24  
สถานะ: **implement Canvas รุ่นแรกใน source แล้ว; รอผู้ใช้ทดสอบตาม `output/phase1-rev24-status.md`**  
ฐานเปรียบเทียบ: Craft UI Runtime source หลัง 0.1.23 และเอกสาร **Craft UI Runtime Requirements** ที่ผู้ใช้ส่งมา

## ขอบเขต Phase 1 ในรุ่นนี้

Phase 1 ครอบคลุมทั้ง Craft Runtime และ package ที่เขียนด้วย Craft. Rev.24 เพิ่ม Canvas/paint primitives ใน Runtime และใช้ความสามารถนี้เป็นฐานให้พัฒนา `packages/gui` **0.2.1** ภายใน Phase 1 เดียวกัน. งาน package แยกหัวข้อและผลตรวจรับจาก Runtime เพื่อเห็นชัดว่าส่วนใดพร้อมแล้ว โดยไม่ได้หมายความว่า Craft เปลี่ยนเป็นเวอร์ชัน 0.2.1.

### การพัฒนา package บนฐาน Rev.24

- นำ `packages/gui` กลับมาเป็น library ที่เลือกใช้ได้ ห่อ `std.ui` ที่ Runtime 0.1.24 เปิดให้ Craft ใช้ และเพิ่ม composition ขั้นต้น เช่น card, toolbar, navigation, field.
- ตัวอย่าง `examples/gui-package-lab` ใช้ package ผ่าน dependency ของ Craft; package มี manifest/lockfile และคู่มือแยกจาก Runtime.
- Package version `0.2.1` ใช้ระบุ dependency เท่านั้น; Craft CLI/Runtime version สำหรับฐานนี้คือ `0.1.24`.
- รายละเอียด implementation อยู่ใน [แผน package 0.2.1](packages/gui-0.2.1-plan.md) และ [สถานะ](output/gui-0.2.1-status.md) ภายใต้ Phase 1 Rev.24.

## ข้อสรุปจากการเทียบ

ข้อกำหนดที่แนบมาเป็น **ภาพปลายทางของ Runtime** สำหรับรองรับแอปขนาดใหญ่ ไม่ใช่รายการที่ 0.1.23 ต้องเสร็จทั้งหมด. ปัจจุบัน `std.ui` มีหน้าต่าง Windows หนึ่งบานต่อ interpreter run, UI tree, Row/Column/Panel/ScrollView/SplitView, sizing/spacing/alignment, VirtualList แนวตั้ง, Text/Button/TextInput, event ขั้นต้น, Direct2D/DirectWrite พร้อม GDI fallback และการ update บางส่วน. 0.1.23 เพิ่งเพิ่ม layout constraints กับ scrollbar แนวตั้ง; [สถานะ 0.1.23](output/phase1-rev23-status.md) บันทึกผล build/check/regression และภาพ Direct2D ที่ DPI 125% แล้ว แต่ยังไม่มีหลักฐานตรวจครบทุก interaction, IME, fallback และ latency. ตารางนี้จึงเป็นการเทียบความสามารถใน source ร่วมกับผลตรวจที่ระบุชัด ไม่ใช่การรับรองทุกพฤติกรรม.

| ระบบตาม Requirements | ที่มีใน 0.1.23 | ช่องว่างหลัก |
| --- | --- | --- |
| 1. Window System / Platform | Native Windows window, title/background, resize/DPI event; หนึ่งหน้าต่างต่อ run | Multi-window, dialog/modal/popup/tooltip/context menu, fullscreen, API สำหรับ monitor, file dialog, clipboard ทั่วไป และ system integration; Linux/macOS GUI |
| 2. Rendering Engine | วาด Text/Button/Panel และ scrollbar ผ่าน Direct2D/DirectWrite หรือ GDI fallback | Public primitives สำหรับ rectangle/rounded rectangle/line/path/image/icon/border/shadow/gradient/transform/opacity, style และ custom scene; renderer parity/asset cache |
| 3. Layout Engine | Row/Column/Panel, fixed/preferred/min/max/weight, padding/gap/align, horizontal SplitView, scroll/clip บางส่วน | Grid/Stack/absolute/relative, margin/justify, grow/shrink policy ที่ชัดเจนทุกกรณี, draggable splitter, horizontal scrollbar/overflow policy; ยังรอ DPI/resize verification |
| 4. State / Reactivity | `mount`/`update` ด้วย node ID, รักษา input/scroll state บางส่วน, rebind/dirty paint บางกรณี | Observable state/dependency tracking, targeted mutation, partial layout ที่รับประกันและวัดได้; ตอนนี้ Craft app ต้องสร้าง tree ใหม่เองเมื่อ state เปลี่ยน |
| 5. Virtualization | `VirtualList` แนวตั้งแถวสูงคงที่และ visible-range event | 2D viewport/VirtualGrid, variable row height, column virtualization, sticky/frozen region, selection/navigation ระดับเซลล์ |
| 6. Text Engine | DirectWrite แสดง/วัดข้อความ, native Windows EDIT รับข้อความบรรทัดเดียวพร้อม IME/caret/selection/clipboard | Custom text shaping/selection API, wrapping/หลายบรรทัด, rich text, font fallback control, undo/redo policy และการตรวจ Thai/CJK/emoji/grapheme อย่างเป็นระบบ |
| 7. Input / Event | Click/button, key down, resize/DPI, scroll, text/focus events; hit testing, pointer capture เฉพาะงาน scroll | Node-targeted pointer down/up/move/double click, key up, propagation/cancel, shortcut routing, drag/drop และ event contract สำหรับ custom Canvas |
| 8. Animation / Compositor | GPU drawing path และ dirty painting บางส่วน | Timeline/easing, composition tree/layers, transform/opacity animation โดยไม่ rebuild tree และ input-to-present instrumentation |
| 9. Custom Drawing API | Renderer ภายในวาด control ที่มีอยู่ | Craft app ยังส่ง draw commands เองไม่ได้; ไม่มี Canvas, image asset หรือ scene primitive สำหรับ chart/diagram/editor |
| 10. Component System | Runtime มี Button/TextInput ขั้นต้น; package GUI ที่เคยทดลองถูกเก็บใน Archive | Reusable widget/component และ design system เป็นหัวข้องาน package ใน Phase 1; Runtime ต้องเปิด primitive กับ event พอให้สร้างได้ โดยไม่ผูกกับ Fluent/Material/Office |

ตัวอย่าง syntax ใน Requirements เช่น `Window { ... }`, `state(0)`, `animate(...)` และ `Canvas { ... }` เป็น **แนวคิด** ยังไม่ใช่ Craft API ที่ใช้ได้. รอบ implement ต้องอ้างอิง syntax/type checker จริงก่อนเผยแพร่ตัวอย่าง.

## การตัดสินใจสำหรับ 0.1.24

โฟกัส **Rendering Engine + Custom Drawing API รุ่นแรก**. เหตุผล: หลัง 0.1.23 จัดพื้นที่ได้ดีขึ้น แต่ผู้พัฒนายังวาด UI อิสระไม่ได้. หากไม่มี paint primitives และ Canvas การสร้าง widget ใหม่, chart, editor หรือ design system จะติดอยู่กับ Text/Button/Panel และต้องเพิ่ม control เฉพาะทางเข้า Runtime เรื่อย ๆ. 0.1.24 จะสร้างสัญญาวาดภาพที่เล็ก ชัดเจน และใช้ backend เดียวกับ UI tree ปัจจุบัน. ไม่สร้าง Craft Office, ERP, spreadsheet หรือ GUI package สำเร็จรูป.

### ผลลัพธ์ที่ต้องส่งมอบ

1. **Retained Canvas node:** Craft ส่งรายการ draw commands แบบข้อมูลผ่าน API ที่ type checker รู้จัก; Runtime clone/validate/bound รายการและผูกกับ node ID. `mount`/`update` เปลี่ยน Canvas โดยไม่เปิด imperative drawing callback บน UI thread. สัญญาต้องใช้ Craft syntax จริง; ชื่อ API เช่น `std.ui.canvas(...)` และ `std.ui.drawRect(...)` เป็นชื่อเสนอจนกว่าจะตรวจข้อจำกัดของภาษา.
2. **Primitive ขั้นต่ำ:** fill rectangle, rounded rectangle, stroke rectangle/line, text และ image bitmap พร้อมตำแหน่ง/ขนาด/สี. มี clip ตาม Canvas/parent viewport และ alpha สำหรับสีพื้นฐาน. กำหนดลำดับการวาด, coordinate system, DPI scaling และ behavior เมื่อขนาดติดลบ/เกินขอบ. Path, gradient, shadow และ arbitrary transform ยังไม่จำเป็นในรุ่นแรก แต่ schema ต้องไม่ปิดทางเพิ่มภายหลัง.
3. **Image asset lifecycle:** โหลดไฟล์ภาพที่ระบุอย่างชัดเจนจาก Craft project/bundle, decode อย่างมีขอบเขต, cache ตาม path/ขนาด/การเปลี่ยนไฟล์ และปล่อย resource เมื่อ node/window ปิด. กำหนดชนิดที่รองรับจริงก่อน implement เช่น PNG/JPEG; รายงาน error แทนภาพว่างเงียบ ๆ. ไอคอนสามารถเริ่มจาก image asset; ยังไม่อ้างว่ารองรับ SVG.
4. **Renderer parity:** Direct2D/DirectWrite เป็นเส้นทางหลัก; GDI fallback แสดง geometry, clipping, text และ bitmap ให้เข้าใจตรงกัน. Device lost/recreate ต้อง rebuild GPU resources จาก retained commands โดยไม่ให้ Craft app ส่งคำสั่งซ้ำ.
5. **Integration กับ layout/event:** Canvas มี preferred/min/max size ตาม API 0.1.23, clip กับ nested scroll ได้, hit test รายงานอย่างน้อย Canvas node ID พร้อม local coordinates. กำหนด event ordering เมื่อคลิก Canvas กับ scroll/resize โดยไม่ทำลาย button/TextInput เดิม. Per-shape hit testing/propagation เต็มรูปแบบเป็นงานถัดไป.
6. **Performance bounds:** จำกัดจำนวน commands/ขนาดภาพ/หน่วยความจำต่อ Canvas และให้ error ที่ชัดเจน. Cache text/image metrics และแยก remeasure/repaint: เปลี่ยนสีหรือ primitive ที่ไม่เปลี่ยนขนาดไม่ควรบังคับ layout ทั้ง tree. เก็บ trace จำนวน draw commands, dirty area, paint time และเวลาจาก input ถึง draw completion; ห้ามอ้าง input-to-present หากยังวัดไม่ได้.

### สัญญา API ที่ต้องออกแบบก่อนลง backend

- เพิ่มชนิด draw command ที่ Craft เก็บใน array ได้ โดยข้อมูลเป็น value/immutable และไม่ถือ native handles. คำสั่งอ้าง asset ผ่าน ID/path ที่ Runtime จัดการ ไม่ให้ Craft ส่ง pointer.
- ระบุ color/alpha แบบเดียวกันทุก backend, logical pixel coordinate และ clipping. ข้อความวาดผ่าน text engine เดิมเพื่อให้ measurement กับภาพตรงกัน.
- นิยามการแก้ Canvas ด้วย `std.ui.update` และ node ID เดิม; การแก้รายการ commands ไม่ล้าง focus/scroll ของ sibling. กำหนด error และขีดจำกัดก่อนเผย API.
- แยก `Canvas` ซึ่งเป็น Runtime primitive ออกจาก `Button`, `Tabs`, `Chart` หรือ theme ซึ่งเป็น component/library. ไม่ใส่ application-specific data model ลง Runtime.

## ลำดับดำเนินงาน

1. รับผลทดสอบ 0.1.23 หากมี และแก้ regression ของ layout/scroll ที่เป็น blocker ก่อนเริ่ม Canvas; หากยังไม่มี ให้ระบุฐานว่าไม่ผ่านการตรวจรับด้วยตา.
2. ออกแบบ type/API, ขีดจำกัด, error contract และตัวอย่าง Craft ที่ checker รับได้; ตรวจวิธี resolve image asset จาก project/bundle.
3. ทำ retained command list กับ scene diff/dirty region และการวัดขนาด Canvas.
4. ทำ Direct2D/DirectWrite rendering, image decode/cache และ device-loss recovery; ตามด้วย GDI fallback.
5. เชื่อม clip/scroll/DPI, Canvas hit testing/event, resource cleanup และ trace.
6. เตรียม **UI Paint Lab** ที่เป็นภาพทดสอบกลาง ๆ: geometric shapes, Thai/English text, images, Canvas หลายชุดใน split/scroll pane และการ update เฉพาะบาง node. ไม่ทำหน้าจอ Craft Office.
7. เขียน API guide, manual test guide และ status ที่แยกสิ่งที่ตรวจจาก source กับสิ่งที่ผู้ใช้ยืนยันจาก GUI จริง.

## เกณฑ์ตรวจรับเมื่อ implement

| หัวข้อ | สิ่งที่ผู้ใช้/ผู้พัฒนาตรวจได้ |
| --- | --- |
| API | ตัวอย่าง Craft `check` ได้; invalid command/asset คืน error ชัดเจน; API เดิมยังใช้ได้ |
| ภาพ | รูปทรง เส้น ข้อความไทย/อังกฤษ และ PNG/JPEG ตรงพิกัด/clip ทั้ง hardware และ fallback; alpha ไม่ทำให้สีผิดหรือภาพค้าง |
| Layout/interaction | Canvas อยู่ใน Row/Column/SplitView/ScrollView ได้; resize, DPI และ scroll แล้วภาพและ local click coordinates ตรงตำแหน่ง |
| Lifecycle | update เปลี่ยน command list ได้; ปิด window และ device recreation ไม่ทำให้ resource leak หรือภาพหาย |
| Performance | รายงานจำนวน command, image memory/cache, p50/p95/p99 ของ layout/paint และ draw completion บน workload เดียวกัน; ระบุเครื่อง/renderer/DPI/ขนาดหน้าต่าง |
| Regression | Button, TextInput/IME, ScrollView, VirtualList และ fallback เดิมไม่เสีย |

การเปิด GUI, ตรวจภาพและความรู้สึกลื่นให้ผู้ใช้ทดสอบตามความต้องการเดิม เว้นแต่ผู้ใช้ขอให้ผู้ช่วยทดสอบในงานนั้นโดยตรง. อย่าระบุว่าเกณฑ์ใดผ่านก่อนมีหลักฐานจริง.

## สิ่งที่ยังขาดหลัง 0.1.24 และลำดับเสนอ

1. **State/reactivity และ targeted updates:** dependency graph, dirty layout ที่พิสูจน์ได้; วางหลัง Canvas เพื่อให้มี node หลากหลายสำหรับทดสอบ.
2. **Input/event contract เต็มรูปแบบ:** pointer move/down/up/double click, key up, capture/propagation, shortcut, drag/drop และ accessibility.
3. **Virtualized 2D viewport:** เป็น primitive สำหรับตารางขนาดใหญ่ ไม่ใช่ spreadsheet engine.
4. **Text engine ขั้นสูง:** multiline, shaping/fallback/selection API, Thai/CJK/emoji verification, undo/redo.
5. **Window/platform:** multi-window, modal/popup, file dialogs, monitor/fullscreen, platform abstraction; Windows ยังเป็นเป้าหมายพัฒนาหลักใน Phase 1 ปัจจุบัน ส่วน Linux/macOS ให้แยก milestone และ backend contract ไม่ประกาศรองรับก่อนทดสอบ.
6. **Animation/compositor และ effects:** opacity/transform/transition, scheduling และวัด frame/input-to-present. Shadow/gradient/path เพิ่มเมื่อ paint foundation เสถียร.
7. **หัวข้อ package ใน Phase 1:** นำ primitive มาสร้าง controls และ design systems ที่เลือกติดตั้งได้; Craft Office/ERP เป็น application layer ในอนาคตและไม่อยู่ในแผนนี้.

ข้อกำหนดใหญ่ทั้งหมดควรถูกติดตามเป็น roadmap ต่อเนื่อง โดย revision หนึ่งต้องมีผลลัพธ์ที่ตรวจรับได้จริง. 0.1.24 จึงแก้ช่องว่าง **วาด UI เองไม่ได้** ก่อน แล้วค่อยเพิ่ม state, event, data และ platform ตามหลักฐานจากการใช้งาน.
