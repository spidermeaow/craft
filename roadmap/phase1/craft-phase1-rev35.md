# Craft Phase 1 Rev.35 (0.1.35) — native TextInput rendering in ScrollView

วันที่: 8 ตุลาคม 2026  
สถานะ: **ผู้ใช้ยืนยันว่าอาการภาพซ้อนหายและความเร็วดีขึ้นหลังปรับ Showcase; การแก้ค่าเซลล์ย้ายไป Rev.36**  
หลักฐาน: ภาพและผลทดสอบ UI Showcase หน้า Excel ที่ผู้ใช้รายงานหลังเปลี่ยนเซลล์ A1–J18 เป็น TextInput 180 ช่อง

## อาการที่ต้องแก้

เมื่อเลื่อนหน้า Excel ลง ข้อความในหลายเซลล์วาดทับกัน ถูกตัดเป็นส่วน ๆ หรือเหมือนมีข้อความจากแถวอื่นค้างอยู่ในช่อง. ภาพแสดงว่าเส้นกรอบและหัวแถวส่วนใหญ่ยังอยู่ตำแหน่งเดิม แต่เนื้อหาของ native TextInput ไม่สัมพันธ์กับตำแหน่งเซลล์. นี่เป็นปัญหาการแสดงผลที่เห็นชัดในกรณี TextInput จำนวนมากภายใน ScrollView; **ยังไม่ยืนยันว่าค่าข้อมูลเสียหรือสาเหตุอยู่ที่ Runtime เพียงฝ่ายเดียว**.

ขณะพบอาการ หน้า Excel ประกอบตารางจาก Row/Column และ native TextInput หนึ่งตัวต่อเซลล์. รูปแบบนี้เป็น stress case สำหรับการวาง native child windows จำนวนมากและไม่ใช่สถาปัตยกรรม cell editor ระยะยาวของ VirtualGrid. อย่างไรก็ตาม control ที่รองรับภายใน ScrollView ต้องไม่ทิ้งภาพค้างหรือแสดงข้อความผิดแถวเมื่อ scroll.

## แยกสาเหตุ

1. สร้างกรณีย่อจาก UI Showcase: TextInput 1, 10 และ 180 ช่องใน ScrollView; เปรียบเทียบกับ TextInput นอก ScrollView และกับตารางที่วาดด้วย Canvas. ตรวจว่าปัญหาเริ่มเมื่อจำนวน control มาก, เมื่อมี style/inset, หรือเมื่อ scroll เท่านั้น.
2. เก็บค่าก่อนและหลังเลื่อนของ layout box ใน logical pixels, พิกัดจริงที่ส่งเข้า `SetWindowPos`, viewport clip และ region ที่ส่งเข้า `SetWindowRgn` สำหรับ input ที่เห็น/เพิ่งพ้นจอ. ตรวจ `positionInputs()` หลัง `scrollBy`, relayout และการเปลี่ยน focus.
3. ตรวจการ hide/show, invalidation ของตำแหน่งเดิมและใหม่, child window z-order, `WS_CLIPCHILDREN`, native EDIT repaint และการใช้ clipping region หลัง style inset. เปรียบเทียบ wheel, scrollbar drag, Page Down, resize, DPI 100%/125%, Direct2D และ GDI fallback.
4. ถ้ากรณีย่อของ native TextInput ยังผิด ให้แก้ที่ `internal/stdlib` แล้วใช้ Showcase เป็น regression scenario. ถ้าผิดเฉพาะองค์ประกอบ Excel ให้แก้ geometry/composition ของแอปด้วย โดยบันทึกหลักฐานแยกกัน. `packages/gui` ไม่ควรซ่อนข้อผิดพลาดของ Runtime ด้วยการวาดทับ.

ไฟล์เริ่มตรวจ: `internal/stdlib/ui_text_windows.go` (`visibleInputBox`, `positionInputs`, `SetWindowPos`, `SetWindowRgn`), `internal/stdlib/ui_scene.go` (`scrollBy`, `shiftBox`), `internal/stdlib/ui_windows.go` (`applyScroll`, repaint) และ `D:\02-Repository\app\src\main.craft` (`editableExcelSheet`). จาก source เห็นว่า Runtime ย้าย child EDIT และตั้ง clipping region ทุกครั้งที่ scroll; จุดนี้เป็น **สมมติฐานตรวจสอบ** ไม่ใช่สาเหตุที่พิสูจน์แล้ว.

## เกณฑ์ตรวจรับ

