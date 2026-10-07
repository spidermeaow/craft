# Craft Phase 1 Rev.25 (0.1.25) — Responsive Layout, Typography และ Canvas Interaction

วันที่: 7 ตุลาคม 2026  
สถานะ: **ลง source แล้ว; รอผู้ใช้ทดสอบตาม `output/phase1-rev25-status.md`**  
ฐาน: `std.ui` หลัง 0.1.24 และการทดลอง `packages/gui` 0.2.1 ใน `D:\02-Repository\craft-test\gui-package-dashboard`

## ขอบเขต Phase 1 ในรุ่นนี้

Craft **0.1.25** เป็นเวอร์ชันเดียวของ CLI/Runtime และเป็นรอบงาน Phase 1 ที่รวมการพัฒนา package ไว้ด้วย. Rev.25 เพิ่ม viewport/style/hit target primitives ใน Runtime และพัฒนา `packages/gui` **0.2.2** ให้ใช้ primitives เหล่านี้ได้. เลข 0.2.2 เป็นเวอร์ชันของ dependency ใน `craft.toml` ไม่ใช่ Craft version หรือ Phase ใหม่.

### การพัฒนา package ใน Rev.25

- เพิ่ม wrapper ของ viewport, text/surface style, styled Canvas text และ hit region ใน `packages/gui` โดยคง API 0.2.1.
- เพิ่ม palette/spacing tokens, composition และตัวช่วย adaptive layout/Canvas target; แอปยังเป็นเจ้าของ state/event และเลือก breakpoint เอง.
- อัปเดต `examples/gui-package-lab` และสำเนา package ใน `craft-test/gui-package-dashboard` ให้ใช้ `gui.*` ใน app source; ล็อก dependency เป็น 0.2.2.
- Runtime และ package มีผลตรวจรับแยกกันภายใน Phase 1 เดียวกัน. อ่าน [สถานะ Runtime Rev.25](output/phase1-rev25-status.md) และ [สถานะ package 0.2.2](output/gui-0.2.2-status.md); ทั้งคู่ยังรอผู้ใช้ทดสอบ GUI จริง.
- [แผน package 0.2.2](packages/gui-0.2.2-plan.md) เป็นแผนย่อยของ Phase 1 Rev.25.

## เหตุผลและหลักฐาน

การทดลอง dashboard ใช้สำเนา `packages/gui` ที่ตรงกับต้นฉบับโดยไม่แก้ Runtime/package. `craft package check` และ `craft check` ผ่าน; `craft run` เปิดหน้าต่างด้วย `direct2d-hardware` และตรวจภาพบนจอ 1536×960 ที่ DPI 125% แล้ว. โครงหน้าและ Canvas วาดได้ แต่แอปตัวอย่างวาด dashboard เป็น Canvas **ขนาดคงที่ 1300×920 logical pixels** จึงมีส่วนถูกตัดเมื่อพื้นที่จริงไม่พอ และ Text/Canvas ยังปรับ typography ได้จำกัด. ปุ่ม ตาราง แท็บ slider และ toggle ในภาพเป็นภาพจำลอง; `CanvasPressed` บอกได้เพียง node กับพิกัดภายใน Canvas.

**การถูกตัดไม่ได้พิสูจน์ว่า layout engine คำนวณผิด**: แอปเลือก Canvas ขนาดคงที่เอง. ช่องว่างของ Runtime คือไม่มีทางที่ชัดเจนและประหยัดในการรับขนาด viewport แล้ววาด Canvas ตามพื้นที่จริง, ไม่มี overflow แนวนอนที่ครบ และไม่มีเครื่องมือ text/style/event พอให้ package สร้าง widget ที่ปรับขนาดและโต้ตอบได้. การทดลองยังไม่ยืนยัน GDI parity, IME ทุกกรณี, input-to-present หรือ performance benchmark. บันทึกผลเดิมอยู่ใน [คู่มือ dashboard test](../../../craft-test/gui-package-dashboard/README.md) ซึ่งเป็นโปรเจกต์ภายนอก repository นี้.

0.1.25 กลับมาเติม **Runtime primitives** ที่ `packages/gui` ต้องใช้ โดยไม่สร้าง Craft Office, ERP, design system สำเร็จรูป หรือเครื่องมือลากวาง. ภาษา Craft และ native Windows renderer ยังคงเป็นฐาน; ไม่ใช้ WebView/React/HTML/CSS. ลำดับตัดสินคือ responsiveness, runtime performance, smoothness, developer experience แล้วจึง resource/binary size.

## เป้าหมายหลักของ 0.1.25

### 1. Layout และ Canvas ที่ปรับตามพื้นที่จริง

