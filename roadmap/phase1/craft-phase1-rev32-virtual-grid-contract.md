# Rev.32 design contract: generic Virtual Grid

วันที่: 8 ตุลาคม 2026  
สถานะ: **สัญญาออกแบบสำหรับ revision ถัดไป; ยังไม่มี public API หรือ implementation**

## จุดประสงค์และขอบเขต

Virtual Grid เป็น component ตารางสองมิติทั่วไปสำหรับข้อมูลจำนวนมาก เช่น database viewer, log explorer และ spreadsheet. Craft UI Engine จัด viewport, พิกัด, การวาด, hit test และ accessibility ของรายการที่มองเห็น. แอปเป็นเจ้าของข้อมูล การเลือกที่มีความหมายต่อโดเมน การแก้ไข และคำสั่งธุรกิจ. Virtual Grid ต้องไม่รู้จัก sheet, สูตร, cell reference, Pivot หรือไฟล์ XLSX.

## สัญญาระหว่างแอปกับ UI host

สัญญานี้เป็นรูปแบบเชิงแนวคิด **ไม่ใช่ syntax ของ Craft ที่ใช้ได้แล้ว**. ก่อนเพิ่ม API ต้องกำหนดชื่อและชนิดจริงใน `std.ui`/`packages/gui` และทำให้ `craft check` ตรวจได้.

| ส่วน | Input/Output | เงื่อนไข |
| --- | --- | --- |
| Grid definition | id, rowCount, columnCount, viewport size, default row height/column width, overscan และ optional fixed rows/columns | จำนวนและพิกัดใช้ integer ที่ไม่ overflow; ไม่สร้าง node ทุกช่อง |
| Size overrides | ช่วงหรือรายการ row height/column width ที่ต่างจากค่าเริ่มต้น | ไม่ต้องส่งขนาดทุกแถว/คอลัมน์; เปลี่ยนขนาดแล้ว viewport และ hit test ต้องตรงกัน |
| Visible range request | grid id, generation, rowStart/rowEnd, columnStart/columnEnd แบบ end-exclusive | ส่งเมื่อ scroll/resize/freeze เปลี่ยนช่วงที่ต้องใช้; รวม overscan; coalesce คำขอที่ถูกแทนที่ |
| Visible data response | generation และข้อมูลเฉพาะช่วงที่ร้องขอ เช่น text, style token, accessible name และสถานะ loading/error | ข้อมูลที่มาช้าจาก generation เก่าต้องไม่ทับ viewport ใหม่; app เลือกแสดง placeholder ได้ |
| Invalidate | id พร้อม row/column หรือ range ที่เปลี่ยน | วาดและ measure เฉพาะส่วนที่ได้รับผลกระทบเมื่อ geometry ไม่เปลี่ยน; geometry เปลี่ยนต้อง relayout เฉพาะขอบเขตที่จำเป็น |
| Hit target | row, column, local position, pointer/keyboard action และ modifiers | ระบุ item จากพิกัดได้โดยไม่สร้าง `UiNode` ต่อช่อง |
| Accessibility | virtual item identity, row/column, label, selected/focused และ action | screen reader ขอข้อมูล item ใน viewport และเดินด้วยคีย์บอร์ดได้โดยไม่ materialize ทั้ง grid |

UI host ส่ง request/event ผ่านคิวตามรูปแบบปัจจุบัน; ไม่เรียก Craft callback ตรงจาก Windows thread. แอปตอบกลับด้วยคำสั่ง public API. ห้ามให้ renderer ถือ reference ไปยัง mutable workbook ของแอป. App state ต้องถูกส่งเป็น snapshot/batch ที่มี generation เพื่อป้องกันผลลัพธ์ย้อนหลัง.

## พิกัดและการวาด

