# Craft 0.1.39 Rendering Foundation — manual handoff

สถานะ: source implementation รอบแรก; **ยังไม่มีผลรับรอง Direct2D/GDI/DPI หรือ benchmark จากผู้ใช้**

## Public Canvas Path API

`std.ui` และ `packages/gui` เปิดชื่อเดียวกัน:

| API | ความหมาย |
| --- | --- |
| `pathMoveTo(x, y)` | เริ่ม subpath ใหม่ |
| `pathLineTo(x, y)` | ต่อเส้นตรง |
| `pathArcTo(cx, cy, radius, startDegrees, sweepDegrees)` | ส่วนโค้งรอบจุดศูนย์กลาง; sweep บวก/ลบกำหนดทิศ |
| `pathQuadraticTo(cx, cy, x, y)` | Quadratic Bézier |
| `pathCubicTo(cx1, cy1, cx2, cy2, x, y)` | Cubic Bézier |
| `pathClose()` | ปิด subpath |
| `fillPath(path, color)` | เติมสีรูปทรง; subpath หลายวงใช้ even-odd fill |
| `strokePath(path, color, thickness)` | วาดเส้นขอบ |

`path` เป็น `UiDraw[]` ที่มีเฉพาะ path segment; `fillPath` หรือ `strokePath` แปลงเป็น `UiDraw` ที่ใส่ใน `canvas` ได้. พิกัดเป็น DIP ใน Canvas. สี path ต้องมี alpha 255; จำกัด 1024 commands, 128 subpaths, พิกัด -8192..8192 และ stroke width 1..256. Path ว่าง ลำดับผิด หรือ arc ที่รัศมี/มุมไม่ถูกต้องคืน `UiPathError` ก่อนวาด. Path segment เดี่ยวใส่ใน Canvas โดยตรงไม่ได้

ตัวอย่างวงแหวน:

```craft
var ring: UiDraw[] = [
    gui.pathArcTo(100, 100, 70, 0, 360), gui.pathClose(),
    gui.pathArcTo(100, 100, 40, 0, -360), gui.pathClose()
]
let shape: UiDraw = gui.fillPath(ring, gui.color(38, 92, 204, 255))
```

## Dirty regions

เมื่อ tree เดิมคง geometry, Runtime เปรียบเทียบ paint commands และรวมพื้นที่เก่า/ใหม่ที่เปลี่ยนไป ก่อนส่ง Win32 invalidation. หาก layout หรือจำนวนคำสั่ง Canvas เปลี่ยน, พื้นที่กว้าง, หรือระบุ bounds อย่างปลอดภัยไม่ได้ จะวาดทั้งพื้นที่ตาม fallback. VirtualGrid ที่ viewport/ขนาดคงเดิมสามารถ redraw ข้อความเซลล์และ selection เฉพาะบริเวณได้; การเปลี่ยนช่วงข้อมูล/ขนาดยัง relayout เต็ม. สิ่งนี้เป็นระบบระดับ Canvas/Component ไม่ใช้โครงสร้าง dirty-cell เฉพาะ Spreadsheet

เปิด `CRAFT_UI_TRACE=1` ก่อนรันแล้วปิดแอปตามปกติเพื่อรับ JSON `craft-ui-engine-timing` ที่ stderr. เพิ่ม `invalidation_regions`, `invalidated_pixels`, `full_redraws`, `renderer` และ `canvas_paint_ops_last_frame`. เวลา paint สิ้นสุดที่ draw call ไม่ยืนยันภาพขึ้นจอจริง และค่าที่ได้ยังไม่ใช่ benchmark ที่ผ่านเกณฑ์ 60 FPS

## ให้ผู้ใช้ตรวจ UI Showcase

### แก้ไขหลังภาพทดสอบ 9 ตุลาคม 2026

ภาพจากผู้ใช้แสดงเส้น Path แบบหยักและวงกลมทดสอบ clipping แสดงเต็มวง. ตรวจ source พบว่า Direct2D เรียก `ID2D1PathGeometry::Open` ที่ COM slot 16 ซึ่งเป็น `Widen`; เมื่อ Path วาดไม่สำเร็จ Runtime fallback ไป GDI ทั้งหน้า. แก้เป็น slot 17 ตาม Windows SDK แล้ว. Showcase กำหนดขนาด Canvas ตัวอย่างเป็น 640×405 DIP แบบตายตัว เพื่อให้วงกลมด้านขวาถูกตัดตามขอบจริง. ปรับตำแหน่งเส้นตรง/กรอบและความหนา stroke ตาม physical pixels ที่ DPI ปัจจุบันด้วย. ยังรอผู้ใช้ยืนยันภาพจริงและ renderer trace หลังติดตั้ง build ใหม่

