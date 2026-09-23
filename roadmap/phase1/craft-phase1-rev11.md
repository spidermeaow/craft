# Craft Phase 1 Rev.11 — GitHub packages และ dependencies แยกรายโปรเจกต์

วันที่: 23 กันยายน 2026  
ฐาน: Rev.10 / Craft 0.1.10  
เป้าหมายรุ่น: 0.1.11  
สถานะ: implementation และ automated acceptance ผ่าน สร้าง Setup แล้ว; ยังไม่ได้รันตัวติดตั้ง ดู [หลักฐาน Rev.11](output/phase1-rev11-status.md) สำหรับผลทดสอบและข้อจำกัด

## 1. เป้าหมายและแนวคิด

ผู้สร้าง library/framework แจก source ผ่าน GitHub พร้อม tag ของแต่ละรุ่น ผู้ใช้เลือก repository และรุ่นด้วยคำสั่งเดียว Craft ดาวน์โหลด ตรวจสอบ และบันทึก dependency ให้โปรเจกต์โดยอัตโนมัติ แต่ละโปรเจกต์เลือกรุ่นของตัวเองได้โดยไม่กระทบโปรเจกต์อื่น และไม่ต้อง activate environment แบบ Python venv

แผนนี้ใช้ Rev.11 สำหรับ GitHub packages ตามทิศทางล่าสุดของผู้ใช้ งาน editor formatting/diagnostics ที่เคยเสนอให้เลื่อนไปแผนแยก ไม่รวมใน acceptance รอบนี้

## 2. Workflow ที่ต้องรองรับ

### ผู้สร้าง package

1. สร้าง Craft project ที่มี `craft.toml`, ชื่อ, version และ `entry = "library"` ที่ repository root
2. เขียน public exports, README, license และ tests แล้วตรวจด้วย Craft
3. Push repository ขึ้น GitHub และสร้าง tag เช่น `v1.2.0` ซึ่ง manifest ระบุ `version = "1.2.0"`
4. ส่ง repository URL และ tag ให้ผู้ใช้ ไม่ต้องลงทะเบียน registry หรือเรียก `craft publish`

### ผู้ใช้ package

คำสั่งเป้าหมาย (ยังไม่มีใน CLI ปัจจุบัน):

```powershell
craft install github.com/dev/mylib@v1.2.0
craft package list
craft check
craft run
```

คำสั่ง install ต้องดาวน์โหลดรุ่นที่เลือก ตรวจ manifest/graph แล้วเพิ่ม dependency ใน `craft.toml` และ commit/checksum ใน `craft.lock` ให้เอง ผู้ใช้ไม่ต้องดาวน์โหลดทุกรุ่นหรือแก้ path cache ด้วยมือ

ให้ alias เริ่มต้นมาจากชื่อ package ที่ผ่าน validation รองรับ `--alias` เพื่อแก้ชื่อชนกัน ห้ามเขียนทับ alias เดิมที่ชี้ไปคนละ source โดยเงียบ ๆ การเลือกรุ่นใหม่ของ source เดิมอย่าง explicit คือการอัปเดต dependency เฉพาะโปรเจกต์นั้น

## 3. Manifest และ lock contract

รูปแบบที่เสนอให้ implement:

```toml
[dependencies]
mylib = { github = "dev/mylib", tag = "v1.2.0", version = "1.2.0" }
greeting = { path = "packages/greeting", version = "1.0.0" }
```

- GitHub dependency ใช้ exact tag เท่านั้นในรอบแรก: `vMAJOR.MINOR.PATCH` ตรงกับ manifest version; ไม่รองรับ branch/latest หรือ version ranges
- `path` และ `github` เป็น source คนละชนิด ใช้ร่วมกันใน dependency เดียวไม่ได้; local dependency เดิมยังทำงานได้
- Lock ระบุ schema version, canonical repository identity, requested tag, resolved full commit SHA, package identity, graph และ source checksum ตามวิธี canonicalization ที่บันทึกชัดเจน
- Lock ไม่เก็บ absolute cache path, token หรือข้อมูล environment; cache ย้ายตำแหน่งได้โดยไม่ทำให้ lock เปลี่ยน
- ตรึง commit หลัง resolve tag: tag ที่ถูกย้ายต้องไม่เปลี่ยน source ของ lock เดิม การ refresh ต้องเป็นคำสั่ง explicit และรายงานการเปลี่ยน identity/checksum
- กำหนด lock schema ใหม่สำหรับ remote sources พร้อมอ่าน local lock เดิมและ upgrade เมื่อผู้ใช้สั่ง install/lock อย่างตั้งใจ
- `craft install` ที่ไม่มี argument restore จาก manifest/lock; ถ้ามี lock ต้องใช้ commit เดิม ถ้า manifest ไม่ตรง lock ให้ fail พร้อมแนวทางแก้ ไม่เลือกเวอร์ชันใหม่เอง
- `craft install --offline` ใช้เฉพาะ source ที่ตรวจ checksum ได้ใน cache; ของขาดต้อง fail โดยไม่เรียก network
- `craft package lock` ไม่เปลี่ยน remote pin เดิมโดยอ้อม; ระบุ semantics ของ graph change และการยอมรับ local source change ให้คง behavior เดิมเท่าที่ทำได้

