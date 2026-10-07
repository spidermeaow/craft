# Craft Phase 1 Rev.25 — แผนพัฒนา `packages/gui` 0.2.2

> แผนย่อยของ [Craft Phase 1 Rev.25](../craft-phase1-rev25.md). `0.2.2` คือเวอร์ชัน package; `craft version` ยังคงเป็น `0.1.25`.

วันที่: 7 ตุลาคม 2026  
สถานะ: **ลง source แล้ว; รอผู้ใช้ทดสอบ**  
ฐาน: `packages/gui` 0.2.1, Craft UI Runtime 0.1.25 ใน source และ `craft-test/gui-package-dashboard`

## เป้าหมาย

ยกระดับ `packages/gui` จาก wrapper และ composition ขั้นต้นให้เป็น library ที่ผู้พัฒนาใช้สร้าง native desktop UI ด้วย Craft ได้สะดวก โดยยังเห็น event/state/lifecycle ที่เกิดขึ้นจริง. ให้แอปเรียก `gui` เป็นทางหลักสำหรับความสามารถของ Runtime 0.1.25, สร้างหน้าจอที่ปรับตาม viewport, ใช้ typography/surface style และมี Canvas targets ที่กดด้วย pointer/keyboard ได้. งานนี้เป็น **หัวข้อ package ภายใน Phase 1 Rev.25**; หากพบ Runtime bug ให้บันทึกและแก้แยกตามขอบเขต ไม่ซ่อนด้วย widget ภาพจำลอง.

Craft GUI ยังคงเป็น native Windows, ไม่มี WebView, React, JSX, HTML/CSS หรือเครื่องมือออกแบบแบบลากวาง. ความไวในการตอบสนองและความเรียบง่ายของ Craft สำคัญกว่าขนาด package. 0.2.2 ไม่ใช่ Craft V1.00 และไม่สร้าง Craft Office หรือ ERP.

## หลักฐานและช่องว่าง ณ จุดเริ่ม

### ผลตรวจ `craft-test` ก่อน implement

ตรวจ source `D:\02-Repository\craft-test\gui-package-dashboard\src\main.craft` แล้วพบ **5 API ของ Runtime ที่แอปยังเรียกตรง**:

| API ที่เรียกตรง | จุดประสงค์ใน dashboard | งาน 0.2.2 ที่ต้องทำก่อน |
| --- | --- | --- |
| `std.ui.viewportWidth` | อ่านความกว้าง client area หลังเปิดหน้าต่างเพื่อเลือก compact/wide | เพิ่ม `gui.viewportWidth` และ `gui.viewportHeight` ด้วย signature ตรง Runtime |
| `std.ui.drawTextStyled` | วาดหัวข้อ/ป้ายบน compact Canvas ด้วยขนาดและน้ำหนักต่างกัน | เพิ่ม `gui.drawTextStyled` |
| `std.ui.hitRegion` | กำหนด target `target-docs`, `target-data`, `target-refresh` | เพิ่ม `gui.hitRegion` และบันทึกสัญญา ID/ลำดับ region |
| `std.ui.textStyle` | ตกแต่ง Text node ในส่วนตรวจ style | เพิ่ม `gui.textStyle` |
| `std.ui.surface` | ตกแต่ง native Button ในส่วนตรวจ style | เพิ่ม `gui.surface` |

`viewportHeight` ยังไม่ได้ถูกใช้ตรงใน dashboard แต่เป็นคู่ API ที่ควรเปิดพร้อม `viewportWidth` เพื่อไม่ให้ package มีช่องว่างด้าน viewport. การอ้างชนิด `UiWindow`, `UiNode`, `UiDraw`, `UiEvent` และ field เช่น `event.kind` เป็นชนิด/ข้อมูลที่ Craft Runtime เปิดให้ Craft ใช้ ไม่ใช่การเรียก `std.ui` ตรง. หลังเพิ่ม wrapper ให้เปลี่ยนห้ากลุ่มคำสั่งใน dashboard เป็น `gui.*` และตรวจด้วยการค้น source ว่า **ไม่มี `std.ui.` เหลือใน app source**. การเรียก `std.ui` ภายใน implementation ของ `packages/gui` ยังจำเป็นและถูกต้อง.