1. Build Craft 0.1.39 แล้วจาก `D:\02-Repository\app` รัน `craft package lock`, `craft check`, `craft run`. หน้า **Rendering Lab** เปิดเป็นค่าเริ่มต้น
2. ดู Circle, Arc, Donut, Quadratic/Cubic และ closed fill/stroke. Donut ต้องมีช่องว่างจริงตรงกลาง ขอบรูปทรงไม่หลุด Canvas
3. กด **Move square**, **Reset position**, **Full redraw**. ตำแหน่งเก่าต้องไม่ทิ้งสีค้าง และ tiles ที่ซ้อนทับต้องยังถูกต้อง. เก็บ trace เพื่อดูว่าการย้ายปกติ invalidates เฉพาะพื้นที่ ส่วน Full redraw เพิ่ม counter
4. รันซ้ำโดยตั้ง `CRAFT_UI_RENDERER=gdi`; เปรียบเทียบรูปทรงและ clip กับ Direct2D ที่ DPI เดียวกัน แล้วลอง 100%/125%/150% ตามจอที่มี
5. ไปหน้า **Rev 37–38 Lab** และ **Excel** เลื่อนเร็ว เลือกช่วง แก้เซลล์ resize และตรวจภาพตกค้าง/ความเร็ว เพื่อจับ regression

บันทึก Windows version, renderer, DPI, ภาพก่อน/หลัง, trace และขั้นตอนทำซ้ำเมื่อพบปัญหา. ยังไม่ได้รัน automated tests หรือเปิด UI ในการเตรียม source รอบนี้ตามข้อตกลงที่ผู้ใช้เป็นผู้ทดสอบ

### ผลทดสอบโดยผู้ช่วยตามคำขอ 9 ตุลาคม 2026

- `craft check` ใน `D:\02-Repository\app` ผ่าน และ `go test ./internal/stdlib` ผ่านหลังปรับ fixture focus ที่ตกค้างให้ระบุ `tabStop` ตามสัญญาปัจจุบัน
- รอบแรก Direct2D พบ `invalidation_regions=0` แม้กด Move เพราะ `rebindLayoutDetailed` บังคับ relayout ทั้งหน้าหากมี adaptive text ที่อื่นใน tree; แก้ให้ rebind ได้เมื่อ adaptive content ไม่เปลี่ยน และ relayout เมื่อข้อความ wrap เปลี่ยนจริง พร้อมเพิ่ม regression test
- รอบ Direct2D หลังแก้: กด Move สองครั้งโดยไม่กด Full ได้ `invalidation_regions=4`, `invalidated_pixels=29770`, `layout.count=1`, `renderer=Direct2D`; ภาพตำแหน่งเก่า/ใหม่ไม่มีสีค้าง
- รอบ Direct2D กด Move หนึ่งครั้งแล้ว Full: `invalidation_regions=2`, `invalidated_pixels=14885`, `full_redraws=4`, `layout.count=2`; ตำแหน่งสี่เหลี่ยมไม่เปลี่ยนหลัง Full
- รอบ GDI กด Move หนึ่งครั้งแล้ว Full: `invalidation_regions=2`, `invalidated_pixels=13740`, `full_redraws=4`, `layout.count=2`, `renderer=GDI`, `renderer_fallback_reason=forced by CRAFT_UI_RENDERER=gdi`; topology, Donut hole, clip และตำแหน่งสี่เหลี่ยมตรงกับ Direct2D ในการตรวจด้วยภาพ. GDI มีขอบโค้งและตัวหนังสือหยาบกว่า

ตัวเลข trace เป็นค่ารวมต่อ session ซึ่งรวม startup, scroll และ interaction ที่ใช้ทดสอบ; เปรียบเทียบเฉพาะภาพและรูปแบบ counter ไม่ใช้เป็น benchmark ที่ควบคุมสภาวะครบถ้วน
