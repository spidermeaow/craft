# Craft Phase 1 Rev.39 (0.1.39) — Rendering Foundation

**สถานะ:** ลง source รอบแรกแล้ว; ยังไม่ผ่าน Definition of Done และรอผู้ใช้ทดสอบ UI  
**ฐาน:** Craft UI Engine 0.1.38  
**เป้าหมาย:** ยกระดับระบบ Rendering ให้รองรับการวาดกราฟ รูปทรงซับซ้อน และการอัปเดตเฉพาะพื้นที่ที่เปลี่ยนแปลง โดยรักษาความเร็วและความเสถียรของ UI Engine

---

## 1. แนวคิดหลัก

Craft UI Engine ต้องเป็น General-purpose Native UI Engine ที่สามารถรองรับการสร้าง Desktop Application ที่มีความซับซ้อนระดับ Microsoft Excel, Visual Studio Code และ PowerPoint ได้

Rev.39 จะมุ่งพัฒนา Rendering Foundation โดยไม่เพิ่ม Component เฉพาะทางหรือ Spreadsheet Logic เข้าไปใน Runtime

**หลักการสำคัญ**

- Engine รับผิดชอบการวาดและจัดการพื้นที่แสดงผล
- Package รับผิดชอบ Component และพฤติกรรมเฉพาะทาง
- Application รับผิดชอบข้อมูลและ Business Logic
- Rendering API ต้องสามารถนำกลับมาใช้ได้กับ Component ทุกประเภท
- หลีกเลี่ยงการสร้าง API ที่ผูกติดกับ Excel หรือ Chart โดยเฉพาะ
- รักษาความเข้ากันได้กับ API เดิมของ 0.1.38

---

## 2. ขอบเขตการพัฒนา

### 2.1 General-purpose Path Rendering

เพิ่ม Drawing Primitives สำหรับการวาดรูปทรงซับซ้อนบน Canvas

**ความสามารถที่ต้องรองรับ**

- `moveTo` — กำหนดจุดเริ่มต้น
- `lineTo` — วาดเส้นตรง
- `arcTo` — วาดส่วนโค้ง
- `quadraticTo` — วาด Quadratic Bézier
- `cubicTo` — วาด Cubic Bézier
- `closePath` — ปิดเส้นทาง
- `fillPath` — เติมสีภายในรูปทรง
- `strokePath` — วาดเส้นขอบรูปทรง

**ข้อกำหนด**

1. รองรับทั้ง Open Path และ Closed Path
2. รองรับหลาย Subpath ภายใน Path เดียว
3. รองรับ Fill และ Stroke
4. รองรับความหนาของเส้น
5. รองรับ Clipping ภายใน Canvas
6. รองรับการวาด Arc, Circle, Ring และ Donut ผ่าน Primitive ทั่วไป
7. ใช้พิกัดแบบ Device-independent Pixel (DIP)
8. จัดการกรณี Path ว่างหรือข้อมูลพิกัดไม่ถูกต้องโดยไม่ทำให้ Runtime Crash

ชื่อ API เป็นข้อเสนอเชิงแนวคิด สามารถปรับให้สอดคล้องกับรูปแบบ API ปัจจุบันของ Craft ได้

### 2.2 Rendering Backend Compatibility

ความสามารถใหม่ต้องรองรับ Rendering Backend ที่มีอยู่ทั้งสองประเภท

| Backend | ข้อกำหนด |
|---|---|
| Direct2D | ใช้ Native Path Geometry และ Hardware Acceleration ตามความพร้อม |
| GDI Fallback | รองรับ Path Rendering ด้วยวิธีที่เหมาะสมกับ Backend |
| ทั้งสอง | ผลลัพธ์ต้องมีรูปทรง ตำแหน่ง ขนาด และ Clipping สอดคล้องกัน |

ไม่จำเป็นต้องให้ Anti-aliasing เหมือนกันทุก Pixel แต่ต้องไม่เกิดความผิดพลาดของรูปทรงหรือการจัดวาง

ห้ามทำให้ระบบเดิมที่วาด Rectangle, Line, Text และ Image เกิด Regression

### 2.3 Dirty-region Rendering

เพิ่มระบบตรวจจับและวาดใหม่เฉพาะพื้นที่ที่เปลี่ยนแปลง เพื่อลดงาน Rendering ที่ไม่จำเป็น

**ความสามารถที่ต้องรองรับ**

- กำหนดพื้นที่ที่ต้องวาดใหม่ (Invalidation Region)
- รวมพื้นที่ที่ซ้อนทับหรืออยู่ใกล้กันเมื่อเหมาะสม
- แปลงพิกัดของ Component เป็นพิกัดหน้าต่างอย่างถูกต้อง
- รองรับ Clipping และ Scroll Offset
- รองรับการเปลี่ยนแปลงของ Text, Canvas และ VirtualGrid
- รองรับการสั่ง Full Redraw เมื่อจำเป็น
- ไม่เกิดภาพตกค้างหรือพื้นที่ที่ไม่ได้รับการวาดใหม่

