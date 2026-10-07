# Craft UI Engine 0.1.19 — ScrollView sizing

Rev.19 แก้กฎ layout จากผลทดสอบ 0.1.18. เมื่อ Row/Column มี ScrollView เป็นลูกตรง ScrollView จะรับพื้นที่ว่างบนแกนหลักก่อน container พี่น้อง. Control panel และ toolbar จึงคงขนาดตามเนื้อหาแทนการขยายกินพื้นที่ที่ควรเป็น viewport.

เนื้อหาภายใน ScrollView คง preferred width/height ของตัวเอง แม้ viewport ใหญ่กว่า. Grid 8×8 จึงไม่ยืดแถวและปุ่มจนใหญ่ผิดธรรมชาติ; grid 32×32 ยังเลื่อนไปดูส่วนที่ล้นได้. ตำแหน่ง scroll ถูก clamp จากความต่างระหว่างขนาดเนื้อหาจริงกับ viewport.

API ไม่เปลี่ยนจาก 0.1.18:

```craft
let view: UiNode = gui.scrollView("items-scroll", gui.column("items", items))
```

แอปที่อาศัยการขยาย Row/Column พี่น้องของ ScrollView แบบเท่ากันอาจเห็นขนาดใหม่. หากต้องการพื้นที่คงที่ในแถบ controls ให้จัดเนื้อหาใน Row/Column/Panel ตามปกติ. รุ่นนี้ยังไม่มี sizing property, scrollbar ที่ลากได้หรือ virtualization.