## 4. ที่เก็บ package และการแยกรุ่น

- ใช้ cache ที่ Craft จัดการใน user cache directory ไม่ติดตั้ง code ลง runtime หรือ global library search path
- Cache key ต้องรวม repository, commit และ checksum ห้ามใช้เพียงชื่อ `mylib1.2` เพราะคนละผู้สร้างอาจใช้ชื่อ/รุ่นเดียวกัน
- เก็บหลายรุ่นร่วมกันได้ เช่น A ตรึง 1.2.0 และ B ตรึง 2.0.0; resolver อ่าน manifest/lock ของ project root เสมอ
- Source ใน cache ถือเป็น immutable ตรวจ integrity ก่อนใช้งาน ไม่แก้ source cache เมื่อแก้ manifest ของโปรเจกต์
- Download ลง staging directory แล้ว promote หลังตรวจครบ รองรับ concurrent installs ด้วย synchronization และป้องกัน cache ครึ่งชุดเมื่อถูกยกเลิก
- ไม่มี recursive cache purge/garbage collection ใน Rev.11; ไม่ลบรุ่นที่โปรเจกต์อื่นอาจใช้อยู่
- ไม่จำเป็นต้อง copy package ลง `packages/` ของแอป; โฟลเดอร์นั้นยังใช้สำหรับ local packages ที่ผู้ใช้จัดการเอง

## 5. Resolver, imports และ bundle

- รองรับ transitive GitHub dependencies และตรวจ cycle, missing package, version mismatch, alias collision และ duplicate identity อย่าง deterministic
- Local paths ภายใน remote package ต้องอยู่ภายใน source tree ของ repository นั้น; ห้าม dependency หนีออกไปอ่าน filesystem ของผู้ใช้
- รอบแรกคงข้อจำกัด resolver เดิม: หลายโปรเจกต์ใช้คนละรุ่นได้ แต่ dependency graph เดียวที่ต้องการหลายรุ่นของ package เดียวให้ fail ด้วย conflict path ที่ชัดเจน ไม่อ้างว่ารองรับ multi-version graph
- ตรวจชื่อ/identity ชนกันแม้มาจากคนละ repository ห้ามรวม package โดยเดาว่าเป็น source เดียวกัน
- `check/run/test/build` ทำงานจาก cache ตาม lock ไม่ดาวน์โหลดเอง; หากขาดให้แนะนำ `craft install`
- `package check/list` แสดง source/tag/commit/status ที่เหมาะสม; ไม่รัน code ของ dependency
- Bundle รวม dependency source ที่ตรวจแล้ว ไม่ต้องใช้ network/cache ของเครื่องเดิมเมื่อรัน bundle และไม่บรรจุ token/cache metadata
- กำหนด bundle language compatibility เมื่อ remote manifest/schema เปลี่ยน ให้ runtime เก่าปฏิเสธอย่างชัดเจนถ้าไม่รองรับ

## 6. Network และความปลอดภัยของ source

- รองรับ public GitHub repositories ผ่าน HTTPS เท่านั้น; private repository/authentication/submodules/Git LFS/monorepo package selection เลื่อนไปก่อน
- กำหนด transport contract ก่อน implement: endpoint และ redirects ที่ยอมรับ, timeout, cancellation, rate-limit diagnostics และเพดาน download/extracted size/file count
- Reject archive traversal, absolute paths, symlink/hardlink, Windows reserved paths, case-collision และ duplicate entries ตาม supported platform
- ไม่มี post-install hook หรือการรัน shell/build/test ของ package อัตโนมัติ
- แยก integrity ออกจาก trust: checksum ยืนยัน source ชุดเดิม ไม่ได้ยืนยันว่า code ปลอดภัย ผู้ใช้ต้องเลือกผู้สร้างที่ไว้ใจได้
- ความล้มเหลวระหว่าง download/validation ต้องไม่เปลี่ยน manifest/lock เดิม; การเขียนสองไฟล์ต้องมี recovery/journal หรือ protocol ที่ตรวจและกู้ภาวะ partial update ได้ ไม่อ้างว่า rename สองไฟล์เป็น transaction เดียว

