# Craft development direction — a language for building tools

วันที่: 20 กันยายน 2026  
สถานะ: ทิศทางหลักฉบับปัจจุบัน

## วิสัยทัศน์

Craft จะเป็นภาษาโปรแกรมที่นักพัฒนาติดตั้งแล้วเริ่มเขียนโปรแกรมได้ง่าย และสามารถนำภาษา, standard library และ package system ไปสร้างเครื่องมือของตนเองได้ เช่น web server, API framework, CLI, worker, database tool, test tool หรือ application library

Craft ไม่ได้มีเป้าหมายหลักเพื่อสร้าง API Framework ของโครงการเอง API Framework ที่อยู่ใน `packages/api-framework` เป็น reference package และหลักฐานว่าผู้ใช้สามารถเขียนเครื่องมือชั้นสูงด้วย Craft ได้เอง ไม่ใช่แกนกลางของภาษาและไม่ใช่เงื่อนไขของ release ถัดไป

## หลักการ

- ภาษาและ runtime ต้องมี behavior ที่คาดเดาได้ อ่านง่าย และตรวจสอบได้
- ความสามารถพื้นฐานอยู่ในภาษาและ standard library; policy ของงานเฉพาะอยู่ใน package ของผู้ใช้
- API ของ standard library ต้องเป็น composable primitives ไม่ผูกกับ router, ORM, authentication หรือ application architecture
- ตัวอย่างต้องใช้งานได้จริงและมี failure behavior ที่ระบุไว้ ไม่ใช้ตัวอย่างเพื่อซ่อนข้อจำกัดของ runtime
- ความง่ายในการเริ่มต้นและข้อความ error สำคัญพอ ๆ กับจำนวน feature
- ทุก resource ที่มี native handle ต้องมี ownership, cancellation, timeout และ cleanup contract
- package ที่เขียนด้วย Craft ต้อง check, test, bundle และนำไปใช้ซ้ำได้โดยไม่แก้ Craft runtime
- ไม่เพิ่ม feature ใหญ่เพียงเพื่อให้ตัวอย่าง framework ดูสมบูรณ์

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

CLI สำหรับ package registry หรือการดึง source จาก GitHub เป็นงานภายหลัง หลัง local package workflow และ lock semantics มีความเสถียร ไม่ให้ registry เป็นเงื่อนไขของการใช้ภาษา

### ระยะ ecosystem

เมื่อ primitives เสถียร ผู้ใช้สามารถสร้าง package ภายนอกเองได้ ตัวอย่างเช่น:

- `api-framework` และ routing/middleware
- database migration/ORM/query builder
- CORS/authentication/OpenAPI policy
- CLI framework และ worker toolkit

Craft core จะรับ feature กลับเข้ามาเฉพาะเมื่อพิสูจน์แล้วว่าเป็น primitive ที่หลายประเภทเครื่องมือต้องใช้ร่วมกัน ไม่รับ policy เฉพาะ framework มาเป็น syntax หรือ runtime behavior โดยตรง

## สิ่งที่ถอดออกจากเป้าหมายหลัก

ยกเลิกการวาง milestone เพื่อทำ API Framework ของ Craft เอง ยกเลิกการถือ router, middleware, CORS, authentication, OpenAPI, migration และ ORM เป็น acceptance ของภาษา แพ็กเกจ `packages/api-framework` คงไว้เป็น reference/ตัวอย่างและใช้ทดสอบ module/import ได้เมื่อจำเป็น แต่ไม่เป็น release gate

การมี HTTP และ Database ใน standard library หมายถึง Craft มี primitives สำหรับให้ผู้ใช้สร้าง framework ได้ ไม่ได้หมายความว่า Craft core ต้องกำหนดรูปแบบ framework ให้ทุกโครงการ

## เกณฑ์คุณภาพของ release

แต่ละ revision ต้องรายงานแยกกันว่า:

- ภาษาและ CLI ทำงานได้หรือไม่
- standard library primitives ใดผ่านจริง
- package/tool example ใดผ่านจริง
- driver, OS หรือ environment ใดมีข้อจำกัด
- tests ใดไม่ได้รัน

Build หรือ example happy path ไม่เพียงพอสำหรับการประกาศเสถียร ต้องมี tests ที่ตรวจ error, cancellation, ownership, cleanup และ compatibility ตาม feature นั้น ๆ

Rev.7 จึงเป็นการ harden HTTP + Database primitives และ test evidence ตามแผนเดิม ส่วน revision ถัดไปให้ใช้เอกสารนี้เป็นทิศทางหลัก โดยไม่เพิ่ม API Framework เป็นงานบังคับ