**ข้อกำหนดด้านความถูกต้อง**

Dirty-region Rendering ต้องคำนึงถึงทั้งพื้นที่เดิมและพื้นที่ใหม่ของวัตถุที่เคลื่อนย้ายหรือเปลี่ยนขนาด รวมถึงวัตถุที่ซ้อนทับกัน

หากไม่สามารถระบุพื้นที่ที่ได้รับผลกระทบได้อย่างปลอดภัย ให้ใช้ Full Redraw แทน

ห้ามใช้ Dirty Cell เป็นโครงสร้างหลักของ Engine เพราะเป็นแนวคิดเฉพาะของ Spreadsheet

### 2.4 Rendering Diagnostics

ต่อยอด Timing Trace ที่มีอยู่ให้สามารถวิเคราะห์ประสิทธิภาพการวาดได้

**ข้อมูลที่ควรวัด**

- Frame Render Time
- Layout Time
- Paint Time
- จำนวน Drawing Operations
- จำนวน Invalidation Regions
- พื้นที่ที่ได้รับการวาดใหม่
- จำนวน Full Redraw
- Rendering Backend ที่ใช้งาน

Diagnostic ต้องสามารถเปิดหรือปิดได้ และไม่ควรเพิ่มภาระที่มีนัยสำคัญเมื่อปิดใช้งาน

---

## 3. ขอบเขตสถาปัตยกรรม

```text
Craft Application
        |
        v
packages/gui
        |
        v
std.ui
        |
        +-- Layout / Input
        |
        +-- Canvas API
        |     |
        |     +-- Basic Primitives
        |     +-- Path Geometry
        |     +-- Fill / Stroke
        |     +-- Clipping
        |
        +-- Invalidation Manager
        |     |
        |     +-- Dirty Regions
        |     +-- Region Merging
        |     +-- Full Redraw Fallback
        |
        +-- Rendering Backend
              |
              +-- Direct2D
              +-- GDI Fallback
```

ต้องรักษาการแยกความรับผิดชอบระหว่าง API, Rendering Backend และ Invalidation Manager

หลีกเลี่ยงการรวมทุกส่วนไว้ในไฟล์หรือฟังก์ชันขนาดใหญ่

---

## 4. สิ่งที่ไม่อยู่ในขอบเขต Rev.39

Rev.39 ไม่ต้องพัฒนาสิ่งต่อไปนี้

- Chart Package เช่น LineChart, BarChart, PieChart
- Spreadsheet Formula Engine
- Variable Row / Column Sizing
- Pinned Rows / Columns
- Chart Animation
- GPU Shader API
- 3D Rendering
- Docking System
- การเปลี่ยนสถาปัตยกรรม UI Engine ทั้งระบบ

ความสามารถเหล่านี้สามารถพัฒนาต่อจาก Rendering Foundation ใน Revision ถัดไป

---

## 5. UI Showcase สำหรับทดสอบ

สร้างหรือขยาย Rendering Showcase โดยไม่ต้องใช้ Database หรือ External API

### Test A — Path Geometry

แสดงรูปทรงต่อไปนี้

- เส้นตรงและเส้นโค้ง
- Circle และ Arc
- Donut / Ring
- Bézier Curve
- รูปทรงปิดที่เติมสี
- รูปทรงหลาย Subpath

**เกณฑ์ผ่าน:** รูปทรงถูกต้อง ไม่มีเส้นผิดตำแหน่งหรือพื้นที่เติมสีผิดพลาด

### Test B — Backend Comparison

แสดง Showcase เดียวกันด้วย Direct2D และ GDI Fallback

**เกณฑ์ผ่าน:** ทั้งสอง Backend แสดงรูปทรงได้ถูกต้องและไม่มี Crash

### Test C — Dirty-region Rendering

สร้าง Canvas ที่มีวัตถุจำนวนมาก โดยเปลี่ยนแปลงเฉพาะวัตถุบางส่วน

ตรวจสอบว่า

- Engine ไม่จำเป็นต้องวาดทุกพื้นที่ใหม่ในทุก Frame
- วัตถุที่ไม่เปลี่ยนแปลงยังแสดงผลถูกต้อง
- วัตถุที่เคลื่อนที่ไม่ทิ้งภาพตกค้าง
- วัตถุที่ซ้อนทับกันไม่เกิดภาพผิดพลาด
- Full Redraw ยังคงให้ผลลัพธ์ถูกต้อง

