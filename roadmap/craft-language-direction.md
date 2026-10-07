# Craft development direction — a general-purpose language

อัปเดต: 7 ตุลาคม 2026

สถานะ: นิยามและทิศทางหลักของโครงการ

นิยามปัจจุบันอยู่ใน [Craft Phase 1](workstreams.md): ภาษา, Runtime, package และการตรวจรับอยู่ใน Phase เดียวกัน โดยแยกหัวข้อในแต่ละ revision. Craft CLI/Runtime เป็น `0.1.25`; GUI package `0.2.2` เป็นเลข dependency ไม่ใช่ Phase ใหม่.

Phase 2 ที่วางไว้จะเน้นความพร้อมใช้งานจริงจากผลแอปจริง ไม่ใช่การย้าย package ไปพัฒนาแยก Phase. การประกาศ V1.00 เป็นการตัดสินใจของเจ้าของโครงการ.

## วิสัยทัศน์

Craft เป็นภาษาโปรแกรมแบบ statically typed สำหรับสร้างโปรแกรมได้หลายประเภท ไม่ผูกนิยามของภาษาไว้กับ API Framework, GUI Framework หรือผลิตภัณฑ์ใดผลิตภัณฑ์หนึ่ง นักพัฒนาควรติดตั้งแล้วเริ่มเขียนโปรแกรมได้ง่าย และใช้ภาษา, runtime, standard library และ package system เป็นฐานสำหรับสร้าง CLI, เครื่องมือ, บริการ และแอปพลิเคชัน

API Framework ต้นแบบอยู่ใน `Archive/packages/api-framework` และไม่ใช่แกนกลางของภาษา. GUI bridge 0.1.13 เป็นประวัติช่วงเริ่มต้น; ปัจจุบัน UI Engine อยู่ใน Runtime และ `packages/gui` เป็น library ที่ใช้ Runtime นี้.

## ขอบเขตปัจจุบันและเป้าหมาย

- **ปัจจุบัน (Craft 0.1.25, GUI package 0.2.2):** Craft มี checker, AST interpreter, CLI, standard library, ระบบ source package จาก public GitHub ตาม exact tag และ Windows UI Runtime ที่มี tree/layout, text/input, scroll/virtual list, retained Canvas และ target events. GUI package ประกอบ API เหล่านี้เป็น library ที่เลือกใช้ได้. ความสามารถ 0.1.25 และ package 0.2.2 ยังรอผู้ใช้ตรวจรับ GUI จริง; `craft build` สร้าง source bundle ที่ต้องใช้ Craft CLI ในการรัน
- **ทิศทางต่อไป:** เปิดทางให้ library เชื่อมกับความสามารถ native อย่างมีสัญญาที่ชัดเจน เช่น event loop, callbacks และ resource lifecycle เพื่อรองรับ GUI และการต่อยอดประเภทอื่น
- **เป้าหมายระยะยาว:** แอป GUI แบบ native ที่ไม่ใช้ WebView และแจกจ่ายให้เปิดได้เอง; สิ่งนี้ยังต้องมี UI engine, platform backend และวิธีบรรจุ runtime หรือ native backend เพิ่มเติม

ขั้นแรกของสายงาน GUI อยู่ใน [แผน Craft 0.1.13](phase1/craft-phase1-rev13.md): ให้ Craft สื่อสารกับ native UI host บน Windows ผ่าน package และ typed runtime boundary ก่อนพัฒนา component หรือ renderer เต็มรูปแบบ ผลตรวจรับจริงอยู่ใน [สถานะ Rev.13](phase1/output/phase1-rev13-status.md)

ขั้นถัดไปอยู่ใน [แผน Craft 0.1.14](phase1/craft-phase1-rev14.md): สร้าง UI tree, layout, text และปุ่มที่รับ input ได้เป็นแนวตั้งเส้นแรก โดยยังคง GUI เป็น package ที่เลือกใช้ได้

การพัฒนา renderer และข้อความในรุ่น 0.1.15 อยู่ใน [แผน Rev.15](phase1/craft-phase1-rev15.md) และ [สถานะ Rev.15](phase1/output/phase1-rev15-status.md)

