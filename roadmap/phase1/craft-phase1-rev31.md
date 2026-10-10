# Craft Phase 1 Rev.31 (0.1.31) — UI controls และ Window state

วันที่: 7 ตุลาคม 2026  
สถานะ: **กำลังนำไปใช้ตามคำขอ; รอผู้ใช้ทดสอบจริง**  
ฐาน: Craft CLI/UI Runtime 0.1.30 และ `packages/gui` 0.2.8

## เป้าหมาย

ปรับคุณภาพ Button และ TextInput ที่พบจาก Craft UI Showcase และให้แอปกำหนดสถานะหน้าต่างเริ่มต้นแบบ maximized ได้อย่างชัดเจน โดยยังคง `openWindow(title, width, height)` สำหรับขนาดปกติ.

## ข้อจำกัดที่พบระหว่างทดสอบแอป

- `packages/gui` 0.2.8 ไม่มี API สำหรับ maximize หน้าต่าง. แอปจึงเปิดด้วยขนาดที่ส่งให้ `gui.openWindow` เท่านั้น; อย่าเพิ่ม `std.ui` call หรือแก้ Runtime/package ระหว่างช่วงทดสอบ. งาน maximize ย้ายมาพิจารณาใน Rev.31 ตามหัวข้อด้านล่าง.
- UI ในแอปตัวอย่างเรียกผ่าน `gui.*` อยู่แล้ว ไม่มีการเรียก `std.ui` โดยตรง. แต่ข้อกำหนด **ใช้เฉพาะ `package/gui`** แบบเคร่งครัดยังทำไม่ได้: โค้ดตารางใช้ `std.convert.toString` สร้าง ID และรหัสแถว; ฟอร์มบันทึกข้อมูลใช้ `std.fs.exists/readText/writeTextAtomic` และ `std.json.quote`. `packages/gui` ปัจจุบันไม่มีความสามารถแปลงค่า สร้าง JSON หรือบันทึกไฟล์.
- ก่อนแก้ใน Rev.31 ต้องกำหนดขอบเขตให้ชัดว่า "เฉพาะ `package/gui`" หมายถึงทุกการเรียกจากแอป หรือเฉพาะ UI API. หากหมายถึงทุกการเรียก ต้องออกแบบ API ระดับ package ที่รองรับรหัสแถวและการบันทึกฟอร์มโดยไม่ให้แอปเรียก `std.*`; หากยังไม่มี API ที่เหมาะสม ให้ระบุว่าฟังก์ชันบันทึกไฟล์ยังไม่รองรับแทนการอ้างว่าทำได้ครบ.
- เมนู Canvas ปัจจุบันใช้ `CanvasPressed` เพื่อไม่ให้แสดง focus ring หลังคลิก จึงไม่ใช่ control ที่เดินด้วย Tab/Enter/Space ได้. หากต้องการทั้งรูปลักษณ์นี้และ keyboard accessibility ให้กำหนด focus-visible policy ในแผนก่อนแก้ Runtime หรือ package.
- ภาพตรวจของผู้ใช้ในหน้า Components และ Contacts แสดงว่า Button กับช่องกรอกข้อมูลยังมีหน้าตาไม่สอดคล้องกับส่วนอื่นของ Showcase. รายละเอียดและเกณฑ์แก้ไขอยู่ในหัวข้อ UI controls ด้านล่าง.

ผู้ใช้อนุญาตให้ implement 0.1.31 แล้ว. ข้อจำกัดเดิมด้านล่างเป็นที่มาของงาน; ผลการนำไปใช้และส่วนที่ยังไม่รองรับระบุท้ายเอกสาร.

## งาน UI controls จากภาพตรวจของผู้ใช้

### 1. Button

