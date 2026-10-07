# Craft Phase 1 Rev.22 — Text Input และ Focus สำหรับ UI Engine

วันที่: 7 ตุลาคม 2026  
ฐาน: Craft 0.1.21  
เป้าหมายรุ่น: Craft 0.1.22  
สถานะ: implement TextInput รุ่นแรกใน source แล้ว; รอผู้ใช้ทดสอบและยังมีข้อจำกัดตาม `roadmap/phase1/output/phase1-rev22-status.md`

## ทิศทาง

Rev.22 พัฒนา **Craft UI Engine ใน Runtime** ต่อจากหน้าต่าง, UI tree, Direct2D/DirectWrite, ScrollView และ Virtual List ที่มีอยู่. เป้าหมายคือให้แอป Craft รับและแก้ไขข้อความในหน้าต่าง native ได้จริง โดยเริ่มจากช่องข้อความบรรทัดเดียวที่ตอบสนองเร็วและรองรับภาษาไทย. งาน UI ยังคงมีลำดับสูงสุดใน Phase 1 ตามการตัดสินใจของเจ้าของโครงการ; เลื่อนการทดสอบ HTTP ร่วมกับ UI ไปหลัง UI Engine พร้อมขึ้น. `std.http` และ HTTP Engine ที่มีอยู่ยังคงเดิม.

`packages/gui` และ `packages/api-framework` อยู่ใน `Archive/` ต่อไป. ตัวอย่าง Rev.22 ต้องเรียก `std.ui` โดยตรง. ไม่เริ่ม Craft Office และไม่สร้างเครื่องมือออกแบบหน้าจอแบบลากวาง.

## ฐานและปัญหาที่ต้องแก้

- `std.ui` ปัจจุบันมี `Text`, `Button`, `Row`, `Column`, `Panel`, `ScrollView`, `VirtualList`, window และ event. ยังไม่มี node สำหรับแก้ไขข้อความ.
- `KeyPressed` บอกปุ่มที่กด แต่ไม่ใช่สัญญารับข้อความที่ผ่าน keyboard layout หรือ IME. การพิมพ์ภาษาไทยและข้อความที่มี Unicode หลาย code point ต้องมี lifecycle ของ composition แยกจาก key event.
- Focus ปัจจุบันผูกกับ button บางส่วน; ต้องขยายเป็นระบบที่ใช้ร่วมกันระหว่าง button และ text input โดยไม่ทำให้ focus หายเมื่อ `std.ui.update` สร้าง tree ใหม่.
- รอบทดสอบ 0.1.21 ยืนยันว่า GUI เปิดด้วย `direct2d-hardware` ได้ แต่ยังไม่ได้ตรวจภาพ, click, scroll และ resize ด้วยตา; ก่อนอ้างว่าฟีเจอร์ใหม่พร้อม ต้องตรวจสิ่งเหล่านี้ร่วมด้วย.

## ผลลัพธ์ที่ Rev.22 ต้องส่งมอบ

ช่องข้อความบรรทัดเดียวใน UI tree ที่คลิกเพื่อ focus, พิมพ์ภาษาอังกฤษ/ไทย, แก้ไข, เลือก, ลบ, copy/cut/paste และส่งค่าให้ Craft application ผ่าน event ได้. Caret, selection และข้อความระหว่าง IME composition ต้องแสดงตรงกับตำแหน่งจริงทั้งบน Direct2D/DirectWrite และ GDI fallback. การ update tree, scroll และ resize ต้องไม่ทำให้ข้อความหรือ focus หายโดยไม่ตั้งใจ.

## สัญญา API และสถานะข้อความ

