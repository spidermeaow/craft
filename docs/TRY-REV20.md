# ตรวจรับ Craft UI Engine 0.1.20

จาก `D:\02-Repository\test-gui` เปิดด้วย `D:\02-Repository\craft\dist\craft.exe run`.

1. กด **Virtual 100,000**: แถว Item 0 เป็นต้นไปต้องปรากฏ; `Rows first..end` ควรเป็นช่วงเล็ก ๆ ตาม viewport ไม่ใช่ 100,000.
2. ชี้ในรายการแล้วใช้ wheel, PageDown และ End: index ของแถวต้องเพิ่มต่อเนื่อง ไปท้ายได้ถึง Item 99999. คลิกแถวหลังเลื่อน; `Clicked` ต้องตรง ID ที่เห็น.
3. กด **Update selected** และ **Update all** หลังเลื่อน: revision/ข้อความแถวที่เกี่ยวข้องเปลี่ยนโดยยังอยู่บริเวณเดิม; เลื่อนต่อแล้วไม่มีแถวซ้ำหรือเลขข้ามผิด.
4. จากท้ายรายการ กด **Virtual 3 rows** และ **Virtual 0 rows**: ตำแหน่งเลื่อนต้อง clamp, ไม่มีภาพแถวเก่าค้าง. กลับไป **Virtual 100,000** และตรวจว่าแถวต้นปรากฏ.
5. ลอง resize หน้าต่างและจอ DPI ต่างกันถ้ามี. ทดสอบฉาก Grid 8/16/32 และ ScrollView เดิมเพื่อดู regression. หากเครื่องใช้ GDI fallback ให้ตรวจภาพ/คลิกซ้ำ.

Trace เสริม: ตั้ง `CRAFT_UI_TRACE=1` ก่อน run แล้วปิดหน้าต่างเพื่อดู JSON ใน terminal. ตรวจ `virtual_rows_last_layout` ว่ายังคงอยู่ระดับแถวที่มองเห็นพร้อม overscan ไม่เพิ่มตาม 100,000. การเทียบความเร็วต้องใช้เครื่อง, ขนาดหน้าต่าง, backend และ workload เดียวกัน.
