# Craft Phase 1 Rev.23 — วิเคราะห์ Layout ตัวอย่างและแผนยกระดับ UI Runtime

วันที่: 7 ตุลาคม 2026  
เป้าหมายรุ่น: Craft 0.1.23  
สถานะ: **implement แล้ว; build/check/regression ผ่านและตรวจภาพ Direct2D ที่ DPI 125% แล้ว; interaction บางส่วนยังรอผลตาม `output/phase1-rev23-status.md`**  
ฐานเปรียบเทียบ: source ของ `std.ui` หลัง Rev.22; การทดสอบ Rev.22 โดยผู้ใช้ยังรอผล

## เป้าหมายและขอบเขต

ใช้ภาพ ERP System และภาพ Craft Office ที่ผู้ใช้ส่งมาเป็น **ภาพอ้างอิงด้านโครงหน้าจอและประสบการณ์ใช้งาน** เพื่อระบุความสามารถที่ Craft UI Engine Runtime ยังต้องมี ภาพเหล่านี้ไม่ใช่ข้อกำหนดให้สร้าง ERP หรือ Craft Office ในรอบนี้ โดยเฉพาะ **ไม่สร้าง Craft Office**. Phase 1 ยังคงโฟกัส Runtime; `packages/gui` และ `packages/api-framework` ยังอยู่ใน Archive. ไม่เปิดงาน HTTP ร่วมกับ UI ในรุ่นนี้.

ประเมินจากสิ่งที่มองเห็นในภาพเท่านั้น: ภาพนิ่งยืนยันพฤติกรรมจริง เช่น resize, keyboard, accessibility, animation หรือความเร็วไม่ได้. เป้าหมาย UX ของ Craft ยังคง **Performance first. Simplicity always.** โดยไม่ใช้ WebView/HTML/CSS/React และไม่สร้างเครื่องมือออกแบบแบบลากวาง.

## ภาพที่ 1: ERP System

| พื้นที่ | องค์ประกอบที่เห็น | ความสามารถที่ต้องรองรับ |
| --- | --- | --- |
| กรอบหน้าต่าง | Title bar พร้อมชื่อระบบ/บริษัท, global search, แจ้งเตือน, ผู้ใช้ และปุ่มหน้าต่าง | App chrome, ไอคอน, การจัดตำแหน่ง, ช่องค้นหา, hit area ของ window controls |
| นำทางซ้าย | Sidebar สีเข้ม, เมนูหลายระดับที่ยุบ/ขยายได้, active item, รายการโปรดด้านล่าง | พื้นที่ความกว้างคงที่/ยืดหยุ่น, tree navigation, selected/hover state, scroll เฉพาะส่วน |
| เมนูและคำสั่ง | Menubar และ ribbon/toolbar ที่มีไอคอน ชื่อคำสั่ง เส้นแบ่ง กลุ่มคำสั่ง และ shortcut | แถบคำสั่งหลายระดับ, ไอคอน/ภาพ, menu popup, tooltip, keyboard shortcut |
| เอกสารที่เปิด | แถบแท็บหลัก มีปุ่มปิดและเพิ่มแท็บ | Tab strip, selected state, overflow เมื่อแท็บมาก, การจัดการ focus |
| พื้นที่งานหลัก | Master–detail สองคอลัมน์; ซ้ายเป็นรายการ ขวาเป็นฟอร์ม | ขนาดแบบสัดส่วนและขั้นต่ำ, splitter, resize, overflow/scroll ที่แยกกัน |
| รายการด้านซ้าย | ค้นหา, ตัวกรองสถานะ, ช่วงวันที่, ตารางข้อมูล, badge สถานะ, แถวที่เลือก, pagination | Form controls, ตารางหลายคอลัมน์, row selection, header, จัดรูปแบบข้อมูล, virtualization |
| รายละเอียดด้านขวา | หัวข้อและ badge, แท็บย่อย, ฟอร์มสองคอลัมน์, lookup, date picker, dropdown, textarea | Grid/form layout, label alignment, input หลายชนิด, validation/error state, popup anchor |
| รายการสินค้าและยอดรวม | Toolbar, ตารางคล้าย spreadsheet, ตัวเลขชิดขวา, หมายเหตุ, summary | Editable grid, column sizing, keyboard navigation, multiline text, sticky totals |
| ท้ายหน้าต่าง | Status bar พร้อมการเชื่อมต่อ เวลา และข้อมูลระบบ | แถบตรึงด้านล่างและ live status ที่ไม่ทำให้ layout หลักกระตุก |

## ภาพที่ 2: Craft Office

