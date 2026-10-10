# แผน Craft 0.2.2 — สร้างแอปสำหรับแจกจ่ายจาก CraftBundle

วันที่: 8 ตุลาคม 2026  
สถานะ: **แผนอนาคต ยังไม่เริ่ม implement**  
ขอบเขตเวอร์ชัน: `0.2.2` ในเอกสารนี้หมายถึง **Craft toolchain** ไม่ใช่เลขเวอร์ชันเดิมของ `packages/gui`

งาน Language Core และ import อยู่ใน Craft 0.1.41: [Multiline Syntax](phase1/craft-phase1-rev41-multiline-language-plan.md) และ [Project Modules](phase1/craft-phase1-rev41-project-modules-plan.md). การแจกจ่ายแอปเป็นงาน 0.2.2 หลังรูปแบบ source และโมดูลของ 0.1.41 นิ่งแล้ว

## เป้าหมาย

ให้ผู้พัฒนาสร้างแอปจาก Craft source แล้วแจกจ่ายเป็นผลลัพธ์ที่รันได้บนสภาพแวดล้อมเป้าหมาย โดยไม่บังคับให้ผู้ใช้ปลายทางติดตั้ง Craft CLI หรือ Craft Runner แยก. `.craftbundle` เป็นรูปแบบกลางที่เก็บ source, dependencies, assets และ metadata ที่จำเป็น; อาจใช้ภายในขั้นตอน build โดยไม่ต้องเป็นสิ่งที่ผู้ใช้ปลายทางเห็น.

Craft เป็นเครื่องมือ build และ runtime ของแอป ไม่สร้างแพลตฟอร์ม Runner กลางที่ต้องติดตั้งเพิ่มเพื่อใช้แอปทั่วไป.

## ผลลัพธ์สองประเภท

| เป้าหมาย | สิ่งที่ส่งมอบ | ขอบเขต Craft | ขอบเขตระบบปลายทาง |
| --- | --- | --- | --- |
| Server / API | executable หรือแพ็กเกจ deploy ที่รวม Craft Runtime กับ bundle | โหลด bundle, รันแอป, รับ configuration, graceful shutdown, health/status และ structured logs | service manager หรือ container จัดการ restart, process supervision, rollout และ log collection |
| Desktop | แอปหรือชุดติดตั้งเฉพาะ OS ที่รวม Craft Runtime, bundle และ assets | เปิดแอปได้เอง, ตรวจ capability, จัดการ lifecycle และข้อผิดพลาดที่ผู้ใช้เห็น | OS จัดการหน้าต่าง, shortcut, การติดตั้ง และบริการ native ผ่าน backend ของ OS นั้น |

คำสั่งในอนาคตอาจอยู่ใต้ `craft build --target server` และ `craft build --target desktop --os windows` แต่ **ชื่อคำสั่งและ flags ยังไม่ใช่ API ที่รับรอง**; ต้องออกแบบ CLI ก่อน implement.

## หลักการออกแบบ

1. รักษา `.craftbundle` เป็น input กลางที่ตรวจสอบ format, Craft version, dependencies, assets และความสามารถที่ต้องใช้ได้ก่อนรัน. Bundle ที่ไม่เข้ากันต้องรายงานสาเหตุชัดเจน.
2. ผลลัพธ์แบบ self-contained ในระยะแรกอาจฝัง **interpreter และ bundle** ใน executable. ผู้ใช้เปิดไฟล์รันได้เอง แต่ไม่อ้างว่า source Craft ถูกคอมไพล์เป็น machine code. Native compilation เป็นหัวข้อแยกในอนาคต.
3. Server target ไม่ทำระบบ orchestration แบบ IIS ใหม่. ใช้ service manager หรือ container ที่มีอยู่ และกำหนดสัญญา startup, readiness, shutdown, exit code, configuration และ logs ให้เชื่อมต่อได้.
4. ไม่ฝัง secrets หรือค่าจำเพาะของเครื่อง deploy ลงใน bundle. รับผ่าน configuration และ environment ตามสัญญาที่กำหนด.
5. Desktop target ต้องมี Runtime และ UI backend ที่รองรับ OS เป้าหมายจริง. Windows เป็น backend ที่มีอยู่; macOS/Linux ต้องพัฒนาและตรวจรับแยกก่อนประกาศรองรับ GUI บน OS นั้น.
6. Artifact แต่ละ OS ต้องสร้างและแจกจ่ายตามรูปแบบของ OS นั้น; การใช้ bundle เดียวกันไม่หมายถึงใช้ binary ไฟล์เดียวกันทุก OS.
7. Desktop artifact ต้องบรรจุไอคอนและ assets ของ **แอปนั้น** ตาม metadata ที่แอปกำหนด เช่นไอคอน `.exe`, shortcut และตัวติดตั้งบน Windows; แยกจาก Window Icon API ของ `package/gui` ใน Rev.40.

## ลำดับงานที่เสนอ

1. กำหนด bundle compatibility, capability manifest, asset resolution และข้อผิดพลาดเมื่อ Runtime ไม่รองรับ.
2. ทำ Desktop Windows แบบ self-contained จาก backend ปัจจุบัน เพื่อพิสูจน์ว่าผู้ใช้เปิดแอปได้โดยไม่ติดตั้ง Craft แยก.
3. ทำ Server target แบบ self-contained พร้อมสัญญาการ deploy โดยใช้ service manager/container ของระบบ.
4. ขยาย Desktop ไปยัง macOS/Linux หลังมี UI backend, packaging และการตรวจรับของแต่ละ OS. CLI/server target อาจรองรับ OS เหล่านี้ก่อน Desktop ได้.

## เกณฑ์ตรวจรับเมื่อเริ่มทำจริง

- ผู้ใช้ปลายทางติดตั้งหรือเปิดแอป Desktop บน OS ที่ประกาศรองรับได้ โดยไม่มี Craft CLI หรือ Runner ติดตั้งอยู่ก่อน.
- Deploy API บนเครื่องที่ไม่มี Craft CLI ได้ โดย configuration และ secrets แยกจาก artifact; process ตอบสนองต่อ shutdown และแสดงสถานะพร้อมใช้งานอย่างถูกต้อง.
- เปรียบเทียบการรันจาก source, bundle และ artifact ที่แพ็กแล้วด้วยแอปตัวอย่างเดียวกัน; ข้อผิดพลาดจาก version/capability ไม่ตรงต้องอ่านเข้าใจ.
- ระบุ OS/architecture, รูปแบบแพ็กเกจ, ขนาด artifact, การอัปเดตและขอบเขตการสนับสนุนจากผลทดสอบจริงก่อนประกาศพร้อมใช้.

## ยังไม่รวมในแผนนี้

- Native compiler ที่แปลงภาษา Craft เป็น machine code โดยตรง.
- Runner แบบ IIS ที่เป็นผลิตภัณฑ์หรือบริการใหม่สำหรับโฮสต์หลายแอป.
- คำรับรองว่า GUI ใช้ได้บน macOS/Linux ก่อนมี backend และผลตรวจรับ.

งานนี้เป็นแผนสำหรับ Craft 0.2.2 ในอนาคต; ยังไม่เปลี่ยนพฤติกรรมของ `craft build` หรือ Setup ใน Craft รุ่นปัจจุบัน.
