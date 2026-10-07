# Craft UI Runtime 0.1.24 — Canvas รุ่นแรก

Canvas เป็น `UiNode` ที่เก็บรายการคำสั่งวาดตามลำดับไว้ใน UI tree. คำสั่งใช้พิกัด **logical pixels ภายใน Canvas** และถูก clip ตาม Canvas และ parent viewport. API ใช้ syntax Craft ปัจจุบัน:

| API | ผลลัพธ์ |
| --- | --- |
| `std.ui.color(r, g, b, a)` | `Int` สี ARGB โดยแต่ละช่องอยู่ในช่วง 0–255 |
| `std.ui.fillRect(x, y, width, height, color, radius)` | `UiDraw` สำหรับพื้นที่ทึบ/โปร่งใสและมุมมน |
| `std.ui.strokeRect(x, y, width, height, color, thickness, radius)` | `UiDraw` สำหรับขอบ |
| `std.ui.drawLine(x1, y1, x2, y2, color, thickness)` | `UiDraw` สำหรับเส้น |
| `std.ui.drawText(text, x, y, width, height, color)` | `UiDraw` สำหรับข้อความบรรทัดเดียว |
| `std.ui.drawImage(path, x, y, width, height)` | `UiDraw` สำหรับ PNG/JPEG |
| `std.ui.canvas(id, width, height, commands)` | `UiNode`; `commands` เป็น `UiDraw[]` |

ตัวอย่าง:

```craft
var commands: UiDraw[] = []
commands.append(std.ui.fillRect(0, 0, 300, 160, std.ui.color(50, 80, 220, 255), 12))
commands.append(std.ui.drawText("สวัสดี Craft", 20, 30, 260, 40, std.ui.color(255, 255, 255, 255)))
let art: UiNode = std.ui.canvas("art", 300, 160, commands)
```

เมื่อคลิก Canvas จะได้รับ `CanvasPressed` พร้อม `nodeId` และ `x/y` ภายใน Canvas; `PointerClicked` เดิมยังส่งพิกัดของหน้าต่างก่อนหน้า. `std.ui.update` ที่เปลี่ยนคำสั่งแต่ขนาดเท่าเดิมจะ invalidate พื้นที่ Canvas โดยไม่ remeasure ทั้ง tree. GPU bitmap cache ถูกปล่อยเมื่อภาพไม่อยู่ใน tree; device loss สร้าง bitmap จาก pixel data เดิมได้.

ภาพอ้าง path เทียบกับ **working directory ตอนรัน**. ไฟล์ต้องเป็น PNG/JPEG ขนาดไม่เกิน 16 MiB และมิติไม่เกิน 2048×2048; cache ภาพต่อ session จำกัด 64 MiB. Canvas จำกัด 2048 commands และ UI tree จำกัดรวม 8192 commands; คำสั่งหนึ่งมีพื้นที่ไม่เกิน 4,194,304 logical pixels. ขอบเขต geometry อยู่ในช่วงที่ API ตรวจและคืน `UiDrawError`/`UiAssetError` หากข้อมูลไม่ถูกต้อง. Source bundle ปัจจุบันยังไม่ฝังไฟล์ภาพไว้ภายใน: ต้องส่ง asset ไว้ข้างแอป/ตั้ง working directory ให้ตรง หรือใช้ระบบ asset ของ bundle ในรุ่นถัดไป.

Direct2D/DirectWrite เป็นเส้นทางหลัก. ตั้ง `CRAFT_UI_RENDERER=gdi` ก่อนรันเพื่อบังคับ GDI fallback สำหรับตรวจภาพ. GDI ใช้ `AlphaBlend` กับ bitmap และรูปทรงโปร่งใส; ข้อความโปร่งใสใน GDI ยังวาดด้วยสีทึบ และความมน/ขอบเส้นใน fallback อาจต่างจาก Direct2D. ยังไม่มี path, gradient, shadow, arbitrary transform หรือ per-shape hit testing.