งาน UI responsiveness ในรุ่น 0.1.16 อยู่ใน [แผน Rev.16](phase1/craft-phase1-rev16.md) และ [สถานะ Rev.16](phase1/output/phase1-rev16-status.md)

งาน layout และ repaint correctness สำหรับ 0.1.17 อยู่ใน [แผน Rev.17](phase1/craft-phase1-rev17.md) และ [สถานะ Rev.17](phase1/output/phase1-rev17-status.md)

งาน ScrollView สำหรับ 0.1.18 อยู่ใน [แผน Rev.18](phase1/craft-phase1-rev18.md) และ [สถานะ Rev.18](phase1/output/phase1-rev18-status.md)

ผลทดสอบและการแก้ sizing สำหรับ 0.1.19 อยู่ใน [แผน Rev.19](phase1/craft-phase1-rev19.md) และ [สถานะ Rev.19](phase1/output/phase1-rev19-status.md)

Virtualized List รุ่นแรกใน 0.1.20 อยู่ใน [แผน Rev.20](phase1/craft-phase1-rev20.md) และ [สถานะ Rev.20](phase1/output/phase1-rev20-status.md)

คุณภาพเป้าหมายของสาย GUI และวิธีวัดความไว ความสวย และความลื่นอยู่ใน [ทิศทาง Craft GUI](craft-gui-direction.md) สำหรับ GUI ให้ความไวจาก input ถึงภาพใหม่และ frame time มาก่อนการประหยัดทรัพยากร เอกสารทิศทางแยกเป้าหมายระยะยาวจากความสามารถที่มีจริงใน Rev.20

## หลักการ

- ภาษาและ runtime ต้องมี behavior ที่คาดเดาได้ อ่านง่าย และตรวจสอบได้
- ความสามารถพื้นฐานอยู่ในภาษาและ standard library; policy ของงานเฉพาะอยู่ใน package ของผู้ใช้
- API ของ standard library ต้องเป็น composable primitives ไม่ผูกกับ router, ORM, authentication หรือ application architecture
- ตัวอย่างต้องใช้งานได้จริงและมี failure behavior ที่ระบุไว้ ไม่ใช้ตัวอย่างเพื่อซ่อนข้อจำกัดของ runtime
- ความง่ายในการเริ่มต้นและข้อความ error สำคัญพอ ๆ กับจำนวน feature
- ทุก resource ที่มี native handle ต้องมี ownership, cancellation, timeout และ cleanup contract
- package ที่เขียนด้วย Craft ต้อง check, test, bundle และนำไปใช้ซ้ำได้โดยไม่แก้ Craft runtime
- ไม่เพิ่ม feature ใหญ่เพียงเพื่อให้ตัวอย่าง framework ดูสมบูรณ์
- ความสามารถเฉพาะทาง เช่น API Framework และ GUI ควรอยู่ใน package เท่าที่ทำได้ ส่วนการเข้าถึงระบบปฏิบัติการหรือ GPU ต้องมี native boundary ที่ตรวจสอบได้
- GUI ใช้ Craft source เป็นวิธีสร้างหน้าจอ ไม่พัฒนาเครื่องมือออกแบบหน้าจอแบบลากวาง; ประสิทธิภาพต้องวัดทั้งเวลาเปิด, RAM, CPU/GPU ตอนอยู่นิ่ง, การตอบสนอง และ frame time ตาม workload จริง

## ขอบเขตของแต่ละชั้น

```text
Craft language
  ├─ compiler/checker/interpreter/CLI
  ├─ standard library primitives
  │    ├─ IO, files, paths, environment, time, process
  │    ├─ JSON, text, bytes, crypto primitives
  │    ├─ HTTP transport/client/server primitives
  │    ├─ database connection/query/transaction primitives
  │    └─ test, logging and diagnostics support
  └─ package/project system
       └─ user packages and tools
            ├─ API framework
            ├─ migrations/ORM
            ├─ CORS/authentication policy
            ├─ GUI library → native UI engine / platform backend (อนาคต)
            └─ domain-specific tools
```

