# Craft Phase 1 Rev.37 (0.1.37) — VirtualGrid range interaction

วันที่: 9 ตุลาคม 2026  
สถานะ: **implement range selection และ Unicode clipboard text ใน source; รอผู้ใช้ทดสอบ UI, accessibility แบบ virtualized ยังไม่เสร็จ**  
ฐาน: ผู้ใช้ทดสอบ UI Showcase หลัง Rev.36 แล้วรายงานว่าเร็ว ภาพสวย และแก้ข้อมูลในเซลล์ได้จริง

## เป้าหมายและลำดับงาน

ต่อยอดจากการแก้ข้อมูลหนึ่งเซลล์ให้ทำงานกับช่วงข้อมูลได้ โดยรักษาความเร็วและภาพที่ผู้ใช้ยืนยันใน Rev.36. Craft ต้องให้ primitive ของ UI ที่ใช้กับตารางทั่วไปได้; การแปลช่วงเซลล์เป็นข้อมูล Spreadsheet เป็นหน้าที่ของแอป. ทำตามลำดับด้านล่างและตรวจรับแต่ละช่วงก่อนขยายงาน.

### 1. ตรวจความเสถียรของ Rev.36 ก่อนเพิ่มพฤติกรรมใหม่

- ตรวจหน้า Excel ที่ DPI 100% และ 125%, resize และการย้ายระหว่างจอที่ DPI ต่างกัน, Direct2D กับ `CRAFT_UI_RENDERER=gdi`, IME ภาษาไทย, scroll เร็วและกลับขึ้น, และสลับหน้าแล้วกลับมา.
- บันทึก renderer, DPI, ขนาดหน้าต่าง และภาพ/ขั้นตอนเมื่อพบปัญหา. ค่า editor, caret, focus, hit region และข้อมูลที่บันทึกต้องตรงกับเซลล์ที่เลือก. หากมีปัญหาให้แก้ Runtime ใน Rev.37 ก่อนเพิ่ม range API.
- การทดสอบชุดนี้เป็น regression gate; ผลที่ผู้ใช้ยืนยันใน Rev.36 ไม่ถูกตีความว่า DPI, GDI หรือ IME ทุกกรณีผ่านแล้ว.

### 2. เลือกช่วงใน VirtualGrid — งานหลักของ Rev.37

| ชั้น | งาน |
| --- | --- |
| Craft UI Engine / `std.ui` | เพิ่ม state และ event แบบทั่วไปสำหรับ anchor/focus ของช่วงสี่เหลี่ยม, hit test แบบลากเมาส์, keyboard navigation พร้อม modifier, การเลื่อน viewport ตามปลายช่วง และการวาด selection โดยไม่สร้าง node ทุกเซลล์. API ต้องใช้กับ data grid ทั่วไปได้ |
| `packages/gui` | เปิด API ของ Engine ด้วยชื่อและสัญญาที่สม่ำเสมอ; เพิ่ม styling ของกรอบ/พื้นเลือกช่วงเฉพาะที่เป็น component ทั่วไป |
| UI Showcase / แอปทดสอบ | ใช้หน้า Excel แสดงการคลิก, Shift+คลิก, ลากเลือก, Shift+Arrow และการกลับไปแก้หนึ่งเซลล์; แอปเก็บนโยบายการเลือกและค่าข้อมูลของตนเอง |

**เกณฑ์ผ่าน:** แสดง anchor และ active cell ชัดเจน; เลือกช่วงได้ทั้งสี่ทิศทางและข้าม viewport; พื้นที่ที่วาดตรงกับ hit region; การเลือกไม่แก้ข้อมูลเอง; editor หนึ่งตัวของ Rev.36 ยังแก้ค่าได้ถูกเซลล์และไม่ทำให้ scroll ช้าลงอย่างเห็นได้ชัด.

### 3. Clipboard สำหรับช่วงข้อมูล

- Engine ให้ primitive อ่าน/เขียน clipboard ข้อความ Unicode และการแจ้งข้อผิดพลาดที่แอปจัดการได้. หากต้องเพิ่มรูปแบบข้อมูล ให้กำหนดสัญญาเป็นชนิดทั่วไป ไม่ผูกกับ workbook.
- Showcase ใช้ช่วงที่เลือกทำ copy/paste เป็นข้อความคั่นด้วย tab/newline. การแปลงพิกัด, validation, การกระจายค่า และ undo ระดับข้อมูลเป็นหน้าที่ของแอป.
- ตรวจ copy ช่วง 2×3 ไปยังโปรแกรมข้อความและ paste กลับเข้าตาราง, Unicode/ภาษาไทย, ข้อมูลหลายบรรทัด, ช่วงที่ตัดขอบตาราง และกรณี clipboard ไม่พร้อมใช้งาน. ค่าเดิมต้องไม่สูญหายเมื่อ paste ล้มเหลว.