| พื้นที่ | องค์ประกอบที่เห็น | ความสามารถที่ต้องรองรับ |
| --- | --- | --- |
| หน้าหลัก | Top bar, sidebar, ข้อความต้อนรับ, quick-action cards, ภาพ hero, gallery template, recent table, quick access | App shell, card/grid, image asset, horizontal gallery, ตาราง, responsive layout |
| หลายหน้าต่าง | หน้าหลักและหน้าต่าง Writer, Spreadsheet, Presentation, PDF ซ้อนกัน | หลาย top-level windows, ownership/focus, lifecycle, DPI และ event routing ต่อหน้าต่าง |
| Writer | Ribbon, หน้ากระดาษ, ข้อความจัดรูปแบบ, ภาพแทรก, scroll, status/zoom | Document canvas, rich text, text selection, pagination, image embedding; เป็น editor เฉพาะทาง |
| Spreadsheet | Ribbon, formula bar, row/column headers, grid จำนวนมาก, sheet tabs, chart | Virtualized 2D grid, cell editing/selection, frozen headers, formula UX, chart; เป็นระบบเฉพาะทาง |
| Presentation | Slide thumbnail rail, slide canvas, ribbon, artwork และข้อความ | Canvas, transform, thumbnail list, object selection; เป็น editor เฉพาะทาง |
| PDF | Toolbar, page thumbnails, document viewport, search, page navigation, zoom | Document rendering, tiled/virtualized pages และ navigation; ตัวอ่าน PDF เป็นระบบเฉพาะทาง |

ภาพที่ 2 แสดงว่าฐาน UI ต้องรองรับ shell, ภาพ, ตาราง, viewport และหลายหน้าต่างได้ แต่ **ไม่จำเป็นต้องใส่ document/spreadsheet/PDF engine ทั้งหมดลงใน UI Runtime**. เครื่องมือเฉพาะทางเหล่านี้ค่อยสร้างบน primitive ของ Runtime ในรุ่นถัดไป.

## เทียบกับ `std.ui` ปัจจุบัน

API ที่มีใน source: `openWindow`, `setTitle`, `setBackground`, `closeWindow`, `nextEvent`, `row`, `column`, `panel`, `scrollView`, `virtualList`, `text`, `button`, `textInput`, `setText`, `mount`, `update`, `renderer`. Renderer มี Direct2D/DirectWrite และ fallback; `TextInput` รุ่นแรกใช้ native Windows EDIT. `VirtualList` เป็นรายการแนวตั้งความสูงแถวคงที่. ชื่อ API ที่มีไม่ได้แปลว่าความสามารถทั้งหมดในภาพพร้อมใช้แล้ว.

| หมวด | สิ่งที่มีแล้ว | ช่องว่างที่ภาพเผยให้เห็น | ชั้นที่เหมาะสม |
| --- | --- | --- | --- |
| โครงหน้าจอ | Row, Column, Panel, ScrollView | ขนาด fixed/flex/min/max, alignment, gap, padding, พื้นที่คงที่, splitter, overlay, overflow ที่คาดเดาได้ | Runtime |
| การวาด | Text, Button, สีพื้นหน้าต่าง, Direct2D/DirectWrite | ภาพ/ไอคอน, border, radius, selected/hover/disabled styles, theme, clipping ที่สม่ำเสมอ | Runtime primitives; theme/components ใน library ภายหลัง |
| การนำทาง | Button และ event | Tabs, tree, menubar, ribbon, popup/menu, tooltip, focus order | Runtime สำหรับ event/overlay/focus; component ใน library ภายหลัง |
| ข้อมูลจำนวนมาก | VirtualList แนวตั้ง | Table/Grid แบบ 2D, header, column resize, row/cell selection, keyboard navigation, variable row height | Runtime สำหรับ viewport/virtualization; table component ภายหลัง |
| ฟอร์ม | TextInput บรรทัดเดียว | Select/ComboBox, checkbox, date picker, textarea, lookup, validation, label alignment | Runtime สำหรับ input/focus/popup; controls ใน library ภายหลัง |
| หน้าต่าง | เปิดหน้าต่างได้หนึ่งบานต่อ interpreter run | หลายหน้าต่าง, owner/modal, routing, cross-window focus | Runtime รุ่นถัดไป |
| การโต้ตอบ | Click, key, scroll และ text events ขั้นต้น | drag & drop, shortcut routing, context menu, accessibility, IME/keyboard ใน control ที่ซับซ้อน | Runtime รุ่นถัดไป |
| ความเร็ว | Partial updates และ VirtualList | virtual grid, trace input-to-present, layout/cache ของ tree ใหญ่, resize/scroll ที่มีหลาย pane | Runtime และเครื่องมือวัด |
| เอกสารเฉพาะทาง | ยังไม่มี | Rich text editor, spreadsheet model/formula, presentation, PDF, chart | Library/แอปในอนาคต; ไม่ใช่งาน Rev.23 |

ข้อจำกัดสำคัญ: TextInput รุ่นปัจจุบันยังเป็น native EDIT จึงต้องออกแบบการจัดขนาด, clipping, focus และภาพลักษณ์ให้ทำงานร่วมกับ renderer หลัก. `VirtualList` ไม่ใช่ตารางสองมิติ. รายละเอียดเชิงพฤติกรรมให้ยึด source และเอกสาร Rev.22 ล่าสุดเมื่อเริ่ม implement.

## ขอบเขตที่เสนอสำหรับ 0.1.23: Layout Foundation

รอบนี้มุ่งทำ **โครงหน้าจอที่ประกอบได้และ resize ได้ถูกต้อง** ก่อนเพิ่มชุด component จำนวนมาก:

