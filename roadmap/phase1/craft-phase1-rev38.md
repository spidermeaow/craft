# Craft Phase 1 Rev.38 (0.1.38) — ปิดงาน Craft UI Engine สำหรับ Windows

วันที่: 9 ตุลาคม 2026  
สถานะ: **กำลัง implement; ยังไม่ผ่านเกณฑ์ปิดรอบและรอผู้ใช้ทดสอบ UI**  
ฐาน: [Rev.37](craft-phase1-rev37.md), [สัญญา VirtualGrid](craft-phase1-rev32-virtual-grid-contract.md) และ [เกณฑ์ Final ของ UI](../craft-ui-final-excel-acceptance.md)

## เป้าหมาย

Rev.38 เป็นรอบเก็บงาน **UI Engine และ GUI package ทั้งระบบ** สำหรับ Windows ให้แอป Desktop ที่ซับซ้อนสร้าง UI ด้วย public API ของ Craft ได้โดยไม่แก้ private Runtime เพื่อรองรับพฤติกรรมเฉพาะแอป เป้าหมายพิสูจน์ด้วย Spreadsheet UI ที่มีประสบการณ์ระดับ Excel และแอปอีกประเภทหนึ่งที่ใช้ primitive เดียวกัน การปิดงานหมายถึงผ่านเกณฑ์ตรวจรับด้านล่างและตรึงขอบเขต API ที่ประกาศรองรับ ไม่ใช่การรับประกันว่าจะไม่มี bug ในอนาคต หากพบ regression หลังปิดรอบให้แก้เป็นงาน maintenance โดยไม่ขยายขอบเขตฟีเจอร์ของ Rev.38

Rev.38 เป็น **milestone ที่แบ่งส่งและตรวจรับได้หลายชุด** ไม่บังคับให้ทุกเรื่องจบใน patch เดียว ห้ามประกาศว่าจบเพียงเพราะ build ผ่านหรือ Showcase หน้าตาดี

## ขอบเขตและเจ้าของงาน

| เจ้าของ | งานใน Rev.38 |
| --- | --- |
| Craft UI Engine / `std.ui` | window lifecycle, event loop, layout/measurement, renderer, scroll/viewport, input/focus/IME, clipboard/drag-drop primitives, DPI, accessibility, diagnostics และ public API ที่จำเป็น |
| `packages/gui` | component ทั่วไปและ styling/state ที่ประกอบจาก Engine เช่น Button, TextInput, virtual table/grid, menu/toolbar, dialog/popup และ theme; API ต้องไม่รู้จัก workbook หรือสูตร |
| UI Showcase และแอปพิสูจน์ | ใช้ public API สาธิตและทดสอบ interaction จริง เก็บข้อมูลและ policy ของแอปเอง พร้อมกรณี Spreadsheet และกรณี non-Spreadsheet |
| ผู้ใช้ | ตรวจผล UI จริงและยืนยันการรับมอบตามชุดทดสอบ; ผลการตรวจของผู้ใช้ต้องบันทึกแยกจาก static/build checks |

สูตร, workbook, XLSX/CSV, recalculation, number format, autofill, paste special และ undo/redo ของข้อมูลเป็นงานแอปหรือ library ไม่เพิ่มเข้า UI Engine. Rev.38 ปิดขอบเขต **Windows desktop** พร้อม Direct2D และ GDI fallback; backend ของ OS อื่นและการแพ็ก/Deploy ตามแผน 0.2.2 มีเกณฑ์แยก ไม่ถูกนับว่าผ่านในรอบนี้

## ลำดับดำเนินงาน

### 1. ทำฐาน Rev.37 ให้ตรวจรับได้ก่อน

- ตรวจและแก้ range selection ด้วยคลิก, Shift+คลิก, ลาก, Shift+Arrow และการเลื่อนข้าม viewport โดยไม่ทำให้ editor เดิมเลือกหรือบันทึกผิดเซลล์
- ตรวจ Unicode clipboard รวมถึงภาษาไทย, copy/paste ช่วง, ขอบตาราง, clipboard ไม่พร้อมใช้ และกรณี paste ล้มเหลวที่ต้องไม่ทำข้อมูลเดิมหาย
- จัดทำรายการ public API ที่มีจริงเทียบกับสัญญา VirtualGrid และบันทึกสิ่งที่ยังขาด ห้ามนับ accessibility ของ virtualized cells ว่าเสร็จจากการมี accessibility ของ TextInput

### 2. ปิดความถูกต้องของ Layout, Rendering และ Window

