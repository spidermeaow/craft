# Craft Phase 1 Rev.28 (0.1.28) — Text reflow และ grid ตามพื้นที่จริง

วันที่: 7 ตุลาคม 2026  
สถานะ: **ลง source แล้ว; รอผู้ใช้ทดสอบ** ([บันทึกสถานะ](output/phase1-rev28-status.md))  
ฐาน: Craft UI Runtime 0.1.27, `packages/gui` 0.2.5 และผลทดสอบ `responsive-tester` ที่ผู้ใช้รายงานว่าใช้งานได้ระดับหนึ่ง

## ปัญหาที่ต้องแก้

Rev.27 อ่าน viewport, เปลี่ยน layout เมื่อ resize/DPI change, จัด Row/Column, เลื่อนเนื้อหา และใช้ `responsiveGrid` ได้แล้ว แต่ยังมีช่องว่างสองจุด:

1. Text วัดขนาดแบบบรรทัดเดียว. ข้อความยาวอาจดันความกว้างของ parent จนเกิด scrollbar แนวนอน หรือถูก clip แม้มีพื้นที่แนวตั้งให้ใช้.
2. `responsiveGrid` ต้องรับ `availableWidth` ที่แอปคำนวณจาก viewport หลังหัก sidebar, padding และ gap เอง. การคำนวณคลาดเคลื่อนทำให้ grid ล้นหรือเหลือพื้นที่ไม่พอดี.

เป้าหมายคือให้ Text และ grid **รับข้อจำกัดความกว้างจาก parent ใน Runtime** และให้ Package เปิด API ที่แอปใช้ได้ง่ายขึ้น โดยยังรักษา state, breakpoint และนโยบาย overflow ที่แอปต้องเลือกเอง.

## งาน Runtime 0.1.28

### 1. Text wrapping ที่สัมพันธ์กับการวัดและวาด

- เพิ่มตัวเลือก wrap ให้ `Text` แบบ opt-in; ค่าเริ่มต้นของ `text` และ `textStyle` เดิมยังเป็นบรรทัดเดียวเพื่อรักษาความเข้ากันได้. ชื่อ API และ signature จะสรุปหลังตรวจข้อจำกัดของ Craft checker; ตัวอย่างชื่อเบื้องต้นคือ `std.ui.textWrap(node, mode)`.
- วัดความสูงของ Text จากความกว้างที่ parent จัดสรรจริง แล้วใช้การแบ่งบรรทัดเดียวกันตอนวาด. ต้องรองรับข้อความอังกฤษ/ไทย, ช่องว่าง, newline และคำหรือช่วงข้อความที่ยาวกว่าหนึ่งบรรทัด โดยกำหนดพฤติกรรมตัดคำยาวให้ชัดก่อน implement.
- เมื่อ resize หรือ DPI เปลี่ยน ให้คำนวณจำนวนบรรทัด ความสูงของ parent และขอบเขต ScrollView ใหม่. หลีกเลี่ยงการวัดซ้ำโดยไม่จำเป็นเมื่อความกว้างและ style ไม่เปลี่ยน.
- ตรวจความสอดคล้องของ DirectWrite และ GDI fallback สำหรับ baseline, clipping และ DPI; ถ้าผลต่างจาก renderer ทำให้ข้อความหายหรือทับกัน ต้องแก้ก่อนปิดงาน.
- จำกัดขอบเขตไว้ที่ `Text` ก่อน. `Button`, `TextInput` และ Canvas text ยังไม่รับประกัน multiline ใน Rev.28.

### 2. Grid ที่ใช้ความกว้างซึ่ง parent จัดให้จริง

- เพิ่ม layout primitive ที่รับ `items`, `minCellWidth`, `maxColumns` และ `gap` แล้วคำนวณคอลัมน์จาก **allocated width** ระหว่าง layout. Package ไม่ต้องเดาความกว้างจาก viewport. ชื่อ API เบื้องต้นคือ `std.ui.grid`; ยังไม่ถือเป็น signature สุดท้าย.
- เมื่อพื้นที่ลดลง ให้ลดจำนวนคอลัมน์ก่อนลด cell ต่ำกว่าค่าขั้นต่ำ. หากพื้นที่แคบกว่าหนึ่ง cell ให้ใช้กติกา overflow/scroll ที่ระบุชัด ไม่สร้างความกว้างติดลบหรือวาง item ทับกัน.
- วัดความสูงแต่ละแถวจาก child ที่สูงที่สุด หลัง Text wrapping; แถวถัดไปต้องเริ่มหลังแถวก่อนหน้าจริง.
- ให้การ resize, DPI change, nested Row/Column, `size`, `maxSize`, `spacing` และ ScrollView ใช้กติกาเดียวกัน. ระบุเมื่อ grid ต้อง remeasure/repaint และรักษา focus/scroll offset ที่เกี่ยวข้อง.

### 3. สัญญา overflow และขนาดขั้นต่ำ

- เขียนกติกาเมื่อ fixed/min/max/weight ขัดกับพื้นที่จริง: ส่วนใดหด ส่วนใด overflow และส่วนใด clip. พื้นที่ที่เล็กกว่าขนาดขั้นต่ำของ control ต้องยังเข้าถึงได้ด้วยการเลื่อนหรือมีข้อจำกัดที่บันทึกชัด.
- ตรวจว่า ScrollView ขยาย content ตามความกว้าง viewport เมื่อมีพื้นที่ และคง overflow แนวนอนเฉพาะเมื่อ content ต้องใช้พื้นที่มากกว่า viewport.
- ไม่เปลี่ยน `Panel` ให้เรียง children; `Panel` ยังคงเป็นการซ้อนชั้น ส่วน content หลายบรรทัดให้ใช้ Column ภายใน Panel ตาม package 0.2.5.

