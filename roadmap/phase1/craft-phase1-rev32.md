# Craft Phase 1 Rev.32 (0.1.32) — Button text alignment และ startup cursor

วันที่: 8 ตุลาคม 2026  
สถานะ: **แก้ Button alignment และ startup cursor ใน source แล้ว; สัญญา Virtual Grid เขียนแล้ว; รอผู้ใช้ทดสอบ Runtime**

## ปัญหา

ใน Craft UI Showcase ปุ่ม Primary, Secondary และ Disabled แสดงข้อความชิดขวาหรือถูกตัดที่ขอบ ขณะที่ Standard Button แสดงข้อความกึ่งกลางตามที่คาด ภาพทดสอบของผู้ใช้แสดงปุ่มทั้งสี่เป็นสองแถวและระบุขนาด 180×44 logical pixels ในแอป

## สาเหตุที่ยืนยันจาก source

- Showcase ใช้ `gui.primaryButton` และ `gui.secondaryButton` ซึ่งเรียก `textStyle` กับ Button; Standard Button เรียก `gui.button` โดยตรง ไม่มี API alignment ที่แอปตั้งต่างกัน
- ใน `internal/stdlib/ui_direct2d_windows.go` ค่า default Button format ส่ง `SetTextAlignment(2)` ซึ่งเป็น `DWRITE_TEXT_ALIGNMENT_CENTER` และแสดงผลตรงกลาง
- Styled Button ส่งค่า `1` ผ่าน `styledTextFormat` ซึ่ง DirectWrite แปลเป็น `DWRITE_TEXT_ALIGNMENT_TRAILING` จึงชิดขวา นี่อธิบายความต่างระหว่างปุ่ม Standard กับปุ่มตกแต่งในภาพได้ตรงจุด
- `internal/stdlib/ui_paint.go` กำหนดค่าภายในสำหรับ Canvas text เป็น start=0, center=1, end=2; ค่าเหล่านี้ตรงกับการตีความของ GDI แต่หากส่งตรงเข้า DirectWrite จะสลับ center กับ trailing เช่นกัน

การตั้งค่า Showcase ไม่ใช่ต้นเหตุของข้อความชิดขวา การยืดขนาดของปุ่มในบางตำแหน่งเป็นประเด็น layout แยกต่างหากและไม่อยู่ในขอบเขต Rev.32 นี้

## งานใน Rev.32

1. แยกค่า alignment ภายใน Craft ออกจากค่า enum ของ DirectWrite ให้ชัดเจน ใช้ชื่อค่าคงที่แทนตัวเลขดิบ
2. แปลง start/center/end เป็นค่า DirectWrite leading/center/trailing เฉพาะตอนสร้าง TextFormat โดยไม่เปลี่ยนความหมายของ GDI
3. ให้ Styled Button และ Standard Button ใช้การจัดข้อความกึ่งกลางแบบเดียวกัน; ตรวจ Canvas text ที่เลือก center ด้วย เพราะใช้เส้นทางแปลงค่าเดียวกัน
4. ปรับคอมเมนต์ที่ระบุค่า alignment ผิด และทบทวน cache key ของ styled TextFormat หลังการแปลงค่า

สถานะ implementation: เพิ่มค่าคงที่ alignment ภายในและแปลงเป็น enum ของ DirectWrite ก่อนตั้ง TextFormat; Styled Button และ Canvas text ใช้เส้นทางเดียวกัน. GDI ใช้ค่าเดิม. Cache key เก็บค่า alignment ภายในที่แยกกันเหมือนเดิม. ยังไม่มีผลทดสอบ GUI จากผู้ใช้สำหรับการแก้นี้.

## เกณฑ์ตรวจรับโดยผู้ใช้

- ปุ่ม Primary, Secondary, Standard และ Disabled ในหน้า Components มีข้อความกึ่งกลางในกรอบที่วาด ไม่ชิดขวาหรือถูกตัด
- Canvas target ที่กำหนด `center` แสดงข้อความกึ่งกลาง
- ตรวจทั้ง Direct2D/DirectWrite และ GDI fallback; การแก้ DirectWrite ไม่ทำให้ GDI เปลี่ยนตำแหน่ง
- ตรวจที่ DPI 100% และ 125% เพื่อยืนยันว่าการจัดตำแหน่งยังคงอยู่กลาง

