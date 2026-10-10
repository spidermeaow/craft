# Craft Phase 1 Rev.41 (0.1.41) — โมดูลภายในโปรเจกต์ที่ประกาศขอบเขตชัดเจน

สถานะ: **ลง source รอบแรกใน 0.1.41 candidate; รอผลทดสอบและ compatibility**  
ขอบเขต: Craft Language Core และระบบโหลดโปรเจกต์; ใช้ได้กับ UI, API, CLI และ library โดยไม่ผูกกับ `package/gui`  
แผนร่วมรุ่น 0.1.41: [Multiline Syntax](craft-phase1-rev41-multiline-language-plan.md)  
งานต่อเนื่อง 0.2.2: [การแพ็กแอป](../craft-0.2.2-deployment-plan.md)

## ปัญหาปัจจุบัน

ไฟล์ `.craft` ทั้งหมดภายใต้ `src` ของโปรเจกต์เดียวกันถูกรวมเป็นโมดูลเดียว. ฟังก์ชันในโฟลเดอร์ย่อยเรียกข้ามไฟล์ได้โดยตรง และชื่อ declaration ต้องไม่ซ้ำกันทั้งโปรเจกต์. วิธีนี้เรียบง่ายสำหรับงานเล็ก แต่เมื่อโค้ดโตขึ้น โฟลเดอร์ไม่บอกขอบเขตการพึ่งพาและไม่ช่วยป้องกันชื่อชนกัน.

## เป้าหมาย

เพิ่ม **โมดูลภายในโปรเจกต์แบบเลือกใช้** เพื่อให้ผู้พัฒนาระบุขอบเขต, import และ API ที่เปิดเผยได้อย่างชัดเจน. โฟลเดอร์เป็นเพียงตำแหน่งเก็บไฟล์; การย้ายไฟล์ต้องไม่เปลี่ยน namespace หรือ visibility โดยอัตโนมัติ. โมดูลภายในโปรเจกต์ไม่ต้องมีเวอร์ชันและ lockfile แยกจากโปรเจกต์; ส่วนที่นำไปใช้ข้ามโปรเจกต์ยังใช้ package dependency ตามเดิม.

ตัวอย่างต่อไปนี้เป็น syntax ที่ลง source รอบแรกของ 0.1.41 candidate และยังรอผลทดสอบ:

```craft
module pages.excel
export func render(): UiNode { ... }
```

```craft
import module excel "pages.excel"
let page: UiNode = excel.render()
```

รอบแรกใช้ `module` declaration ในแต่ละไฟล์: ไฟล์ที่ประกาศชื่อเดียวกันอยู่ในโมดูลเดียวกันโดยไม่อิงชื่อโฟลเดอร์. การ import ใช้ชื่อโมดูลที่ประกาศจริง และยังต้องตรวจ diagnostics/compatibility จากการทดสอบ.

การออกแบบ import ทำให้ผู้ใช้เห็นที่มาได้ชัด: Standard Library ใช้ `std`, dependency ภายนอกใช้ `import package gui "gui"`, และโค้ดในโปรเจกต์ใช้ `import module excel "pages.excel"`. `import gui "gui"` เดิมยังหมายถึง package. Diagnostics/editor ยังต้องตรวจว่าบอกชนิดของ import และตำแหน่ง declaration ได้ชัดเพียงพอ.

## ความเข้ากันได้และขอบเขต

1. โปรเจกต์เดิมที่ไม่ประกาศโมดูลยังรวมไฟล์ใต้ `src` แบบเดิมและทำงานเหมือนเดิม; ไม่บังคับให้ UI Showcase หรือแอปเก่าย้ายทันที.
2. Declaration ที่ไม่ `export` ต้องถูกจำกัดในโมดูลที่ประกาศ. โมดูลอื่นเรียกผ่าน import และชื่อโมดูล/alias เท่านั้น; รายงานการเข้าถึงชื่อ private และชื่อชนกันให้เข้าใจง่าย.
3. Import ภายในโปรเจกต์กับ import package ภายนอกต้องแยกและ resolve ได้แน่นอน. ป้องกันวงจร import, alias ชนกัน และโมดูลชื่อซ้ำ.
4. `std` และ package เดิมยังมีความหมายเดิม. การเพิ่มโมดูลต้องไม่เปลี่ยนการทำงานของ API/CLI หรือการบรรจุ source bundle โดยไม่ตั้งใจ.
5. `craft fmt`, `craft check`, `craft run`, `craft build` และ diagnostics ต้องเข้าใจ source หลายโมดูลและรักษาเส้นทางไฟล์/ตำแหน่ง error.

## งานและเกณฑ์ตรวจรับ

- กำหนดสเปก syntax, identity, visibility และ import resolution พร้อมตัวอย่าง valid/invalid ก่อนแก้ Lexer/Parser.
- เพิ่ม module graph สำหรับโมดูลภายในโปรเจกต์ใน loader/type checker/interpreter/bundle โดยไม่ใช้พฤติกรรมของ path เป็น namespace โดยปริยาย.
- เพิ่ม regression cases: หลายไฟล์ในโมดูลเดียว, export/private, alias, declaration ชื่อซ้ำต่างโมดูล, วงจร import, package import ร่วมกับ internal import และโค้ดเดิมที่ไม่ประกาศโมดูล.
- ตรวจ compatibility กับ Standard Library, GUI Showcase, ตัวอย่าง API/CLI และ Craft packages ที่มีอยู่; บันทึก migration หากจำเป็น.
- ผู้ใช้ทดลองจัด UI Showcase บางส่วนเป็นโมดูลจริงและยืนยันว่า `craft check`/`craft run` ใช้งานได้โดยไม่มีการเปลี่ยนพฤติกรรมหน้าเดิม.

มี implementation รอบแรกใน Parser และ project module resolver พร้อม regression cases ที่ยังไม่ได้รันตามความต้องการให้ผู้ใช้ทดสอบเอง. ยังไม่รับรองผล compatibility ของโมดูลรูปแบบใหม่.