## งาน GUI Package 0.2.6

- ห่อ Text wrapping และ grid primitive ใหม่โดยคง signature ของ `gui.responsiveGrid(id, items, availableWidth, minCellWidth, maxColumns, gap)` เดิมไว้เพื่อ compatibility. เพิ่ม helper ใหม่ที่ **ไม่ต้องรับ `availableWidth`** และใช้ขนาดที่ parent จัดให้; ชื่อ API จะกำหนดหลัง Runtime ตกลง signature.
- เพิ่ม composition สำหรับหน้าแบบ TopBar + Sidebar ที่สลับ Sidebar เต็ม/ย่อ/เมนูบนตาม breakpoint โดยให้ caller ระบุ breakpoint, sidebar width, menu items และ state ที่เลือก. Package ไม่เป็นเจ้าของ state หรือ event loop.
- ให้ `surfaceCard` รองรับ Text ที่ wrap ภายใน `Panel` + `Column` โดยไม่เปลี่ยนข้อจำกัดว่า `surface` รับ `Panel` หรือ `Button`.
- อัปเดต `examples/gui-package-lab`, `gui-package-dashboard` และ `responsive-tester` ให้ใช้ grid ที่รับความกว้างจาก parent และแสดงข้อความยาวจริง; สำเนา package, manifest และ lockfile ต้องตรงกับ source.
- อัปเดตคู่มือ migration: `responsiveGrid` แบบเดิมยังใช้ได้, helper ใหม่เหมาะกับเนื้อหาที่ต้องเติมพื้นที่, และ Canvas พิกัดคงที่ยังต้องปรับด้วยแอปเอง.

## ลำดับดำเนินงาน

1. บันทึกสัญญา API/หน่วย logical pixels และชุดกรณีขอบก่อนแก้ Runtime: ความกว้าง 0/น้อยมาก, ข้อความไทย/อังกฤษ, newline, DPI 100%/125%, renderer ทั้งสองชนิด.
2. ทำ Text measurement, wrapping และ paint ให้ตรงกัน พร้อม layout ที่คำนวณความสูงหลังทราบ allocated width.
3. ทำ Grid primitive จาก parent constraints และตรวจ nested layout/scroll/focus.
4. ห่อ API ใน Package 0.2.6 แล้วเปลี่ยนตัวอย่างสามแอป โดยแยกให้เห็นว่าส่วนใดเป็นความสามารถ Runtime และส่วนใดเป็น composition ของ Package.
5. อัปเดต docs, version/manifest/lockfile และ Setup หลัง code พร้อม; บันทึกผล build แยกจากผลทดสอบ GUI.

## เกณฑ์ตรวจรับ

| ส่วน | ผลที่ต้องเห็น |
| --- | --- |
| Runtime Text | ข้อความยาวไทย/อังกฤษใน Text wrap ตามความกว้างจริง; ความสูงเพิ่มโดยไม่ทับ item ถัดไป; newline และ DPI ทำงานตามสัญญา |
| Runtime Grid | ย่อ/ขยายหน้าต่างแล้วจำนวนคอลัมน์เปลี่ยนจากความกว้าง parent โดยไม่มี `availableWidth` จากแอป; cell และแถวไม่ทับกัน |
| Overflow | หน้าต่างแคบกว่าขนาดขั้นต่ำยังเข้าถึงเนื้อหาที่เกินผ่าน ScrollView; ไม่มี scrollbar แนวนอนเมื่อเนื้อหาพอดี |
| Package | `gui.responsiveGrid` เดิมยังใช้ได้; helper ใหม่และ shell composition ใช้ในตัวอย่างโดยไม่เรียก `std.ui` ตรงจาก app source |
| Compatibility | ตัวอย่าง 0.1.27 และ package 0.2.5 ไม่เปลี่ยนพฤติกรรมโดยไม่ตั้งใจ; หาก API ต้องเปลี่ยน ให้มี migration ชัดเจน |
| การตรวจจริง | ผู้ใช้ตรวจ `craft package check`, `craft check`, `craft run`, resize ผ่าน breakpoint, ย้ายจอ DPI และลองกด/เลื่อน/พิมพ์ด้วยตนเอง |

## ขอบเขตที่ไม่รวม

Rev.28 ไม่ทำ multiline TextInput, rich text, CSS-like layout, reflow ของ Canvas แบบพิกัดคงที่, table/grid ข้อมูลเต็มรูปแบบ หรือคำรับประกันว่าข้อความทุกความยาวจะเห็นพร้อมกันในหน้าต่างที่เล็กกว่าขนาดขั้นต่ำ. งาน performance จะวัดเฉพาะผลของการเปลี่ยน layout/paint; ยังไม่ประกาศตัวเลข input-to-present โดยไม่มีเครื่องมือวัดถึงจอจริง.

การเปิดแอปและทดสอบ GUI ให้ผู้ใช้ทำเองตามความต้องการปัจจุบัน. ผล implementation และข้อจำกัดจริงอยู่ใน [บันทึกสถานะ](output/phase1-rev28-status.md); เกณฑ์ในแผนนี้ยังต้องรับผลจากผู้ใช้.