- หน้า Components แสดง Primary และ Secondary ที่มีสี/ขอบ/รูปแบบ focus ไม่เป็นชุดเดียวกัน; หลังคลิก Secondary เห็นกรอบสีน้ำเงินเด่นและพื้นปุ่มเปลี่ยนจนแย่งความสนใจจากสถานะปกติ. ต้องกำหนด visual states ของปุ่มให้ครบ: default, hover, pressed, keyboard focus, disabled และ loading หาก API รองรับ.
- ระบุสัญญา style ที่แอปใช้ผ่าน `packages/gui`: สีพื้นและข้อความ, ขอบ, radius, padding, ความสูง, การจัดข้อความ และ variant เช่น primary/secondary. หลีกเลี่ยงการให้แอปวาด Canvas แทน Button เพียงเพื่อซ่อนกรอบ focus เพราะจะเสียการใช้งานด้วยคีย์บอร์ด.
- แยก focus ที่มาจากคีย์บอร์ดกับการคลิกเมาส์ให้เหมาะสม: focus ต้องเห็นชัดเมื่อเดินด้วย Tab แต่ไม่ควรมีกรอบสีน้ำเงินค้างหลังคลิกเมื่อมี Active/Pressed style อยู่แล้ว. รักษา Enter/Space, disabled policy และ UI Automation Invoke.

### 2. TextInput และ field

- หน้า Contacts แสดงช่องกรอกแบบ native EDIT กรอบเหลี่ยมเส้นบางและกว้างเต็มการ์ดโดยไม่มีลำดับความสำคัญทางสายตา. `gui.field` ยังควบคุมพื้นหลัง ขอบ radius ระยะในช่อง ความสูง สี placeholder/focus/error และความกว้างสูงสุดของ input ไม่ได้ในระดับเดียวกับ Panel/Button.
- กำหนด API styling สำหรับ TextInput/field โดยรักษา IME, caret, selection, clipboard, undo และการเชื่อม label กับ accessibility. ให้แยก layout ของ label, helper text, validation message และความกว้างของฟอร์มจากการวาดตัว input.
- ตรวจ placeholder, focus, disabled และ error state ที่ DPI หลายระดับ. ช่อง Note ในตัวอย่างยังเป็น TextInput บรรทัดเดียว; หากต้องการหลายบรรทัด ให้แยกเป็นงาน control ใหม่และระบุข้อจำกัดอย่างชัดเจน.

เกณฑ์ตรวจรับร่วม: หน้า Components และ Contacts ใช้ `packages/gui` อย่างเดียวสำหรับ UI, ปุ่มและช่องกรอกดูเป็นระบบเดียวกับการ์ด/ตัวอักษร, ไม่มีการแก้หน้าตาด้วย Canvas ที่ทำให้ keyboard accessibility ถอยหลัง และผู้ใช้ตรวจภาพจริงก่อนถือว่าผ่าน.

## งาน Runtime 0.1.31

1. ออกแบบ `std.ui.maximizeWindow(window: UiWindow)` หรือทางเลือกที่กำหนดสถานะตั้งแต่ `openWindow` โดยระบุ signature และพฤติกรรมที่แน่นอนก่อนลง source. หากใช้คำสั่งแยก ต้องเรียกหลัง `openWindow` และก่อนหรือหลัง `mount` ได้อย่างคาดเดาได้.
2. บน Windows ให้ใช้การ maximize ของระบบปฏิบัติการ ไม่ใช้ขนาดหน้าจอที่คำนวณเอง. ต้องเคารพ work area, taskbar, DPI และจอที่หน้าต่างอยู่; เมื่อ restore ให้กลับสู่ขนาดปกติที่กำหนดไว้.
3. หลังเปลี่ยนสถานะ `viewportWidth/viewportHeight`, layout, native TextInput และ event `Resized` ต้องสะท้อนขนาดจริง. การเรียกซ้ำต้องปลอดภัย และการเรียกกับหน้าต่างที่ปิดแล้วต้องคืนข้อผิดพลาดตามรูปแบบ UI Runtime.
4. ระบุพฤติกรรมบนแพลตฟอร์มที่ยังไม่มี native UI host และอย่าให้ API ที่ประกาศใช้ได้แต่ทำงานเงียบ ๆ โดยไม่มีผล.

## งาน GUI package

