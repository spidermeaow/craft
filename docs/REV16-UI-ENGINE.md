# Craft 0.1.16 — responsive UI host และ partial repaint

Rev.16 เปลี่ยน Windows UI host จากการ polling ทุก 8 ms เป็นการรอ Win32 input หรือ wake event จากคำสั่ง Craft โดยตรง คำสั่ง `mount`, `update`, `setTitle` และคำสั่งอื่นยังใช้ API เดิมผ่าน `packages/gui` แอป GUI ยังคงเขียนด้วย Craft และใช้ Direct2D/DirectWrite เมื่อมี hardware target หรือ GDI fallback

เมื่อ `gui.update` ส่ง tree ที่ ID และชนิด node อยู่ในโครงสร้างเดิม Engine ใช้ layout geometry เดิมและ invalidate เฉพาะ node ที่ข้อความเปลี่ยน หากเปลี่ยนมากกว่า 16 node หรือพื้นที่รวมมากกว่าครึ่งหน้าต่างจะวาดทั้งหน้าต่างแทน หากโครงสร้างเปลี่ยนจะจัด layout และวาดใหม่ทั้งหน้าต่าง Hover/focus เปลี่ยนเฉพาะพื้นที่ปุ่ม Direct2D ใช้ clip ตาม dirty rect และข้าม subtree ที่อยู่นอกพื้นที่; GDI ยังใช้ Win32 invalid region เช่นกัน การ resize หรือ DPI change ยังคง full layout/paint

เปิด diagnostics โดยตั้ง `CRAFT_UI_TRACE=1` ก่อน `craft run` เมื่อ session ปิดจะมี JSON หนึ่งบรรทัดบน stderr ประกอบด้วยจำนวน frame และค่า p50/p95/p99 เป็น microseconds ของ command queue, layout, paint ถึงจบคำสั่งวาด และ input ถึงจบคำสั่งวาดสำหรับ event ล่าสุดที่ตามด้วย `gui.update` ฝั่ง Direct2D จุดสิ้นสุดคือ `EndDraw`; ฝั่ง GDI คือจบ draw calls ค่าเหล่านี้เป็นเวลา **ภายใน Engine** ไม่ใช่เวลาที่พิกเซลปรากฏบนจอจริง และยังไม่ได้วัดกับแอป React ใน workload เดียวกัน

ยังไม่มี text measurement, animation, ScrollView หรือ virtualization; layout แบบ equal-share เดิมยังอยู่ และ source bundle ยังต้องใช้ Craft CLI เพื่อรัน ผล latency และภาพบนเครื่องจริงต้องบันทึกใน [สถานะ Rev.16](../roadmap/phase1/output/phase1-rev16-status.md)