- นิยามสัญญาขนาดของ child ใน Row/Column/ScrollView ให้ชัดเมื่อมี fixed, min/max, weight และพื้นที่ไม่พอ: ใครหดได้, ใคร overflow, ใครถูก clip. ตรวจ nested layout ที่ DPI ต่างกันและขณะ resize.
- ให้ Craft รับ **viewport ขนาด logical pixels** หรือ event เปลี่ยนขนาดที่มีขนาด content area ที่ใช้งานได้ เพื่อสร้าง draw commands ใหม่ตามพื้นที่จริง. พิจารณา Canvas แบบ fill/weight ที่ได้ขนาดจาก layout โดยไม่ต้องกำหนด width/height คงที่; API ที่เลือกต้องไม่ส่ง native pointer หรือ callback ที่รันบน UI thread ให้ Craft.
- เพิ่ม overflow แนวนอนที่กำหนดได้ รวมถึง scrollbar/การเลื่อนที่สอดคล้องกับการ clip และ nested scroll. Fixed Canvas ที่ใหญ่กว่า viewport ต้องเลื่อนดูได้หรือมีนโยบาย clip ที่ชัด ไม่ถูกตัดเงียบ ๆ.
- ทำ layout/paint invalidation เฉพาะส่วนที่กระทบเมื่อ resize หรือเปลี่ยน DPI; รักษา scroll offset และ focus เท่าที่สัญญาเดิมอนุญาต. ทดสอบกรณีย้ายหน้าต่างระหว่างจอ DPI ต่างกัน.

### 2. Typography และ style primitives ขั้นต้น

- เพิ่ม text style แบบข้อมูล: font size, weight, alignment, wrapping/overflow และสี ให้ Text/Button/Canvas text ใช้สัญญาการวัดและวาดเดียวกันทั้ง DirectWrite กับ GDI. กำหนดหน่วย logical pixels และ DPI scaling ให้ชัด.
- เพิ่ม style สำหรับพื้นหลัง, border, radius และ state พื้นฐานของ node ที่ package นำไปประกอบ widget ได้ โดยไม่ฝัง palette หรือแบรนด์ของแอปไว้ใน Runtime.
- แยก hover, pressed, focus และ disabled ในสัญญาของ control ที่รองรับจริง; หาก state ใดยังไม่พร้อม ให้ API/เอกสารระบุไว้ ไม่ให้ package อ้างว่ามีแล้ว.
- ตรวจข้อความไทย/อังกฤษและ fallback font ในขนาดที่ต่างกัน รวมถึง clipping เมื่อพื้นที่ไม่พอ. การเปลี่ยน style ที่ไม่กระทบ geometry ไม่ควรบังคับ remeasure ทั้ง tree.

### 3. Event/hit test สำหรับ widget ที่วาดด้วย Canvas

- เพิ่ม pointer down/up/move/enter/leave, double click และ key up ตามความจำเป็น พร้อม node ID, local coordinates, modifiers และลำดับ event ที่แน่นอน. กำหนด capture, cancel เมื่อออกนอก window, resize หรือปิด node.
- ให้ Canvas ระบุ **hit regions หรือ child targets** เป็นข้อมูลที่ Runtime ตรวจขอบเขตได้ เพื่อให้ปุ่ม/แท็บหลายตัวใน Canvas เดียวมี ID แยกกัน. หากยังไม่ทำ per-shape hit test อัตโนมัติ ให้กำหนด region ที่แอประบุเองอย่างชัดเจน.
- เพิ่ม focus/keyboard activation สำหรับ widget แบบ custom เฉพาะขั้นต่ำที่ใช้งานได้จริง; ไม่ให้การวาดเองทำให้ click ได้แต่ keyboard ใช้ไม่ได้. Event queue ต้องไม่ทำ click/key/close หายระหว่าง pointer move จำนวนมาก.
- ตรวจว่า event ใหม่ไม่ชนกับ Button, TextInput/IME, scrollbar drag, scroll wheel และ `CanvasPressed` เดิม; กำหนด compatibility/migration เมื่อเปลี่ยนสัญญา.

## งานที่ต้องติดตามหลัง 0.1.25

| ลำดับ | งาน | ขอบเขต Runtime ที่ต้องมี | สิ่งที่ควรอยู่ใน `packages/gui` |
| --- | --- | --- | --- |
| ถัดไป | Form controls | Input/focus/accessibility primitives สำหรับ checkbox, radio, select, multiline text, IME และ selection | รูปลักษณ์, validation, form composition |
| ถัดไป | Data table และ 2D virtualization | column/row viewport, cell hit test, keyboard navigation, selection และการวาดเฉพาะช่วงที่เห็น | table/grid component, sorting, pagination, data binding |
| ถัดไป | Animation และการวัด | scheduling/compositor, frame timing, input-to-present เมื่อวัดได้จริง | transition presets และ animation API ที่อ่านง่าย |
| ระยะต่อไป | Window/popup/accessibility | popup/modal ownership, focus routing, accessibility tree และ multi-window lifecycle | menu, dialog, tooltip และ app shell |