## ปัญหาเคอร์เซอร์หมุนค้างเมื่อเปิดแอป

ผู้ใช้พบว่าเคอร์เซอร์เมาส์แสดงสถานะกำลังโหลดต่อเนื่องหลังเปิด Craft UI Showcase. การแก้ครั้งแรกเพิ่ม `GetMessageW` หนึ่งครั้ง เพราะเอกสาร Windows ระบุว่า feedback cursor ตอนเริ่มโปรแกรมจะหยุดหลังการเรียกนี้. **ผลทดสอบของผู้ใช้ยืนยันว่าแนวทางนี้อย่างเดียวยังไม่แก้อาการ:** เคอร์เซอร์ยังหมุนอยู่ และหยุดเมื่อคลิกหน้าต่างอื่นแล้วกลับมา.

ตรวจเพิ่มพบว่า class ของหน้าต่าง Craft ลงทะเบียนโดยปล่อย `hCursor` เป็น null. การสลับหน้าต่างทำให้ Windows ตั้งเคอร์เซอร์ใหม่ จึงเป็นคำอธิบายที่สอดคล้องกับอาการ แต่ยังไม่ใช่สาเหตุที่ยืนยันด้วยการทดสอบแยก. รอบก่อนปรับ Runtime ให้โหลด `IDC_ARROW`, กำหนดเป็น class cursor และเรียก `SetCursor` หลังเปิดหน้าต่าง. **ผู้ใช้รายงานว่า 0.1.32 รอบนั้นยังไม่ผ่าน**. การเรียก `SetCursor` เดิมเกิดก่อน maximize, mount และ paint เฟรมแรก จึงอาจถูกสถานะ startup หรือการเปิดหน้าต่างทับภายหลัง.

รอบแก้ใหม่ให้ Runtime จัดการ `WM_SETCURSOR` สำหรับ client area ของหน้าต่าง Craft โดยตรง และคืน arrow cursor หลัง paint เฟรมแรกเมื่อเมาส์อยู่เหนือ client area ของหน้าต่างนั้น. ไม่ทับ cursor ของ native TextInput, พื้นที่ขอบหน้าต่าง หรือหน้าต่างอื่น. คง `GetMessageW` เพื่อจัดการ startup feedback ตามเอกสาร Windows. สาเหตุระดับ OS ที่แน่ชัดยังต้องยืนยันจากผลทดสอบผู้ใช้.

เกณฑ์ตรวจรับ: ติดตั้ง Setup ที่สร้างจาก source รอบแก้ใหม่ เปิด UI Showcase โดยไม่คลิกหน้าต่างอื่นก่อน และตรวจว่าเคอร์เซอร์กลับเป็นปกติเมื่อหน้าต่างพร้อมใช้งาน. เลื่อนเมาส์เหนือพื้นหน้าต่างและ TextInput, สลับออก/กลับ, คลิก และปิดหน้าต่างเพื่อตรวจว่าเคอร์เซอร์ยังเหมาะกับพื้นที่นั้น. ผู้ใช้เป็นผู้ทดสอบ; ยังไม่มีผลทดสอบของรอบแก้ใหม่นี้.

## เทียบข้อเสนอ UI Engine สำหรับแอประดับ Excel กับ source ปัจจุบัน

ข้อเสนอที่แนบมาเป็นเป้าหมายเชิงสถาปัตยกรรม ไม่ใช่รายการ API ที่ Craft รองรับแล้ว. การเทียบนี้อ้างอิง `packages/gui/src/gui.craft` และ `internal/stdlib/ui*.go` ใน source ปัจจุบัน. หน้า Excel ใน UI Showcase เป็นตาราง Canvas ขนาดเล็กที่แอปถือข้อมูลเอง; ไม่ใช่หลักฐานว่า Engine รองรับข้อมูลหลักล้านแถว.

เกณฑ์รอบ Final ที่แยกงาน Craft UI ออกจากงานแอป Spreadsheet อยู่ใน [Craft UI Final: Excel acceptance](../craft-ui-final-excel-acceptance.md). Craft เป็นเครื่องมือสร้าง UI; สูตร, workbook, XLSX, pivot และ workflow เฉพาะ Excel อยู่ในแอป/Core/library ของแอป.

