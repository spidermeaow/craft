# Craft Phase 1 Rev.15 — GPU renderer และ text path ขั้นต้น

วันที่: 6 ตุลาคม 2026  
ฐาน: Craft 0.1.14  
เป้าหมายรุ่น: Craft 0.1.15  
สถานะ: implementation และ Windows smoke test ขั้นต้นแล้ว; ดู [สถานะ Rev.15](output/phase1-rev15-status.md)

## เป้าหมาย

พัฒนา UI Engine ตาม [ทิศทาง Craft GUI](../craft-gui-direction.md) โดยให้ผู้สร้างแอปยังเขียนด้วย Craft เท่านั้น และเริ่มใช้ renderer แบบ GPU กับข้อความที่คมชัดบน Windows โดยไม่ใช้ WebView งานนี้เป็นฐานของประสบการณ์ที่สวยและลื่น ไม่ใช่คำประกาศว่า 60/120 FPS หรือ component framework ครบแล้ว

## ขอบเขตของรุ่น

1. เพิ่ม Direct2D HWND render target แบบ hardware และ DirectWrite text format ให้ UI tree เดิม โดยเก็บ factory, target, brush และ format ตลอดอายุหน้าต่างและคืนเมื่อปิด
2. รองรับ device loss ด้วยการสร้าง render target และ resource ที่ผูกกับ device ใหม่; เมื่อสร้าง hardware target ไม่ได้ ให้ GDI fallback เพื่อให้แอปยังเปิดได้
3. ให้ `gui.renderer(window)` รายงาน backend จริง (`direct2d-hardware` หรือ `gdi`) เพื่อใช้ตรวจผล ไม่ถือว่าเห็น UI แล้วแปลว่า GPU ทำงาน
4. แก้ปุ่มให้ activate เมื่อปล่อยเมาส์ในปุ่ม, ยกเลิกเมื่อปล่อยนอกปุ่ม, คืนสี hover เมื่อเมาส์ออก และรักษา keyboard focus หลัง tree update ที่ ID ปุ่มยังอยู่
5. ปรับ DPI ที่ Windows boundary และไม่วาดต่อเนื่องเมื่อแอปอยู่นิ่ง; tree ที่ไม่เปลี่ยนไม่ควรกระตุ้น relayout/paint ใหม่
6. รักษา source bundle, package และ GUI bridge เก่าให้ runtime รุ่นนี้ยังอ่านได้; GUI package 0.1.15 อยู่ใน CLI/Setup

## ลำดับตรวจรับ

| กรณี | ผลที่ต้องเห็น |
| --- | --- |
| GPU path | เครื่องที่รองรับแสดง `direct2d-hardware`; ปุ่มมุมโค้งและข้อความวาดผ่าน Direct2D/DirectWrite |
| Fallback | เมื่อ hardware target ใช้ไม่ได้ รายงาน `gdi` และแอปยังตอบสนอง |
| Device loss | หลังเปลี่ยน display/device target ถูกสร้างใหม่โดยไม่ค้างหรือ leak |
| Input | กดแล้วลากออกไม่ activate; click, Tab/Shift+Tab, Enter/Space และ focus หลัง update ทำงานตาม contract |
| DPI/text | ข้อความไทย/Unicode และ hit test ถูกตำแหน่งบน DPI 100%, 150%, 200% รวมถึงย้ายข้ามจอ |
| Idle/performance | เมื่อไม่มี event/animation ไม่ render วน; บันทึก CPU/GPU idle, RAM, เวลาเปิด, frame time ตาม workload และเครื่องที่ใช้ |
| Compatibility | `go test ./...`, source/bundle, CLI/API ที่ไม่ import GUI, การติดตั้ง/ถอนติดตั้ง และ checksum ไม่ถดถอย |

หลักฐานจริงต้องลงใน [สถานะ Rev.15](output/phase1-rev15-status.md) แยก compile, automated test, Windows visual test และ benchmark ไม่ใช้ build สำเร็จเป็นหลักฐานว่าภาพสวยหรือลื่น

หลักการตัดสินด้านประสิทธิภาพหลัง Rev.15: ให้ความไวจาก input ถึงภาพใหม่และความนิ่งของ frame มาก่อนการลด CPU/RAM/GPU หากการประหยัดทรัพยากรเพิ่ม latency ต้องปรับวิธีกำหนดเวลาและวัดซ้ำ ตัวเลข idle CPU ที่ดีขึ้นเพียงอย่างเดียวไม่ใช่หลักฐานว่า GUI ตอบสนองเร็วขึ้น ดู [ทิศทาง Craft GUI](../craft-gui-direction.md)

## งานที่ยังไม่รวม

Craft Office, visual designer แบบลากวาง, layout แบบ content/fill/min-max ที่สมบูรณ์, text measurement/editing/IME, ScrollView/virtualization, animation engine, blur/shadow เต็มรูปแบบ, accessibility เต็มรูปแบบ, Linux/macOS GUI backend และ standalone app executable งานเหล่านี้ยังเป็นเป้าหมายต่อเนื่องก่อนเรียก Craft GUI ว่าพร้อมสร้างแอปทั่วไป
