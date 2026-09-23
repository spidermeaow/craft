# Craft Phase 1 Rev.9 — Safe tooling primitives

วันที่: 21 กันยายน 2026  
ฐาน: Rev.8 / Craft 0.1.8  
เป้าหมายรุ่น: 0.1.9  
สถานะ: พร้อมติดตั้ง 0.1.9; ดูผลทดสอบและขอบเขตการรับรองใน [สถานะ Rev.9](output/phase1-rev9-status.md)

## 1. เป้าหมาย

Rev.9 ทำให้ Craft เหมาะสำหรับสร้าง command-line tools และ package ที่ทำงานกับไฟล์ได้อย่างคาดเดาได้ โดยทำให้ filesystem, path, runtime configuration และ diagnostics ที่มีอยู่เป็น **safe tooling primitives** ที่มี contract ชัดเจน

Rev.2 มี `std.fs`, `std.path`, `std.env` และ arguments อยู่แล้ว ดังนั้นรอบนี้ไม่สร้าง namespace ซ้ำ แต่เติมช่องว่างที่ทำให้เครื่องมือใช้งานจริงเสี่ยง เช่น การเขียนไฟล์ที่ไม่ atomic, path ที่ยัง normalize/resolve ไม่ครบ และ error ที่ยังไม่จำแนก operation ได้ดี

เป้าหมายคือปิดชุด API ที่รับเข้าขอบเขตให้ stable ได้ ไม่ใช่อ้างว่า standard library ทั้งหมดเสร็จใน Rev.9

## 2. ขอบเขตที่รับเข้า

### 2.1 Path contract

- เพิ่ม `std.path.clean(path)`, `std.path.absolute(path)` และ `std.path.relative(base, target)` หรือ API ที่เทียบเท่าซึ่ง type signature และ failure behavior ระบุชัด
- กำหนด semantics ของ separator, volume/drive, UNC path, empty path, `.` และ `..` บน Windows target; `clean` เป็น lexical normalization และต้องไม่อ้างว่าตรวจการมีอยู่หรือแก้ symlink
- ระบุชัดว่าทุก filesystem API ยังคง resolve relative path จาก process working directory ไม่ใช่ project root โดยปริยาย
- รักษา `join`, `fileName`, `extension`, `parent`, `isAbsolute` เดิมให้ compatible หรือมี migration note หากมีความจำเป็นต้องเปลี่ยน behavior

### 2.2 Filesystem contract

- เพิ่ม `std.fs.writeTextAtomic(path, text)` สำหรับเขียน UTF-8 ผ่าน temporary file ใน directory เดียวกัน แล้ว replace destination เฉพาะเมื่อเขียนและปิดไฟล์สำเร็จ
- กำหนด policy การเขียนทับ, permission ของไฟล์ใหม่, cleanup ของ temporary file, behavior เมื่อ destination ถูกล็อก และการรายงานกรณี replace/cleanup ล้มเหลว
- เพิ่ม API bytes เฉพาะเมื่อ type `Bytes` ที่มีอยู่รองรับ contract ได้ครบ: `readBytes` และ `writeBytesAtomic`; ทั้งคู่ต้องมีขนาดสูงสุดที่ระบุและไม่ทำ text decoding โดยปริยาย
- ทบทวน API เดิม `readText`, `writeText`, `appendText`, `copyFile`, `moveFile`, `removeFile`, `createDirectory`, `listDirectory`, `exists` และ `isDirectory` ให้ระบุ precondition, symlink behavior, overwrite rule และ limit ไว้ในเอกสารเดียว
- รักษา `removeFile` ให้ลบได้เฉพาะ regular file และไม่เพิ่ม recursive delete, wildcard delete หรือ directory removal ใน Rev.9
- ไม่อ้างว่า API filesystem เป็น sandbox หรือป้องกัน TOCTOU race จาก process อื่น; error จาก OS ยังคงเกิดได้หลัง validation

### 2.3 Runtime configuration และ command arguments

- ทำให้ `std.env.args`, `std.env.get` และ `std.env.lookup` มี contract เดียวกัน: แยก unset ออกจาก present-empty ด้วย `lookup` และยืนยันว่าไม่สามารถแก้ parent process environment ได้
- เพิ่ม helper ขนาดเล็กสำหรับการ validate required environment/argument เฉพาะเมื่อออกแบบ type/error semantics ได้ชัด; หลีกเลี่ยง parser framework หรือ global mutable configuration
- ไม่โหลด `.env` อัตโนมัติ และไม่เพิ่ม secret store; environment และ source bundle ต้องไม่ทำให้ secret ปรากฏใน diagnostics, lockfile หรือ bundle

### 2.4 Diagnostics และ logging ขั้นพื้นฐาน

- ทำ errors ของ filesystem/path ให้บอก operation และ path ที่เกี่ยวข้องโดย sanitize ข้อความจาก OS และไม่รวมเนื้อหาไฟล์, environment value หรือ secret
- logging ถูกเลื่อนออกจาก release gate: ยังไม่มี contract สำหรับ level, structured fields, redaction และ output destination ที่เล็กพอและไม่ผูก policy ของ application

### 2.5 ตัวอย่าง เอกสาร และ compatibility