- ใช้กติกา measurement/alignment/padding/clip ที่สอดคล้องกันระหว่าง Button, TextInput, Canvas และ VirtualGrid; พื้นที่วาดต้องตรงกับ hit region และ focus region
- ตรวจ native child control เมื่อ scroll, resize, เปลี่ยนหน้า, ย้ายหน้าต่างข้ามจอ และ DPI 100%/125%/150% รวมถึง mixed DPI; แก้ stale frame, cell text ซ้อน, caret/editor หลุดตำแหน่ง และ wait cursor ค้าง
- ทำ dirty region/cell invalidation, text measurement/cache และ viewport update ที่ไม่วาดทั้งตารางเมื่อเปลี่ยนหนึ่งเซลล์; Direct2D และ GDI ต้องแสดงข้อมูลเดียวกันภายใต้สัญญาที่ประกาศ
- รองรับแถว/คอลัมน์ขนาดต่างกัน, พิกัดไกล, scroll สองแกน และ fixed/sticky regions แบบ generic ที่ใช้ประกอบ frozen panes ได้ โดยไม่สร้าง node ตามจำนวนเซลล์ทั้งหมด

### 3. ปิด Input, Text และ Component API

- กำหนด event routing, pointer capture, focus transfer, keyboard navigation, command scope และ modifier ให้แน่นอน; editor ต้อง commit/cancel, selection ต้องไม่เปลี่ยนข้อมูลโดยไม่ตั้งใจ
- ตรวจ caret, selection, composition/IME ภาษาไทย, Unicode, multiline หรือ rich text เฉพาะส่วนที่ประกาศรองรับ; ข้อจำกัดของ TextInput ต้องเขียนชัดและไม่ทำให้ use case ที่รับรองติดขัด
- ให้ public primitive ของ clipboard และ drag/drop พร้อม error/cancellation ที่แอปจัดการได้; นโยบาย copy/paste และการ validate ข้อมูลยังอยู่ในแอป
- ปรับ Button, TextInput และ disabled/focus/hover/pressed state ให้ alignment, ขนาด, contrast และ hit target สม่ำเสมอ; เพิ่ม generic menu/toolbar/dialog/popup และ primitive การวาด text/image/shape/chart เท่าที่จำเป็นต่อ UI ระดับเป้าหมาย ไม่ผูก component กับ Spreadsheet

### 4. ปิด VirtualGrid, Responsiveness และ Accessibility

- ขอ/ยกเลิกข้อมูลตาม viewport ล่าสุด รองรับ async source, generation/backpressure และไม่ให้ผลตอบกลับเก่าทับ viewport ใหม่; งานหนักของแอปต้องไม่หยุด UI thread
- รองรับ selection/editing/keyboard/clipboard ผ่าน public API พร้อมการเลื่อนข้าม viewport และขนาดแถวคอลัมน์ต่างกัน; ใช้หน่วยความจำตามข้อมูลจริงหรือพื้นที่ที่มองเห็น ไม่ตามจำนวนเซลล์สูงสุด
- เพิ่ม UI Automation แบบ virtualized สำหรับ Grid/Table/Selection และ item lookup ตามความเหมาะสม: screen reader ต้องอ่านชื่อ ค่า พิกัด active cell และช่วงที่เลือกได้โดยไม่สร้าง automation node ทั้งตาราง ตรวจ keyboard-only, high contrast และ large text
- เปิด trace ที่วัด frame time, input-to-draw latency, scroll/update queue, memory และ renderer path เพื่อวิเคราะห์ปัญหาได้โดยไม่แก้ private Runtime ของแอป

### 5. ตรึง Public API และเอกสาร

- ตรวจทุก API ที่ Showcase/แอปพิสูจน์เรียก: ต้องเป็น `std.ui` หรือ `packages/gui` ที่มีสัญญาและตัวอย่างใช้งาน; ห้ามพึ่ง private Runtime หรือ workaround เฉพาะแอป
- ทบทวนชื่อ, event payload, error semantics, lifecycle และ compatibility; เขียน migration note เมื่อเปลี่ยน API จาก Rev.37
- เขียนคู่มือ component, VirtualGrid, accessibility, DPI, renderer fallback และ diagnostics พร้อมข้อจำกัดที่ยอมรับได้; บันทึกผลการตรวจรับและ revision ของ Craft/GUI/แอปที่ใช้จริง

## เกณฑ์ตรวจรับก่อนประกาศว่า UI Engine จบ

