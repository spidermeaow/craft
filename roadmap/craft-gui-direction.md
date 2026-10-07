# Craft UI Engine — เป้าหมายและทิศทางผลิตภัณฑ์

วันที่: 6 ตุลาคม 2026  
สถานะ: แนวทางระยะยาว ไม่ใช่รายการความสามารถที่เสร็จแล้ว

นิยามปัจจุบันอยู่ใน [Craft Phase 1](workstreams.md): UI Engine ระดับ native อยู่ใน Runtime ส่วน GUI library/component ที่ใช้ซ้ำได้อยู่ใน `packages/gui`; ทั้งคู่เป็นหัวข้อพัฒนาของ Phase เดียวกัน. Craft CLI/Runtime `0.1.25` และ GUI package `0.2.2` มีเลขเวอร์ชันแยกเพื่อจัดการ dependency แต่ไม่ใช่คนละ Phase.

ใน Phase 2 ที่วางไว้ จะใช้แอปจริงตรวจและปรับ responsiveness, frame stability, accessibility และการแจกจ่ายของ GUI. อ่าน [นิยาม Phase](craft-phases.md) สำหรับขอบเขตล่าสุด.

## Goal

**Craft UI Engine** คือ Native GUI Framework สำหรับภาษา Craft ที่มุ่งสร้าง desktop application ซึ่ง **เร็ว ลื่น ตอบสนองทันที และให้ประสบการณ์ใช้งานที่ดี** โดยผู้พัฒนายังเขียนโค้ดที่เรียบง่าย อ่านง่าย และไม่ซับซ้อนเกินจำเป็น

> **Performance first. Simplicity always.**

Craft UI Engine ไม่ได้มุ่งเป็น GUI framework ที่เบาที่สุด ให้ความเร็วและความลื่นมาก่อน แล้วลดการใช้ทรัพยากรเฉพาะส่วนที่ทำได้โดยไม่ทำให้การตอบสนองแย่ลง

ผู้พัฒนาเขียน desktop application ด้วย **ภาษา Craft เพียงภาษาเดียว** โดยใช้รูปแบบการเขียน GUI ที่อ่านง่าย เช่น สร้างหน้าต่าง วางองค์ประกอบ กำหนดรูปลักษณ์ และตอบสนองต่อ event แอปที่ได้ควรมีหน้าตาสวย ตอบสนองเร็ว เลื่อนและเคลื่อนไหวลื่นเหมือนแอปสมัยใหม่ โดยทำงานแบบ native และไม่ใช้ WebView

แรงบันดาลใจจากแอปที่สร้างด้วย React หมายถึง **คุณภาพของประสบการณ์ใช้งาน** ไม่ใช่การนำ React, JSX, hooks, HTML, CSS หรือ JavaScript มาเป็นวิธีเขียน Craft GUI ไม่มีแผนทำเครื่องมือออกแบบหน้าจอแบบลากวาง

```text
Craft application (UI, state, event handling)
  → GUI library ที่เขียนด้วย Craft
  → native UI engine / runtime ที่พัฒนาด้วย Go
  → renderer และ API ของระบบปฏิบัติการ
  → หน้าต่าง desktop แบบ native
```

GUI เป็น library ที่โปรเจกต์เลือกใช้ได้ Craft ยังเป็นภาษาทั่วไปสำหรับ CLI, API และงานอื่น ส่วนที่ต้องติดต่อหน้าต่าง, input, font หรือ GPU อยู่หลัง native boundary ผู้สร้างแอปไม่ต้องเขียน Go

## Core goals