Go ยังคงเป็น implementation language ของ Craft runtime, CLI และ native standard library ส่วนเครื่องมือระดับ policy ควรเขียนด้วย Craft เพื่อให้ผู้ใช้ศึกษา แก้ไข และต่อยอดได้ใน source เดียวกับแอปของตน

## ลำดับการพัฒนา

### ระยะพื้นฐานภาษา

รักษา static typing, functions, structs, collections, modules/import, errors, tasks, timers, formatter, checker, test blocks และ bundle compatibility ให้เสถียร พร้อม diagnostics ที่บอกวิธีแก้ได้

### ระยะ standard library

เติมของที่โปรแกรมหลายประเภทใช้ร่วมกัน:

- logging และ structured diagnostics ที่ไม่เปิดเผย secrets
- file/path/directory และ atomic write
- configuration primitives จาก environment/arguments พร้อม validation
- HTTP request/response, headers, cookies, `OPTIONS`, client, server lifecycle และ test transport
- database pool, typed values, transactions, timeout, cancellation และ schema metadata
- bytes/text/encoding, JSON, time และ random/crypto primitives
- process/signal, graceful shutdown และ test utilities

แต่ละความสามารถต้องมีตัวอย่าง console หรือ library ที่ไม่ต้อง import API Framework เพื่อพิสูจน์ว่าเป็น primitive ที่ใช้ได้ทั่วไป

### ระยะ package และเครื่องมือผู้ใช้

ทำให้ผู้ใช้สร้าง package ได้ง่าย: manifest, local modules, dependency lock, version compatibility, bundle, documentation, test discovery และ diagnostics ของ import/name collision

Rev.11 รองรับการดึง source package จาก public GitHub ด้วย exact tag และ lockfile แล้ว ส่วน package registry และการโหลด native extension ยังเป็นงานอนาคต

### ระยะ ecosystem

เมื่อ primitives เสถียร ผู้ใช้สามารถสร้าง package ภายนอกเองได้ ตัวอย่างเช่น:

- `api-framework` และ routing/middleware
- database migration/ORM/query builder
- CORS/authentication/OpenAPI policy
- CLI framework และ worker toolkit
- GUI library และ native UI engine สำหรับ desktop application

Craft core จะรับ feature กลับเข้ามาเฉพาะเมื่อพิสูจน์แล้วว่าเป็น primitive ที่หลายประเภทเครื่องมือต้องใช้ร่วมกัน ไม่รับ policy เฉพาะ framework มาเป็น syntax หรือ runtime behavior โดยตรง

## สิ่งที่ถอดออกจากเป้าหมายหลัก

ยกเลิกการวาง milestone เพื่อทำ API Framework ของ Craft เอง ยกเลิกการถือ router, middleware, CORS, authentication, OpenAPI, migration และ ORM เป็น acceptance ของภาษา. ต้นแบบ `Archive/packages/api-framework` เก็บไว้เป็น reference/ตัวอย่างและใช้ศึกษาสัญญา module/import ได้เมื่อจำเป็น แต่ไม่เป็น release gate

การมี HTTP และ Database ใน standard library หมายถึง Craft มี primitives สำหรับให้ผู้ใช้สร้าง framework ได้ ไม่ได้หมายความว่า Craft core ต้องกำหนดรูปแบบ framework ให้ทุกโครงการ

## เกณฑ์คุณภาพของ release

แต่ละ revision ต้องรายงานแยกกันว่า:

- ภาษาและ CLI ทำงานได้หรือไม่
- standard library primitives ใดผ่านจริง
- package/tool example ใดผ่านจริง
- driver, OS หรือ environment ใดมีข้อจำกัด
- tests ใดไม่ได้รัน

Build หรือ example happy path ไม่เพียงพอสำหรับการประกาศเสถียร ต้องมี tests ที่ตรวจ error, cancellation, ownership, cleanup และ compatibility ตาม feature นั้น ๆ

Rev.7 เป็นงาน harden HTTP และ Database primitives ตามแผนเดิม Revision ถัดไปให้ใช้เอกสารนี้เป็นทิศทางหลัก โดยไม่เพิ่ม API Framework หรือ GUI เป็นเงื่อนไขของแกนภาษา
