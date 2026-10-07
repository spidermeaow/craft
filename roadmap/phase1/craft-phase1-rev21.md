# Craft Phase 1 Rev.21 — โฟกัส Craft Runtime และพัก package ระดับสูง

วันที่: 7 ตุลาคม 2026  
ฐาน: Craft 0.1.20  
เป้าหมายรุ่น: Craft 0.1.21  
สถานะ: implement ใน source แล้ว; รอผู้ใช้ทดสอบตาม `docs/TRY-REV21.md`

## เหตุผลและขอบเขต

Rev.21 มุ่งทำให้ภาษา, Craft Runtime และ native engines มีสัญญาที่ชัดเจนและเสถียรก่อนขยาย API ระดับสูง. `packages/gui` เป็น wrapper ของ `std.ui`; `packages/api-framework` ประกอบ router, middleware และ response helpers บน `std.http` และ standard library. ทั้งสองเป็นต้นแบบงาน package ภายใน Phase 1 และทำให้การพัฒนา Runtime ต้องดูแล package, ตัวอย่าง และ distribution ไปพร้อมกัน.

Rev.21 จะพักการแจกและการพัฒนาทั้งสอง package ใน repository หลัก โดยเก็บ source เดิมไว้ที่ `Archive/packages/gui` และ `Archive/packages/api-framework`. การพักนี้ไม่เปลี่ยนสถานะของ package manager: Craft ยังต้องติดตั้งและ import library ทั่วไปได้. งาน package กลับมาพัฒนาใน Phase 1 Rev.24 เมื่อ Runtime รองรับ use case ที่ต้องการได้เพียงพอ.

## สิ่งที่ต้องคงไว้

- `std.ui` และ UI Engine ใน Runtime: window, event, input, layout, render, scrolling และ virtual list ที่มีอยู่ พร้อมสัญญา Craft ↔ native.
- `std.http` และ HTTP Engine: request/response, server และความสามารถพื้นฐานที่มีอยู่.
- Language core, checker/interpreter, standard library, package manager และ CLI ที่ไม่ขึ้นกับสอง package นี้.
- ตัวอย่างขนาดเล็กที่เรียก Runtime โดยตรง เพื่อให้ตรวจพฤติกรรมของ engine ได้โดยไม่ต้องพึ่ง package ระดับสูง.
- โปรเจกต์ `D:\02-Repository\test-gui` เป็นพื้นที่ทดสอบภายนอก repository นี้; อย่าย้ายหรือลบโดยอัตโนมัติ. ระบุให้ชัดหากตัวอย่างนั้นยังใช้สำเนา GUI package ของตัวเอง.

## งาน implementation