1. **Fast:** เปิดแอป วาด UI คำนวณ layout และเปลี่ยนหน้า/state ให้เร็ว ลดงานระหว่าง render ที่ไม่จำเป็น และรองรับ UI ซับซ้อนโดยไม่ให้ความเร็วตกอย่างรุนแรง
2. **Responsive:** Click, scroll, drag, resize, typing, navigation และการใช้หน้าต่างต้องตอบสนองทันที งาน UI ทั่วไปไม่ควรทำให้ผู้ใช้รู้สึกว่าแอปกำลังรอคิด
3. **Smooth:** Scrolling, animation, transitions และการ resize ต้องลื่น โดย frame time สม่ำเสมอ ความลื่นเป็นส่วนหนึ่งของ performance
4. **Native performance:** ให้ Craft application ผ่าน UI engine ไปยัง native renderer และ platform โดยตรงในระดับที่เหมาะสม หลีกเลี่ยง layer หรือ runtime ขนาดใหญ่ที่เพิ่ม startup time, memory หรือ latency โดยไม่จำเป็น และไม่พึ่ง browser engine เพื่อแสดง GUI
5. **Efficient, not just lightweight:** ใช้ CPU, RAM และ GPU อย่างมีเหตุผล แต่ไม่ลดการใช้ทรัพยากรจนแอปกระตุก หากการใช้ RAM เพิ่มขึ้นอย่างสมเหตุสมผลช่วยให้แอปลื่นขึ้น ให้เลือกความลื่น
6. **Simple developer experience:** Syntax อ่านง่าย, boilerplate น้อย, state และ event เข้าใจง่าย, สร้าง component ได้ง่าย ผู้พัฒนาไม่ต้องจัดการ low-level UI หรือเข้าใจ renderer ก่อนสร้างแอปได้ Engine รับความซับซ้อนด้าน performance ไว้ภายใน

ลำดับตัดสินเมื่อเกิด trade-off:

```text
1. Responsiveness
2. Runtime performance
3. Smoothness
4. Developer experience
5. Resource efficiency
6. Binary size / lightweight
```

ลำดับนี้ใช้กับงานที่ผู้ใช้สัมผัสได้จริง การลดขนาด binary หรือ memory ต้องไม่แลกกับเวลาตอบสนองที่แย่ลง และ engine ภายในซับซ้อนได้เมื่อ API ที่ผู้พัฒนาใช้ยังเรียบง่าย

ตัวอย่างแนวคิดของ Developer API (ยังไม่ใช่ syntax หรือ API ที่ Craft รุ่นปัจจุบันรองรับ):

```craft
import craft.ui

func main() {
    let window = Window(title: "Craft App", width: 1200, height: 800)
    window.content = Column {
        Text("Hello Craft")
        Button("Click me") { print("Hello!") }
    }
    window.show()
}
```

## นำเป้าหมายไปใช้ในการออกแบบ

เป้าหมายคือให้ผู้ใช้รู้สึกว่า Craft GUI ตอบสนองได้รวดเร็วเทียบกับแอปสมัยใหม่ที่สร้างด้วย React เมื่อใช้เครื่อง งาน และ UI ที่เทียบกันได้ ผู้ใช้ยอมให้แอปใช้ CPU, RAM หรือ GPU มากขึ้น หากช่วยลดเวลาจากการกดหรือคลิกจนเห็นผล และช่วยให้การพิมพ์ เลื่อน และ animation ไม่สะดุด คำว่า “เร็วเท่า React” เป็นเป้าหมายด้านประสบการณ์ ไม่ใช่ตัวเลขหรือคำรับประกัน เพราะความเร็วขึ้นกับ workload และเครื่องที่ใช้

เมื่อการลดการใช้ทรัพยากรทำให้ input รอ, frame มาช้า หรือเกิดอาการหน่วง ให้เลือกความไวและความต่อเนื่องของภาพก่อน แล้วค่อยหาวิธีลดต้นทุนโดยไม่ทำให้ประสบการณ์แย่ลง การปรับช่วง polling, caching, batching หรือ scheduling ต้องวัด latency ก่อนและหลังเสมอ ห้ามสรุปว่าการใช้ CPU ต่ำลงแปลว่าตอบสนองเร็วขึ้น

## หลักการออกแบบประสิทธิภาพ

