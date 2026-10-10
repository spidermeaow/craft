# Craft Phase 1 Rev.40 (0.1.40) — UI Engine Final gate

**สถานะ:** กำลังดำเนินการ; 0.1.40 เป็น candidate และยังไม่ผ่าน Final  
**เป้าหมาย:** ปิด Craft UI Engine สำหรับ Windows ด้วยหลักฐานตาม [เกณฑ์ Final](../craft-ui-final-excel-acceptance.md) ไม่ใช่เพียงเพิ่มเลขเวอร์ชัน

## ขอบเขตการประกาศ

ประกาศได้สองผลแยกกัน:

1. **UI Engine Final:** แอป Spreadsheet ที่เขียนด้วย Craft ใช้ public API สร้าง interaction, scale, accessibility และ component ทั่วไปได้โดยไม่แก้ private Runtime เฉพาะแอป
2. **Spreadsheet App Final:** แอปและ library ของแอปผ่านทุกฟีเจอร์ที่ประกาศในหมวด A–L, สูตร, ไฟล์, print และ integration พร้อมผลทดสอบรายรายการ

Rev.40 ตั้งใจปิดผลแรกเท่านั้น ผลแรกยังต้องมีแอป Spreadsheet ติดตั้งและใช้งานได้เป็นหลักฐาน. UI Showcase เป็น regression harness ไม่ใช่หลักฐานแทนแอปจริง.

## สถานะที่ตรวจได้จาก source

| ด้าน | หลักฐานปัจจุบัน | สิ่งที่ต้องทำก่อนผ่าน |
| --- | --- | --- |
| Canvas และ dirty region | Rev.39 มี Path, Direct2D/GDI และ trace; ผู้ใช้รายงานว่าภาพและตัวหนังสือคมหลังแก้ | ปิด Definition of Done Rev.39 ด้วย DPI, memory และ benchmark ที่ควบคุมสภาวะ |
| VirtualGrid | public API มี viewport สองแกน, selection/range, editor และ clipboard; Showcase แก้ข้อมูลได้ตามผลผู้ใช้ | ตรวจ scale ของ sparse data, การเลื่อนเร็ว, focus/IME/GDI/DPI กับแอปจริง |
| ขนาดและพื้นที่ตรึง | `virtualGrid` รับ row height และ column width ค่าคงที่ | เพิ่ม generic variable-size geometry และ pinned/sticky regions พร้อม viewport/hit test ที่ถูกต้อง |
| Input | มี native TextInput, focus, keyboard และ clipboard | เพิ่ม public pointer capture และ drag-drop primitives; ตรวจ IME ไทย, caret, multiline และการรับ input ระหว่างงานหนัก |
| Component | `packages/gui` มี theme, button, field, shell, toolbar และ grid wrappers | เพิ่ม generic menu/dialog และ DataGrid ที่ประกอบ editor/selection ได้; API ต้องไม่รู้จัก workbook/formula |
| Window/Taskbar icon | Runtime สร้าง Win32 window แต่ยังไม่มี public API สำหรับกำหนดไอคอนของหน้าต่าง | เพิ่ม `gui.setWindowIcon(window, path)` ผ่าน `std.ui` ให้ Windows ใช้กับ title bar, Alt+Tab และ Taskbar พร้อม fallback และการคืนทรัพยากร |
| Accessibility | UIA provider มี virtual grid และ visible cell ขั้นต้น | เพิ่ม/ตรวจ Grid, Table, Selection และ VirtualizedItem patterns รวมถึง offscreen items, high contrast และ keyboard-only |
| Rendering/diagnostics | มี backend switch และ timing trace | วัด input-to-present, frame time, redraw area, CPU/RAM และรายงานสเปก/ข้อมูล/วิธีวัด |
| โปรแกรมพิสูจน์ | `D:\02-Repository\app` เป็น Showcase; หน้า Excel เก็บค่าในหน่วยความจำ | สร้างแอป Spreadsheet ที่ติดตั้งได้ ใช้ public API เท่านั้น และตรวจ workflow เปิด/บันทึก, สูตร, sort/filter, copy/paste, undo/redo, chart, import, print/export |

ช่องที่ระบุว่า “ต้องทำ” ไม่สามารถถือว่าผ่านด้วยภาพหน้าจอหรือ `craft check` เพียงอย่างเดียว. หากข้อใดไม่จำเป็นต่อ UI Engine Final ให้บันทึกเหตุผลพร้อมวิธีที่แอปทำ workflow เดียวกันผ่าน public API.

## ไอคอนของแอป GUI ใน Rev.40

