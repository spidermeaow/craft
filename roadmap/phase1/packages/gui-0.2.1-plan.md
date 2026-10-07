# Craft Phase 1 Rev.24 — แผนพัฒนา `packages/gui` 0.2.1

> แผนย่อยของ [Craft Phase 1 Rev.24](../craft-phase1-rev24.md). `0.2.1` คือเวอร์ชัน package ไม่ใช่ Craft version.

วันที่: 7 ตุลาคม 2026  
สถานะ: **ลง source รอบแรกแล้ว; รอผู้ใช้ทดสอบตาม `../output/gui-0.2.1-status.md`**  
ฐาน: Craft UI Runtime 0.1.24 และต้นแบบ `Archive/packages/gui`

## เป้าหมาย

พัฒนา `packages/gui` ในหัวข้อ package ของ Phase 1 Rev.24 ให้เป็น **library ที่เขียนด้วย Craft** สำหรับประกอบ UI จาก `std.ui` ให้ผู้พัฒนาเขียน desktop GUI ได้ง่าย อ่านง่าย และใช้ซ้ำได้ แอปยังรันบน native Windows UI Engine ของ Craft ไม่มี WebView, React, HTML/CSS หรือเครื่องมือออกแบบแบบลากวาง เป้าหมายคือความไวและความลื่นขณะใช้งาน ไม่ใช่ลดขนาด package ให้เล็กที่สุด

การพัฒนา package ไม่ได้ประกาศว่า Craft พร้อม V1.00 หาก package พบว่าขาด native primitive ให้บันทึกเป็นงาน Runtime และเพิ่มเฉพาะส่วนที่จำเป็นต่อผลลัพธ์ 0.2.1

## จุดตั้งต้นและข้อจำกัดที่ต้องยอมรับ

- `Archive/packages/gui` เป็น wrapper รุ่น 0.1.20: ยังไม่ครอบคลุม `size`/`splitView`, `textInput`, Canvas และ event ใหม่ จึงใช้เป็นข้อมูลอ้างอิง ไม่คัดกลับมาทั้งชุดโดยไม่ทบทวน
- Runtime มี UI tree และ `mount`/`update`, layout แนวแถว/คอลัมน์, scroll/virtual list, text/button/text input และ Canvas ตาม source 0.1.24 แต่ Canvas ยังรอการตรวจรับ GUI จริงตาม [สถานะ Rev.24](../output/phase1-rev24-status.md)
- ตอนนี้หน้าต่างบน Windows จำกัดหนึ่งบานต่อ interpreter run; `craft build` ยังเป็น source bundle ที่ต้องใช้ Craft runtime ไม่ใช่ standalone executable
- ปัจจุบัน Runtime ไม่มี hover/pressed/focus state สำหรับ custom widget ครบ, event propagation, general popup, animation, styled editable text หรือ per-shape hit test ดังนั้น package ต้องไม่โฆษณาว่ามีความสามารถเหล่านี้แล้ว

## ขอบเขต 0.2.1

### 1. เปิด package กลับมาอย่างเป็นระบบ

สร้าง `packages/gui` เป็น package ปกติที่เลือกติดตั้งได้ด้วย manifest/version ของ Craft แยก public API จาก helper ภายในให้ชัด กำหนด minimum Craft Runtime ที่รองรับเป็น 0.1.24 และให้ version ของ package เริ่มที่ 0.2.1 บันทึก dependency ใน `craft.toml`/`craft.lock` ตามระบบ package ที่มีอยู่ หลีกเลี่ยงการผูกแอปทั่วไปกับ GUI โดยอัตโนมัติ

ทบทวน `craft gui init` ที่ถูกพักใน Rev.21: ให้เปิดกลับเฉพาะเมื่อการคัดลอก package, การเพิ่ม dependency และการปฏิเสธการเขียนทับข้อมูลเดิมทำได้ปลอดภัย ถ้า package install แบบทั่วไปให้ประสบการณ์ดีกว่า ให้ใช้ทางนั้นเป็นวิธีหลักและระบุเหตุผลในคู่มือ ไม่สร้างช่องทางติดตั้งสองแบบที่ให้ผลต่างกันโดยไม่จำเป็น