| ระบบ | สิ่งที่มีแล้ว | ช่องว่างที่ต้องออกแบบ/พัฒนา |
| --- | --- | --- |
| ตารางและ virtualization | `Grid` จัดตำแหน่ง children และ `VirtualList` ขอช่วงแถวที่มองเห็น | ยังไม่มี Virtual Grid สองมิติที่ virtualize ทั้งแถวและคอลัมน์โดยไม่สร้าง `UiNode` ต่อเซลล์; frozen panes, ขนาดแถว/คอลัมน์ต่างกัน และ merged cells ยังไม่มีสัญญารองรับ |
| Layout และ scrolling | มี measure/place, scroll แนวตั้ง/แนวนอน, เก็บตำแหน่ง scroll และ reuse geometry ใน update บางกรณี | การเปลี่ยนข้อมูลตารางจำนวนมากยังต้องมี viewport model, พิกัดขนาดใหญ่, hit test ตาม row/column และการปรับขนาดที่ไม่วัดทั้งตารางใหม่ |
| Rendering | มี Direct2D/DirectWrite, GDI fallback, Canvas, invalidation เฉพาะกรอบบางกรณี, TextFormat/ภาพ cache บางส่วน | ยังไม่มี pipeline สำหรับ dirty cells, การรวมคำสั่งวาดและแยกชั้น grid/selection/editor ที่กำหนดสัญญาชัด; ไม่ควรอ้างว่าทุกการแก้เซลล์วาดเฉพาะเซลล์นั้นแล้ว |
| Input และ focus | มี pointer events/capture, keyboard events, shortcut, focus scope, native `EDIT` และ Canvas hit region | ยังไม่มี selection model สำหรับช่วงเซลล์, keyboard navigation ของตาราง, drag selection/autofill หรือ editor ที่ยึดตำแหน่งเซลล์และรักษา IME/selection ระหว่าง scroll |
| ข้อมูลและ state | แอปเป็นเจ้าของ state; `VirtualList` ขอ visible range | ยังไม่มี data source contract สำหรับขอ cell ตามช่วงที่มองเห็น, แจ้ง cell/range ที่เปลี่ยน, lazy loading หรือ cancellation; หลีกเลี่ยงการคัดลอกทั้ง sheet เป็น UI tree |
| Clipboard และประวัติ | native TextInput ใช้พฤติกรรมแก้ข้อความของ Windows | ยังไม่มี copy/paste ช่วงเซลล์ข้ามแอป, รูปแบบข้อมูลตาราง หรือ undo/redo ระดับคำสั่งของตาราง |
| Responsiveness | UI host อยู่คนละ OS thread กับ Craft interpreter; มีคิวคำสั่ง/เหตุการณ์และ `CRAFT_UI_TRACE` วัดเวลา engine | ยังไม่มีสัญญาส่งผลลัพธ์งานข้อมูลหนักกลับ UI อย่างปลอดภัย, ยกเลิกงานเก่า และจำกัด backlog; `nextEvent` ที่ไม่ถูกอ่านนานอาจทำให้คิวเต็ม |
| Accessibility / DPI / platform | มี UI Automation provider, focus policy, logical pixels และ per-monitor DPI บน Windows | ต้องออกแบบ virtual cell accessibility, การอ่านช่วงที่เลือกด้วย screen reader และตรวจหลาย DPI; backend native ปัจจุบันมีเฉพาะ Windows |

ข้อเสนอที่ระบุว่า GPU เป็นงานอนาคตต้องปรับให้ตรง source: Craft ใช้ Direct2D/DirectWrite แล้ว และ fallback เป็น GDI เมื่อใช้ GPU path ไม่ได้. งานต่อไปคือพิสูจน์ความถูกต้องและประสิทธิภาพของทั้งสองเส้นทางในตารางขนาดใหญ่ ไม่ใช่เพิ่ม GPU เป็นครั้งแรก.

### สิ่งที่ต้องกำหนดใน Rev.32

