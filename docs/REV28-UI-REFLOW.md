# Craft UI 0.1.28 — Text reflow และ Grid ตามพื้นที่จริง

Runtime 0.1.28 เพิ่ม `std.ui.textWrap(node)` สำหรับ `Text` และ `std.ui.grid(id, items, minCellWidth, maxColumns, gap)`. GUI Package 0.2.6 ห่อ API เป็น `gui.textWrap`, `gui.grid` และ `gui.autoGrid`; `gui.responsiveGrid` เดิมยังใช้ได้.

## Text

```craft
let description: UiNode = gui.textWrap(gui.text("description", "Long English or Thai text ..."))
```

Wrapping เป็น opt-in เฉพาะ `Text`. Runtime แบ่งบรรทัดตามความกว้างที่ parent จัดให้ แล้วปรับความสูงก่อนวาง node ถัดไป. Newline เป็นการขึ้นบรรทัดใหม่. เมื่อไม่มีช่องว่างให้ตัด จะตัดตามตัวอักษร; การแบ่งคำภาษาไทยยังไม่ใช่ Unicode line-breaking ที่สมบูรณ์. `Button`, `TextInput` และ Canvas text ยังคงพฤติกรรมเดิม. เมื่อหน้าต่างแคบจน glyph เดียวกว้างเกินพื้นที่ ต้องใช้ overflow/scroll.

## Grid

```craft
let cards: UiNode = gui.autoGrid("cards", items, 180, 4, 12)
```

Grid คำนวณจำนวนคอลัมน์จากความกว้างที่ parent จัดให้จริง: อย่างน้อยหนึ่งคอลัมน์ ไม่เกิน `maxColumns`, ช่องว่าง `gap` logical pixels. แต่ละแถวสูงเท่ากับ child ที่สูงที่สุดในแถวนั้น. หาก viewport แคบกว่า `minCellWidth` ให้ห่อ Grid ด้วย `ScrollView` หรือเลือก min width ที่เล็กลง; Runtime ไม่บีบข้อความหรือปุ่มให้ต่ำกว่าขั้นต่ำโดยอัตโนมัติ.

## Shell

`gui.adaptiveShell(id, header, sidebar, mobileNav, content, viewportWidth, breakpoint, sidebarWidth)` ใช้ Sidebar เมื่อ viewport กว้างถึง breakpoint และใช้ mobileNav เมื่อแคบกว่า. แอปยังเป็นเจ้าของ state, breakpoint และ event loop. เมื่อ `Resized` หรือ `DpiChanged` ให้เรียก `gui.update` ด้วย tree ใหม่.

## ตรวจด้วยตนเอง

ใช้ `D:\02-Repository\responsive-tester` และ `D:\02-Repository\gui-package-dashboard`: รัน `craft package check`, `craft check`, `craft run`; ย่อ/ขยายผ่าน breakpoint, อ่านข้อความที่ wrap, ดูจำนวนคอลัมน์และการเลื่อน, กดปุ่มและลอง DPI ต่างกัน. ผู้ใช้เป็นผู้ทดสอบ GUI ตามความต้องการปัจจุบัน. การ build source/Setup ไม่ใช่ผลยืนยันพฤติกรรมบนหน้าจอ.
