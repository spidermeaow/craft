# Craft Phase 1 Rev.12 — ความเสถียรของ GitHub source packages

วันที่: 6 ตุลาคม 2026  
ฐาน: Craft 0.1.11 / Phase 1 Rev.11  
เป้าหมายรุ่น: Craft 0.1.12  
สถานะ: ดำเนินการแล้ว ดูผลและขอบเขตการตรวจรับใน [สถานะ Rev.12](output/phase1-rev12-status.md)

## 1. เป้าหมาย

ทำให้ workflow ติดตั้งและจัดการ Craft library จาก **public GitHub** ที่เริ่มใน Rev.11 ใช้งานได้อย่างคาดเดาได้ ปลอดภัยต่อข้อมูลโปรเจกต์ กู้จากข้อผิดพลาดได้ และมีข้อความแนะนำที่ผู้ใช้เข้าใจ โดยยังคงรูปแบบ source package, exact tag และ lockfile เดิม

Rev.12 เป็นรอบ hardening และตรวจรับ ไม่ถือว่าฟีเจอร์ Rev.11 ทุกกรณีผ่านเพียงเพราะมี implementation หรือเคยผ่าน fixture test แยกผลของ unit/integration tests, live GitHub workflow, portable CLI และ Windows installer ออกจากกัน

การขยับเลข Phase หรือ version ในแผนนี้ไม่ได้หมายถึง Craft 1.00; ผู้ใช้เป็นผู้ตัดสินใจว่าจะปล่อย 1.00 เมื่อใด

## 2. ฐานที่มีแล้วและช่องว่างที่ต้องปิด

Rev.11 มี `craft install`, `craft package check/list/lock`, exact-tag resolution, lock format 2, checksum cache, offline restore, transitive dependencies, source bundle และ journal/recovery แล้ว ดู [สัญญาการใช้งาน](../../docs/REV11-PACKAGES.md) และ [ผล Rev.11](output/phase1-rev11-status.md)

ช่องว่างที่ต้องตรวจและแก้ใน Rev.12:

- Windows Setup 0.1.11 ถูกสร้างแล้ว แต่ยังไม่มีผล **ติดตั้ง อัปเกรด ถอน และเรียก CLI ผ่าน PATH** จากการตรวจรับโดยผู้ใช้
- เมื่อ network, GitHub API, cache หรือ project files มีปัญหา ผู้ใช้ต้องแยกสาเหตุและทำขั้นตอนกู้คืนได้โดยไม่ต้องแก้ไฟล์ภายใน cache เอง
- ต้องพิสูจน์ว่า lock และ bundle ให้ผลซ้ำได้บนเครื่องหรือ cache ใหม่ รวมถึงการเปลี่ยน version ของ dependency โดยตั้งใจ
- ต้องป้องกัน regression ของ local packages, การ import/export และโปรเจกต์ที่ไม่ใช้ GitHub packages

## 3. ขอบเขตงาน

### 3.1 สัญญา CLI และ diagnostics

- ตรวจและทำให้ exit code, stdout/stderr และข้อความ error ของ `install`, `install --offline`, `install --refresh`, `install --recover`, `package check/list/lock` สม่ำเสมอ
- Error ต้องระบุ package/alias/tag และขั้นตอนที่ล้มเหลว พร้อมคำสั่งแก้ไขที่ใช้ได้จริง โดยไม่พิมพ์ secret, cache internals ที่ไม่จำเป็น หรือข้อมูล response ที่ควบคุมโดยภายนอกโดยตรง
- แยกกรณีให้ชัด: manifest ผิด, lock ไม่ตรง, tag ไม่พบ, rate limit, network timeout, cache ขาด, cache เสีย, version conflict, import ไม่ public, journal ค้าง และ recovery ถูกปฏิเสธเพราะไฟล์ถูกแก้ภายหลัง
- คำสั่งตรวจ (`package check/list`, `check`, `run`, `test`, `build`) ต้องไม่ดาวน์โหลดหรือแก้ manifest/lock โดยอ้อม
- `install` ซ้ำเมื่อ dependency ตรง lock และ cache ถูกต้องต้องไม่เปลี่ยนไฟล์โปรเจกต์โดยไม่จำเป็น; ระบุผลว่าใช้ pin เดิมหรือมีการเปลี่ยนแปลง

### 3.2 ความทนทานของ download และ cache

- ตรวจ timeout/cancellation และการจัดการ HTTP failures ของ GitHub API กับ archive download ให้ครบ รวมถึง rate limit, 404, truncated response และ redirect ที่ไม่อนุญาต
- ถ้าจะ retry ให้ใช้เฉพาะความล้มเหลวชั่วคราวที่นิยามไว้ มีจำนวนครั้งและเวลารอจำกัด เคารพ cancellation และไม่ retry validation/security failure; บันทึกพฤติกรรมจริงในคู่มือ
- ตรวจ checksum และ source boundaries ทุกครั้งก่อนใช้ cache; ห้ามใช้ข้อมูลที่ตรวจไม่ผ่านเพียงเพราะมีไฟล์อยู่แล้ว
- กำหนด **วิธีกู้ cache ที่เสียอย่าง explicit และปลอดภัย** โดยไม่ลบ cache ทั้งก้อนหรือข้อมูลของโปรเจกต์อื่นอัตโนมัติ หากเพิ่มคำสั่ง repair ให้แสดง target และผลที่เปลี่ยนอย่างชัดเจน
- คงเพดานขนาด archive, file count, graph และ redirect ของ Rev.11 เว้นแต่มีหลักฐานว่าต้องปรับ; ทุกการปรับต้องมีเหตุผลและ regression coverage

