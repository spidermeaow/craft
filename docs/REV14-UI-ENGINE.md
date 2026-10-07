# Craft 0.1.14 — UI tree ขั้นต้น

Rev.14 เพิ่ม UI tree ขนาดเล็กบน Windows GUI bridge จาก Rev.13 ผ่าน `packages/gui` ซึ่งเป็น package ที่เลือกใช้ได้ ไม่เปลี่ยน Craft ให้เป็นภาษาเฉพาะ GUI

## API

| API | ผล |
| --- | --- |
| `gui.text(id, value): UiNode` | ข้อความหนึ่งบรรทัด |
| `gui.button(id, label): UiNode` | ปุ่มพื้นฐาน |
| `gui.row(id, children): UiNode` | เรียงลูกแนวนอน |
| `gui.column(id, children): UiNode` | เรียงลูกแนวตั้ง |
| `gui.panel(id, children): UiNode` | วางลูกซ้อนในพื้นที่ภายใน panel |
| `gui.mount(window, root)` | ติดตั้ง tree ครั้งแรก |
| `gui.update(window, root)` | แทน tree เดิมและวาดใหม่ |

`children` มีชนิด `UiNode[]` และแต่ละ node ต้องมี `id` ไม่ซ้ำใน tree เดียวกัน สูงสุด 4096 nodes และลึกไม่เกิน 64 ระดับ `ButtonPressed` ส่ง ID ใน `UiEvent.nodeId` กลับไปยัง Craft; application state ยังอยู่ใน Craft แล้วเรียก `update` เมื่อ state เปลี่ยน

Layout ตอนนี้แบ่งพื้นที่ลูกของ Row/Column เท่า ๆ กัน มี padding 16 logical pixels และ gap 8 logical pixels Panel วางลูกในพื้นที่ภายใน 16 logical pixels Resize จะคำนวณ layout ใหม่ Hit test ใช้พื้นที่ของ node หลัง layout

Renderer รุ่นนี้ใช้ Windows GDI วาดพื้นหลัง, rectangle และ text หนึ่งบรรทัด ไม่มี WebView และไม่มี GPU renderer ข้อความใช้ฟอนต์ระบบและยังไม่มีการวัดขนาดข้อความเพื่อกำหนด content size ปุ่มส่ง event จาก click และ Enter/Space ไปยังปุ่มที่ focus ขั้นต้น Tab/Shift+Tab เปลี่ยนปุ่มที่ focus และเมาส์มีสี hover ขั้นต้น ยังไม่มี pressed visual, configurable style, clipping ของลูก, DirectWrite, IME หรือ accessibility เต็มรูปแบบ

GUI package 0.1.13 ที่คัดลอกไว้ในโปรเจกต์เดิมไม่ถูกเขียนทับอัตโนมัติ ให้เก็บการแก้ไขของตัวเองก่อน แล้วนำ package 0.1.14 เข้าโปรเจกต์พร้อมแก้ dependency version ใน `craft.toml` และ lockfile โปรเจกต์ใหม่ใช้ `craft gui init` ได้ตามปกติ

`craft build` ยังสร้าง source bundle ที่ต้องรันด้วย Craft CLI 0.1.14 หรือใหม่กว่า Bundle รุ่นก่อนยังอยู่ในรายการรูปแบบที่ runtime อ่านได้