1. ออกแบบ `std.ui.textInput(id, initialValue, placeholder): UiNode` หรือชื่อที่สอดคล้องกับ API ปัจจุบัน ก่อนลงมือ native code. เขียนตัวอย่าง Craft ที่ใช้ syntax จริงและ `craft check` ได้; ไม่ใช้ callback/lambda ที่ภาษายังไม่มี.
2. ให้ engine เป็นเจ้าของ **editing state ชั่วคราว** ของ node ID: committed text, caret, selection, focus และ IME composition. `initialValue` ใช้เมื่อ node ID ปรากฏครั้งแรก; `std.ui.update` ที่ใช้ ID เดิมต้องไม่ทับข้อความระหว่างพิมพ์. กำหนด API สำหรับเปลี่ยนค่าจากโปรแกรมอย่างชัดเจน เช่น `std.ui.setText(window, id, value)` พร้อมนโยบายเมื่อกำลัง composition; ห้ามใช้ค่าเก่าจาก tree update ทับ input ใหม่.
3. เพิ่ม event ที่มี `nodeId` และข้อความที่ commit แล้ว เช่น `TextChanged` และ `TextSubmitted`. ระบุลำดับระหว่าง focus, input, composition, update และ close. Composition ที่ยังไม่ commit ไม่ออกเป็น `TextChanged`; event ที่ส่งถึง Craft ต้องเป็น UTF-8 ที่ถูกต้อง. ขยาย `UiEvent` แบบคงความเข้ากันได้กับ field/event เดิม.
4. จำกัดความยาวข้อความและจำนวน event/หน่วยความจำตามเหตุผล พร้อม error ที่อ่านเข้าใจได้. ระบุพฤติกรรมเมื่อ node หาย, เปลี่ยน ID หรือ window ปิดระหว่าง composition.

## งาน UI Engine

1. **Focus และ keyboard:** mouse focus, Tab/Shift+Tab ระหว่าง button/text input, focus ring, Enter/Space ของ button และ Enter ของช่องข้อความ. Focus ต้องคงอยู่บน node ID เดิมหลัง update; ถ้า node ถูกลบให้ย้ายหรือ clear focus ตามกฎที่ระบุ. `KeyPressed` เดิมยังใช้ได้และไม่ถูกส่งซ้ำจนเกิด action สองครั้ง.
2. **การแก้ข้อความ:** caret, click เพื่อวาง caret, Home/End, Left/Right, Backspace/Delete, Shift selection, Ctrl+A/C/X/V และ paste จาก clipboard. การแก้ไขต้องไม่แยก surrogate pair หรือชุดอักขระที่ผู้ใช้มองเป็นหน่วยเดียวโดยไม่ตั้งใจ; ระบุ Unicode boundary ที่รองรับจริงในเอกสาร.
3. **IME:** เชื่อม Windows text input/composition กับ engine โดยไม่อนุมานตัวอักษรจาก virtual key. แสดง pre-edit, caret และ candidate window ใกล้ตำแหน่งพิมพ์; commit/cancel อย่างถูกต้องเมื่อ focus เปลี่ยน, scroll, resize, update หรือปิด window. ทดสอบภาษาไทยและ IME อย่างน้อยหนึ่งภาษาที่มี composition. เลือก Windows API ที่เหมาะกับ custom text control จากผล spike และบันทึกเหตุผลก่อน implement เต็ม.
4. **Layout และ renderer:** วัดข้อความและ placeholder ตาม font/DPI จริง; clip ข้อความใน input, เลื่อนข้อความแนวนอนเมื่อ caret ออกนอกช่อง, วาด caret/selection/composition โดย redraw เฉพาะพื้นที่ที่เปลี่ยน. ตรวจ Direct2D/DirectWrite และ GDI fallback ให้ได้พฤติกรรมเดียวกัน. Cursor blink ต้องหยุดเมื่อไม่ focus และไม่ทำให้ idle render ต่อเนื่องทั้งหน้าต่าง.
5. **Event loop และ responsiveness:** UI thread ห้ามรอ Craft callback ระหว่าง OS input/composition. Coalesce การวาดหรือ event ที่ทำได้โดยไม่ทิ้ง committed text, click, key, focus หรือ close. เก็บ trace ของเวลา input ถึงจบ draw, จำนวน layout/paint pass, input queue และ frame ที่เกิน budget; แยก draw completion ออกจากเวลาภาพปรากฏจริง.
6. **ความเข้ากันได้:** Button, ScrollView, Virtual List และหน้าต่างเดิมต้องทำงานเหมือนเดิม. Text input ที่อยู่ใน ScrollView ต้อง hit test/focus ถูกหลัง scroll. `craft build` ยังเป็น source bundle; CLI/package system ไม่เปลี่ยนขอบเขต.

## ตัวอย่างและการตรวจรับ

เพิ่มหน้า **UI Input Lab** ใน `D:\02-Repository\craft-test\runtime\08-ui` หรือโปรเจกต์ UI ใหม่ใต้ `craft-test/runtime/` พร้อมคู่มือแยกหัวข้อ; ใช้ `std.ui` โดยตรง. หน้าเดียวควรมีหลายช่อง, ปุ่ม, scroll area และส่วนแสดงค่า `TextChanged`/`TextSubmitted` เพื่อสังเกตผล. `D:\02-Repository\test-gui` ใช้ต่อได้หากต้องการ stress UI เดิม แต่ไม่เป็น dependency ของ Rev.22.