- เพิ่มตัวอย่าง CLI ที่อ่าน configuration จาก argument/environment, validate ค่า, แล้วเขียนผลแบบ atomic โดยไม่พึ่ง API Framework
- อัปเดต `docs/LANGUAGE.md` และสร้างคู่มือ Rev.9 ที่ระบุ API, error/cancellation/cleanup semantics, Windows caveats, ขนาดสูงสุด และตัวอย่าง failure handling
- เพิ่ม migration note เฉพาะเมื่อ API เดิมมี behavior ที่ต้องเปลี่ยน; bundle format และ local package/lock semantics ของ Rev.8 ต้องไม่เปลี่ยนโดยไม่จำเป็น

## 3. สิ่งที่ไม่ทำใน Rev.9

- package registry, GitHub/remote dependency, publish, update หรือ semantic-version range
- process execution, shell invocation, signals, REPL, file watching หรือ scheduled jobs
- recursive directory deletion, filesystem sandbox, permission/ACL management, transactional multi-file update หรือ crash-proof durability guarantee ระดับ filesystem/volume
- `.env` auto-loading, secret store, configuration framework, CLI parser framework หรือ application-wide dependency injection
- API Framework, ORM, migrations, CORS, authentication, OpenAPI หรือ database driver ใหม่
- HTTP streaming/client/TLS feature, native compiler backend, VM, lambda/closure หรือ Enum

## 4. แผนดำเนินงาน

1. **M0 — Baseline และ contract:** inventory API Rev.2 ที่มีอยู่, รวบรวม compatibility tests และกำหนด Windows/path/error/secret policy ก่อนเปลี่ยน code
2. **M1 — Path:** เพิ่ม path normalization/resolution ที่เลือก, tests สำหรับ drive, UNC, separator, relative paths และ documentation ของ working-directory semantics
3. **M2 — Atomic filesystem I/O:** implement text atomic write ก่อน; เพิ่ม bytes API เฉพาะเมื่อ limits, memory behavior และ error handling ครบ; ทดสอบ replace, locked destination และ temp cleanup
4. **M3 — Configuration/diagnostics:** ปรับ errors ให้ระบุ operation โดยไม่รั่ว secrets, ยืนยัน argument/environment contract และเลื่อน logging ออกจาก release gate
5. **M4 — Tool example และ docs:** เพิ่ม example package/CLI, acceptance script และคู่มือ troubleshooting; ตรวจ bundle และ local package use case
6. **M5 — Acceptance:** รัน regression ที่เกี่ยวข้อง, Windows filesystem acceptance, race check สำหรับ implementation ที่มี concurrent state, build/installer และบันทึกหลักฐานใน `output/phase1-rev9-status.md`

## 5. เกณฑ์ผ่าน Rev.9

- API path และ filesystem ที่รับเข้า scope มี signature, limit, relative-path rule, symlink rule, overwrite rule และ failure behavior บันทึกชัดเจน
- `writeTextAtomic` ไม่ทำลาย destination เดิมเมื่อการเขียนหรือปิด temporary file ล้มเหลว; ผลกรณี OS/process crash ต้องระบุเป็น best-effort ไม่ใช่ durability guarantee
- การลบยังเป็น non-recursive และ regular-file only; ไม่มี destructive default ใหม่
- text/bytes limits ป้องกันการอ่านหรือ materialize ข้อมูลไร้ขอบเขตตาม contract ที่ประกาศ
- error ระบุ operation/path ได้โดยไม่เปิดเผย file contents, environment values, credentials หรือ secrets
- ตัวอย่าง CLI ใช้งาน `args`/`env` และ atomic write ได้โดยไม่ import API Framework และ check/test/build/run ผ่านตามหลักฐานจริง
- full regression ของ Rev.1–Rev.8, package/lock/bundle compatibility และ installer/build ที่เกี่ยวข้องผ่านบน Windows target ที่รองรับ โดยรายงานสิ่งที่ไม่ได้รันอย่างชัดเจน

## 6. หลักฐานและการทดสอบ

ก่อนประกาศ Rev.9 ให้เตรียม script/fixtures ที่ใช้ temporary directory เฉพาะการทดสอบและไม่ลบ path นอก fixture โดยต้องครอบคลุมอย่างน้อย:

- read/write UTF-8, invalid UTF-8, size limit, missing path, denied/locked destination และ directory ที่ถูกส่งแทน file
- atomic write success/failure, destination ที่มีอยู่และไม่มีอยู่, temporary-file cleanup, replace failure และ behavior หลัง retry
- path cases ของ Windows: relative, drive-rooted, drive-relative, UNC, mixed separator, `.`/`..`, trailing separator และ base/target ที่ไม่สามารถทำ relative ได้
- environment absent/present-empty, arguments, diagnostic redaction และ source/bundle/lockfile ที่ไม่มี secret
- existing filesystem API compatibility, project discovery จาก subdirectory, local package check/lock/build/run และ formatter write behavior

ผู้ใช้เป็นผู้รันทดสอบและบันทึกผล acceptance ตาม working preference ของ repository; ห้ามประกาศผ่านเพียงเพราะ build สำเร็จ

## 7. งานหลัง Rev.9

หลัง Rev.9 API ที่ปิดรับรองจะเป็น stable surface: รับ bug, security issue และ compatibility regression โดยไม่ขยาย semantics แบบไม่จำเป็น ส่วน feature ใหม่ เช่น process/signal, watcher, crypto/random, HTTP client/streaming, advanced encoding และ test utilities ต้องมี revision, contract และ acceptance แยกต่างหาก

ดู [สถานะ Rev.8](output/phase1-rev8-status.md), [แผน Rev.8](craft-phase1-rev8.md) และ [ทิศทางการพัฒนา Craft](../craft-language-direction.md)