## 7. GitHub ที่ต้องเตรียม

**ไม่จำเป็นต้อง release ตัว Craft ขึ้น GitHub ก่อนเริ่มพัฒนา** ใช้ CLI ที่ build ในเครื่องทำ implementation และ fixture tests ได้

ก่อน live acceptance ต้องมี public repository สำหรับ library ทดสอบจริง พร้อม manifest, exported API และ tags อย่างน้อย `v1.0.0`, `v1.2.0`, `v2.0.0` โดย version ใน source แต่ละ tag ตรงกัน ผู้ใช้เป็นผู้ระบุ repository/บัญชีปลายทางก่อน push หรือ publish; การเขียนแผนนี้ไม่อนุญาตให้อัปโหลดเอง

เริ่มด้วย fixtures/local test server ได้จนถึงจุดนี้ แล้วแจ้งผู้ใช้ว่าต้องเตรียม URL/tags อะไร ไม่สร้างบัญชีหรือ publish repository โดยอนุมาน

GitHub Releases ของ **ตัว Craft** เป็นช่องทางแจก CLI/Setup ให้ผู้ใช้ภายนอก จัดทำได้ภายหลังหรือก่อนเปิดใช้งานจริง ส่วน **library** ต้องมี repository/tag แต่ไม่บังคับอัปโหลด binary หรือ Release asset เพิ่ม

## 8. Milestones

1. **M0 — Contracts:** สรุป install syntax, exact-tag policy, alias conflicts, lock migration, cache layout, transport และ limits
2. **M1 — Local model:** เพิ่ม source model/schema/resolver โดยคง local package regression และตรวจ graph conflicts
3. **M2 — Fetch/cache:** downloader, archive validation, immutable cache, cancellation/concurrency และ integrity
4. **M3 — CLI:** install/add/restore/offline semantics, manifest editing ที่รักษาข้อมูลผู้ใช้ และ transactional recovery
5. **M4 — Integration:** check/list/run/test/build/bundle, transitive packages และ diagnostics
6. **M5 — Acceptance/release:** fixture tests, public GitHub repository test, regression และ Windows Setup 0.1.11 พร้อมคู่มือผู้สร้าง/ผู้ใช้

## 9. เกณฑ์รับงาน

- สองโปรเจกต์ใช้ library คนละรุ่นได้จริงโดยไม่กระทบกัน; import behavior ตรงกับ lock
- Install explicit tag เพิ่ม manifest/lock ถูกต้อง; restore ใช้ commit/checksum เดิม; offline restore ทำงานเมื่อ cache ครบ
- ทดสอบ moved tag, corrupted cache, rate limits, cancellation, truncated/oversized/malicious archive และ concurrent installs
- Fault injection ระหว่าง update manifest/lock กู้คืนได้ ไม่ทำให้โปรเจกต์ใช้ dependency ผิดชุดโดยเงียบ ๆ
- Transitive/cycle/version/identity conflict diagnostics ชี้ dependency chain ที่เกี่ยวข้อง
- Bundle รันได้โดยไม่ใช้ network/cache ต้นทาง; local package และ Rev.1–Rev.10 regression ผ่าน
- Live acceptance ใช้ GitHub repository ที่ผู้ใช้เตรียมจริง พร้อมบันทึก URL/tag/commit ที่ไม่ลับ; ไม่แทนหลักฐานนี้ด้วย mock tests
- Setup/portable CLI 0.1.11 ผ่าน installed workflow และมี checksums; build สำเร็จอย่างเดียวไม่ใช่ acceptance
- บันทึกผลและรายการที่ยังไม่รันใน `output/phase1-rev11-status.md` เมื่อเริ่มส่งมอบ โดยค่าเริ่มต้นให้ผู้ใช้ทดสอบเอง เว้นแต่ผู้ใช้อนุญาตให้ agent รัน

## 10. นอกขอบเขต

Central registry, package search/publish, authentication/private GitHub, semver ranges/automatic updates, package signing infrastructure, multi-version graph, `craft upgrade`, editor protocol/LSP และ process execution

ดู [local packages Rev.8](../../docs/REV8-PACKAGES.md), [Rev.10](craft-phase1-rev10.md) และ [ทิศทางภาษา](../craft-language-direction.md)