### 4. Accessibility ของ VirtualGrid

- ให้ screen reader รับรู้จำนวนแถว/คอลัมน์ ตำแหน่งและค่าของเซลล์ที่มองเห็น active cell และช่วงที่เลือก โดยไม่สร้าง UI node ตามขนาดข้อมูลทั้งหมด.
- ศึกษา UI Automation สำหรับ virtualized items และกำหนด provider/lookup contract ที่ใช้กับตารางทั่วไป. Showcase เป็นกรณีตรวจด้วย keyboard-only และ screen reader; ต้องไม่ลดประสิทธิภาพของ VirtualGrid.
- งานนี้อาจแยกเป็น revision ถัดไปหากการตรวจรับ range และ clipboard ใช้เวลามาก; ห้ามนับว่าผ่าน Rev.37 จาก accessibility ของ TextInput เพียงตัวเดียว.

## สิ่งที่ยังไม่อยู่ใน Craft UI Engine

สูตร, workbook, การอ่าน/เขียน XLSX, number format, paste special, autofill และ undo/redo ของข้อมูลเป็นงานของแอป Spreadsheet หรือ library. Frozen panes และขนาดแถว/คอลัมน์ต่างกันเป็นงาน generic UI ในรอบถัดไปหลัง range interaction มั่นคง; ไม่รวมเป็นเกณฑ์ปิด Rev.37 นี้.

## หลักฐานก่อนประกาศผล

บันทึก public API ที่เพิ่ม, ตัวอย่างที่ใช้ API โดยไม่แตะ private Runtime, ผลการทดสอบของผู้ใช้แยกตาม DPI/renderer/IME, การทดสอบเลือกช่วงและ clipboard และข้อจำกัดที่ยังเหลือ. การ build หรือ `craft check` อย่างเดียวไม่ถือว่าผ่าน UI จริง. งาน accessibility ให้ระบุสถานะเป็นผ่านหรือค้างอย่างชัดเจน ไม่รวมกับผลผ่านของ range selection.

## Implementation รอบแรก

- เพิ่ม `std.ui.gridRangeSelection` / `gui.gridRangeSelection` โดยเก็บ anchor กับ focus ใน VirtualGrid node และวาดพื้นที่เลือกด้วย Direct2D/GDI โดยไม่สร้าง node เพิ่มตามจำนวนเซลล์. เพิ่ม `gridRangeStyle` ให้แอปกำหนดสีพื้นและกรอบช่วงที่เลือกได้.
- เพิ่ม `GridRangeDragged` สำหรับการลากเมาส์ข้ามเซลล์ และส่ง Shift+Arrow จาก editor ไปยังการเดินเซลล์ของ grid. UI Showcase ใช้ event นี้รวมกับ `GridCellPressed`/`GridSelectionMoved` เพื่อจัดการ anchor และ focus.
- เพิ่ม `std.ui.clipboardReadText`, `std.ui.clipboardWriteText` และ wrapper ใน `packages/gui` สำหรับ Unicode text บน Windows. Showcase มีปุ่ม Copy/Paste และ Ctrl+C/Ctrl+V เฉพาะหน้า Excel ที่แปลงช่วงเป็นข้อความคั่นด้วย tab/newline โดยเก็บข้อมูลในแอป.
- เพิ่ม[คู่มือและขั้นตอนทดสอบ](../../docs/REV37-UI-VIRTUAL-GRID-RANGE.md). Build portable CLI และ `craft check` ของ Showcase สำเร็จ; ยังไม่ได้เปิดแอปหรือทดสอบ interaction ตามความต้องการให้ผู้ใช้ทดสอบเอง.
- **ค้าง:** UI Automation Grid/ItemContainer/VirtualizedItem provider, การวัดผล DPI/IME/GDI/scroll จริง และผลตรวจรับ drag/clipboard จากผู้ใช้. ยังไม่ประกาศว่า Rev.37 ผ่านทั้งหมด.
