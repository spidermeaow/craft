# Craft Phase 1 Rev.17 — Content-aware layout และ repaint correctness

วันที่: 6 ตุลาคม 2026  
ฐาน: Craft 0.1.16  
เป้าหมายรุ่น: Craft 0.1.17  
สถานะ: **implemented in source; รอผู้ใช้ทดสอบภาพจริงและรับรอง**

## เป้าหมาย

ทำให้ Craft UI Engine จัดพื้นที่ให้ Text, Button, Row, Column และ Panel ตามเนื้อหาจริง โดยรักษาการตอบสนองที่ผู้ใช้ยอมรับใน 0.1.16 ตามแนวคิด [Performance first. Simplicity always.](../craft-gui-direction.md) หน้าจอที่มี node จำนวนมากต้องไม่ทำให้ลูกหายไปเงียบ ๆ เพราะ padding และการหารพื้นที่เท่ากัน และการเปลี่ยนข้อความหนึ่งจุดต้องปรากฏครบโดยไม่ทิ้งภาพเก่า

0.1.17 ยังเป็น Windows native GUI ผ่าน Craft package และ Direct2D/DirectWrite โดยมี GDI fallback เป้าหมายของรุ่นนี้คือ **ความถูกต้องของการวัด จัดตำแหน่ง วาด และ hit test**; ไม่ใช่การเพิ่ม component จำนวนมาก

## หลักฐานจาก 0.1.16

- ผู้ใช้ทดลองแล้วให้ 0.1.16 ผ่านด้านการใช้งานและความรู้สึกตอบสนองเร็ว
- ตัวอย่าง stress ที่มีตาราง 16×16 และ 32×32 สร้าง tree และรับ event ได้ แต่พื้นที่ลูกบางชั้นถูกคำนวณเหลือศูนย์ จึงไม่เห็นตาราง ส่วน 4×4 มองเห็นได้
- พบกรณีเปลี่ยนข้อความเซลล์เดียวแล้ว status เปลี่ยน แต่ glyph ในเซลล์ค้าง เพราะตัวอักษรวาดล้น layout box ขนาดเล็กและ dirty rect ไม่ครอบคลุม paint bounds ทั้งหมด มีการแก้ขยาย invalidation/culling ใน source หลังการทดลอง แต่ยังต้อง build และตรวจซ้ำก่อนถือว่าแก้แล้ว
- Trace `input_to_draw_complete` วัดเวลาถึงจบคำสั่งวาด ไม่ใช่เวลาที่พิกเซลปรากฏบนจอ; ผลการทดลอง stress บางรอบอยู่ราว 14–15 ms แต่ยังใช้รับรองความลื่นของตารางใหญ่ไม่ได้ เพราะตารางไม่ปรากฏ

## ขอบเขต implementation

### 1. Measurement pass ที่ถูกต้อง

- แยก **measure → layout → paint** ให้ชัด โดย leaf แจ้งขนาดที่ต้องการก่อน parent จัดพื้นที่; เก็บขนาดใน logical pixels และคำนวณใหม่เมื่อข้อความ ความกว้าง font หรือ DPI เปลี่ยน
- ใช้ DirectWrite วัดข้อความ Unicode/ภาษาไทยและปุ่มใน Windows path พร้อม cache ที่มี key ครบตามข้อความ, font/ขนาด, width constraint และ DPI; GDI fallback ต้องได้ geometry ที่สอดคล้องพอให้ paint/hit test ตรงกัน
- กำหนดขนาดขั้นต่ำของ Button รวมข้อความและ padding; Text ไม่ควรได้ box สูงศูนย์เมื่อมีพื้นที่ให้วาด และ glyph ต้องไม่วาดนอก paint bounds ที่ engine รู้จัก
- กันขนาดติดลบ, overflow และการคำนวณซ้ำเกินจำเป็น; บันทึกจำนวน measure/layout pass ใน diagnostics เมื่อเปิด trace

