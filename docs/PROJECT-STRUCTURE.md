# โครงสร้างโปรเจกต์ Craft

Craft เป็นภาษาโปรแกรมทั่วไปและอยู่ใน **Phase 1** ซึ่งรวมทั้ง Runtime และ package. Source ปัจจุบันคือ **0.1.50 candidate**; `gui`, `gui-template`, `api` และ `db` ใช้เวอร์ชันเดียวกับ Craft. ดู [README หลัก](../README.md), [Rev.50](REV50-RESPONSIVE-GUI.md) และ [หัวข้อพัฒนา Phase 1](../roadmap/workstreams.md).

Phase 2 ที่วางไว้เป็นช่วงตรวจรับและปรับ Craft ให้พร้อมใช้งานจริงกับแอปจริง; ยังไม่เริ่ม. การประกาศ V1.00 แยกจากการเปลี่ยน Phase.

| ตำแหน่ง | หน้าที่ |
| --- | --- |
| `cmd/craft/` | จุดเริ่มของ Craft CLI |
| `internal/` | parser/checker/interpreter, standard library, native UI/HTTP/database runtime และระบบ project/package |
| `packages/gui/` | GUI library ที่เขียนด้วย Craft; มี `craft.toml`, `src/` และ README ของ package |
| `packages/gui-template/` | Enterprise Light compositions ที่ต่อยอดจาก `packages/gui` |
| `packages/api/`, `packages/db/` | HTTP และ database helpers บน standard library |
| `examples/` | แอปตัวอย่างที่แยกจาก Runtime; `gui-package-lab` ทดสอบวิธีใช้ `packages/gui` |
| `templates/` | โครงโปรเจกต์เริ่มต้นของ CLI |
| `tests/` | การตรวจสอบ Craft CLI/Runtime ใน repository |
| `docs/` | สัญญาความสามารถ, คู่มือ และขั้นตอนตรวจรับ; `REV*`/`TRY-REV*` เป็นเอกสารตามรุ่นย้อนหลัง |
| `roadmap/` | ทางเข้าของแผนและสถานะ; `phase1/` รวมแผน Runtime และ `phase1/packages/` สำหรับแผน package |
| `packaging/windows/`, `scripts/` | สูตรสร้างและบรรจุ Craft CLI, package, ตัวอย่างและเอกสาร |
| `craft-vscode/`, `craft-file-icon-theme/` | VS Code extensions ที่มีเวอร์ชันของตัวเอง |
| `Archive/` | snapshot ของ package/งานเก่า ไม่ใช่ source ที่ติดตั้งเป็นค่าเริ่มต้น |
| `dist/`, `build/`, `.local/` | ผล build, เครื่องมือ build และข้อมูลท้องถิ่น; ไม่ใช่ source canonical |

`D:\02-Repository\craft-ui` และ `D:\02-Repository\InventoryManagement` เป็นแอปแยกจาก repository นี้; ทั้งสองอ้างอิง package จาก Craft checkout ในเครื่อง. แอป Showcase ที่รวมใน repository นี้อยู่ภายใต้ `examples/`.

## กติกาเวอร์ชันและเอกสารใหม่

1. `craft version`, `internal/toolchain/version.go` และ `packaging/windows/craft.iss` ระบุเวอร์ชัน Craft เดียวกัน (`0.1.50`).
2. แต่ละ package ระบุเวอร์ชันที่ `packages/<name>/craft.toml`; โปรเจกต์ที่ใช้ package ระบุเวอร์ชันเดียวกันใน dependency และล็อกด้วย `craft package lock`.
3. แผนใหม่อยู่ภายใต้ Phase 1 Rev. ถัดไป พร้อมหัวข้อ Runtime และ package ที่แยกชัด ระบุเวอร์ชันของ Craft และ dependency แต่ละตัว รวมถึงสถานะตรวจรับของทั้งสองส่วน.
4. ตัวอย่างและคู่มือต้องบอกว่าใช้ Runtime และ package รุ่นใด. ผล build สำเร็จไม่เท่ากับผลทดสอบ GUI หรือความพร้อมใช้งานจริง.
5. งาน package 0.2.1/0.2.2/0.2.3 เป็นประวัติ Rev.24/25/26 ที่ `roadmap/phase1/packages/`. `REV*` เป็นเอกสารตามรุ่น; ใช้แผนและสถานะ Rev. ล่าสุดเป็นจุดเริ่มอ่าน.

สูตร Setup ปัจจุบันบรรจุ `craft.exe` 0.1.50 และ source ของ `gui`, `gui-template`, `api`, `db`. ตัวติดตั้งใน `dist/` เป็นผล build เฉพาะเครื่องและไม่อยู่ใน Git; การ push source ไม่ได้เผยแพร่ GitHub Release. `craft build` ยังให้ source bundle ที่ต้องใช้ Craft Runtime; ยังไม่ใช่ standalone native executable ของแอปผู้ใช้.