1. **Sizing contract:** กำหนด fixed/fill หรือ weight, min/max, padding, gap และ alignment สำหรับ Row/Column/Panel/ScrollView โดยระบุการคำนวณเมื่อพื้นที่ไม่พอ, DPI เปลี่ยน และค่าไม่ถูกต้อง. API ต้องใช้ได้จาก Craft syntax จริงและรักษา API เดิม.
2. **Nested pane layout:** รองรับ sidebar/content, header/body/footer และ master/detail หลายชั้นโดยไม่ hardcode พิกัด. เพิ่ม SplitView หรือ primitive ที่ให้สัดส่วนและขนาดขั้นต่ำ พร้อม drag divider หากออกแบบและ hit test พร้อม. ต้องไม่ทำให้ TextInput หลุด focus ขณะ resize.
3. **Clipping/scroll contract:** แต่ละ pane ที่ scroll ต้อง clip, hit test และจัด scrollbar ตาม viewport ของตนเอง; nested scroll ต้องส่ง wheel ไปยังพื้นที่ที่เหมาะสม. ระบุพฤติกรรมเมื่อ content สั้น/ยาวและระหว่าง resize.
4. **Node identity และ event stability:** เมื่อต้นไม้ UI เปลี่ยนหรือปรับขนาด ให้ selected/focus/scroll state ของ node ID เดิมอยู่ถูกที่. หลีกเลี่ยง full relayout/paint ใน pane ที่ไม่ได้รับผลเมื่อทำได้อย่างปลอดภัย.
5. **ตัวอย่าง Layout Lab:** สร้างตัวอย่างสังเคราะห์โดยใช้ `std.ui` ตรง ๆ: top bar, sidebar, tab-like row, master/detail, form placeholders, list และ status bar. ใช้ชื่อ/ข้อมูลกลาง ๆ; ไม่สร้าง ERP หรือ Craft Office. เอกสารบอกส่วนที่เป็นเพียง placeholder ชัดเจน.

ก่อนล็อก API ให้ทำตาราง behavior ของ layout ทุกกรณีและตัวอย่าง Craft ที่ `craft check` ได้เมื่อถึงรอบ implement. อย่าเพิ่มชื่อ control ที่ยังไม่มี implementation ลงในเอกสารผู้ใช้เหมือนพร้อมใช้.

## เกณฑ์รับงานเมื่อ implement Rev.23

- Layout Lab แสดง shell หลายชั้นและ master/detail ที่ 30/70 โดยมี min width; resize ตั้งแต่หน้าต่างแคบถึงกว้างแล้วไม่มีส่วนทับกันหรือหายโดยไม่มี overflow ที่นิยามไว้.
- Sidebar, list และ detail เลื่อนแยกกันตามขอบเขต viewport; clipping, click และ TextInput focus ตรงกับภาพหลัง scroll/resize.
- DPI หลายระดับและ fallback renderer ให้พิกัด layout/hit test สอดคล้องกัน; ไม่เกิด blank frame หรือภาพค้างที่เห็นได้.
- เปรียบเทียบ baseline เดียวกันด้วยเวลาตอบสนอง input, layout/paint time และ frame time p50/p95/p99 พร้อมรายงาน renderer, ขนาดหน้าต่าง, DPI และจำนวน node. ไม่รับประกัน FPS หรือ “เร็วเท่า React” จากภาพนิ่ง.
- API เดิมและตัวอย่าง UI เดิมยังทำงาน; รายงานข้อจำกัดจริงใน status และคู่มือให้ผู้ใช้ทดสอบด้วยตนเอง.

## ลำดับต่อหลัง 0.1.23

| ลำดับ | งาน | เหตุผล |
| --- | --- | --- |
| ถัดไปสูง | Theme/style state พื้นฐาน, image/icon, overlay/popup, tabs/menu composition | ทำ app shell และ action bar ในทั้งสองภาพได้จริง |
| ถัดไปสูง | Table/VirtualGrid, selection, column sizing, headers | ERP list และ spreadsheet-like data เป็นจุดหนักด้าน performance |
| ถัดไป | Form controls เพิ่ม, textarea, date/select, shortcut/focus/accessibility | ทำฟอร์มหนาแน่นให้ใช้งานด้วยเมาส์และคีย์บอร์ดได้ครบ |
| ระยะหลัง | หลายหน้าต่าง, drag & drop, animation/effects, chart/document canvases | รองรับ workflow ซับซ้อนโดยไม่ขยาย Rev.23 จนตรวจรับไม่ได้ |
| แยกจาก Runtime | Rich text, spreadsheet/formula, presentation, PDF, ERP workflow | เป็น domain engines หรือ library ที่สร้างบน UI Runtime |

เมื่อเริ่ม implement ให้ปรับลำดับตามผลทดสอบ Rev.22 และต้นทุนจริงของ Windows backend. เอกสารฉบับนี้เป็นแผนความสามารถ ไม่ใช่การประกาศว่าองค์ประกอบในภาพใช้งานได้แล้ว.