- เลื่อนหน้า Excel ช้าและเร็ว รวมทั้งเลื่อนกลับขึ้น: ข้อความแต่ละช่องอยู่ในเซลล์ของตน ไม่มีตัวอักษรซ้อน ค้าง ตัดผิดตำแหน่ง หรือข้อความจากแถวอื่น.
- คลิกช่องที่เห็นหลังเลื่อนแล้ว caret และพื้นที่รับคลิกตรงกับกรอบ; ช่องที่อยู่นอก viewport ไม่รับคลิกหรือทับเนื้อหาอื่น.
- ค่า A2/A10 ที่แก้แล้วไม่สูญหายเมื่อเลื่อน สลับหน้า หรือ resize; การทดสอบนี้แยก data integrity จาก paint defect.
- ตรวจบน DPI 100% และ 125% รวมทั้ง Direct2D และ GDI fallback. บันทึก renderer, ขนาดหน้าต่าง, วิธี scroll และภาพก่อน/หลังเพื่อเทียบผล.

ผู้ใช้เป็นผู้ทดสอบ UI จริงตามแนวทางโครงการ. การแก้ source หรือ build Setup เพียงอย่างเดียวไม่ถือว่าผ่านเกณฑ์. งาน cell editor เดียวที่วางทับ VirtualGrid และ selection หลายช่องเป็นประเด็นสถาปัตยกรรมต่างหาก; ไม่ใช้เพื่อปิดบังบั๊กการ scroll ของ native TextInput ในรอบนี้.

## Implementation 0.1.35

- สร้าง native EDIT แบบซ่อนไว้ก่อน แล้วให้ `positionInputs()` กำหนดตำแหน่งและ clipping region ก่อนแสดงครั้งแรก เพื่อไม่ให้ control ใหม่ปรากฏชั่วคราวที่พิกัดเริ่มต้น.
- เมื่อ control เลื่อนตำแหน่ง ใช้ `SWP_NOCOPYBITS` กับ `SetWindowPos` และ invalidate เนื้อหาของ EDIT เพื่อบังคับวาดข้อความจากข้อมูลปัจจุบัน ไม่คัดลอกภาพเก่าจากตำแหน่งก่อน scroll.
- จำกัด clipping region ให้อยู่ภายในขนาดจริงของ child control และเก็บ geometry/clip ล่าสุด เพื่อลดการย้าย, show/hide และสร้าง region ซ้ำเมื่อค่าไม่เปลี่ยน.
- UI Showcase หน้า Excel ใช้ 180 TextInputs เป็นกรณีตรวจรับแรก และมี [ขั้นตอนทดสอบ 0.1.35](../../../app/TEST-REV35.md) สำหรับ scroll, focus และข้อมูลเซลล์. หลังผู้ใช้รายงานเรื่องความเร็ว ได้ปรับองค์ประกอบแอปตามหัวข้อถัดไป.

การวิเคราะห์ source สนับสนุนให้แก้เส้นทาง reposition/repaint นี้. ผู้ใช้รายงานภายหลังว่าอาการภาพค้าง/ซ้อนในภาพเดิมหายแล้ว แต่ยังไม่มีผลทดสอบครบทุก DPI และ renderer.

## ผลทดสอบและการลดภาระของ Showcase

ผู้ใช้รายงานว่าอาการข้อความซ้อน/ค้างในภาพเดิมหายแล้ว แต่หน้า Excel ทำงานช้ามาก. สาเหตุที่น่าจะมีส่วนสำคัญคือ Showcase สร้าง native TextInput 180 ตัวและ Runtime ต้องย้าย/วาด control ที่เห็นทุกครั้งที่ scroll. จึงปรับ Showcase ให้แสดงเซลล์ที่ไม่ถูกเลือกด้วย Canvas และสร้าง native TextInput เฉพาะเซลล์ที่กำลังแก้ (อีกหนึ่งตัวคือ value bar). แอปยังเก็บค่าทั้ง 180 เซลล์แยกจากภาพและให้คลิกเซลล์ใดก็ได้เพื่อพิมพ์ในตำแหน่งนั้น. การเปลี่ยนนี้ลดงาน native child windows ในกรณี Excel; **ยังไม่ได้พิสูจน์ว่า Runtime รองรับการเลื่อน TextInput 180 ตัวได้เร็วพอ**. ผู้ใช้จะตรวจความเร็วและความถูกต้องอีกครั้งตาม [ขั้นตอนทดสอบ 0.1.35](../../../app/TEST-REV35.md).