- Scroll สองแกนใช้ logical pixels และช่วงแถว/คอลัมน์แบบ 64-bit; แปลงเป็นพิกัดจอเฉพาะช่วงที่มองเห็นก่อนส่ง Direct2D หรือ GDI. การเลือก item ที่พิกัดไกลต้องไม่คูณ row × column เพื่อหาตำแหน่งใน array ขนาดทั้งตาราง.
- Fixed rows/columns เป็น generic pinned regions. แอป Spreadsheet ประกอบเป็น frozen panes เอง. การซ่อน การจัดกลุ่ม และ merge มี model ในแอป; Engine ต้องเปิดทางให้แอป map visible index และวาด span โดยไม่ทำลาย hit test.
- ชั้นวาดแยก background, item content, separators, hover/focus และ overlay. Hit region ต้องใช้ geometry เดียวกับสิ่งที่วาด. การเปลี่ยนข้อความของ item เดียวไม่ควร invalidate ทั้ง viewport หากขนาดไม่เปลี่ยน.
- Text measurement/cache ผูกกับ font, DPI, width และ content. DirectWrite กับ GDI fallback ต้องใช้ alignment และ clip ตามสัญญาเดียวกัน. การย้ายจอ/เปลี่ยน DPI ต้องล้าง cache ที่ขึ้นกับ scale.

## Input และ ownership

Craft จัด pointer capture, focus routing, keyboard events, IME/text editor primitive และ clipboard payload ทั่วไป. แอปกำหนดกฎ selection, commit/cancel, undo/redo, copy/paste และทางลัดของโดเมน. Spreadsheet app จึงกำหนด F2, Shift+Arrow, autofill และ paste special โดยไม่เพิ่มเงื่อนไขเหล่านี้ใน Runtime.

Editor เป็น native text input หรือ component ที่รองรับ IME และวางทับ geometry ของ item ที่แอปเลือก. เมื่อ scroll/resize ระหว่างแก้ไข แอปตัดสินใจว่าจะเลื่อน editor ตาม item, commit หรือ cancel; Engine ต้องรายงาน viewport/ตำแหน่งให้ตัดสินใจได้. Clipboard ของหลาย item ใช้ข้อมูลที่แอป serialize; Engine รับผิดชอบการอ่าน/เขียน clipboard ของระบบโดยไม่ตีความข้อมูล Spreadsheet.

## Responsiveness และขอบเขตทรัพยากร

- คำขอ visible range ที่เก่าต้องยกเลิกหรือข้ามได้; แอปที่โหลดข้อมูลช้าต้องไม่ทำให้ UI thread รอ. จำกัดจำนวน request ที่ค้างและแจ้ง overflow อย่างชัดเจน.
- Memory ของ UI tree, layout boxes และ draw commands ต้องสัมพันธ์กับ viewport + overscan, fixed regions และ cache ที่มีเพดาน ไม่สัมพันธ์กับ `rowCount × columnCount`. ข้อมูลจริงของแอปอาจโตตามจำนวน item ที่มีค่า.
- เพิ่ม trace สำหรับ data request latency, queue delay, measure, layout, draw และ input-to-draw. `CRAFT_UI_TRACE` ปัจจุบันจบที่ draw completion จึงไม่ใช่ FPS/display presentation.

## เกณฑ์ตรวจรับเมื่อเริ่ม implement

1. ใช้ component เดียวกันทำ database viewer และ Spreadsheet demo โดยไม่มีชื่อ/API ที่อ้างสูตรหรือ Workbook ใน Engine.
2. เปิดตารางเชิงตรรกะ 1,048,576 × 16,384 และเลื่อนไปปลายช่วงได้ โดยจำนวน node/draw commands คงอยู่ตาม viewport. ไม่ต้องเก็บทุกช่องในหน่วยความจำ.
3. เปลี่ยนค่า item เดียวแล้วได้ dirty region ตาม item; เปลี่ยนความกว้างคอลัมน์แล้ว text, hit target, editor และ accessibility ตรงตำแหน่งใหม่.
4. เลื่อนสองแกนเร็ว ๆ ระหว่าง data source ตอบช้าแล้วไม่แสดงข้อมูล generation เก่าใน viewport ใหม่; UI ยังรับ input และปิดหน้าต่างได้.
5. ผู้ใช้ตรวจ mouse, keyboard, IME ไทย, screen reader, DPI 100%/125%, หลายจอ และ GDI fallback. เก็บ trace พร้อมสเปกเครื่องก่อนตั้งเป้าตัวเลข frame time เป็นเกณฑ์ผ่าน.

การออกแบบนี้ปิดงานสัญญาใน Rev.32 เท่านั้น. การเพิ่ม type, event, renderer และ component จริงเป็นงาน implementation ที่ต้องแยก revision และให้ผู้ใช้ทดสอบ.