- **เปิดเร็วและใช้หน่วยความจำพอดี:** เริ่มเฉพาะระบบที่แอปใช้ โหลดทรัพยากรเมื่อจำเป็น และคืน font, image, GPU object และ OS handle ตาม lifecycle ที่ชัดเจน
- **อยู่นิ่งแล้วประหยัด:** render ตามการเปลี่ยนแปลงและตาม animation ที่กำลังทำงาน ไม่มี frame loop ต่อเนื่องเมื่อหน้าจอไม่เปลี่ยน
- **ตอบสนองทันที:** งานของแอปที่ใช้เวลานานต้องไม่ปิดกั้น event loop; input, layout และ paint มีขอบเขตงานที่วัดได้ในแต่ละ frame
- **เปลี่ยนเฉพาะส่วนที่กระทบ:** ติดตาม dirty state, เลี่ยง layout/paint/GPU upload ซ้ำ, cache ทรัพยากรที่ใช้ซ้ำ และ coalesce event ที่ทำได้โดยไม่ทำ click/key/close หาย
- **รองรับข้อมูลจำนวนมาก:** List, Table, Grid และ Tree ต้องออกแบบให้สร้าง วัด และวาดเฉพาะบริเวณที่มองเห็นเมื่อถึงขั้นพัฒนา component เหล่านั้น
- **สวยโดยควบคุมต้นทุน:** typography, สี, มุมโค้ง, เงา, blur และ animation ต้องมีลำดับการวาดที่ชัด Effect ที่แพงควรมีขอบเขต, cache หรือระดับคุณภาพที่ปรับตามอุปกรณ์ และเคารพการตั้งค่าลดการเคลื่อนไหว
- **ความคมชัดและความถูกต้อง:** logical pixels, DPI, text shaping, clipping, focus, accessibility และ input ต้องทำงานสอดคล้องกันบนจอและอุปกรณ์ที่รองรับ
- **ใช้ GPU เมื่อคุ้มค่า:** renderer เป้าหมายใช้ GPU สำหรับงานวาดและ animation ที่เหมาะสม โดยมีการจัดการ device loss และ fallback ที่ระบุชัด การมี GPU ไม่ใช่ข้ออ้างให้ใช้ CPU, RAM หรือ GPU มากตอนแอปอยู่นิ่ง

ขนาดไฟล์ที่เล็กเพียงอย่างเดียวไม่พอจะเรียกว่า “เบา” และภาพหนึ่งภาพที่ดูสวยไม่พอจะเรียกว่า “ลื่น” ต้องประเมินการเปิดแอป การใช้งานจริง และช่วงที่แอปไม่ได้ทำงานร่วมกัน โดยให้ผลตอบสนองขณะใช้งานมีน้ำหนักสูงกว่าตัวเลข idle CPU/RAM

## วิธีพิสูจน์ก่อนอ้างผล

ทุก revision ที่เพิ่มความสามารถ GUI ต้องบันทึก baseline และผลหลังเปลี่ยนบนเครื่องและสถานการณ์ที่ระบุ โดยอย่างน้อยวัด:

| ด้าน | สิ่งที่วัด |
| --- | --- |
| การเริ่มต้น | เวลาเปิดจนหน้าต่างใช้งานได้, ขนาดไฟล์แจกจ่าย |
| ทรัพยากร | RAM ตอนเปิด/หลังใช้งาน, CPU/GPU ตอนอยู่นิ่ง, จำนวน OS/GPU resources หลังเปิดและปิดซ้ำ |
| การตอบสนอง (เกณฑ์หลัก) | เวลาจาก input ถึงภาพใหม่ปรากฏบนจอ ทั้งค่าทั่วไปและช่วงช้า เช่น p95/p99; เวลาใน input queue, event ที่ตกหล่น และอาการค้างเมื่อมีงานยาว |
| การวาด (เกณฑ์หลัก) | frame time, frame ที่เกินเวลาของจอ, อาการสะดุดระหว่าง scroll/animation, จำนวน layout/paint pass และพื้นที่ที่วาดซ้ำ |
| ข้อมูลมาก | เวลาและหน่วยความจำของการเลื่อนรายการ/ตารางขนาดใหญ่ พร้อมจำนวน element ที่ถูกสร้างจริง |
| คุณภาพภาพ | ความคมชัดบนหลาย DPI, text/fallback font, clipping, animation และ accessibility |

ทดสอบทั้งเครื่องระดับทั่วไปและเครื่องที่มีจอ refresh rate สูงเมื่อมีอุปกรณ์ การตั้งเป้า 60/120 Hz เป็น **เป้าหมายของสถานการณ์ที่ hardware รองรับ** ไม่ใช่คำรับประกันของทุกแอปหรือทุกเครื่อง ให้กำหนดงบและเกณฑ์ผ่านของแต่ละ scenario จาก baseline ก่อนประกาศผล และรายงานค่าเฉลี่ยควบคู่กับช่วงที่ช้าที่สุดอย่างเหมาะสม หากจะเทียบกับ React ต้องใช้แอปตัวอย่าง งาน ข้อมูล เครื่อง และวิธีวัดเดียวกัน แล้วระบุผลตามสถานการณ์จริง

ตัวอย่างตรวจรับควรเป็นโปรแกรม GUI ขนาดเล็กและ workload ที่สร้างซ้ำได้ เช่น เปิดหน้าต่าง, คลิกปุ่ม, resize, scroll, animation, รายการจำนวนมาก และเปิด/ปิดซ้ำ ไม่ใช้ Craft Office เป็นตัวอย่างหรือเงื่อนไขในช่วงนี้