- เพิ่ม wrapper `gui.maximizeWindow(window)` พร้อมคู่มือและตัวอย่างการเปิดหน้าต่างแบบ maximized. GUI package ที่มากับ Rev.31 ใช้เวอร์ชัน 0.1.31 เดียวกับ Craft.
- เพิ่ม style/variant API ของ Button และ TextInput/field ตามสัญญา UI controls ข้างต้น โดยไม่ให้แอปต้องเรียก `std.ui` ตรงหรือเขียน Canvas แทน control พื้นฐาน.
- ตัวอย่างแอปที่ใช้ API ใหม่ต้องใช้ Runtime 0.1.31 ขึ้นไป; รุ่นก่อนหน้าคงใช้ GUI package 0.2.8.

## เกณฑ์ตรวจรับที่ผู้ใช้ทดสอบเอง

| กรณี | ผลที่ต้องเห็น |
| --- | --- |
| เปิดและ maximize | หน้าต่างขยายเต็ม work area ของจอปัจจุบัน โดยไม่บัง taskbar |
| Restore | กลับสู่ขนาดปกติของหน้าต่างโดยไม่สูญเสียเนื้อหาหรือสถานะของแอป |
| Resize/DPI | layout, ScrollView และ TextInput อยู่ในตำแหน่งถูกต้อง; viewport/event รายงานขนาดจริง |
| หลายจอ | maximize บนจอที่หน้าต่างอยู่ รวมถึงจอที่มี DPI ต่างกัน |
| Lifecycle | เรียกซ้ำได้; หน้าต่างที่ปิดแล้วคืน error ที่ชัดเจน |
| Button visual states | Primary/Secondary สอดคล้องกัน; mouse click ไม่ทิ้งกรอบ focus เด่น; Tab มี focus ที่มองเห็น; Enter/Space และ disabled ยังถูกต้อง |
| TextInput/field | ขอบ ระยะในช่อง ขนาด สี placeholder/focus/error และความกว้างฟอร์มเป็นระบบเดียวกับ UI; IME/caret/selection/label ยังใช้งานได้ |

ผู้ใช้เป็นผู้ตรวจพฤติกรรม GUI จริง. ผล build หรือ static check ไม่แทนการตรวจหน้าต่างบน Windows.

## สถานะ implementation

- เพิ่ม `std.ui.maximizeWindow` บน Windows และ wrapper `gui.maximizeWindow`; ตัวอย่าง Craft UI Showcase เรียกหลังเปิดหน้าต่าง. Windows คำนวณ work area, taskbar และจอที่ใช้ตามกลไก maximize ของระบบ. Runtime ปรับ layout จากข้อความ `WM_SIZE`.
- เพิ่ม `std.ui.inputStyle` และ `gui.inputStyle` สำหรับ TextInput ที่ยังใช้ native EDIT; `gui.field` ได้ขอบมน ระยะในช่อง ความสูง และความกว้างสูงสุด. เพิ่ม `gui.secondaryButton` และจำกัด focus ring ของ Button/Canvas target ให้แสดงจากการเดินด้วยคีย์บอร์ด.
- ยังไม่มี API สีพื้น input อื่นนอกจากขาว, สี placeholder, TextInput หลายบรรทัด, loading state ของ Button หรือ field helper สำหรับ validation message โดยตรง. Error state ใช้สีขอบที่ส่งให้ `inputStyle` ขณะไม่ได้ focus ได้. ต้องตรวจ mouse/keyboard focus, IME, DPI, หลายจอ, restore, accessibility และสถานะหลังปิดหน้าต่างด้วยผู้ใช้.
- แอป Showcase ใช้ `gui.*` สำหรับ UI; `std.convert`, `std.fs` และ `std.json` ยังใช้กับการแปลงค่าและบันทึกข้อมูลตามเดิม.
- Dependency `gui = { path = "..." }` ไม่ต้องระบุเวอร์ชันและไม่รับ `version` ที่เลือกเอง; parser เติมเวอร์ชัน Craft 0.1.31 ให้ แล้ว resolver ตรวจว่า `packages/gui/craft.toml` ตรงกัน. Lock file ยังบันทึก `gui@0.1.31` และ checksum เพื่อให้ตรวจความตรงกันได้. Package อื่นยังต้องระบุเวอร์ชันตามเดิม.
