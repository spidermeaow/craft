# Craft Phase 1 Rev.18 — ScrollView และ viewport-aware rendering

วันที่: 6 ตุลาคม 2026  
ฐาน: Craft 0.1.17  
เป้าหมายรุ่น: Craft 0.1.18  
สถานะ: implement ใน source แล้ว; รอผู้ใช้ทดสอบและรับรอง

## เป้าหมาย

ทำให้เนื้อหาที่ใหญ่กว่าหน้าต่างเข้าถึงได้จริงผ่าน `gui.scrollView(id, child)` โดยรักษาความถูกต้องของตำแหน่งภาพ, click/hit test และความไวที่เป็นเป้าหมายหลักของ Craft UI Engine. ใช้ Windows native renderer เดิม; ไม่ใช้ WebView.

## ขอบเขต

1. เพิ่ม ScrollView node ที่มีลูกหนึ่ง node, viewport ใน logical pixels และ scroll offset ที่ถูก clamp ตามขนาดเนื้อหา. รักษา offset ตาม ID เมื่อ update/resize และ clamp ใหม่เมื่อขนาดเปลี่ยน.
2. รองรับ wheel แนวตั้ง, Shift+wheel และ horizontal wheel แนวนอน; รองรับ PageUp/PageDown/Home/End และ arrow keys บน scroll view ที่ active. วงล้อต้องสะสม delta ขนาดเล็กจากอุปกรณ์ละเอียด ไม่ทิ้ง input เงียบ ๆ.
3. clip ภาพและ hit test ภายใน viewport ทั้ง Direct2D และ GDI; update/hover/focus หลังเลื่อนต้องไม่วาดรั่วออกนอก parent. เลื่อนด้วยการย้าย geometry ที่วัดแล้ว ไม่ remeasure ข้อความทุกครั้ง.
4. ข้าม subtree ที่อยู่นอก dirty/viewport ระหว่าง paint. เพิ่ม trace สำหรับ scroll input-to-draw, จำนวน node ที่วาดใน frame ล่าสุด และจำนวน scroll event; ไม่มีการสร้าง node แบบ virtualized ในรุ่นนี้.
5. ปรับตัวอย่าง Stress Lab ให้ grid อยู่ใน ScrollView, อัปเดต API docs, version/bundle/installer และคำแนะนำทดสอบด้วยมือ.

## เกณฑ์ตรวจรับโดยผู้ใช้

- 8×8, 16×16 และ 32×32 สามารถเลื่อนดูเซลล์ที่ล้นได้; คลิกเซลล์หลังเลื่อนตรงกับภาพ.
- Wheel/Shift+wheel/horizontal wheel และ keyboard เลื่อนได้โดยไม่กระโดดผิดแกนหรือออกนอกขอบ.
- Resize/DPI change, update หนึ่งเซลล์และ update ทั้งตารางไม่ทิ้งภาพเก่า; offset ถูก clamp เมื่อเนื้อหาหด.
- Direct2D และ GDI fallback ให้ภาพและ hit test ตรงกัน; เปรียบเทียบ trace กับ 0.1.17 ใน workload เดียวกัน.
- Source package, CLI, source bundle และ installer ใช้งานได้โดยยังไม่อ้างว่าเป็น standalone app binary.

## นอกขอบเขต

Virtualized List/Grid, momentum animation, overlay scrollbar, touch gesture ผ่าน DirectManipulation, TextBox/IME และ Craft Office. งานเหล่านี้ควรต่อยอดหลัง ScrollView ผ่านการตรวจรับภาพจริง.