## สถานะปัจจุบัน

ณ Craft Runtime `0.1.25` มี viewport query, text/surface styling และ Canvas hit regions เพิ่มจาก renderer, layout, scroll, virtual list และ TextInput รุ่นก่อน. `packages/gui` `0.2.2` ห่อ API เหล่านี้และเพิ่ม composition ที่เขียนด้วย Craft; ผลตรวจรับ GUI ของรุ่นนี้ยังรอผู้ใช้ทดสอบ. รายละเอียดอยู่ใน [Runtime status](phase1/output/phase1-rev25-status.md) และ [GUI package status](phase1/output/gui-0.2.2-status.md). ย่อหน้าถัดไปเป็นลำดับประวัติการพัฒนา ไม่ใช่รายการขอบเขตปัจจุบันทั้งหมด.

Craft 0.1.13 พิสูจน์ bridge ระหว่าง Craft กับ Windows host และ 0.1.14 เพิ่ม UI tree/layout ขั้นต้น Rev.15 เริ่มมีทางวาดด้วย Direct2D hardware และ DirectWrite พร้อม GDI fallback ตาม [สถานะ Rev.15](phase1/output/phase1-rev15-status.md) และตัวอย่างรายงานว่าใช้ hardware renderer ได้บนเครื่องที่ทดลอง แต่ยังไม่มีการวัด input-to-present latency หรือเทียบกับแอป React ใน workload เดียวกัน API แนวคิดด้านบน, text/layout engine ที่สมบูรณ์, virtualization และ animation ยังเป็นงานต่อไป เอกสารนี้เป็นเกณฑ์ชี้ทิศทาง ไม่ใช่ประกาศว่าความสามารถเหล่านั้นมีครบแล้ว

> **Craft UI should feel fast before it feels lightweight.**

งาน [Craft 0.1.16: responsiveness และ frame pipeline](phase1/craft-phase1-rev16.md) เพิ่ม event-driven host, layout reuse และ partial repaint ใน source แล้ว ผล latency และการตรวจภาพจริงยังอยู่ใน [สถานะ Rev.16](phase1/output/phase1-rev16-status.md)

[Craft 0.1.17: content-aware layout และ repaint correctness](phase1/craft-phase1-rev17.md) ถูก implement ใน source แล้ว และรอการทดสอบภาพจริงและ performance จากผู้ใช้ตาม [สถานะ Rev.17](phase1/output/phase1-rev17-status.md)

[Craft 0.1.18: ScrollView และ viewport-aware rendering](phase1/craft-phase1-rev18.md) ถูก implement ใน source แล้ว และรอการตรวจรับภาพจริง/ความไวจากผู้ใช้ตาม [สถานะ Rev.18](phase1/output/phase1-rev18-status.md). Virtualization ยังเป็นงานถัดไป

[Craft 0.1.19: ScrollView sizing correctness](phase1/craft-phase1-rev19.md) แก้ปัญหาที่พบจากภาพ 0.1.18 ให้ viewport รับพื้นที่ว่างก่อน controls และไม่ยืด content โดยไม่จำเป็น ผลตรวจในเครื่องพัฒนาอยู่ใน [สถานะ Rev.19](phase1/output/phase1-rev19-status.md); ยังต้องรับรองบนเครื่องผู้ใช้

[Craft 0.1.20: Virtualized List รุ่นแรก](phase1/craft-phase1-rev20.md) เพิ่มการสร้างและวัดเฉพาะแถวที่อยู่ใน viewport พร้อม overscan โดยเริ่มจากแถวแนวตั้งที่สูงเท่ากัน; ยังรอผู้ใช้ตรวจรับ GUI จริงตาม [สถานะ Rev.20](phase1/output/phase1-rev20-status.md)

[Craft 0.1.22: Text Input และ Focus](phase1/craft-phase1-rev22.md) เพิ่มช่องข้อความบรรทัดเดียวใน `std.ui` ด้วย Windows EDIT control เพื่อรับ IME, caret, selection และ clipboard พร้อม event กลับ Craft; ยังรอผู้ใช้ตรวจรับตาม [สถานะ Rev.22](phase1/output/phase1-rev22-status.md). หน้าตาของช่องนี้ยังผูกกับ native control จึงยังไม่ถึงเป้าหมาย styling อิสระระยะยาว.
