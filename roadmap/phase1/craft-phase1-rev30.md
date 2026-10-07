# Craft Phase 1 Rev.30 (0.1.30) — ปิดงาน Focus/Accessibility และทำ Command ให้พร้อมใช้

วันที่: 7 ตุลาคม 2026  
สถานะ: **ลง source แล้ว; รอผู้ใช้ทดสอบ GUI/IME/UIA จริง**  
ฐาน: Craft CLI/UI Runtime 0.1.29 และ `packages/gui` 0.2.7

## เป้าหมาย

ทำให้ desktop interaction ที่เริ่มใน Rev.29 ใช้งานได้ครบก่อนเพิ่ม control ชุดใหม่: focus ต้องไม่หลุดไปส่วนที่ใช้งานไม่ได้, shortcut ต้องมีขอบเขตและจัดการคำสั่งชนกันอย่างคาดเดาได้, และ Button/Canvas target ที่ Runtime วาดเองต้องถูกค้นพบผ่าน Windows UI Automation. Package ต้องห่อความสามารถเหล่านี้โดยไม่ย้าย state หรือ callback ของแอปเข้า UI thread.

Rev.30 เป็นรอบ **ปิดหนี้ของ Rev.29**. ห้ามนับ semantic metadata ที่อยู่ใน UiNode ว่าเป็น accessibility ที่พร้อมใช้ จนกว่าเครื่องมือ UI Automation อ่านได้จริง. หากงาน UIA หรือ focus scope ไม่ผ่านเกณฑ์ ให้รายงาน Rev.30 ว่ายังค้างแทนการเลื่อนเกณฑ์ตรวจรับออกเงียบ ๆ.

## ฐานที่มีแล้ว

- Runtime 0.1.29 มี Tab/Shift+Tab, `focus`, `focusPolicy` สำหรับ Button/TextInput, focus ring ของ Button/Canvas, การพยายามเลื่อน ScrollView ตาม focus, shortcut ระดับหน้าต่าง และ event `CommandInvoked`.
- `std.ui.accessible` เก็บ name/role บน UiNode แต่ยังไม่มี provider สำหรับ custom controls; Canvas target ยังไม่มี policy/semantic API ราย target.
- ยังไม่มี focus scope, การ restore focus ของ overlay, หรือ priority ของ shortcut ระหว่าง scope ที่ซ้อนกัน. TextInput เป็น native Windows EDIT.
- Package 0.2.7 มี wrapper และ command helper; ตัวอย่าง `gui-package-dashboard` รับ Ctrl+S/F5 แล้ว. ผล GUI/IME/assistive technology ยังรอผู้ใช้ทดสอบ.

## งาน Runtime 0.1.30

### P0 — Focus contract ให้ครบ

1. เพิ่ม focus scope แบบ stack ที่ผูกกับ subtree ID: เปิด scope แล้วบันทึก focus เดิม, เลือก entry target ที่เปิดใช้งาน, จำกัด Tab/Shift+Tab และ programmatic focus ให้อยู่ใน scope, ปิดแล้วคืน focus เดิมถ้ายังอยู่และใช้ได้. Scope ใช้เป็นฐานของ Dialog/Popup ในอนาคต; รอบนี้ยังไม่เพิ่ม widget เหล่านั้น.
2. เพิ่ม policy ของ Canvas hit region ทีละ target: enabled, tab stop, accessible name/role. Target ที่ disabled ต้องไม่รับ pointer/Enter/Space และไม่อยู่ใน traversal. Target ที่ถูกลบหรือซ่อนหลัง `update` ต้องไม่ค้างเป็น focused ID.
3. นิยาม `focus(window, id)` สำหรับ ID ที่ไม่มี, disabled, อยู่นอก scope, ไม่ปรากฏใน layout หรืออยู่ใน VirtualList ที่ยังไม่ materialize ให้คืน error แยกสาเหตุ. แก้ลำดับ focus หลัง mount/update/remove/resize/DPI change และกรณีหน้าต่างเสียแล้วได้ focus กลับ.
4. ตรวจ focus visibility สำหรับ ScrollView ซ้อนกันและ Canvas target: เลื่อนตามระยะจริงเมื่อ scroll ถูก clamp; focus ring ต้องอยู่ใน viewport และไม่ถูก scrollbar/parent clipping บัง. ค่า `tabStop=false` ยังอนุญาต programmatic focus ถ้า enabled และอยู่ใน scope.