ผลตรวจนี้เป็นการอ่าน source เท่านั้น; ยังไม่ได้รัน dashboard หลัง 0.1.25. จึงให้การปิดช่องว่างทั้งห้าเป็น **ลำดับแรกและ gate ก่อนเพิ่ม component ใหม่**. หาก Craft checker หรือ GUI เปิดจริงพบข้อผิดพลาดของ Runtime ให้แยกบันทึก ไม่ดัดแปลง package เพื่อปิดบังผล.

| ด้าน | ที่มีใน source | ช่องว่างสำหรับ package |
| --- | --- | --- |
| API Runtime 0.1.25 | `viewportWidth/Height`, `textStyle`, `surface`, `drawTextStyled`, `hitRegion` และ event ของ Canvas targets | `packages/gui` 0.2.1 ยังไม่ export API เหล่านี้; dashboard ต้องเรียก `std.ui` ตรง |
| Layout | Row/Column/Panel/SplitView/ScrollView, sizing/weight และ scrollbar แนวนอนใน Runtime | แอปยังต้องอ่าน viewport และเลือก layout เอง; fixed Canvas ไม่ reflow อัตโนมัติ |
| Component | `card`, `toolbar`, `navigation`, `field` ประกอบ node ได้ | ไม่มี style/theme ที่ควบคุมภาพได้จริง, ไม่มี ID/event convention สำหรับ custom target, ไม่มี pattern สำหรับหน้าจอ adaptive |
| State/event | `nextEvent`, `update` และ node ID | แอปต้องตรวจ `event.kind`/`nodeId` และ rebuild tree เอง; ต้องมีตัวช่วยที่ลด boilerplate โดยไม่กลืน event หรือทำให้ update เกินจำเป็น |
| การตรวจรับ | 0.2.1 เคย `package check`/`craft check` และเปิด dashboard บน Direct2D ก่อนแก้ Runtime | 0.1.25 และ dashboard รุ่น compact **ยังรอผู้ใช้ทดสอบ**; ห้ามนำผลก่อนหน้าไปอ้างว่า API ใหม่ผ่านแล้ว |

## ขอบเขต 0.2.2

### 1. ความเข้ากันได้และ public API

ปรับ `packages/gui/craft.toml` เป็น 0.2.2 และกำหนด **minimum Runtime 0.1.25** ในเอกสาร. เพิ่ม wrapper ที่ signature ตรงกับ `std.ui` สำหรับ viewport, text/surface style, styled Canvas text และ hit region. ตรวจรายการ public signatures ของ Runtime เทียบกับ package โดยแยกสามกลุ่ม: มี wrapper, ตั้งใจให้เป็น Runtime-only และยังไม่พร้อมเปิดใช้; ห้ามอ้างว่า “ครอบคลุมทั้งหมด” จากจำนวนฟังก์ชันอย่างเดียว.

คง API 0.2.1 ที่ยังมีสัญญาเดิมไว้. การเปลี่ยนชื่อหรือพฤติกรรมที่หลีกเลี่ยงไม่ได้ต้องมี migration และตัวอย่างก่อน/หลัง. ตรวจ Craft syntax/type checker จริงก่อนยืนยันชื่อ helper ใหม่; ตัวอย่างในแผนเป็นแนวทาง ไม่ใช่คำสั่งที่ใช้ได้แล้ว.

### 2. Styling ที่ package ควบคุมได้จริง

เพิ่ม helper สำหรับสร้าง Text/Button/Panel ด้วย `textStyle` และ `surface` โดยรับค่าสี ขนาด น้ำหนัก ระยะห่าง/มุมที่ Runtime รองรับ. จัด palette และ spacing tokens เป็นค่าหรือฟังก์ชัน Craft ที่อ่านง่าย และเปิดให้แอป override ได้. `card`, `toolbar`, `navigation`, `field` ควรใช้ tokens สม่ำเสมอโดยไม่ฝังสีเฉพาะ dashboard.