0.1.25 **ไม่สัญญา** ว่าจะมี checkbox, dropdown, date picker, text editor, table/grid, slider, popup หรือ animation สำเร็จรูปทั้งหมด. สิ่งเหล่านี้ต้องตามหลัง primitive และการตรวจรับที่เกี่ยวข้อง. Source bundle ยังไม่ใช่ native standalone executable และ Windows ยังเป็น platform ที่ทดสอบหลัก.

## สัญญา API ก่อน implement

1. ตรวจข้อจำกัด syntax/type checker ปัจจุบัน แล้วออกแบบ signature ของ viewport, flexible Canvas, style และ hit region เป็น value types ที่ Craft เก็บ/ส่งได้. ชื่อ API ใดในแผนนี้เป็นแนวคิดจนกว่าจะตรวจว่า `craft check` รับจริง.
2. ระบุ behavior ของ fixed/min/max/weight, overflow, DPI, hit testing, pointer capture และ error bounds ในเอกสารก่อนเขียน backend.
3. กำหนด migration ของ `std.ui.canvas`, `drawText`, `mount/update` และ event เดิมให้แอป 0.1.24 ยังทำงานได้; เพิ่ม API ใหม่แบบต่อยอดเมื่อเป็นไปได้.
4. แยก Runtime `std.ui` ออกจาก `packages/gui`: Runtime ดูแล layout/paint/input ที่ต้อง native; package ดูแล component/theme/state policy. ตัวอย่างตรวจรับใช้ package แต่ไม่แก้ package เพื่อปิดบังช่องว่างของ Runtime.

## ลำดับดำเนินงาน

1. เก็บ baseline ของ dashboard test ที่ DPI 100%/125%, ขนาด window สองระดับ และ renderer Direct2D/GDI; บันทึกส่วนที่ clip และ frame/paint trace ตามที่วัดได้จริง.
2. ลงสัญญา layout/viewport/overflow และตัวอย่าง Craft ที่สร้าง responsive Canvas อย่างน้อยสองขนาด.
3. ลง text style, measurement/paint parity และ state primitives ขั้นต้น; ตรวจข้อความไทย/อังกฤษ.
4. ลง event/hit regions/focus ตามขอบเขตขั้นต่ำ; ตรวจ overlap, scroll, capture และ keyboard.
5. ปรับ dashboard test **เฉพาะใน `craft-test`** เพื่อแสดง responsive layout และปุ่ม/แท็บ custom ที่คลิกเป็นรายตัวได้ ไม่แก้ `packages/gui` ระหว่างการทดสอบแยกชั้นนี้.
6. เขียน API guide, migration, manual test guide และ status ที่แยกสิ่งที่ลง source จากผลทดสอบจริง.

## เกณฑ์ตรวจรับ

| หัวข้อ | ต้องพิสูจน์ |
| --- | --- |
| Layout/DPI | Dashboard ตัวอย่างสองขนาดหน้าต่างและ DPI 100%/125% ไม่มีพื้นที่สำคัญถูกตัดโดยไม่สามารถเลื่อนหรือจัดใหม่; resize/ย้ายจอไม่ทำภาพค้าง |
| Typography | Text/Button/Canvas text วัดและวาดตรงกันในขนาด/น้ำหนักที่รองรับ; ไทย/อังกฤษไม่ถูก clip ผิดจาก layout |
| Interaction | Canvas มี target แยกอย่างน้อย 3 จุด; pointer/keyboard เปลี่ยน state ถูก, event ไม่หาย; Button/TextInput/IME/scroll เดิมไม่ถอยหลัง |
| Rendering | Direct2D และ GDI แสดงพื้นที่, clip, text และ hit region ตรงสัญญา โดยบันทึกความต่างที่ยอมรับได้ |
| Performance | บันทึก startup, idle CPU/RAM, layout/paint pass, frame time p50/p95/p99 และเวลาจาก input ถึง draw completion; **ไม่เรียก input-to-present** จนมีเครื่องมือวัดถึงจอจริง |
| Compatibility | ตัวอย่าง Rev.22–24 และ `packages/gui` 0.2.1 เดิม check/run ได้ หรือมี migration ชัดเจน |

การเปิด GUI และทดสอบภาพจริงให้ผู้ใช้ทำเองตามความต้องการปัจจุบัน เว้นแต่ผู้ใช้สั่งให้ผู้ช่วยทดสอบในงาน implement นั้นโดยตรง. สถานะสิ่งที่ลง source และข้อจำกัดของรอบ implement นี้อยู่ใน `output/phase1-rev25-status.md`.