| กรณี | เกณฑ์ผ่าน |
| --- | --- |
| English/Thai/Unicode | พิมพ์ แก้ ลบ เลือก copy/paste ได้ตรง; ข้อความใน Craft event ตรงกับสิ่งที่เห็น ไม่มีตัวอักษรซ้ำหรือหาย |
| IME composition | pre-edit และ candidate อยู่ใกล้ caret; commit/cancel ไม่เสียข้อความเดิมหรือส่ง `TextChanged` ซ้ำ |
| Focus | Tab/Shift+Tab, คลิก, update tree, ลบ node และปิดหน้าต่างให้ผลตามกฎ; focus ring ตรง node จริง |
| Editing | Home/End, arrows, Backspace/Delete, Shift selection, Ctrl+A/C/X/V และ Enter ใช้ได้; selection ไม่ทำให้ Unicode เสีย |
| Scroll/resize/DPI | caret, selection, hit test และ candidate position ถูกหลัง scroll/resize และบน DPI ต่างกัน; ไม่มีภาพค้าง |
| Regression | Button, ScrollView, Virtual List 100,000 แถวและ renderer fallback ยังใช้ได้; `craft check/run/build` ของตัวอย่างเดิมไม่เสีย |
| Performance | เทียบ baseline 0.1.21 กับ 0.1.22 บนเครื่อง/renderer/workload เดียวกัน; รายงาน p50/p95/p99 ของ input-to-draw, frame time, RAM และ idle CPU โดยไม่สรุปจากค่าเฉลี่ยเดียว |

ระหว่าง implement ให้เพิ่ม unit/integration checks สำหรับ text state, focus order, event ordering, Unicode boundary และกรณี node เปลี่ยน. การเปิด GUI เพื่อตรวจภาพ, IME, DPI และความรู้สึกลื่นให้ผู้ใช้ทดสอบตามความต้องการเดิม เว้นแต่ผู้ใช้ขอให้ผู้ช่วยทดสอบในงานนั้นโดยตรง. ระบุผลที่ตรวจได้และข้อจำกัดไว้ใน `roadmap/phase1/output/phase1-rev22-status.md` พร้อมคู่มือ `docs/TRY-REV22.md`.

## ลำดับทำงาน

1. ตรวจ baseline 0.1.21 จาก `craft-test`, พฤติกรรม focus/button ปัจจุบัน และ native message flow; ทำ spike ของ Windows text input/IME ก่อนล็อก API.
2. กำหนดสัญญา `TextInput`, editing state และ event ด้วยตัวอย่าง Craft ที่ type-check ได้.
3. ทำ focus และ editing core รวม Unicode, selection และ clipboard; เชื่อม native composition.
4. เชื่อม layout/rendering ทั้ง Direct2D และ GDI; ตรวจ scroll, resize, DPI และ device/fallback path.
5. เตรียม UI Input Lab, trace, เอกสารและ manual acceptance; รับผลทดสอบจากผู้ใช้แล้วแก้ regression ก่อนปิดรุ่น.

## นอกขอบเขต 0.1.22

TextArea หลายบรรทัด, rich text, syntax highlighting, password field, spellcheck, drag & drop, animation/styling API เต็มรูปแบบ, Virtualized Grid/Table, accessibility provider แบบครบ, Linux/macOS GUI, standalone native application binary, GUI library ใน revision ถัดไปของ Phase 1, Craft Office และการทดสอบ HTTP ร่วมกับ UI. สิ่งเหล่านี้เป็นงานต่อไปและต้องไม่ถูกอ้างว่าเสร็จจาก Text Input รุ่นแรก.

## แหล่งอ้างอิงสำหรับงาน Windows text input

- [Microsoft Text Services Framework](https://learn.microsoft.com/en-us/windows/desktop/TSF/text-services-framework) สำหรับประเมินเส้นทาง IME ของ custom text control.
- [Microsoft UI Automation control patterns](https://learn.microsoft.com/en-us/windows/win32/winauto/uiauto-controlpatternsoverview) สำหรับออกแบบข้อมูล text/focus ที่จะต้องเปิดให้ accessibility ในระยะถัดไป.
