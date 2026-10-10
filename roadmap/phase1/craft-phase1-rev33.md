# Craft Phase 1 Rev.33 (0.1.33) — UI Engine correctness และ component consistency

วันที่: 8 ตุลาคม 2026  
สถานะ: **ผู้ใช้รับรองให้ผ่าน 0.1.33 และอนุญาตให้เดินต่อ 0.1.34**  
ฐาน: Craft 0.1.32 และผลทดสอบ UI Showcase โดยผู้ใช้

## เป้าหมาย

ทำให้ UI Engine บน Windows แสดงผลและตอบสนองได้สม่ำเสมอก่อนเพิ่ม component ขนาดใหญ่. Rev.33 เป็นรอบแก้คุณภาพของ primitive ทั่วไปที่ใช้ได้กับหลายแอป; ไม่เพิ่มตรรกะ Spreadsheet ลงใน Craft.

## สถานะ implementation

- เพิ่ม `std.ui.virtualGrid` และ `gui.virtualGrid` แบบ uniform cell size พร้อม scroll สองแกน, request ช่วง visible rows/columns และ cell press event. Showcase หน้า Layouts ใช้ข้อมูล mock 100,000 × 1,000 โดยสร้าง node เฉพาะช่วงที่ร้องขอ. รายละเอียด API และข้อจำกัดอยู่ใน [คู่มือ prototype](../../docs/REV33-UI-VIRTUAL-GRID.md).
- ปรับ disabled native TextInput ให้ใช้พื้นหลังเดียวกับ styled surface และสีข้อความสำหรับ disabled; ลดการเรียก `EnableWindow` ซ้ำเมื่อ state ไม่เปลี่ยน.
- การแก้ Button alignment และ cursor จาก 0.1.32 ยังคงอยู่ใน source. ยังไม่มีผลผู้ใช้ยืนยันว่าปัญหา cursor หมุนค้างถูกปิดแล้ว; Gate A และ Gate B จึงยังรอทดสอบ.
- VirtualGrid รอบนี้ยังไม่ครบสัญญาระยะยาวเรื่อง asynchronous generation/cancellation, dirty-cell invalidation และ virtual accessibility. Gate C ต้องวัดจำนวน node/หน่วยความจำและพฤติกรรม scroll จากผลทดสอบจริงก่อนสรุปผ่าน.
- ผลทดสอบผู้ใช้: ตาราง mock 100,000 แถวเลื่อนได้ลื่น แต่บางครั้งเนื้อหาวาดไม่ทันเมื่อเลื่อนเร็วมาก. ตรวจ source พบว่า request ช่วงใหม่ถูกระงับขณะรอ snapshot เดิม จึงปรับให้ Runtime เก็บคำขอช่วงล่าสุดแยกตาม Grid และแทนคำขอที่ยังไม่ถูกอ่าน. ผลทดสอบซ้ำรายงานว่าดีขึ้นเล็กน้อย; ยังไม่มีตัวเลข frame time หรือ latency จากเครื่องทดสอบ.
- ผลตรวจรับเพิ่มเติม 8 ตุลาคม 2026: ผู้ใช้ทดสอบการเลื่อนเร็วแล้วรายงานว่าดีขึ้นเล็กน้อย แต่บางช่วงยังเรนเดอร์ตามไม่ทัน และยอมรับข้อจำกัดนี้. ผู้ใช้ให้ 0.1.33 ผ่านโดยยังไม่ได้ทดสอบทุกส่วนอย่างละเอียด และขอให้ย้ายเวลาไปพัฒนาส่วนอื่นใน 0.1.34. การรับรองนี้ไม่ใช่หลักฐานว่า Gate A/B/C ผ่านทุกกรณีหรือมี benchmark ประสิทธิภาพ.

## งานที่ต้องปรับปรุง

### 1. ปิดปัญหาที่ค้างจาก 0.1.32