### P0 — Windows UI Automation ขั้นแรก

1. ทำ provider สำหรับ tree ที่ Runtime วาดเองบน HWND หลัก: Text, Button, Panel/กลุ่มที่มีชื่อ และ Canvas target ที่ระบุ semantic metadata. ส่ง Name, ControlType/role, IsEnabled, IsKeyboardFocusable, HasKeyboardFocus, BoundingRectangle และลำดับ parent/child/sibling ที่ตรงกับภาพ. Native EDIT ต้องยังปรากฏเป็น TextInput ของ Windows โดยไม่เกิด node ซ้ำ.
2. Button และ Canvas target ที่ enabled เปิด Invoke pattern; invoke จาก assistive technology ต้องเข้า event queue เดียวกับการกดปุ่ม โดยส่ง ID เดียวและไม่เรียก Craft code บน UI thread. Label ของ `gui.field` ต้องเชื่อมกับ TextInput ให้ screen reader อ่านชื่อช่องได้.
3. เมื่อ `mount/update`, focus, enabled state, scroll, resize หรือ close เปลี่ยน ให้ provider ตอบค่าปัจจุบันและแจ้ง UIA event ที่จำเป็น. Provider ที่อ้าง node ถูกลบต้องไม่ชี้ข้อมูลเก่า; กำหนด owner/lifetime และการคืน COM/native resources ชัดเจน. การ query ต้องไม่รอ Craft event loop จน deadlock.
4. แยกการตรวจ `Narrator`/Accessibility Insights หรือ UIA client จาก static build. บันทึกข้อจำกัดของ provider รุ่นแรก เช่น ยังไม่มี custom Text pattern, selection/table patterns หรือ high-contrast theme เต็มระบบ.

### P1 — Command dispatch และ TextInput

1. เพิ่ม command binding ตาม scope โดยยังคง binding ระดับหน้าต่างเดิม. Priority ที่ต้องใช้: native EDIT/IME composition → focus navigation/system key → scope ที่อยู่บนสุด → window binding → raw key. `KeyPressed` เดิมยังส่งตามสัญญา แต่ `CommandInvoked` ต้องเกิดครั้งเดียวต่อการกดและไม่เกิดเมื่อ scope/command ใช้ไม่ได้.
2. กำหนด conflict diagnostics สำหรับ key/modifier เดียวกันใน scope เดียว; scope บนสุดอาจบัง window binding ได้อย่างตั้งใจ. Unbind/close/update ต้องลบ binding ตาม ownership โดยไม่ทิ้ง shortcut ค้าง. ระบุการจัดการ key repeat, Alt/system key, keyup และ modifier ค้างเมื่อหน้าต่างเสีย focus.
3. ตรวจ Ctrl+A/C/X/V/Z/Y, caret, selection, Enter และการพิมพ์ไทยผ่าน IME ใน TextInput. หากต้องให้ command แย่งคีย์แก้ไขข้อความ ให้มี opt-in เฉพาะ binding และค่าเริ่มต้นต้องปลอดภัยกับ EDIT.

## งาน GUI Package 0.2.8

- เพิ่ม wrapper/helper สำหรับ focus scope, Canvas target policy/semantics และ scoped command ด้วยชื่อที่สอดคล้องกับ API 0.2.7. Signature สุดท้ายต้องยืนยันกับ Craft checker ก่อนถือเป็น API เผยแพร่; ชื่อในแผนยังเป็นชื่อเชิงแนวคิด.
- ปรับ `field`, `primaryButton`, `navigation`, `appShell`, `adaptiveShell` ให้ส่ง semantic name/role และ enabled/tab policy ที่ถูกต้องโดยไม่สร้าง ID ซ้ำ. `field` ต้องเชื่อม label กับ native TextInput; Canvas `targetButton` ต้องมีชื่อที่ screen reader อ่านได้.
- เพิ่มตัวอย่างใน `gui-package-dashboard` สำหรับ scope ที่เปิด/ปิดได้, Canvas target disabled, Save/Refresh/Undo และสถานะ focus ปัจจุบัน; `responsive-tester` ใช้ตรวจ resize/scroll/focus. แอปยังรับ `CommandInvoked` ใน event loop แล้วแก้ state เอง.
- คง API 0.2.7 ที่ใช้ได้อยู่หรือเขียน migration ที่ชัดเจน. อัปเดต package copy, manifest, lockfile, คู่มือ และ Setup ให้เวอร์ชันตรงกัน.

