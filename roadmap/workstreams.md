# Craft Phase 1: Runtime และ package

นิยามปัจจุบัน ณ 7 ตุลาคม 2026: **Craft ยังอยู่ใน Phase 1**. Phase 1 รวมการพัฒนาภาษา, Craft Runtime, standard library, package ที่เขียนด้วย Craft, ตัวอย่าง และงานตรวจรับ/ชุดติดตั้ง. งานทุกด้านทำควบคู่กันได้ภายใน Phase เดียว. การเปลี่ยนเป็น V1.00 จะเกิดเมื่อเจ้าของโครงการตัดสินใจ.

| หัวข้อพัฒนาใน Phase 1 | สิ่งที่รับผิดชอบ | ที่อยู่หลัก | เวอร์ชันที่เกี่ยวข้อง |
| --- | --- | --- | --- |
| ภาษาและ Runtime | checker/interpreter, CLI, standard library, UI/HTTP/database engines, native boundary | `cmd/`, `internal/`, `docs/REV*.md` | Craft `0.1.25`; `craft version` แสดงค่านี้ |
| Package | Craft API/composition และ developer experience บน runtime primitives | `packages/gui/`, `examples/gui-package-lab/` | `gui` `0.2.2` ใน `craft.toml`; เป็นเลข dependency ไม่ใช่ Craft version |
| คุณภาพและการแจกจ่าย | compatibility, diagnostics, performance, accessibility, installer, คู่มือและผลทดสอบ | `packaging/`, `scripts/`, `docs/TRY-*.md` | ใช้เวอร์ชัน Craft ของ artifact ที่ส่งมอบ |

**Craft 0.1.25 บรรจุ GUI package 0.2.2**. Setup แสดงเวอร์ชัน Craft `0.1.25` และรวม package `gui` `0.2.2`; ไม่เปลี่ยน `craft version` เป็น `0.2.2`. Package เป็นส่วนหนึ่งของงาน Phase 1 แต่ยังเลือกใช้เป็น dependency ได้. การทำ package ไม่ได้แปลว่า Runtime หรือ Craft ทั้งหมดพร้อมใช้จริง.

## แผนที่ใช้อ้างปัจจุบัน

- [Phase 1 Rev.24 (Craft 0.1.24)](phase1/craft-phase1-rev24.md): Canvas/paint Runtime และหัวข้อ package `gui` 0.2.1
- [Phase 1 Rev.25 (Craft 0.1.25)](phase1/craft-phase1-rev25.md): viewport/style/interaction Runtime และหัวข้อ package `gui` 0.2.2
- [สถานะ Rev.24](phase1/output/phase1-rev24-status.md) และ [สถานะ Rev.25](phase1/output/phase1-rev25-status.md) แยกผลของ Runtime/package ให้เห็นชัด

แผน package รุ่น 0.2.1/0.2.2 อยู่ใน `roadmap/phase1/packages/` และสถานะอยู่ใน `roadmap/phase1/output/`. แผนใหม่ให้ใส่หัวข้อ Runtime หรือ package ภายใต้ Phase 1 Rev. ถัดไป. `Archive/packages/` เก็บต้นแบบที่ไม่ใช่ package active.

หลักการวางงาน: สิ่งที่ต้องเข้าถึง OS, renderer หรือ native handle อยู่ที่ Runtime; composition และ policy ที่เขียนด้วย Craft ได้อยู่ที่ package; แอปตัวอย่างและ `craft-test` แยกจาก implementation. ดู [โครงสร้าง repository](../docs/PROJECT-STRUCTURE.md).

## ขั้นถัดไป

[Phase 2](craft-phases.md) เน้นทดสอบกับแอปจริงและปรับความเสถียร, compatibility, diagnostics, performance, accessibility และการแจกจ่าย. ยังไม่เริ่มจนกว่าเจ้าของโครงการจะตัดสินใจ. V1.00 เป็นการตัดสินใจแยกจากเลข Phase.