1. **ความถูกต้อง:** รายการ interaction ที่รับรองทำงานตรงกันใน Showcase และแอปพิสูจน์ ไม่มี text ซ้อน/หาย, editor ผิดเซลล์, hit region เหลื่อม, focus ค้าง หรือ cursor busy ค้างใน workflow ที่กำหนด
2. **API และความทั่วไป:** แอป Spreadsheet ทำ grid, editor, selection, toolbar/dialog และ accessibility ด้วย public API เท่านั้น และตัวอย่าง non-Spreadsheet ใช้ primitive หลักซ้ำได้ ไม่มีการเพิ่มสูตรหรือ workbook ลง Craft
3. **ขนาดและความเร็ว:** ตรวจ sparse grid อย่างน้อยถึงพิกัด 1,048,576 × 16,384 โดยไม่สร้าง node ตามจำนวนเซลล์ทั้งหมด; ระบุเครื่อง, renderer, DPI, ชุดข้อมูลและวัด frame time, input-to-draw, CPU/memory, queue backlog ขณะ scroll เร็ว/แก้ไข/resize เป้าหมายเริ่มต้นคือ scroll ใกล้ 60 FPS และการแก้เซลล์ทั่วไปตอบสนองราว 16–50 ms; กำหนด threshold ที่ใช้ตัดสินจริงจาก baseline ที่วัดและบันทึกก่อนประกาศผ่าน
4. **สภาพแวดล้อม:** ตรวจ Windows, Direct2D/GDI, DPI 100%/125%/150% และ mixed-monitor, IME ไทย, keyboard-only, screen reader, high contrast และ large text ตาม matrix ที่เผยแพร่ กรณีไม่ผ่านต้องมี owner และแผนแก้ก่อนปิดรอบ
5. **ความเสถียร:** ไม่มี defect ระดับบล็อกการใช้งานหรือข้อมูลผิด/สูญหายที่เปิดอยู่; defect รองต้องมีผลกระทบและข้อจำกัดระบุชัด ผู้ใช้ตรวจ UI จริงและลงผลรับ/ไม่รับของแต่ละ workflow
6. **หลักฐานส่งมอบ:** public API matrix, คู่มือ/migration, regression matrix, performance trace, บันทึกผลผู้ใช้ และ build/installable artifact ของ revision เดียวกันครบ ผล build/check ไม่ใช้แทนผลทดสอบ UI

หากเกณฑ์ข้อใดไม่ผ่าน สถานะ Rev.38 ยังคงเป็น **อยู่ระหว่างดำเนินการ** แม้ source จะ implement ครบรายการแล้ว เมื่อผ่านทุกข้อจึงบันทึก UI Engine baseline และหยุดขยายขอบเขตของ revision นี้ ส่วนความสามารถแอป Spreadsheet ตามหมวด A–L ใน [เกณฑ์ Final](../craft-ui-final-excel-acceptance.md) ยังมีการตรวจรับของแอป/Core/library แยกต่างหาก

## Implementation ที่เริ่มแล้ว

- UI Automation ของ VirtualGrid ระบุ role เป็น DataGrid และ cell ที่โหลดอยู่เป็น DataItem พร้อมชื่อที่มีพิกัดแถว/คอลัมน์และค่าเซลล์; grid ที่ไม่มี accessible name ระบุจำนวนแถว/คอลัมน์
- UI Automation `IsKeyboardFocusable` และการสั่ง focus ใช้กติกา focus เดียวกับ Runtime จึงรวม VirtualGrid ที่ focus ได้และไม่อ้างว่า Text ทุกตัว focus ได้
- `CRAFT_UI_TRACE=1` จำกัดตัวอย่างเวลาล่าสุดต่อ metric เพื่อไม่ให้หน่วยความจำเติบโตตามเวลาที่เปิดแอป
- แก้ `relayout()` ที่เคยเรียก scroll-to-focus ทุกครั้งหลัง `gui.update`: เมื่อกริดล่างขอข้อมูลช่วงใหม่ หน้า ScrollView ภายนอกจะคงตำแหน่งที่ผู้ใช้เลื่อนไว้; การสั่ง focus โดยตรงยังเลื่อนหาเป้าหมายตามเดิม
- **ยังค้าง:** UIA Grid/Table/Selection/VirtualizedItem patterns และการเรียกเซลล์นอก viewport, variable row/column size, fixed regions, dirty-cell rendering, menu/dialog/popup, drag/drop, renderer/DPI/IME matrix, performance threshold และหลักฐานรับมอบจากผู้ใช้ รายการเหล่านี้ห้ามนับว่าผ่านจากการเปลี่ยนแปลงข้างต้น