### 2. กฎจัดพื้นที่ Row, Column และ Panel

- ให้ `row(id, children)` และ `column(id, children)` จัดลูกโดยใช้ขนาดตามเนื้อหาเป็นฐานก่อน แล้วแจกพื้นที่เหลือแก่ลูกที่ขยายได้; Panel มี padding ที่คำนึงถึง available size ไม่หักจนลูกเหลือศูนย์โดยไม่รายงาน
- นิยามกฎ `content`, `fill`, minimum และ maximum size ที่เรียบง่ายและสม่ำเสมอทั้งสองแกน ค่าเริ่มต้นของ Text/Button ต้องใช้งานได้โดยไม่ต้องกำหนดทุก property เอง
- เมื่อผลรวม minimum size มากกว่าพื้นที่หน้าต่าง ให้ใช้ overflow policy ที่ระบุได้ เช่น clip พร้อมรายงาน diagnostics ไม่บีบต่ำกว่า minimum จนหาย; ไม่มี ScrollView ในรุ่นนี้ จึงไม่รับประกันว่าเห็นทุกเซลล์พร้อมกัน
- รักษา ID, ลำดับลูกและพื้นที่ hit test ให้ตรงกับภาพที่วาด โดยเฉพาะหลัง resize และย้ายหน้าต่างข้าม DPI
- รักษาชื่อและชนิดของ API เดิม (`row`, `column`, `panel`, `text`, `button`, `mount`, `update`) แต่ยอมรับว่าขนาด/ตำแหน่งภาพเปลี่ยนจาก equal-share; จัดทำ migration note สำหรับแอปที่อาศัยรูปทรงเดิม

### 3. Repaint และ invalidation

- ตรวจและปิด bug ข้อความเซลล์ค้างด้วย paint bounds ที่ได้จาก measurement ไม่ใช้ margin คงที่ 24 px เป็นคำตอบถาวร; invalidation ต้องครอบคลุม glyph overhang, focus/hover effect และพื้นหลังที่ถูกเปิดเผย
- เมื่อข้อความเปลี่ยน ให้แยกกรณี **ขนาดเท่าเดิม** (paint เฉพาะบริเวณที่กระทบ) กับ **ขนาดเปลี่ยน** (จัด layout เฉพาะบรรพบุรุษ/พี่น้องที่ได้รับผล แล้ว invalidate พื้นที่เก่าและใหม่)
- ตรวจลำดับการวาดและ clip ของ Direct2D รวมถึง GDI fallback เมื่อ dirty regions ทับกัน, มี parent พื้นหลัง, เปลี่ยนโครงสร้าง tree, resize, DPI change หรือ device loss
- ถ้า partial repaint ซับซ้อนหรือแพงกว่าการวาดใหม่ในกรณีใด ให้ fallback เป็น full repaint โดยยังได้ภาพถูกต้องก่อนเพิ่ม optimization

### 4. ตัวอย่างและ developer experience

- ปรับ [Craft UI Stress Lab](../../examples/gui-stress/src/main.craft) และตัวอย่างใน `D:\02-Repository\test-gui` ให้แสดงโหมดที่ดูภาพจริงได้ และมีโหมดจำนวน node สูงเพื่อวัดภาระ; บอกผู้ใช้เมื่อเนื้อหาล้นพื้นที่
- ผู้เขียนแอปยังสร้าง tree และจัดการ event ด้วย Craft ไม่ต้องจัดการ font metrics, dirty rect หรือ Direct2D resource เอง
- เพิ่มเอกสาร API sizing/overflow พร้อมตัวอย่างที่ check/run ได้จริง; ไม่ใช้ syntax จำลองในคู่มือใช้งาน

## ลำดับทำงาน