- แอป GUI เรียก `gui.setWindowIcon(window, path)` เป็น public API แบบเลือกใช้; `packages/gui` ส่งต่อให้ native Runtime ตั้งไอคอนขนาดเล็ก/ใหญ่ของหน้าต่าง. แอปที่ไม่เรียก API ใช้ไอคอนเริ่มต้น.
- Runtime ต้องจัดการ path/รูปแบบไฟล์ที่ไม่ถูกต้องอย่างชัดเจน, คงไอคอนหลัง `mount`/`update`, และปล่อย native icon handles เมื่อเปลี่ยนไอคอนหรือปิดหน้าต่าง. ใช้ไอคอนเดียวกันใน title bar, Alt+Tab และ Taskbar เท่าที่ Windows กำหนดสำหรับหน้าต่างนั้น.
- เพิ่มหน้า UI Showcase ให้ผู้ใช้ตรวจไอคอนจริง รวมถึงกรณีไฟล์หายและการเปลี่ยนไอคอน. ตรวจผลบน Windows ด้วยตนเอง; `craft check` ยืนยันได้เพียงชนิด/API.
- โปรแกรม API/CLI ไม่ต้อง import `gui` ไม่ต้องมีไฟล์ไอคอน และพฤติกรรม console เดิมไม่เปลี่ยน.
- ไอคอนที่ฝังใน `.exe`, shortcut และตัวติดตั้งของ **แอปผู้ใช้** เป็นงานของระบบ build/deploy. การตั้ง window icon ระหว่าง `craft run` ไม่ถือว่าฝังไอคอนใน binary หรือให้แอปมี identity แยกจาก `craft.exe` แล้ว.

### ความคืบหน้า candidate

- เพิ่ม `gui.setWindowIcon(window, path)` และ `std.ui.setWindowIcon` สำหรับไฟล์ `.ico` บน Windows. Runtime ตั้งไอคอนใหญ่/เล็กด้วย `WM_SETICON` และคืน icon handles เมื่อเปลี่ยนหรือปิดหน้าต่าง.
- UI Showcase มีหน้า Window Icon พร้อมไฟล์ตัวอย่างสองสีและกรณีไฟล์หาย. ผลการแสดงใน Title bar, Alt+Tab และ Taskbar ยังรอผู้ใช้ตรวจด้วยตนเอง.
- Gate ที่เหลือในตารางและ checklist ด้านล่างยังไม่ผ่าน จึงยังไม่ประกาศ UI Engine Final.

## ลำดับการทำงาน Rev.40

1. ปิด Rev.39: รวบรวมผลทดสอบ Rendering Lab, GDI, DPI, memory และ benchmark; แก้ regression ที่พบ
2. สร้าง Spreadsheet acceptance app แยกจาก Showcase ใช้ public API และ sparse data; ทำรายการ workflow ที่พิสูจน์ UI Engine ก่อน
3. สำหรับแต่ละ workflow ที่ติดขัด ระบุ generic primitive ที่ขาด และทดสอบกับ UI อีกประเภทหนึ่ง เช่น data viewer หรือ IDE ก่อนเพิ่มใน Runtime
4. เติม public API, package component, เอกสารและตัวอย่างตาม blocker จริง; รักษา API เดิม
5. จัดทำ traceability matrix ตามเกณฑ์ Final: owner, API, supported/partial/unsupported, ขั้นตอนทดสอบ, ผลและข้อจำกัด
6. ส่ง build ให้ผู้ใช้ทดสอบบน Windows พร้อม Direct2D/GDI, DPI, IME, keyboard และ screen reader; แก้ข้อบกพร่องแล้วทดสอบซ้ำ

## Gate สำหรับประกาศว่า “ผ่าน Final”

- [ ] มีแอป Spreadsheet ที่ติดตั้งและเปิดได้จริง ใช้ Craft public UI API โดยไม่แก้ private Runtime เฉพาะแอป
- [ ] Workflow หลักและ file round-trip ตามเกณฑ์ Final มีผลทดสอบ ไม่สูญเสียข้อมูลเงียบ ๆ
- [ ] Sparse workbook พิกัดไกลไม่สร้าง node ตามจำนวนเซลล์ทั้งหมด; มีตัวเลข scale/latency/frame/memory
- [ ] Keyboard-only, IME ไทย, accessibility, DPI หลายค่า, Direct2D/GDI และ high contrast มีผลตรวจ
- [ ] GUI App ตั้งไอคอนหน้าต่างผ่าน `gui` ได้ และผู้ใช้ตรวจ title bar, Alt+Tab, Taskbar, fallback และ resource cleanup; API/CLI ไม่มี dependency นี้
- [ ] Rev.39 ผ่าน Definition of Done และ UI Showcase ไม่เกิด regression
- [ ] Traceability matrix แยกสิ่งที่ Engine, package, แอปและ library รับผิดชอบ; ไม่มี partial/unsupported ที่ถูกนับเป็นผ่าน
- [ ] ผู้ใช้ตรวจผลใช้งานจริงและยืนยันการรับงาน

ห้ามเปลี่ยนสถานะเป็น “Final ผ่านแล้ว” หรือเผยแพร่คำรับรองจนกว่าทุก gate มีหลักฐาน. การไม่รัน automated/UI tests โดยผู้ช่วยตาม preference ของผู้ใช้หมายความว่ารายการตรวจต้องค้างจนได้รับผลทดสอบจากผู้ใช้.
