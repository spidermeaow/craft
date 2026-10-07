# Craft UI Runtime 0.1.26: responsive composition

Rev.26 เพิ่มการวัดข้อความและตัวช่วยจัดคอลัมน์สำหรับหน้าต่างหลายขนาด โดยยังใช้ native Windows UI Runtime และ Craft package ตามเดิม

## สิ่งที่ทำ

- `std.ui.measureTextWidth(window, value, size, weight): Int` วัดข้อความด้วย renderer/font/DPI เดียวกับ UI host และคืนความกว้างเป็น logical pixels. ขนาดตัวอักษร 8–96, weight 100–900 ทีละ 100, ข้อความ UTF-8 ไม่เกิน 4096 bytes และไม่มี NUL. เรียกหลังเปิดหน้าต่างและก่อนปิด.
- `gui.measureTextWidth(...)` เป็น wrapper สำหรับแอปที่ใช้ package. ใช้ตั้ง `minWidth` ของปุ่ม/เมนูจากความกว้างข้อความบวก padding แทนกำหนดกว้าง 94 px แล้วข้อความตัดบรรทัด.
- `gui.responsiveGrid(id, items, availableWidth, minCellWidth, maxColumns, gap)` จัด node เดิมเป็น Row หลายแถวตามพื้นที่ที่แอปส่งให้. จำนวนคอลัมน์อยู่ระหว่าง 1 ถึง `maxColumns`; cell กว้างเท่ากัน. แอปต้องคำนวณพื้นที่หลังหัก sidebar/padding และ rebuild เมื่อได้ `Resized` หรือ `DpiChanged`.
- `examples/gui-package-lab` ใช้ grid สี่การ์ด; คอลัมน์ลดตามหน้าต่าง. CLI/installer เป็น 0.1.26 และ GUI package เป็น 0.2.3.

```craft
let labelWidth: Int = gui.measureTextWidth(window, "Secondary", 16, 600)
let buttonWidth: Int = labelWidth + 24
let button: UiNode = gui.size(gui.button("secondary", "Secondary"), buttonWidth, 36, buttonWidth, 0, 0)

let width: Int = gui.viewportWidth(window) - sidebarWidth - 24
let cards: UiNode[] = [firstCard, secondCard, thirdCard, fourthCard]
let grid: UiNode = gui.responsiveGrid("cards", cards, width, 180, 4, 12)
```

## ขอบเขต

Grid เป็น composition ใน package; Runtime ยังคงจัด Row/Column และวัดข้อความ. มันไม่เปลี่ยน Canvas ที่วาดด้วยพิกัดตายตัวให้ reflow อัตโนมัติ. ถ้า `availableWidth` น้อยกว่า `minCellWidth` grid เหลือหนึ่งคอลัมน์และอาจต้องใช้ `scrollView` สำหรับเนื้อหาที่มีความกว้างขั้นต่ำ. การวัดข้อความเป็น single-line metric; line wrapping, font fallback และผลบน Direct2D/GDI ยังต้องตรวจบนหน้าจอจริง. ไม่มีคำรับประกันว่าหน้าจอทุกแบบพอดีโดยไม่กำหนด minimum size และ overflow policy.

ดู [วิธีตรวจรับ](TRY-REV26.md) และ [สถานะ Rev.26](../roadmap/phase1/output/phase1-rev26-status.md).