### 2. API ระดับใช้งานจริงที่ยังเป็น Craft

ให้ public API ช่วยลด boilerplate ของ `std.ui` โดยไม่ซ่อน event loop จนแก้ปัญหายาก:

- window/session helper สำหรับเปิด, mount, update, รับ event และปิด พร้อม lifecycle/error ที่ชัดเจน
- layout helpers สำหรับ row, column, panel, spacing, sizing, alignment, split, scroll และ virtual list
- component พื้นฐาน: text, button, text input, card/surface, toolbar และรายการนำทางอย่างง่าย โดยประกอบจาก native node ที่มีจริง
- custom drawing helper สำหรับสี, Canvas และ image เมื่อตรวจรับ 0.1.24 แล้ว; ระบุทาง fallback และข้อจำกัดเรื่อง path ของภาพ
- theme tokens ขั้นต้น เช่น palette, spacing และขนาดตัวอักษร/องค์ประกอบ เท่าที่ Runtime ปรับภาพได้จริง; token ที่ยังไม่มี native support ต้องไม่ทำให้ผู้ใช้เข้าใจว่าปรับได้

ชื่อและ signature ในหัวข้อนี้เป็น **เป้าหมาย API** ไม่ใช่คำสั่ง Craft ที่ใช้ได้แล้ว ก่อน implement ต้องตรวจ syntax/type checker จริงและเขียนตัวอย่างที่ `craft check` ผ่าน หลีกเลี่ยง DSL แบบ JSX/hooks หรือ callback syntax ที่ภาษายังไม่รองรับ

### 3. State และ event แบบตรงไปตรงมา

รุ่นแรกให้แอปเก็บ state ใน Craft และสร้าง tree ใหม่ด้วย function เมื่อ event ทำให้ state เปลี่ยน โดยใช้ node ID คงที่และ `update` ของ Runtime เพื่อรักษา scroll/focus ตามขอบเขตที่รองรับ จัด event routing ตาม `nodeId` และ helper สำหรับ `ButtonPressed`, `TextChanged`, `VisibleRangeChanged`, `CanvasPressed`, `CloseRequested` โดยไม่ทำ event สูญหาย

ไม่ประกาศ observable/reactive state อัตโนมัติใน 0.2.1 จนกว่าจะมี primitive และต้นทุนที่พิสูจน์ได้ ถ้า helper ลด boilerplate ได้โดยไม่เพิ่มการ rebuild ที่ไม่จำเป็น ให้ใส่ได้พร้อม benchmark

### 4. Styling ที่มีขอบเขตจริง

สร้างรูปลักษณ์เริ่มต้นที่สม่ำเสมอด้วย palette, spacing, icon/image และ Canvas ในส่วนที่เหมาะสม พร้อมกำหนดว่า control ใดมี native appearance เช่น TextInput และ control ใดวาดเองได้ Package ไม่ควรจำลอง input ด้วย Canvas เพียงเพื่อให้สวย หากยังขาด IME, caret, selection หรือ accessibility ที่จำเป็น

เป้าหมายด้านภาพของ 0.2.1 คือ **shell และ components พื้นฐานที่ดูเป็นระบบ** ไม่อ้างว่ามี animation, shadow, blur, gradient, dropdown หรือ table/grid สมบูรณ์แล้ว จัดรายการดังกล่าวเป็นงานต่อไปพร้อม native dependencies ที่ต้องเพิ่ม

## ตัวอย่างผลลัพธ์ที่ต้องมี