- ตรวจผลจริงของ Button text alignment ทั้ง Standard/Styled/Disabled และ Canvas text ที่เลือก center บน DirectWrite และ GDI fallback. ตำแหน่งข้อความต้องตรงกับกรอบที่วาดในทุกสถานะ.
- ตรวจอาการ busy cursor หลังเปิดแอปโดยไม่ต้องคลิกหน้าต่างอื่น. หากยังเกิด ให้เก็บลำดับ window activation, message loop, `WM_SETCURSOR`, mount และ paint ก่อนแก้ซ้ำ; ตรวจว่า TextInput ยังใช้ I-beam และขอบหน้าต่างใช้ cursor ของระบบ.
- ตรวจ Disabled TextInput หลังปรับสีพื้นของ native EDIT ให้ตรงกับพื้นหลังของกรอบ. ต้องไม่มีสี่เหลี่ยมสีเทาซ้อน, สีข้อความอ่านได้, รับ focus/พิมพ์ไม่ได้ และสถานะ accessibility ระบุ disabled ถูกต้อง.
- แก้ error จาก UI Showcase ที่ใช้ API ปัจจุบัน และใช้ผลทดสอบของผู้ใช้ยืนยัน ไม่ถือว่าการ build Setup อย่างเดียวผ่านเกณฑ์ UI.

### 2. Rendering และ layout ของ component

- กำหนด measurement, minimum/preferred size และ padding ของ Button/TextInput ให้สอดคล้องกันเมื่อวางใน Row, Column, Grid และ container ที่ยืดได้. แยกกรณี component กว้างตามเนื้อหากับกรณีที่ parent กำหนดความกว้าง.
- ตรวจให้ visual bounds, text bounds, focus bounds และ hit region ของ Button, TextInput และ Canvas target ตรงกัน. เมื่อ DPI เปลี่ยนหรือ resize ต้องไม่เลื่อนคนละตำแหน่ง.
- กำหนดและตรวจสี/ขอบ/ข้อความของ normal, hover, pressed, focused, disabled และ validation state โดยให้ Direct2D กับ GDI fallback แสดงความหมายเดียวกัน. ตรวจ high contrast และความชัดของข้อความด้วย.
- ลดการวาดซ้ำเมื่อเปลี่ยน state เฉพาะ component; บันทึกว่ากรณีใด relayout ทั้งหน้าจอ และแก้เฉพาะกรณีที่ trace ชี้ว่าเป็นปัญหา.

### 3. Input, focus และ accessibility

- ตรวจ native TextInput สำหรับ caret, selection, clipboard, IME ไทย, placeholder, disabled/read-only behavior, Tab order และการคืน focus หลัง rerender. กำหนดขอบเขต API ของ disabled กับ read-only ให้ต่างกันชัดเจนหากต้องใช้ทั้งคู่.
- ตรวจการใช้งาน Button/Canvas target ด้วยเมาส์และคีย์บอร์ด, focus visible, shortcut และ UI Automation. Control ที่คลิกได้ต้องมี hit region ตรงภาพและชื่อ/สถานะที่ screen reader อ่านได้.
- ตรวจ resize, maximize/restore และ DPI 100%/125% บนหน้าจอจริง โดย layout, native child control และตำแหน่ง pointer ยังตรงกัน.

### 4. ตารางข้อมูล: เริ่มจาก foundation ที่เป็น generic

- ใช้ [สัญญา Virtual Grid ของ 0.1.32](craft-phase1-rev32-virtual-grid-contract.md) กำหนด public API, event, ownership ของข้อมูลและ capability gate ให้ใช้งานได้กับตารางทั่วไป.
- Prototype viewport สองแกนที่ร้องขอข้อมูลเฉพาะช่วงที่เห็น, คำนวณพิกัด/hit test ตามแถวและคอลัมน์ และไม่สร้าง `UiNode` ทุกเซลล์. วัดจำนวน node, หน่วยความจำ และเวลา input-to-draw จากชุดข้อมูลที่กำหนดก่อนขยาย scope.
- งาน selection หลายช่วง, editor overlay, clipboard ช่วงเซลล์, frozen panes, merged cells, virtual accessibility และ scale ระดับ Spreadsheet เป็นขั้นต่อไปตามผล prototype; **ไม่ถือว่ามีใน 0.1.33 เพียงเพราะมีเอกสารหรือ Canvas demo**.