### 3.3 ความถูกต้องของ manifest, lock และ recovery

- Install, refresh และ lock ต้องคง comments/entries ที่ไม่เกี่ยวข้องใน `craft.toml` และไม่ทำ lock เปลี่ยนโดยไม่มีเหตุผล
- การเขียน manifest/lock ต้องคง invariants ของ journal: สำเร็จทั้งคู่ หรือปฏิเสธ/กู้คืนอย่างชัดเจน; ห้ามเลือก dependency ชุดใหม่เงียบ ๆ หลัง partial failure
- ตรวจ crash/interruption points ก่อนและหลัง staging, cache promotion, manifest replacement, lock replacement และ cleanup
- Recovery ต้องไม่เขียนทับไฟล์ที่ผู้ใช้แก้หลัง interruption; กรณี journal เสียหรือผลลัพธ์กำกวมให้หยุดพร้อมขั้นตอนตรวจด้วยมือ
- กำหนด compatibility ของ lock format 1/2 และ bundle รุ่นก่อน รวมถึงผลเมื่อ runtime เก่าอ่าน bundle ใหม่; เปลี่ยน schema เฉพาะเมื่อจำเป็นและมี migration ที่ทดสอบได้

### 3.4 Workflow ที่ผู้ใช้ทำจริง

- ติดตั้ง library จาก public GitHub ด้วย exact tag, import, `check/test/run/build`, แล้วรัน bundle โดยไม่มี cache/network ต้นทาง
- Restore โปรเจกต์จาก manifest/lock บน cache ว่าง และ restore แบบ offline บน cache ที่ครบ โดยต้องได้ commit/checksum เดิม
- เปลี่ยน tag/version โดย explicit; ตรวจว่ารุ่นเก่าในอีกโปรเจกต์ไม่ถูกเปลี่ยน และ refresh tag เดิมไม่เกิดโดยอ้อม
- ตรวจ transitive dependencies, local dependency ภายใน remote package, alias/identity/version conflicts และ public export diagnostics
- ทดสอบ Windows portable CLI และ installer workflow แยกกัน รวมถึงติดตั้ง/อัปเกรด/ถอนและการเรียก `craft` จาก terminal ใหม่

## 4. Milestones

1. **M0 — Baseline และ contract:** รวบรวม failure cases จาก Rev.11, ระบุข้อความ/exit code/ผลต่อไฟล์สำหรับทุกคำสั่ง และกำหนด acceptance matrix ก่อนแก้ implementation
2. **M1 — CLI และ recovery:** ปรับ diagnostics, idempotence, journal handling และแนวทางกู้ cache เฉพาะรายการที่เสีย
3. **M2 — Transport และ integrity:** ปรับ bounded retry หากจำเป็น, cancellation, archive validation และ cache verification โดยไม่เปลี่ยน trust model
4. **M3 — Compatibility และ regression:** ตรวจ local/GitHub graphs, lock/bundle รุ่นเดิม, formatting และ behavior ของโปรเจกต์ที่ไม่ใช้ package
5. **M4 — Distribution acceptance:** จัดทำคู่มือทดลอง Rev.12, สร้าง artifacts และบันทึกผล portable CLI, live GitHub และ Windows installer ตามหลักฐานที่รันจริง

## 5. เกณฑ์รับงาน

- ทุก failure case ใน acceptance matrix มีผลที่คาดหวังสำหรับ exit code, diagnostics และสถานะ manifest/lock/cache; ไม่มีกรณีที่ใช้ dependency ผิดชุดโดยไม่แจ้ง
- การติดตั้งซ้ำและ restore จาก lock ให้ commit/checksum เดิม; offline ไม่เรียก network; cache เสียถูกปฏิเสธและมีขั้นตอนกู้ที่จำกัดผลกระทบ
- Network interruption, invalid archive และ interrupted project update ไม่ทำให้ไฟล์โปรเจกต์เปลี่ยนเป็นชุดครึ่ง ๆ กลาง ๆ หรือทำให้ cache ที่ยังไม่ตรวจผ่านถูกใช้งาน
- Bundle ที่สร้างจาก locked graph รันได้โดยไม่ใช้ cache/network ต้นทาง; local packages และโปรเจกต์รุ่นก่อนยังทำงานตาม compatibility contract
- มีผลตรวจ live public GitHub อย่างน้อยหนึ่ง repository/tag ที่มีอยู่จริง โดยบันทึก tag, commit, checksum และขั้นตอนที่ทำ; fixture test ใช้แทนผลนี้ไม่ได้
- มีผลตรวจ portable CLI และ Windows installer ตามรายการใน 3.4 หรือบันทึกชัดว่า **ยังไม่ผ่านการตรวจรับ** ห้ามสรุปจาก build สำเร็จอย่างเดียว
- บันทึกผลที่รันจริง ข้อจำกัด และรายการที่ยังไม่รันใน `output/phase1-rev12-status.md` เมื่อเริ่มส่งมอบ การทดสอบสำหรับงาน implementation ให้ผู้ใช้เป็นผู้ดำเนินการ เว้นแต่ผู้ใช้สั่งให้ agent ทดสอบงานนั้นโดยตรง

## 6. นอกขอบเขต

Private GitHub/authentication, central registry, package search/publish, semver ranges, automatic upgrades, package signing infrastructure, native/binary libraries, multi-version graph, Linux/macOS release acceptance, GUI engine และ native application executable

Rev.12 ไม่เปลี่ยน Craft ให้เป็นภาษาเฉพาะสำหรับ package management และไม่เริ่มพัฒนา Craft Office