สร้างตัวอย่าง Craft GUI ขนาดเล็กที่ไม่ใช่ Craft Office หรือ ERP: หน้าต่างหนึ่งบาน มี header/sidebar/content, ปุ่มเปลี่ยน state, text input, scroll และรายการจำลองจำนวนมาก; เพิ่ม Canvas/image เฉพาะเมื่อ Rev.24 ผ่านการตรวจรับ ตัวอย่างต้องแสดงการแยก `screen(state)` ออกจาก event loop และใช้ package import จริง ไม่เรียก `std.ui` ตรงจากโค้ดแอปใน flow หลัก

จัดทำคู่มือเริ่มต้น, API reference, ตารางเทียบ `gui` ↔ `std.ui`, วิธีติดตั้ง/อัปเกรด, ข้อจำกัดและ troubleshooting สำหรับ Windows รวมถึงข้อเท็จจริงว่า bundle ยังต้องใช้ Craft Runtime

## ลำดับ implement

1. ตรวจ source/API และผลทดสอบ Rev.24 ล่าสุด; กำหนด minimum runtime และรายการ native blockers ที่จำเป็นจริง
2. ออกแบบ public surface และ signature ด้วย Craft syntax ปัจจุบัน; ทำตัวอย่างเล็กที่ type checker รับได้
3. สร้าง `packages/gui` จาก API ใหม่ ย้ายส่วนที่ยังเหมาะจาก Archive และกำหนด version/dependency
4. ทำ lifecycle, layout, event/state และ components รุ่นแรก; ต่อ Canvas/styling เท่าที่ backend รองรับ
5. เชื่อม workflow ติดตั้งหรือ `craft gui init` ที่เลือก พร้อมป้องกันการเขียนทับ package ในโปรเจกต์ผู้ใช้
6. เพิ่มตัวอย่าง, คู่มือ, migration จาก wrapper 0.1.20 และ status ที่แยกสิ่งที่ implement/ตรวจแล้ว/ยังรอผู้ใช้ตรวจ

## เกณฑ์ตรวจรับ

| ด้าน | ผลที่ต้องพิสูจน์ |
| --- | --- |
| Package | โปรเจกต์ใหม่ติดตั้ง `gui` ได้โดยไม่แก้ source ของ Runtime; dependency/version/lock ตรงกัน; ไม่เขียนทับ package เดิม |
| API | ตัวอย่างใช้ `import gui "gui"` และ `craft check` ผ่าน; public API ไม่ต้องเรียก native detail สำหรับ UI พื้นฐาน |
| Interaction | คลิก, พิมพ์ผ่าน IME, scroll, virtual list, resize และปิดหน้าต่างได้; state เปลี่ยนแล้ว UI ตรงโดยไม่สูญ scroll/focus เกินสัญญา Runtime |
| ภาพ | Shell, typography/spacing และ Canvas ที่เปิดใช้แสดงถูกบน DPI ที่ตรวจ; บอกความต่างของ Direct2D/GDI และข้อจำกัดภาพ |
| Performance | เทียบ workload เดียวกันระหว่างใช้ `std.ui` ตรงกับผ่าน package: startup, input-to-draw, frame/paint time, idle CPU/RAM และจำนวน node; package ไม่เพิ่มงานซ้ำอย่างมีนัยสำคัญโดยไม่มีเหตุผล |
| เอกสาร | ขั้นตอนติดตั้ง/รัน/อัปเกรดทำตามได้ และระบุ Windows, single-window, source bundle และ native blockers ตรงความจริง |

ตามความต้องการปัจจุบัน **ผู้ใช้เป็นผู้ทดสอบ GUI เอง** ในรอบ implement ถัดไป เว้นแต่จะสั่งให้ผู้ช่วยทดสอบเป็นการเฉพาะ แผนนี้ยังไม่ได้รันทดสอบหรือรับรองผลข้างต้น

## หลัง 0.2.1

ค่อยเพิ่ม component ที่ซับซ้อน (tabs, menu/popup, table/grid, dialogs), theme/styling ที่ปรับได้จริง, state/reactivity, animation และ accessibility จากผลทดสอบของ package และ Runtime อย่าใช้ตัวอย่าง Craft Office เป็นงานตรวจรับของช่วงนี้