## ขอบเขตความรับผิดชอบ

| ชั้น | งานในแผน |
| --- | --- |
| `std.ui` / Runtime | การวัดและวาด, layout, event/hit test, cursor, TextInput native, focus, DPI, accessibility และ primitive ของ virtual viewport |
| `packages/gui` | style/variant และ component ทั่วไปที่ประกอบจาก primitive ได้; ไม่ทำ workaround เพื่อซ่อนข้อผิดพลาดของ Runtime |
| UI Showcase | ตัวอย่างและกรณีตรวจรับที่เรียก public API ปัจจุบัน แสดงสถานะและขนาดที่ต่างกันอย่างชัดเจน; ไม่ถือเป็นหลักฐานว่า Engine ถูกต้องจนกว่าผู้ใช้ตรวจจริง |
| แอป Spreadsheet | ข้อมูลเซลล์, สูตร, workbook, selection policy และ workflow เฉพาะแอป ไม่ย้ายเข้า UI Engine |

## ลำดับและเกณฑ์ตรวจรับ

1. **Gate A — ปัญหาเดิม:** ผู้ใช้เปิด Showcase จาก Setup 0.1.33 แล้วตรวจ Button alignment, Canvas target, Disabled TextInput และ cursor ตั้งแต่หน้าต่างเปิด โดยไม่ต้องสลับหน้าต่างเพื่อให้ cursor ปกติ. หากข้อใดไม่ผ่าน ให้เก็บผลและแก้ก่อนประกาศ Rev.33 ผ่าน.
2. **Gate B — component/layout:** แสดง Button/TextInput หลายขนาดใน Row/Column/Grid และหน้าต่างที่ resize ได้; ข้อความอยู่ในกรอบ, padding สม่ำเสมอ, hit region ตรงภาพ, native input ไม่ทับขอบ. ตรวจ Direct2D และ GDI fallback, DPI 100%/125%, mouse และ keyboard.
3. **Gate C — foundation ตาราง:** มีตัวอย่างตารางทั่วไปที่ใช้ viewport สองแกนและข้อมูล mock จำนวนมาก โดยจำนวน element/หน่วยความจำโตตามพื้นที่ที่เห็น ไม่โตตามจำนวน cell ทั้งหมด. ระบุ metric, ขนาดชุดข้อมูล, เครื่องทดสอบและข้อจำกัด; หากยังไม่ implement ให้คงสถานะเป็นแผน ไม่อ้างว่าผ่าน.

ผู้ใช้เป็นผู้ทดสอบ UI จริงตามแนวทางโครงการ. งานเอกสารและ static source review ไม่ใช่ผลตรวจรับ Runtime. ก่อนออก Setup ให้ระบุเลข Craft/GUI package ที่ตรงกัน และบันทึกผลทดสอบที่ผู้ใช้รายงานแยกจากสิ่งที่ implement แล้ว.

## ยังไม่รวม

- GUI backend ของ macOS/Linux และการแพ็กแอปแบบ self-contained; ดู [แผน Craft 0.2.2](../craft-0.2.2-deployment-plan.md).
- Spreadsheet engine, สูตร, XLSX, ribbon และฟังก์ชันเฉพาะ Excel; ดู [เกณฑ์ Final](../craft-ui-final-excel-acceptance.md).
- การรับรอง Virtual Grid ระดับ Final หรือประสิทธิภาพระดับล้านแถวจาก prototype ในรอบนี้.