**เกณฑ์ผ่าน:** ผลลัพธ์จาก Partial Redraw ต้องสอดคล้องกับ Full Redraw และมีหลักฐานจาก Diagnostics ว่าลดพื้นที่หรืองานวาดได้จริงในกรณีที่เหมาะสม

### Test D — VirtualGrid Regression

นำ VirtualGrid ของ 0.1.38 มาทดสอบอีกครั้ง

- เลื่อนแนวตั้งและแนวนอน
- เลือกเซลล์และช่วงเซลล์
- แก้ไขข้อมูลภายในเซลล์
- เปลี่ยนขนาดหน้าต่าง
- ใช้คีย์บอร์ดและเมาส์
- ตรวจสอบ Rendering ระหว่างเลื่อนข้อมูลจำนวนมาก

**เกณฑ์ผ่าน:** ความสามารถเดิมต้องไม่เสีย และต้องไม่เกิดภาพตกค้างหรืออาการกระตุกผิดปกติจากการเปลี่ยนแปลงครั้งนี้

---

## 6. Performance Requirements

การประเมินประสิทธิภาพต้องใช้การวัดจริง ไม่ใช่เพียงความรู้สึกว่า UI ลื่น

| รายการ | เกณฑ์ |
|---|---|
| Rendering Target | 60 FPS สำหรับ Showcase ที่กำหนด บนเครื่องทดสอบเป้าหมาย |
| Frame Time | เป้าหมายประมาณ 16.67 ms ต่อ Frame ที่ 60 Hz |
| Partial Redraw | ต้องลดงานวาดได้ในกรณีที่เปลี่ยนแปลงเฉพาะพื้นที่ |
| Full Redraw | ต้องแสดงผลถูกต้อง |
| Memory | ไม่เพิ่มขึ้นต่อเนื่องโดยไม่มีเหตุผล |
| Resize | ไม่มีภาพผิดพลาดหรือ Crash |
| Direct2D / GDI | ผ่าน Functional Tests ทั้งสอง Backend |

ค่า 16.67 ms เป็นงบเวลาต่อ Frame ที่ 60 Hz ไม่ใช่เวลาที่อนุญาตให้ Rendering ใช้ทั้งหมดเสมอไป

ต้องเก็บผล Baseline จาก 0.1.38 และเปรียบเทียบกับ Rev.39 ภายใต้เงื่อนไขเดียวกัน

---

## 7. Implementation Guidelines

1. ตรวจสอบโครงสร้าง Source Code ปัจจุบันก่อนแก้ไข
2. ใช้ API และสถาปัตยกรรมที่มีอยู่เป็นฐาน
3. แยก Path Geometry ออกจาก Backend Implementation
4. แยก Dirty-region Management ออกจาก VirtualGrid
5. หลีกเลี่ยงการสร้างไฟล์ขนาดใหญ่ที่รวมหลายความรับผิดชอบ
6. แบ่งฟังก์ชันให้มีหน้าที่ชัดเจนและตรวจสอบได้ง่าย
7. จัดรูปแบบโค้ดให้อ่านง่ายและสามารถเพิ่มบรรทัดหรือแก้ไขเฉพาะส่วนได้โดยไม่ทำให้ Syntax เสีย
8. เพิ่ม Automated Tests สำหรับความสามารถใหม่
9. รักษา Backward Compatibility กับ Rev.38
10. ห้ามประกาศว่าฟีเจอร์ผ่านก่อนมีผลทดสอบรองรับ

ไม่ควร Refactor ส่วนที่ไม่เกี่ยวข้องกับเป้าหมาย Rev.39 โดยไม่มีเหตุผลทางเทคนิคที่ชัดเจน

---

## 8. Definition of Done

Rev.39 จะถือว่าผ่านเมื่อ

- [ ] Canvas รองรับ Path Geometry ตามขอบเขต
- [ ] สามารถวาด Arc, Bézier และ Donut ได้
- [ ] Direct2D รองรับความสามารถใหม่
- [ ] GDI Fallback รองรับความสามารถใหม่
- [ ] Dirty-region Rendering ทำงานถูกต้อง
- [ ] ไม่มีภาพตกค้างระหว่าง Partial Redraw
- [ ] มี Full Redraw Fallback ที่ปลอดภัย
- [ ] Rendering Diagnostics ใช้งานได้
- [ ] UI Showcase ผ่านการทดสอบ
- [ ] VirtualGrid เดิมไม่เกิด Regression
- [ ] ผ่านการทดสอบ DPI ที่กำหนด
- [ ] ไม่มี Memory Leak ที่ตรวจพบจากชุดทดสอบ
- [ ] มีผล Benchmark เปรียบเทียบก่อนและหลัง
- [ ] ผู้ใช้ตรวจสอบผลลัพธ์จริงและยืนยันการผ่าน