กำหนด style states ของ custom Canvas target จาก **state ที่แอปส่งเข้ามา** เช่น normal/hover/focus/pressed เท่าที่ event จริงบอกได้; package ต้องไม่สมมติว่ามี disabled/accessibility state สมบูรณ์. TextInput ยังเป็น Windows EDIT; ไม่ทำ Canvas input ปลอมเพื่อให้หน้าตาสวยแต่เสีย IME/caret/selection. ไม่ประกาศ shadow, gradient, animation หรือ font family selection หาก Runtime ยังไม่มี.

### 3. Adaptive layout และ Canvas composition

เพิ่มรูปแบบ API ที่อ่าน viewport ตอนเปิดหน้าต่างและใช้ `Resized`/`DpiChanged` เพื่อเลือก layout variant หรือคำนวณขนาด Canvas อีกครั้งอย่างชัดเจน. ตัวช่วยต้องไม่สร้าง frame loop หรือ rebuild เมื่อขนาดไม่เปลี่ยน. ให้แอปกำหนด breakpoint เอง; package ไม่ผูกกับขนาดหน้าจอของ dashboard. กำหนดกติกา fixed content ที่ใหญ่กว่า viewport: ใช้ ScrollView/scrollbar แนวนอนหรือจัดใหม่ ไม่ clip แบบเงียบ.

ทำ composition ที่ใช้ซ้ำได้ เช่น app shell แบบ header/sidebar/content, section/card และ action row **เฉพาะส่วนที่ประกอบจาก node ปัจจุบันได้จริง**. ให้ node ID เป็นอินพุตหรือมี convention ที่ไม่ชนกัน และไม่สร้าง UI tree เกิน limit. หลีกเลี่ยง API declarative แบบ JSX/hooks ที่ภาษา Craft ยังไม่มี.

### 4. Canvas target และ event helpers

ให้แอปสร้างพื้นที่กดใน Canvas พร้อม ID ผ่าน `gui.hitRegion` และ helper ระดับสูงสำหรับ “ภาพปุ่ม/แท็บ + hit region” ที่คืน `UiDraw[]` หรือ composition ตามชนิดที่ Craft รองรับจริง. ประกาศ event ที่เกี่ยวข้องในคู่มือ: `CanvasTargetPressed`, pointer down/up/move/enter/leave, `FocusChanged`, `KeyPressed/Released`, `Resized`/`DpiChanged`. Helper สำหรับเปรียบเทียบ event ตาม `kind`/`nodeId` ทำได้ แต่ต้องไม่ consume event สำคัญหรือเปลี่ยนลำดับของ Runtime.

การคลิก target และการกด Enter/Space ต้องใช้ ID เดียวกัน. Canvas target เป็นพื้นที่โต้ตอบที่แอปต้องกำหนด label/ภาพเอง; ยังไม่ใช่ native Button และไม่มี accessibility tree. หากความต้องการ keyboard/focus ใช้งานไม่ได้จริงจาก Runtime ให้บันทึก bug แทนการประกาศ component พร้อมใช้.

### 5. การติดตั้งและตัวอย่าง

ใช้ระบบ dependency ของ Craft เป็นเส้นทางหลัก: คัดลอก package เข้าโปรเจกต์หรืออ้าง local path ใน repository, ระบุ version 0.2.2 และรัน `craft package lock`. **ยังไม่เปิด `craft gui init`** จนกว่าจะมีเหตุผลด้าน UX และสัญญาไม่เขียนทับโปรเจกต์ที่ชัดเจน. ไม่แก้ Runtime เพียงเพื่อเปิดคำสั่งนี้.

อัปเดต `examples/gui-package-lab` ให้ใช้ API ใหม่ผ่าน `gui` อย่างเดียวใน flow หลัก. ปรับ `craft-test/gui-package-dashboard` เป็น test application แยก โดยเปลี่ยนการเรียก primitive 0.1.25 จาก `std.ui` เป็น `gui`, เปรียบเทียบหน้า compact/wide, hit targets อย่างน้อย 3 จุด, native TextInput, scroll แนวนอน และ style. ตัวอย่างนี้ใช้ทดสอบ package ไม่ใช่ข้อกำหนดให้สร้างแอปจากภาพให้ครบทุก control; ตาราง/slider/editor ที่ยังเป็นภาพจำลองต้องติดป้ายให้ชัด.

