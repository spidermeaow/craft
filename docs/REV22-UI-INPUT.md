# Craft UI Engine 0.1.22 — Text Input รุ่นแรก

`std.ui.textInput(id, initialValue, placeholder): UiNode` เพิ่มช่องข้อความบรรทัดเดียวใน UI tree. Windows host ใช้ native EDIT control ที่จัดตำแหน่งและ clip ตาม layout ของ Craft UI Engine เพื่อให้รับข้อความผ่าน keyboard layout และ IME ของ Windows พร้อม caret, selection, copy/cut/paste และ horizontal scrolling ของข้อความ. หน้าตาของ control ยังเป็นของ Windows; ยังไม่มี styling API สำหรับช่องนี้.

`initialValue` ใช้ตอนสร้าง node ID ครั้งแรก. เมื่อ `std.ui.update` ใช้ ID เดิม ค่าที่ผู้ใช้กำลังแก้จะคงอยู่ ไม่ถูกข้อความจาก tree ทับ. เปลี่ยนค่าจากโปรแกรมด้วย `std.ui.setText(window, id, value)`; คำสั่งนี้ไม่ส่ง `TextChanged` และจะปฏิเสธเมื่อ IME composition ยังทำงาน. เมื่อ node ถูกลบออกจาก tree editing state ของ ID นั้นถูกปล่อย; หากนำ ID เดิมกลับมาจะเริ่มจาก `initialValue` ใหม่.

`std.ui.nextEvent` เพิ่ม `event.text: String` ให้ `UiEvent` ทุกชนิด (ว่างเมื่อไม่เกี่ยวกับข้อความ). `TextChanged` ส่งค่าปัจจุบันหลังการแก้ไข; ระหว่าง IME composition จะรอจนจบ composition ก่อนส่ง. `TextSubmitted` ส่งค่าปัจจุบันเมื่อกด Enter ในช่อง. `FocusChanged` ส่ง ID ของ node ที่ได้รับ focus; ID ว่างเมื่อ focus ถูก clear. `KeyPressed` จาก input มี `nodeId` ของช่องนั้น. App ควรตอบสนองต่อ `TextChanged` และ `TextSubmitted` ตาม `nodeId` และ `event.text`.

ช่องรับข้อความจำกัด 1024 UTF-16 code units ต่อช่อง, UTF-8 ต้องถูกต้อง, ห้าม NUL และบรรทัดใหม่. ข้อความยาวเกินหรือการตั้งค่าที่ไม่ถูกต้องคืน `UiArgumentError`. รุ่นนี้รองรับหนึ่งหน้าต่างต่อ interpreter run ตามข้อจำกัดเดิม.

## ข้อจำกัดที่ยังมี

- เป็น native EDIT control ภายใน UI tree จึงยังวาดหน้าตาได้ไม่อิสระเท่า Text/Button ที่ engine วาดเอง. เป้าหมาย component ที่ customize ได้เต็มรูปแบบยังเป็นงานต่อไป.
- ยังไม่มี TextArea, rich text, password mode, undo policy ของ Craft, selection index API หรือ accessibility provider ของ Craft เอง. Native EDIT ให้พฤติกรรม Windows พื้นฐาน.
- การตัดข้อความและ composition ใช้กฎของ Windows EDIT/IME; ยังไม่ได้อ้างว่าทุกภาษาและทุก grapheme cluster ผ่านการตรวจรับ.
- ยังไม่มีการวัด input-to-present latency; trace เดิมจบที่ draw completion ของ parent renderer.

ทดสอบด้วย `D:\02-Repository\craft-test\runtime\11-ui-input` ตาม `D:\02-Repository\craft-test\guides\12-ui-input.md`. โปรเจกต์นี้เรียก `std.ui` โดยตรง ไม่มี GUI package และไม่มี HTTP.
