# Craft Phase 1 Rev.42 — UI Engine Final acceptance workbench

**สถานะ:** สร้างแอปทดสอบใน `D:\02-Repository\craft-sheet` แล้ว; ยังไม่ผ่านเกณฑ์ Final
และยังไม่มีผลทดสอบจากผู้ใช้สำหรับรอบนี้

## วัตถุประสงค์

ใช้ Spreadsheet ที่เขียนด้วย Craft และ `packages/gui` ผ่าน Public API เป็นสนามสอบ
UI Engine. ตรรกะ workbook, สูตร และไฟล์อยู่ในแอปหรือไลบรารีของแอป. ห้ามเพิ่ม
พฤติกรรมเฉพาะ Excel เข้า private UI Runtime เพื่อให้แอปตัวอย่างผ่านเท่านั้น.

ระหว่างสร้างแอปพบว่า VirtualGrid รับ Text node เท่านั้น แต่เดิม Text ใส่พื้นสี/กรอบ
ด้วย `surface` ไม่ได้. Rev.42 จึงขยาย `surface(Text, ...)` และการวาดใน Direct2D/GDI
ให้เป็น generic Text styling; VirtualGrid ยังคงรับ Text node และใช้ Public API เดิม.
เปิด `textWrap` บน Text cell ได้เพื่อทดสอบ clipping โดยความสูงแถวปัจจุบันยังคงที่.
การจัดแนวนอนของ Text ที่ส่งผ่าน `align` ถูกนำไปใช้กับการวาดจริงทั้งสอง backend.

## บันทึกการแก้ไขรอบ 0.1.42

### Craft UI Engine และ package/gui

- ขยาย `std.ui.surface` ให้รับ `Text` นอกเหนือจาก `Panel` และ `Button` เพื่อให้เซลล์ของ `VirtualGrid` กำหนดพื้นและกรอบผ่าน Public API ได้ โดย `gui.surface` ใช้ wrapper เดิม.
- ปรับการตรวจสอบ `VirtualGrid` ให้รับ Text cell ที่ใช้ `textWrap`; ข้อจำกัดปัจจุบันคือความสูงแถวยังคงที่ จึงอาจตัดข้อความที่ยาวเกินแถว.
- ปรับตัววาด Direct2D และ GDI ให้แสดงพื้น/กรอบของ Text และใช้ horizontal alignment ของ Text ตามที่กำหนดไว้.
- ปรับหมายเลข Runtime และ `packages/gui` เป็น **0.1.42 candidate** พร้อมเอกสาร package; ยังไม่ถือเป็น release ที่ผ่านเกณฑ์.

### แอปสนามสอบ `D:\02-Repository\craft-sheet`

- สร้าง workbook model, หน้า spreadsheet ด้วย `packages/gui`, formula bar, sheet tabs, VirtualGrid และ cell editor.
- เพิ่มการทำงานเบื้องต้นของชีต การแก้เซลล์ clipboard, undo/redo, find/replace, sort, รูปแบบเซลล์ และสูตรตามระดับที่ระบุในตารางด้านล่าง.
- เพิ่มการบันทึก/เปิดไฟล์ `.csheet` และ CSV import/export. `.csheet` เป็นรูปแบบทดลองของแอป **ไม่ใช่ XLSX**.
- เพิ่ม README ที่แยกความสามารถปัจจุบัน ข้อจำกัด และขั้นตอนทดสอบด้วยตนเอง.

### การตรวจสอบและสถานะ

- สร้าง `craft.exe` candidate ด้วย `go build` และสร้าง `craft.lock` ของแอปด้วย `craft package lock`; สองขั้นตอนนี้ไม่ใช่ผลทดสอบการทำงาน.
- **ยังไม่ได้รัน** automated tests, `craft check` หรือ `craft run` และยังไม่ได้ทดสอบ UI จริง ตามความประสงค์ที่ผู้ใช้ทดสอบด้วยตนเอง.
- ปัญหาและรายการที่ยังทำไม่ได้บันทึกในตารางด้านล่าง; **0.1.42 ยังไม่ผ่าน UI Engine Final gate**.

### ผลทดสอบจากผู้ใช้: การ parse struct field

ผู้ใช้รัน `craft check` กับ `craft-sheet` และพบ `E1001` ที่ `src/model.craft:5`
(`separate fields with a newline or semicolon`). สาเหตุจาก Lexer ใน Craft 0.1.42
มอง `>` ท้ายชนิด `Map<String, String>` เป็น operator ที่ต้องต่อ expression จึงไม่ส่ง
newline ให้ parser ระหว่าง field. ปรับแอปให้ใช้ `;` หลัง field นี้ตาม grammar เดิม
เพื่อให้ผู้ใช้ทดสอบต่อได้โดยไม่ต้องติดตั้ง Craft ใหม่. Language Core ควรแยก `>`
ที่ปิด generic type ออกจาก operator เปรียบเทียบ และเพิ่ม regression case สำหรับ
struct field ที่มี generic type โดยไม่ต้องใช้ `;`. **ยังรอผู้ใช้รัน `craft check` ซ้ำ**.

## สถานะรายการที่ผู้ใช้กำหนด