---

## 9. Expected Outcome

เมื่อ Rev.39 เสร็จสมบูรณ์ Craft UI Engine ควรมี Rendering Foundation ที่รองรับการสร้าง Custom UI Components ที่ซับซ้อนมากขึ้น โดยไม่ต้องเพิ่ม Rendering Logic เฉพาะทางเข้าไปใน Runtime ทุกครั้ง

ตัวอย่างสิ่งที่สามารถนำไปต่อยอด ได้แก่

- Chart Package
- Diagram Editor
- Advanced Dashboard
- Custom Controls
- Interactive Canvas
- Spreadsheet Visualization

**เป้าหมายสุดท้ายของ Rev.39 ไม่ใช่การสร้างกราฟสำเร็จรูป แต่เป็นการทำให้ Craft UI Engine มีความสามารถพื้นฐานเพียงพอที่จะสร้างกราฟและ UI ซับซ้อนประเภทอื่นได้อย่างมีประสิทธิภาพ**

---

## 10. ข้อกำหนดสำหรับการส่งมอบ

หลัง Implement ให้สรุป

1. ไฟล์ที่เพิ่มและแก้ไข
2. API ที่เพิ่มหรือเปลี่ยนแปลง
3. ผล Automated Tests
4. ผล Direct2D / GDI Comparison
5. ผล Performance Benchmark
6. ข้อจำกัดหรือปัญหาที่ยังคงเหลือ
7. วิธีเปิด UI Showcase เพื่อให้ผู้ใช้ทดสอบด้วยตนเอง

**ห้ามระบุว่า Rev.39 เสร็จสมบูรณ์ หากยังไม่ได้ทดสอบและยืนยันตาม Definition of Done**

---

## Implementation รอบแรก (9 ตุลาคม 2026)

- เพิ่ม Canvas Path API ตามข้อ 2.1 ผ่าน `std.ui` และ wrapper ใน `packages/gui`; arc แปลงเป็น cubic segments ที่ใช้ร่วมกันทั้งสอง backend, quadratic แปลงเป็น cubic, และหลาย subpath ใช้ alternate/even-odd fill เพื่อวาด Donut ได้
- เพิ่ม Direct2D path geometry กับ GDI native path drawing ภายใต้ clip ของ Canvas. Public API รับเฉพาะสีทึบในรอบนี้เพื่อไม่ให้ alpha ของสอง backend ต่างกัน; จำกัด path 1024 commands และ 128 subpaths พร้อมตรวจข้อมูลไม่ถูกต้องก่อนส่ง renderer
- เพิ่มตัวจัดการ invalidation regions: เปรียบเทียบคำสั่ง Canvas ก่อน/หลัง, เก็บขอบเขตเก่าและใหม่, รวมบริเวณที่ซ้อน/ใกล้กัน, และ fallback เป็น full redraw เมื่อขอบเขตไม่แน่นอนหรือพื้นที่กว้าง. เปิดทางให้ VirtualGrid ที่ geometry คงเดิมใช้ partial redraw กับข้อความเซลล์/selection
- Timing trace เพิ่มจำนวน invalidation regions, พื้นที่พิกเซลที่ invalidated, full redraws และ backend ที่ใช้จริง
- เพิ่ม Go test source สำหรับ path compiler และการรวม dirty regions แต่ยัง **ไม่ได้รัน tests** ตาม preference ที่ผู้ใช้ทดสอบเอง; `go build ./cmd/craft` สำเร็จในขั้น compile เท่านั้น
- เพิ่มหน้า Rendering Lab ใน UI Showcase; ยังไม่ได้เปิดแอปหรือเปรียบเทียบ Direct2D/GDI/DPI/performance จริง ผล baseline, benchmark, memory และการรับรองจากผู้ใช้ยังค้าง จึงยังไม่ถือว่า Rev.39 เสร็จ

## ผลทดสอบเพิ่มเติม (9 ตุลาคม 2026)

ผู้ช่วยตรวจ Direct2D และ GDI ใน UI Showcase ตามคำขอผู้ใช้ พบและแก้เงื่อนไขที่ทำให้ adaptive text ที่ไม่เปลี่ยนบังคับ full relayout เมื่อ Canvas เปลี่ยน. หลังแก้ trace แสดง partial invalidation จากการย้ายสี่เหลี่ยมและ full redraw เมื่อกดปุ่ม; `go test ./internal/stdlib` และ `craft check` ใน Showcase ผ่าน. รายละเอียดตัวเลขอยู่ที่ [คู่มือการตรวจ Rendering](../../docs/REV39-UI-RENDERING.md). ยังขาดการตรวจหลาย DPI, benchmark ที่ควบคุมสภาวะ และการรับรอง Definition of Done ทั้งหมด