## สัญญาที่ต้องตัดสินก่อน implement

1. รูปแบบ API ของ scope, target policy และ scoped shortcut ต้องใช้ชนิด Craft ที่มีจริง ไม่เพิ่ม generic หรือ callback โดยไม่จำเป็น.
2. เลือก UIA provider architecture และ threading model พร้อมเอกสาร ownership: HWND, provider object, UiNode snapshot, event emission และ lifetime ตอน close. ตรวจว่า native EDIT ไม่ถูก expose ซ้ำ.
3. กำหนดนิยาม `visible` ของ focus target เมื่ออยู่ใน ScrollView, VirtualList และ subtree ที่สลับ responsive layout; error ต้องแยก target ไม่มีใน tree กับยังไม่ materialize.
4. ระบุ event ordering ระหว่าง `KeyPressed`, `CommandInvoked`, `ButtonPressed` และ UIA Invoke ให้คงที่ และอย่าให้ source เดิมเกิด action สองครั้ง.

## ลำดับดำเนินงาน

1. รับผลทดสอบ 0.1.29 จากผู้ใช้เป็น baseline; แก้ regression ของ Tab, shortcut, TextInput/IME และ resize ก่อนขยาย API.
2. ปิด focus scope/Canvas target policy/visibility แล้วทบทวน GDI และ Direct2D ด้วย source/diff; ให้ผู้ใช้ตรวจ GUI จริง.
3. ลง UIA provider และตรวจ lifecycle/query/focus/Invoke ด้วยเครื่องมือ Windows ที่ผู้ใช้เลือก. หาก provider ไม่พร้อม ให้หยุดการอ้างว่าผ่าน accessibility.
4. ลง scoped command และ TextInput priority; ปรับ Package 0.2.8, ตัวอย่าง, เอกสาร, lockfile และ Setup หลัง source พร้อม.
5. แยกรายงานผล build/static check ออกจากผลทดสอบ GUI/IME/UIA ของผู้ใช้.

## เกณฑ์ตรวจรับที่ผู้ใช้ทดสอบเอง

| ด้าน | ผลที่ต้องเห็น |
| --- | --- |
| Focus scope | เปิด scope แล้ว Tab/Shift+Tab วนเฉพาะข้างใน; focus ไปพื้นหลังถูกปฏิเสธ; ปิด scope แล้วคืน focus เดิมเมื่อยังใช้ได้ |
| Focus/scroll | ScrollView ซ้อนกันเลื่อนจน Button/Canvas target ที่ focus ปรากฏ; update/remove/disable ไม่ทิ้ง focus บน ID เก่า; ring ไม่ถูกตัดผิดที่ DPI 100%/125% |
| Commands | Ctrl+S, Ctrl+Shift+S, F5 และ scope shortcut ส่ง ID ถูกต้องครั้งเดียว; conflict คืน error; unbind/close ไม่มี command ค้าง |
| TextInput/IME | คีย์แก้ไขข้อความ, undo, caret และการพิมพ์ไทยทำงาน แม้มี shortcut ของแอป; ไม่เกิด command ระหว่าง IME composition |
| UI Automation | Narrator/เครื่องมือ UIA อ่านชื่อ role enabled/focus/ตำแหน่งของ Button และ Canvas target; invoke ได้; update/remove ไม่แสดง node ค้าง; TextInput อ่าน label ได้ |
| Compatibility | แอปบน Package 0.2.7 ยัง check/run ได้หรือมี migration ที่ระบุจุดแก้ชัด; Setup รายงาน 0.1.30 และ package เป็น 0.2.8 |

ผู้ใช้เป็นผู้ทดสอบแอปและ screen reader เองตามความต้องการปัจจุบัน. `go build`, `craft check` และการประกอบ Setup เป็นหลักฐานคนละส่วนกับการทดสอบพฤติกรรมจริง.

## นอกขอบเขต Rev.30

ยังไม่เพิ่ม reactive `State<T>`, DataGrid, multiline/password TextInput, Menu/Dialog widget, Drag & Drop, global hotkey, custom control SDK หรือ accessibility pattern ครบทุกชนิด. งานเหล่านี้ค่อยต่อเมื่อ focus, command และ UIA foundation ผ่านเกณฑ์ข้างต้น.

