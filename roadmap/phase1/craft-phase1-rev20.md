# Craft Phase 1 Rev.20 — Virtualized List รุ่นแรก

วันที่: 7 ตุลาคม 2026  
ฐาน: Craft 0.1.19  
เป้าหมายรุ่น: Craft 0.1.20  
สถานะ: implement ใน source แล้ว; รอผู้ใช้ทดสอบ GUI จริง

## เหตุผล

`ScrollView` ใน 0.1.18–0.1.19 clip และข้ามการวาด node นอก viewport ได้ แต่แอปยังต้องสร้างทุก node และ engine ยังต้อง validate, clone, measure และ layout ทั้ง tree. กริด 32×32 มี 1,024 ปุ่มจริง และ UI tree ปัจจุบันจำกัด 4,096 node. วิธีนี้ยังไม่เหมาะกับรายการระดับหมื่นหรือแสนแถว แม้การเลื่อนส่วนที่มองเห็นจะเร็วขึ้นแล้วก็ตาม.

Rev.19 แก้กฎขนาดของ viewport/content แล้ว จึงเป็นฐานสำหรับรายการที่มีความสูงรวมมาก แต่สร้าง node เพียงช่วงที่มองเห็น. ก่อนเริ่ม implement ให้ใช้ผลทดสอบ `test-gui` ของผู้ใช้ตรวจว่า Rev.19 แสดงภาพ, scroll, hit test และ resize ถูกต้อง; หากพบข้อผิดพลาดที่กระทบฐานนี้ ให้แก้ก่อน.

## เป้าหมาย

สร้าง **Virtualized List แนวตั้งที่ทุกแถวสูงเท่ากัน** เป็น component แรกสำหรับข้อมูลจำนวนมาก. Craft application ยังคงกำหนดข้อมูลและหน้าตาของแถวด้วย Craft; engine ดูแล viewport, ช่วง index ที่ต้องใช้, offset, clipping, layout และ hit test. เมื่อรายการมี 100,000 แถว จำนวน `UiNode` ที่สร้างจริงต้องขึ้นกับจำนวนแถวใน viewport และ overscan ไม่ขึ้นตามจำนวนข้อมูลทั้งหมด.

ลำดับความสำคัญคือความถูกต้องของภาพและการตอบสนองต่อ input ก่อนลด RAM หรือขนาด binary. ไม่เพิ่ม frame loop ขณะหน้าต่างอยู่นิ่ง.

## ขอบเขต implementation

1. **กำหนดสัญญา API ก่อนลงมือ:** เพิ่ม API ใน GUI package/`std.ui` สำหรับ `virtualList` ที่ระบุ ID, จำนวนรายการ, ความสูงแถวใน logical pixels และชุด `UiNode` เฉพาะช่วงที่ร้องขอ. ระบุการส่ง visible range จาก engine ไปยัง Craft ผ่าน `UiEvent` และการ update ช่วงกลับเข้า engine ให้ชัด พร้อมตัวอย่าง Craft ที่ `craft check` ได้จริง. ห้ามอ้าง callback/lambda หรือ syntax ที่ Craft ยังไม่รองรับ. ให้ index เป็นแบบเริ่มที่ 0 และนิยามช่วง `[first, end)` เพื่อเลี่ยงความกำกวม.
2. **Windowing:** คำนวณ visible range จาก offset และ viewport พร้อม overscan ขนาดจำกัด; สร้าง/clone/measure/layout เฉพาะแถวในช่วงนั้น. พื้นที่ scroll ทั้งหมดต้องสะท้อน `itemCount × rowHeight` โดยไม่สร้าง node placeholder ทุกแถว. ตรวจ integer overflow และขีดจำกัดจำนวนรายการ.
3. **Event และ identity:** เมื่อ range เปลี่ยน ส่ง event ที่มี list ID และช่วง index; coalesce range event ที่ล้าสมัยได้ แต่ห้ามทำ click, key หรือ close หาย. รักษา stable item ID/index, selection และ hit test ให้ตรงกับแถวที่เห็นหลัง wheel, PageDown, Home/End, resize และ update. กำหนดพฤติกรรมช่วงที่ข้อมูลยังไม่ส่งมาอย่างชัดเจน ไม่แสดงแถวเก่าผิด index.
4. **การเปลี่ยนข้อมูล:** เมื่อ item count เพิ่ม/ลด ให้ clamp offset และช่วง index; เมื่อข้อความ/สถานะของแถวที่เห็นเปลี่ยน ให้ update เฉพาะส่วนที่กระทบ. รองรับรายการว่างและรายการสั้นกว่า viewport. การสลับ source/list ID ต้องไม่เอา offset หรือ node เก่ามาปะปน.
5. **Native paths และ diagnostics:** ใช้ได้กับ Direct2D/DirectWrite และ GDI fallback โดย clip/dirty region/DPI ตรงกัน. เพิ่ม trace เช่น visible range, จำนวนแถวที่สร้าง, จำนวน range request, เวลา input ถึงจบ draw และจำนวน frame ที่เกิน budget; ระบุชัดว่า draw completion ไม่ใช่เวลาภาพปรากฏจริงบนจอ.
6. **ตัวอย่างและ distribution:** เพิ่มโหมด Virtual List 100,000 แถวใน `examples/gui-stress` และ `D:\02-Repository\test-gui` พร้อมปุ่ม jump/resize data/update visible item เท่าที่ API รองรับ. อัปเดต GUI package/template, เอกสาร, version/bundle metadata และ Windows installer เมื่อ implement จริง. `craft build` ยังคงเป็น source bundle ตามสถานะปัจจุบัน.

