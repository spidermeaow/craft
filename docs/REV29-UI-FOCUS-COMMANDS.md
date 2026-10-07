# Craft UI 0.1.29 — Focus และ keyboard commands

Runtime เพิ่ม `std.ui.focusPolicy(node, tabStop, enabled)`, `std.ui.focus(window, id)`, `std.ui.bindShortcut(window, commandId, key, modifiers)` และ `std.ui.unbindShortcut(window, commandId)`. Package 0.2.7 ห่อ API เหล่านี้เป็น `gui.*` พร้อม `gui.commandInvoked(event, id)`.

Tab/Shift+Tab เดินตามลำดับ UiNode. `tabStop=false` ข้ามใน traversal แต่ยังขอ focus ด้วยโปรแกรมได้. `enabled=false` ทำให้ Button/TextInput ไม่รับ focus; Button ไม่ถูกกดด้วย pointer/keyboard. การขอ focus ไปยัง ID ที่ไม่มีอยู่หรือ disabled คืน `UiFocusError`. การเปลี่ยน focus จะพยายามเลื่อน ScrollView ให้ control ปรากฏ. Focus ring แสดงบน Button และ Canvas target ใน Direct2D/GDI. TextInput ใช้ focus ของ native EDIT.

Shortcut ผูกกับหน้าต่างเดียว. `key` เป็น Windows virtual key code; `modifiers` ใช้ Shift=1, Ctrl=2, Alt=4 (บวกค่าเมื่อใช้ร่วมกัน). ตัวอย่าง Ctrl+S คือ `(83, 2)` และ F5 คือ `(116, 0)`. เมื่อกด Runtime ส่ง `CommandInvoked` โดย `event.nodeId` คือ command ID, `event.key` และ `event.modifiers` คือคีย์ที่กด. `KeyPressed` ยังคงส่งก่อน command; key repeat ไม่สร้าง `CommandInvoked` ซ้ำ. Shortcut ที่ชนกันคนละ ID คืน `UiShortcutError`. TextInput เก็บ Ctrl+A/C/V/X/Z/Y และ IME composition ไว้กับ EDIT ก่อน. ไม่มี global hotkey.

`std.ui.accessible(node, name, role)` เก็บ semantic metadata ขั้นต้นใน tree; role คือ `text`, `button`, `textbox`, `group`. Runtime 0.1.29 **ยังไม่มี UI Automation provider สำหรับ custom Button/Canvas** และยังไม่มี focus scope ของ Dialog/Popup. Metadata นี้จึงยังไม่ปรากฏกับ screen reader. Native EDIT ยังคงใช้ provider ของ Windows. งานสองส่วนนี้ค้างจากแผน Rev.29 และไม่ควรถือว่าผ่านเกณฑ์ accessibility.

Build source ทำได้ แต่ยังไม่ได้รัน automated tests หรือทดสอบ GUI ตามความต้องการของผู้ใช้.
