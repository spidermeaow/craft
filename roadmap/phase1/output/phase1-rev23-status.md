# Craft Phase 1 Rev.23 — สถานะ implementation

วันที่: 7 ตุลาคม 2026  
สถานะ: **build/check/automated regression ผ่าน; เปิด Layout Lab ตรวจภาพบน Direct2D ที่ DPI 125% แล้ว; ยังต้องรับผลทดสอบ interaction เพิ่ม**

## สิ่งที่เพิ่ม

- `std.ui.size`, `std.ui.maxSize`, `std.ui.align`, `std.ui.spacing`, `std.ui.splitView` พร้อมตรวจช่วง argument และการ clone node.
- Layout แบบ fixed/preferred, minimum, weight และระยะห่างที่กำหนดได้; SplitView สอง pane ที่คุมสัดส่วนและขั้นต่ำ.
- แสดง scrollbar แนวตั้งที่ลากได้ และส่ง wheel จาก scroll pane ชั้นในไปยัง pane ชั้นนอกเมื่อชั้นในเลื่อนต่อไม่ได้.
- ตรวจ layout spec ในการตัดสินใจ reuse layout เพื่อไม่ให้ API ใหม่ใช้ geometry เก่า.
- Layout Lab, คู่มือทดลอง, version ของ CLI/installer เป็น 0.1.23.

## ข้อจำกัดและงานค้าง

- ยังไม่มี splitter ที่ลากได้, scrollbar แนวนอนแบบเห็น/ลาก, theme/style, icon/image, tab/table จริง, accessibility provider และ multi-window.
- เมื่อหน้าต่างเล็กกว่าขั้นต่ำรวม ระบบย่อ pane; content ควรอยู่ใน ScrollView เพื่อเข้าถึงได้.
- ยังไม่ได้ตรวจ IME/ภาษาไทยด้วยมือ, scrollbar drag แบบ end-to-end, GDI fallback, DPI อื่น หรือ input-to-present latency. อย่าอ้างว่าเกณฑ์เหล่านี้ผ่านก่อนมีหลักฐานจริง.
- งาน HTTP และ Craft Office ไม่อยู่ในรอบนี้.

วิธีทดสอบด้วยตา: `docs/TRY-REV23.md`.

## ผลตรวจ 7 ตุลาคม 2026

- `go build -o dist\craft.exe .\cmd\craft` ผ่าน; `dist\craft.exe version` รายงาน `Craft 0.1.23 / Phase 1 Rev.23` หลังแก้ metadata ที่ยังพิมพ์ Rev.22.
- `examples/gui-layout-lab`: `craft check` ผ่าน และ `craft run` เปิดหน้าต่างด้วย `direct2d-hardware`.
- ตรวจภาพจริงบน Windows DPI 125% หลังย่อหน้าต่าง: top bar, sidebar, master/detail, scrollbar แนวตั้งทั้งสอง pane และ status bar แสดงอยู่; ข้อความใน TextInput ยังปรากฏหลัง resize. ภาพนี้ไม่ยืนยันความถูกต้องของ IME หรือการลาก scrollbar.
- `go test ./...` ครั้งแรกไม่ผ่านเพราะ `TestCLIEndToEnd` ยังยึดข้อความ version `0.1.19`; แก้เป็น `0.1.23` และเพิ่ม regression test ของ split/weighted layout กับ layout-spec invalidation แล้วรันซ้ำ **ผ่านทั้งหมด**.
- หน้าต่าง Layout Lab ที่เปิดครั้งแรกกินพื้นที่จอมากจนภาพจับหน้าจอเห็นส่วนล่างไม่ครบ; การย่อหน้าต่างทำให้เห็น status bar. ยังไม่สรุปว่าเป็น bug ของ initial window sizing เพราะอาจมีการปรับหน้าต่างระหว่างทดสอบ.