## เกณฑ์ตรวจรับ

| กรณี | ผลที่ต้องเห็น |
| --- | --- |
| 0, 1 และรายการสั้น | ไม่มีช่องหลอกหรือ crash; แถวไม่ยืดเต็ม viewport; offset อยู่ในขอบ |
| 100,000 แถว | ไปต้น/กลาง/ท้ายด้วย wheel, keyboard หรือ jump แล้วเลข index และภาพถูกต้อง; จำนวน UI node/แถวที่สร้างขึ้นกับ viewport + overscan ไม่ขึ้นตาม 100,000 |
| Scroll และ click | ไม่มีแถวเก่าค้างหรือแถวซ้ำ; คลิกหลังเลื่อนได้ index ตรงภาพ; input event สำคัญไม่หาย |
| Update และ resize | อัปเดตแถวที่เห็น/จำนวนรายการแล้วภาพถูกต้อง; offset/range clamp เมื่อรายการหด; DPI และ resize ไม่ทำให้ hit test ผิด |
| Responsiveness | เก็บ baseline 0.1.19 และผล 0.1.20 บนเครื่อง, backend, viewport และ workload เดียวกัน; รายงาน p50/p95/p99 ของ input-to-draw, frame time, RAM, จำนวน node และจังหวะกระตุก โดยไม่สรุปว่า “เร็วกว่า” จากค่าหนึ่งค่า |
| Compatibility | `ScrollView` และ GUI ตัวอย่างเดิมยังทำงาน; `craft check/run/build`, package/template และ installer metadata ตรงรุ่น |

ทดสอบเชิงโครงสร้างด้วย unit/integration tests สำหรับ range math, overscan, overflow, ID, clamp, event ordering และ hit test. การตรวจภาพจริง, DPI, GDI fallback และความรู้สึกลื่นให้ผู้ใช้ทดสอบตามความต้องการเดิม เว้นแต่ผู้ใช้ขอให้รันทดสอบในงาน implementation นั้นโดยตรง.

## ลำดับทำงาน

1. รับผลทดสอบ Rev.19 จากผู้ใช้และบันทึก baseline ของ Stress Lab 32×32 กับรายการจำลอง 100,000 แถว.
2. ออกแบบ API และ lifecycle ของ range event; เขียนตัวอย่าง Craft ที่ type-check ได้ก่อนเพิ่ม native implementation.
3. ทำ range math, logical content extent และการ reuse แถวที่ยังอยู่ในช่วง; เพิ่ม unit tests ของกรณีขอบ.
4. เชื่อม input/update/paint กับ Direct2D และ GDI; ตรวจ hit test และ event ordering ผ่าน integration tests.
5. เพิ่มโหมดใน test-gui, เอกสารการใช้และข้อจำกัด; เทียบ trace บน workload เดียวกัน แล้วแก้ regression ก่อนส่งให้ผู้ใช้ตรวจรับ.

## นอกขอบเขต 0.1.20

Virtualized Grid/Table/Tree, แถวสูงไม่เท่ากัน, scrollbar ที่ลากได้, momentum/touch animation, TextBox/IME, styling/animation API, Craft Office และ native standalone application binary. สิ่งเหล่านี้ต้องไม่ถูกสื่อว่าเสร็จจาก Virtualized List รุ่นแรก.