## ลำดับดำเนินงาน

1. ปิด direct Runtime calls ทั้งห้ากลุ่มที่พบใน dashboard ด้วย wrapper 0.1.25 และ `viewportHeight` คู่กัน; คง 0.2.1 compatibility.
2. รับผลตรวจ 0.1.25 จากผู้ใช้ถ้ามี; หากยังไม่มีให้บันทึก “รอทดสอบ” และหลีกเลี่ยงการพึ่ง behavior ที่ยังไม่ยืนยันโดยไม่ระบุความเสี่ยง.
3. ทำ API inventory `std.ui` ↔ `gui` และตรวจ syntax/type ของ helper ที่เสนอ. ตัดสิน public surface ก่อนเพิ่ม component ใหม่.
4. ตามด้วย style tokens/compositions และ adaptive/event helpers ที่มีต้นทุนต่ำ.
5. อัปเดต example และ dashboard test โดยไม่เพิ่ม app-specific code เข้า Runtime หรือ package. ล็อก dependency/version ใหม่และเขียน migration จาก 0.2.1.
6. เขียน API reference, manual test guide, status และผลเปรียบเทียบภาพ/interaction/performance **ตามหลักฐานที่ทดสอบจริง**.

## เกณฑ์ตรวจรับ

| ด้าน | สิ่งที่ต้องพิสูจน์ |
| --- | --- |
| Package/compatibility | `gui@0.2.2` ติดตั้งผ่าน dependency/lock ได้; ตัวอย่าง 0.2.1 ยัง check/run ได้ หรือมี migration เฉพาะจุดที่อธิบายชัด |
| API | ตัวอย่างใหม่ import `gui "gui"` และไม่ต้องเรียก `std.ui` โดยตรงสำหรับ viewport/style/Canvas targets; Craft checker รับ signature จริง |
| Adaptive UI | เปิดที่สองขนาดหน้าต่างและ DPI 100%/125%; header/content สำคัญมองเห็นหรือเลื่อนถึงได้, resize แล้ว layout/Canvas ตรง viewport |
| Interaction | pointer และ Enter/Space ใช้ target ID เดียวกันอย่างน้อย 3 จุด; hover/focus/pressed state ตรง event; TextInput/IME, Button และ scroll เดิมไม่ถอยหลัง |
| Renderer | Direct2D และ GDI แสดง style, typography, clip และ hit areas ตามสัญญา; บันทึกความต่างที่ยังมี |
| Performance | เทียบ workload เดียวกันระหว่าง `std.ui` ตรงกับ `gui`: startup, layout/paint, input-to-end-draw, idle CPU/RAM และจำนวน rebuild; รายงานเครื่อง/DPI/renderer. ไม่อ้าง input-to-present หากยังวัดไม่ได้ |
| เอกสาร | มี quick start, API inventory, migration 0.2.1→0.2.2, ข้อจำกัด Windows/single-window/source bundle และขั้นตอนทดลองที่ทำซ้ำได้ |

ตามความต้องการปัจจุบัน ผู้ใช้เป็นผู้ทดสอบแอป/GUI เอง เว้นแต่จะสั่งให้ผู้ช่วยทดสอบในงาน implement นั้นโดยตรง. รายการ implementation และผลที่ยังรอตรวจรับอยู่ใน `../output/gui-0.2.2-status.md`; ยังไม่ได้รัน `craft check` หรือ GUI สำหรับ 0.2.2.

## งานที่ยังอยู่นอก 0.2.2

Native checkbox/radio/select/date picker, multiline text editor, table/grid virtualization 2D, popup/dialog, animation/compositor, reactive state อัตโนมัติ, accessibility ที่ครบ และการแจกจ่าย standalone binary ต้องมี Runtime primitive/การตรวจรับเพิ่ม. Rev. ถัดไปจึงค่อยนำ primitive ที่พร้อมมาทำ component ระดับสูง โดยไม่โยน application-specific logic ลง Runtime.