1. เขียนสัญญา **generic Virtual Grid/DataGrid** ระหว่าง UI กับ data source: จำนวนแถว/คอลัมน์, ช่วง visible พร้อม overscan, ค่าที่ร้องขอ, การแจ้ง item/range ที่เปลี่ยน และนโยบายเมื่อข้อมูลยังโหลดไม่เสร็จ. ต้องใช้กับตารางข้อมูลทั่วไปได้โดยไม่รู้จัก worksheet หรือสูตร. ตัวอย่าง callback ในข้อเสนอเป็นเพียงแนวคิด; ต้องออกแบบให้ใช้ได้จริงกับชนิดข้อมูลและรูปแบบ event ของ Craft ก่อนประกาศ API.
2. กำหนด **viewport และพิกัด** สำหรับการ scroll สองแกน, frozen rows/columns, ขนาดที่เปลี่ยนได้ และ hit testing โดยไม่สร้าง widget ทุกเซลล์. ระบุว่าฟีเจอร์ merged cells จะเข้ามาในขั้นใด เพราะมีผลต่อทั้ง layout, selection และ accessibility.
3. กำหนด **generic selection/editor/input primitives** แยกจาก renderer: hit target, ช่วงที่เลือก, การลาก, focus กับ IME และการแก้ข้อมูลแบบ commit/cancel. แอป Spreadsheet เป็นผู้กำหนด active cell, Shift+Arrow, F2, autofill, paste special และ undo/redo ของข้อมูล; Craft จัด keyboard routing และ clipboard primitives ทั่วไป.
4. กำหนด **ขอบเขต invalidation และตัววัดผล**: เมื่อ cell เปลี่ยนให้รู้ว่าต้องวาด region ใด; วัดเวลา data request, layout, draw และการตอบสนองต่อ input แยกกัน. `CRAFT_UI_TRACE` ปัจจุบันวัดถึง draw completion ไม่ใช่เวลาที่ภาพปรากฏบนจอ จึงห้ามแปลงตัวเลขนี้เป็น FPS โดยตรง.
5. บันทึก **ขอบเขตความรับผิดชอบ**: UI Engine จัด viewport, rendering, input และ accessibility แบบทั่วไป; แอปสร้าง formula bar, sheet tabs, selection policy และ workflow เฉพาะ Spreadsheet; การคำนวณสูตร, dependency graph, workbook และ `.xlsx` เป็นของ Core/library ของแอป.

สัญญาที่กำหนดตามรายการนี้อยู่ใน [Rev.32 generic Virtual Grid contract](craft-phase1-rev32-virtual-grid-contract.md). เป็นเอกสารออกแบบ ไม่ใช่ public API ที่ใช้ได้แล้ว.

Rev.32 ยังไม่ถือว่าได้ implement Virtual Grid หรือ Excel เต็มรูปแบบจากการเขียนสัญญาเหล่านี้. งานแก้ Button alignment และ startup cursor ด้านบนยังเป็นงานแก้พฤติกรรมปัจจุบันของรุ่นนี้.

### ลำดับ implementation หลังสัญญาชัดเจน

1. **Foundation:** Virtual Grid สองมิติ, data request เฉพาะ viewport, scroll สองแกน, hit test, active cell และ keyboard navigation.
2. **Editing:** overlay editor ที่รองรับ IME, selection หลายเซลล์, clipboard ช่วงเซลล์, resize/frozen panes และ command history สำหรับ undo/redo.
3. **Scale:** dirty cell rendering, text/layout cache, variable sizes, lazy/background data loading, cancellation และ virtual accessibility.
4. **Application proof:** ตัวอย่าง spreadsheet หลายชีตที่ใช้ formula engine และไฟล์แยกจาก UI Engine เพื่อทดสอบสัญญาในงานจริง.

เกณฑ์ประสิทธิภาพจากข้อเสนอ เช่น 1 ล้านแถว, เลื่อนใกล้ 60 FPS และการแก้เซลล์ภายใน 16–50 ms เป็น **เป้าหมายสำหรับออกแบบและวัดบนเครื่องที่ระบุสเปกภายหลัง** ไม่ใช่ผลที่ Craft ทำได้แล้ว. เกณฑ์ที่ต้องรักษาตั้งแต่ Foundation คือหน่วยความจำและจำนวน node/draw commands โตตามพื้นที่ที่มองเห็น ไม่โตตามจำนวนเซลล์ทั้งหมด. ผู้ใช้เป็นผู้ทดสอบพฤติกรรม GUI; ยังไม่มีผลทดสอบประสิทธิภาพสำหรับรายการนี้.