1. บันทึกภาพและ trace ของ 0.1.16 บนหน้าต่าง/เครื่อง/DPI ที่กำหนด โดยแยก 4×4, 8×8, 16×16 และ 32×32
2. สร้าง measurement contract และ text metrics/cache; ตรวจข้อความไทย, อังกฤษ, Unicode, ข้อความยาวและ DPI
3. เปลี่ยน layout algorithm เป็น content-aware พร้อม minimum/overflow policy และ hit test ที่ตรงตำแหน่ง
4. แก้ paint bounds และ partial repaint; ตรวจ one-cell update, update-all, hover/focus และ resize
5. เปรียบเทียบ latency/frame time กับ 0.1.16 บน workload เดียวกัน; ปรับก่อนรับงานหาก responsiveness ถดถอย
6. ตรวจ compatibility ของ `craft check/run/build`, source bundle, GUI package/template, Windows build/installer และบันทึกผลจริงใน status Rev.17

## เกณฑ์ตรวจรับ

| กรณี | ผลที่ต้องเห็น |
| --- | --- |
| หน้าจอปกติ | Text/Button ขนาดตามข้อความและ padding; ไม่มีปุ่มขยายกินครึ่งหน้าต่างโดยไม่จำเป็น |
| ตาราง 8×8 | เซลล์แสดงข้อความอ่านได้และคลิกตรงตำแหน่งบนหน้าต่างอ้างอิง |
| ตาราง 16×16 | เซลล์ไม่หายเพราะ box สูง/กว้างศูนย์; หากพื้นที่ไม่พอ ต้องแสดงส่วนที่อยู่ใน viewport และรายงาน overflow ตาม contract |
| ตาราง 32×32 | ไม่ crash, ไม่ค้าง, ไม่สร้าง box ขนาดติดลบ; clip/overflow ชัดเจนและไม่มีภาพหลอกว่าตารางทั้งหมดถูกแสดง |
| Update หนึ่งเซลล์ | ข้อความเซลล์ที่เลือกและ status เปลี่ยนครบ ไม่มี glyph เก่าค้างหรือรอย repaint; hit test ยังตรง |
| Update ทั้งตาราง | ภาพถูกต้องทุกส่วนที่มองเห็น และ dirty/full repaint fallback ทำงานตาม threshold |
| DPI/resize | layout, font, clip และ hit test สอดคล้องกันที่ 100%, 150%, 200% เมื่อมีอุปกรณ์; resize ต่อเนื่องไม่ทิ้งภาพเก่า |
| Performance | รายงาน p50/p95/p99 ของ input ถึง draw completion, measure/layout/paint และ frame time เทียบ 0.1.16 ใน workload เดียวกัน; งาน UI เล็กต้องไม่ถดถอยอย่างมีนัยสำคัญโดยไม่มีเหตุผลที่อธิบายได้ |
| Compatibility | `go test ./...`, ตัวอย่าง GUI รุ่นก่อน, `craft check/run/build`, bundle 0.1.16 และ Windows installer ยังทำงาน |

การยอมรับรุ่นต้องอาศัยผลทดสอบภาพจริงของผู้ใช้และตัวเลขที่วัดซ้ำได้ ไม่อนุมานจากการคอมไพล์หรือจากการที่ปุ่มยังคลิกได้อย่างเดียว หากต้องเปลี่ยนพฤติกรรมของ API เดิม ให้ระบุผลต่างและวิธีย้ายโค้ดในเอกสารก่อนประกาศผ่าน

## งานนอกขอบเขต 0.1.17

- Craft Office และ visual designer แบบลากวาง
- ScrollView, virtualization, Table/Grid component สำเร็จรูป, TextBox/IME/editing, animation engine, blur/shadow และ accessibility เต็มรูปแบบ
- Linux/macOS GUI backend, หลายหน้าต่าง, standalone application binary ที่ไม่ต้องใช้ Craft CLI และ V1.00

0.1.17 จะทำให้ **พื้นฐาน layout และภาพถูกต้อง** ก่อนขยายความสามารถด้าน scrolling, ข้อมูลมาก และความสวยใน revision ถัดไป