| กลุ่ม | รายการ | สถานะใน Craft Sheet รอบนี้ | งานต่อ/ตัวตัดสิน |
| --- | --- | --- | --- |
| ตาราง | สร้าง เปิด บันทึก XLSX | **ยังไม่ได้** | ต้องมี ZIP/OpenXML library ที่แอปเรียกจาก Craft ได้ และ round-trip จริง |
| ตาราง | เพิ่ม ลบ เปลี่ยนชื่อ สลับ Sheet | **มี** | จำกัด 4 ชีต; ทดสอบ data retention และ undo |
| ตาราง | เพิ่ม/ลบ Row และ Column | **มีบางส่วน** | ขยับข้อมูลใน fixed 60 × 18; ยังไม่เลื่อนสูตร/รูปแบบตาม Excel |
| ตาราง | ปรับความกว้าง Column/ความสูง Row | **ยังไม่ได้** | VirtualGrid Public API ยังใช้ขนาดสม่ำเสมอ |
| ตาราง | Copy/Cut/Paste/Paste Special | **มีบางส่วน** | ช่วงแบบ TSV และ Paste Values; ไม่มี styles/formats/transpose |
| ตาราง | Undo/Redo | **มีบางส่วน** | snapshot workbook สูงสุด 20 ขั้น; ตรวจ memory/latency |
| ตาราง | Fill Handle | **ยังไม่ได้** | ปุ่ม Fill Down/Right คัดลอก literal; ต้องมี pointer drag target และ relative reference rewrite |
| ตาราง | Find & Replace | **มีบางส่วน** | ค้นหา/แทนที่ข้อความในชีตปัจจุบัน; ยังไม่มี options/regex |
| รูปแบบ | Font, size, bold, italic, underline | **มีบางส่วน** | 12/14/18 pt และ bold; font family/italic/underline ยังไม่มี API ครบ |
| รูปแบบ | สีอักษร/พื้น, border | **มีบางส่วน** | blue/yellow/border preset ต่อเซลล์; ยังไม่มี picker และ palette |
| รูปแบบ | Merge & Center, alignment, wrap | **มีบางส่วน** | alignment/wrap toggle; fixed row height อาจ clip; merge ยังไม่มี |
| รูปแบบ | Number format, Format Painter | **มีบางส่วน** | general/currency/percent และ copy/apply style; date/custom formats ยังไม่มี |
| สูตร | SUM/AVERAGE/MIN/MAX/COUNT/COUNTA | **มีบางส่วน** | rectangular range ตัวเลขพื้นฐาน; ยังไม่มี dependency graph/error semantics |
| สูตร | IF/AND/OR/SUMIF/COUNTIF/VLOOKUP/XLOOKUP | **มีบางส่วน** | syntax พื้นฐาน, exact vertical lookup; ยังไม่มี full Excel semantics, dependency graph/error propagation |
| สูตร | A1, $A$1, ข้าม Sheet | **มีบางส่วน** | direct reference หนึ่งอักษร, แสดงค่าจากชีตอื่น; ยังไม่ครบการคำนวณ/ย้ายสูตร |
| ข้อมูล | Sort A-Z/Z-A | **มีบางส่วน** | sort ทั้งแถวตามคอลัมน์ที่เลือก; ยังไม่มีช่วง/หลายเงื่อนไข |
| ข้อมูล | Filter, Freeze Panes, Table, validation dropdown | **ยังไม่ได้** | แอปต้องกำหนด data policy; frozen region/variable viewport ต้องประเมิน generic UI primitive |
| ข้อมูล | AutoSum | **มีบางส่วน** | เตรียมข้อความสูตร; ต้องเลือกช่วงและยืนยันค่าที่ได้ |
| ไฟล์/พิมพ์ | CSV import/export | **มีบางส่วน** | quoted fields; จำกัดขนาด workbook และ active sheet |
| ไฟล์/พิมพ์ | Export PDF/OS print | **ยังไม่ได้** | ต้องมี print/PDF library หรือ generic print API แยกจาก UI Engine |
| ไฟล์/พิมพ์ | Print Preview/A4/orientation/print area | **มีบางส่วน** | basic A4 landscape on-screen preview; ไม่มี orientation/area controls หรือ pagination |

## Gate สำหรับประกาศ UI Engine Final

1. ผู้ใช้รัน `craft check` และ `craft run` ในสนามสอบแล้วส่งผล. แอปต้อง compile,
   เปิดได้ และไม่มี busy cursor, stale paint, selection drift หรือ edit loss.
2. ทำ manual interaction matrix บน Direct2D/GDI, DPI 100/125/150%, IME ไทย,
   keyboard-only, screen reader และหน้าต่างหลายขนาด. เก็บ trace ของ renderer
   และผลวัด scroll/edit latency. ไม่มีการอ้างว่าผ่านจาก screenshot อย่างเดียว.
3. แยก defect ที่เกิดจาก Public UI API ออกจาก Spreadsheet Core ของแอปด้วย
   repro ที่ใช้ component ทั่วไป. เพิ่ม UI Engine เฉพาะ primitive ที่ใช้ได้กับ
   แอปอื่นด้วย; ไม่ย้าย XLSX หรือสูตรเข้า Craft.
4. ทุกรายการในตารางต้องเปลี่ยนจาก “มีบางส่วน/ยังไม่ได้” เป็นหลักฐานตรวจรับ
   หรือข้อยกเว้นที่ผู้ใช้อนุมัติก่อนประกาศแอป Excel-class Final.

รอบนี้ไม่ได้รัน automated tests, `craft check` หรือเปิดแอป เพราะผู้ใช้เป็นผู้ทดสอบเอง.