1. **เก็บต้นฉบับ:** ตรวจว่าทุกไฟล์ใน `packages/gui` และ `packages/api-framework` มีสำเนาครบใน `Archive/packages/`. บันทึก version และที่มาของสำเนา. คง archive เป็น reference ที่ไม่ถูก installer หรือ default build เก็บไปแจก.
2. **ถอด package ที่ใช้งานอยู่:** นำ `packages/gui` และ `packages/api-framework` ออกจากตำแหน่ง active หลังตรวจสำเนา. อย่าลบ `Archive` หรือแก้ source ใน archive เพื่อให้กู้คืนได้.
3. **จัดการตัวอย่าง:** `examples/api-server` และ `examples/gui-bridge`, `examples/gui-layout`, `examples/gui-stress` อ้าง path ใต้ `packages/`. ย้ายตัวอย่างที่ต้องเก็บไว้เข้า archive หรือเขียนใหม่ให้ใช้ `std.http`/`std.ui` โดยตรง; ห้ามปล่อย manifest ที่ชี้ไป path ที่หายไป. คงตัวอย่าง Runtime ที่ใช้ได้จริงอย่างน้อยหนึ่งชุดสำหรับ HTTP และ UI.
4. **จัดการ `craft gui init`:** คำสั่งนี้ใช้ GUI template ที่ฝังใน binary และยังสร้าง package ได้แม้ถอด `packages/gui`. สำหรับขอบเขต Rev.21 ให้พักคำสั่งและ template สำหรับการแจกเป็นค่าเริ่มต้น พร้อมปรับ help, diagnostics และเอกสาร. เก็บ template เดิมใน archive หากต้องแก้หรือลบ source ก่อนเริ่มงาน package รอบใหม่. ไม่ให้ CLI ชวนใช้ package ที่ประกาศว่าพักไว้.
5. **ตัวติดตั้งและ artifact:** แก้ `packaging/windows/craft.iss` ที่ดึง `packages/*` และตัวอย่างที่พึ่ง package. ปรับรายการไฟล์ใน release script/manifest ที่เกี่ยวข้อง. Binary และ installer ที่ build ไว้ก่อนหน้าไม่เปลี่ยนตาม source โดยอัตโนมัติ; artifact ใหม่ต้องสะท้อนขอบเขต Rev.21.
6. **เอกสารและการทดสอบใน source:** ปรับ README, roadmap/current status, คู่มือ GUI/API และ test ที่อ้าง path package จริง (เช่น `tests/rev4_test.go`) ให้บอกสถานะ archived หรือตรวจ Runtime ผ่าน fixture อิสระ. คงเอกสารประวัติรุ่นเก่าไว้ได้ แต่ติดป้ายว่าเป็นประวัติ ไม่ใช่คำแนะนำติดตั้งปัจจุบัน.
7. **กลับไปโฟกัส Runtime:** ใช้ตัวอย่างที่เรียก `std.ui`/`std.http` โดยตรงเพื่อแยกปัญหา engine จากความซับซ้อนของ wrapper. จัดลำดับงาน UI ตาม responsiveness, rendering correctness, input/text และ workload จริง; จัดลำดับ HTTP ตาม lifecycle, errors, concurrency และ resource handling. งานเพิ่มความสามารถจริงให้กำหนดเกณฑ์วัดเฉพาะก่อน implement.

## เกณฑ์ตรวจรับ

| ประเด็น | ผลที่ต้องได้ |
| --- | --- |
| Archive | มี source, manifest, README และ tests ของทั้งสอง package ครบ; เปิดอ่านและนำกลับมาใช้ได้ |
| Active tree | ไม่มี dependency path ที่ชี้ไป `packages/gui` หรือ `packages/api-framework` ที่ถูกถอดแล้ว |
| Runtime | โปรแกรมที่ใช้ `std.ui` และ `std.http` โดยตรงยังใช้ API เดิมได้; package manager ยังรองรับ library อื่น |
| CLI และ installer | help, GUI init, ตัวอย่าง และรายการไฟล์ที่แจกตรงกับสถานะ package; installer ไม่อ้าง glob ที่ว่าง |
| เอกสาร | ผู้ใช้เห็นชัดว่า Rev.21 เน้น Runtime และทั้งสอง package ถูกพักไว้จนกว่า Runtime จะพร้อมขึ้น |

การตรวจ source และ diff ทำได้ระหว่าง implement. การรันทดสอบอัตโนมัติหรือเปิด GUI ให้ผู้ใช้เป็นผู้ทำ เว้นแต่ผู้ใช้ขอทดสอบสำหรับงานนั้นโดยตรง. ส่งขั้นตอนทดสอบและผลที่คาดหวังให้ผู้ใช้เมื่อส่งมอบ.

## นอกขอบเขต 0.1.21

ไม่ลบ UI Engine หรือ HTTP Engine ออกจาก Runtime, ไม่ยกเลิกระบบ package ทั่วไป, ไม่เริ่ม Craft Office, ไม่สร้าง GUI component library หรือ API Framework รุ่นใหม่ และไม่ถือว่าการจัดระเบียบขอบเขตครั้งนี้เป็นการประกาศ V1.00.
