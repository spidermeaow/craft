# เกณฑ์ Final: Craft UI Engine ต้องใช้สร้างโปรแกรม Spreadsheet ระดับ Excel ได้

วันที่: 8 ตุลาคม 2026  
สถานะ: **เป้าหมายและเกณฑ์ตรวจรับระยะ Final; ยังไม่ใช่ความสามารถที่มีแล้ว**

## เป้าหมาย

เมื่อประกาศ Craft UI Engine ระดับ Final ต้องมีโปรแกรม Spreadsheet Desktop **ที่พัฒนาโดยใช้ Craft** เพื่อพิสูจน์ว่า Craft เป็นเครื่องมือสร้าง UI ของแอปซับซ้อนได้. โปรแกรมตัวอย่างรับผิดชอบฟังก์ชันตามหมวดในไฟล์อ้างอิงที่ผู้ใช้ให้ (A–L, กลุ่มสูตร และขนาดข้อมูล). Craft ให้ API และเครื่องมือ UI ทั่วไป; ไม่ต้องมีความรู้เรื่อง workbook, สูตร หรือเมนูเฉพาะ Excel. แอปอาจใช้ library และบริการเสริม แต่ต้องไม่แก้ private UI Runtime เพื่อทำฟีเจอร์เฉพาะแอป. หน้าตาเหมือน Excel หรือ Canvas mock เพียงอย่างเดียวไม่ถือว่าผ่าน.

คำว่า “ระดับ Excel” ในเอกสารนี้หมายถึง **โจทย์ทดสอบความสามารถของเครื่องมือพัฒนา UI** ไม่ใช่รายการฟีเจอร์ที่ต้องเพิ่มเข้า Craft. การรองรับไฟล์ สูตร หรือบริการของ Microsoft เป็นงานของแอปและ library ที่แอปเลือกใช้. แต่ละความสามารถในไฟล์อ้างอิงต้องมีเจ้าของและผลตรวจรับของตัวเองก่อนกล่าวว่าเสร็จ.

## ขอบเขตความรับผิดชอบ

| ชั้น | ต้องรับผิดชอบ |
| --- | --- |
| Craft UI Engine / `std.ui` | หน้าต่าง, layout, renderer, scroll, viewport, Canvas, hit test, input/focus, text/IME, clipboard และ drag-drop primitives, accessibility, DPI, event loop และการวาดอย่างมีประสิทธิภาพ. เพิ่ม primitive ใหม่เมื่อใช้ได้กับแอปหลายประเภทและ public API เดิมประกอบไม่ได้อย่างเหมาะสม |
| `packages/gui` | component ทั่วไป เช่น Button, TextInput, menu, dialog, toolbar, generic virtual table/grid, chart primitive, theme และ state styling. DataGrid ที่อยู่ใน package ต้องไม่รู้จักสูตรหรือ Workbook |
| แอปทดสอบ Spreadsheet / Craft Office | หน้าตาและ workflow เฉพาะ Spreadsheet: sheet tabs, formula bar, name box, ribbon commands, cell selection policy, editor behavior, autofill, print setup, query/pivot panels และการประกอบ generic components ให้เป็นหน้าจอจริง |
| Spreadsheet Core ของแอป | cell model, parser/evaluator สูตร, dependency graph, recalculation, named ranges, validation, sort/filter, analysis, pivot/model และ undo/redo ระดับข้อมูล |
| Library ที่แอปใช้ | CSV/XLSX และรูปแบบที่ประกาศรองรับ, PDF/print layout, encoding และ round-trip. จะพัฒนาเป็น package แยกได้ แต่ไม่อยู่ใน UI Engine |
| บริการ/ส่วนขยายที่แอปเลือก | Sync/co-authoring, AI, web data, add-ins, automation/ภาษาอื่น และความเข้ากันได้กับบริการเฉพาะราย |

**กฎตัดสินขอบเขต:** ถ้าความสามารถยังมีความหมายในโปรแกรมจัดการข้อมูล, IDE หรือ ERP ที่ไม่ใช่ Spreadsheet ให้พิจารณาเป็น generic UI API/component. ถ้าต้องรู้จัก cell reference, sheet, formula, pivot หรือรูปแบบไฟล์ Workbook ให้เก็บไว้ในแอป/Core/library ของแอป. แอปมีสิทธิ์สร้าง component เฉพาะเองด้วย Canvas และ public UI API; ไม่จำเป็นต้องย้ายทุก component เข้าสู่ `packages/gui`.

## สิ่งที่ต้องพัฒนาใน Craft UI Engine และ GUI package

1. **Generic virtual viewport/grid:** ให้ component ขอข้อมูลเฉพาะช่วงที่มองเห็นทั้งสองแกน วาดและ hit test ตามพิกัดใหญ่ได้ โดยไม่สร้าง `UiNode` ทุกช่อง. API ใช้กับตารางฐานข้อมูลหรือ log viewer ได้ ไม่ผูกกับ worksheet.
2. **Layout, scroll และ renderer:** scroll สองแกน, fixed/sticky region ที่ประกอบเป็น frozen pane ได้, ขนาดแถว/คอลัมน์ต่างกัน, clip, dirty region และ text measurement/cache ที่ถูกต้อง. การ merge/hide/group แบบ Spreadsheet เป็น policy ในแอปบน primitive เหล่านี้.
3. **Input และข้อความทั่วไป:** keyboard/pointer routing, capture, focus, selection/editor primitive, IME ภาษาไทย, caret, clipboard และ drag-drop. Craft ไม่เป็นเจ้าของกฎ `Shift+Arrow`, F2, autofill หรือ paste special ของ Spreadsheet.
4. **Styling และการวาด:** font, สี, border, alignment, rotation, multiline, shapes/images และ chart drawing primitive. Number format, conditional rules และชุดกราฟเฉพาะอยู่ในแอปหรือ library.
5. **Responsiveness:** UI host ต้องรับ input ต่อได้เมื่อแอปทำงานหนัก; public API ต้องรองรับการส่งอัปเดตอย่างปลอดภัย, backpressure และ cancellation ที่เหมาะกับ async data source. งานคำนวณสูตร/โหลดไฟล์เป็นของแอป.
6. **Accessibility และ DPI:** generic virtual items ต้องเข้าถึงด้วย screen reader/keyboard; ตรวจ high contrast, zoom, หลายจอ/DPI และ GDI fallback. Final สำหรับ Windows ต้องผ่านก่อน; platform อื่นมีเกณฑ์แยกเมื่อประกาศรองรับ.

## สิ่งที่พัฒนาในแอปทดสอบ Spreadsheet

แอปเป็นเจ้าของ UI เฉพาะงานทั้งหมด: worksheet grid composition, sheet tabs, formula bar, name box, ribbon/toolbar commands, cell/range selection, merge/freeze/autofill behavior, editor commit/cancel, paste special, validation, charts, query/pivot dialogs, page layout/print preview และสถานะการทำงานร่วมกัน. แอปเรียก generic API ของ Craft และสร้าง component เฉพาะใน repository ของแอปได้.

Spreadsheet Core และ library ของแอปเป็นเจ้าของสูตรทุกชนิด, workbook, file formats, import/export, analysis, pivot, automation, sync, permissions และ history. การแบ่งโมดูลภายในแอปไม่ทำให้สิ่งเหล่านี้กลายเป็นหน้าที่ของ Craft UI Engine.

## แผนที่ความสามารถจากไฟล์อ้างอิง

ทุกหมวดด้านล่างเป็นรายการตรวจรับของ **แอป Spreadsheet Final** ไม่ใช่ backlog ของ Craft. คอลัมน์ UI คือหน้าจอ/interaction ที่แอปสร้างด้วย Craft; คอลัมน์ Logic คือ Core, library หรือ integration ของแอป. Craft รับผิดชอบเฉพาะ generic primitives ในหัวข้อก่อนหน้า. รายการเหล่านี้ไม่ใช่คำกล่าวว่าปัจจุบันมีแล้ว.

| หมวด | UI ที่แอปสร้างด้วย Craft | Logic / library / service ของแอป |
| --- | --- | --- |
| A. ตารางและเซลล์ | Worksheet Grid; cell และ multi-range selection; editor; resize/AutoFit; wrap; freeze/split; hide/group; autofill handle; find/Go To UI; clipboard/undo UI | Insert/delete cells, rows, columns; merge; hide/group model; AutoFill/Flash Fill; find/replace; Go To Special; copy/cut/paste, Paste Special และ command history |
| B. การจัดรูปแบบ | Font, fill, border, alignment แนวนอน/ตั้ง, rotation, wrap, data bars, icon sets, theme และ preview | Number/currency/date/custom formats; cell styles; Format Painter; conditional rules/color scales; clear formatting |
| C. Workbook/Worksheet | หลาย sheet tabs, tab color, view, multiple windows, hyperlink/comment/note controls | New/rename/move/copy/hide sheets, template, named ranges, protection, workbook views และ lifecycle |
| D. สูตร | Formula Bar, reference highlight, spill display, auditing/error UI และ Evaluate Formula view | Parser/evaluator, relative/absolute/cross-sheet/external references, named formulas, dynamic arrays, calculation modes, dependencies, circular reference และ LAMBDA |
| E. ข้อมูล | Table controls, sort/filter UI, dropdown validation, slicer/timeline และ refresh state | Structured references, single/multi sort, Auto/Advanced Filter, remove duplicates, validation, Text to Columns, consolidate, subtotal, Quick Analysis และ linked data types |
| F. กราฟ | Chart/sparkline/shape/image/text-box components; axes, legend, labels, trendline และ error-bar interaction | ชุดข้อมูลและการคำนวณสำหรับ bar/line/pie/area/scatter/histogram/waterfall/funnel/treemap/stock/surface/radar/combo/map/PivotChart |
| G. Pivot/Data Model | Pivot field panel, drag/drop, drill-down view, slicers และ refresh feedback | Grouping/aggregation/calculated fields, PivotTable/PivotChart, relationships, data model, measures/KPI และทางเลือกที่ประกาศสำหรับ Power Pivot/DAX |
| H. นำเข้า/แปลงข้อมูล | Import wizard, Query Editor, preview, progress, cancel และ connection UI | CSV/TXT, XLSX/XLS, XML/JSON, database/web connectors, merge/append, pivot/unpivot, type conversion, refresh และ query parameters |
| I. วิเคราะห์เชิงลึก | Scenario, Forecast และผลวิเคราะห์ในรูปตาราง/กราฟ | Goal Seek, What-If tables, Solver, trend/regression, descriptive statistics และเครื่องมือเสริมที่ประกาศ |
| J. Automation | Macro/action recorder UI, editor, add-in management และผลการรัน | Custom functions, automation runtime, extension contract; VBA, Office Scripts/TypeScript, Python, Power Automate และ Copilot ต้องระบุว่าเป็นการรองรับตรง, compatibility layer, integration หรือทางเลือกของ Craft |
| K. พิมพ์/ส่งออก | Print Preview, page layout/break UI, margins/header/footer/print titles และ progress | Print area/orientation/paper/scaling, export PDF/CSV, save XLSX/XLSB ตาม format ที่ประกาศ |
| L. ร่วมงาน/ความปลอดภัย | Presence, comments/@mentions, changes, read-only/protected state และ accessibility checker UI | Sharing, co-authoring, autosave, version history, encryption, permissions และ sync service |

### สูตรและความเข้ากันได้

Spreadsheet Core ต้องมีทะเบียนฟังก์ชันและผลตรวจรับแยกตามกลุ่มในไฟล์อ้างอิง: mathematical, statistical, logical, lookup/reference, text, date/time, financial, database, information, engineering, dynamic arrays, lambda/functional, cube, web และ compatibility. ตัวอย่างเช่น `SUM`, `IF`, `XLOOKUP`, `FILTER` และ `LAMBDA` เป็นกรณีตรวจ; ไม่ใช่รายการครบทุกสูตร. ต้องประกาศรายฟังก์ชันว่า supported, partial หรือ unsupported พร้อม semantics, error behavior, locale และกรณี round-trip ของสูตรในไฟล์. การมี Formula Bar โดยไม่มี evaluator ไม่ถือว่าผ่านหมวด D.

### ขนาดข้อมูลเป้าหมาย

ไฟล์อ้างอิงให้ตัวเลขศึกษา: 1,048,576 แถว × 16,384 คอลัมน์ต่อ sheet, 32,767 อักขระต่อเซลล์ และสูตรยาว 8,192 อักขระ. ใช้เป็น **เป้าหมาย compatibility ของแอป** ไม่ใช่ข้อจำกัดที่ Craft UI Engine ต้องฝังไว้. แอป Final ต้องเปิด workbook แบบ sparse ที่มีพิกัดถึงขอบดังกล่าวโดยไม่จัดสรรหน่วยความจำสำหรับ 17,179,869,184 เซลล์ทั้งหมด; Craft ต้องให้ viewport/พิกัดที่ไม่บังคับสร้าง UI ตามจำนวนเซลล์. ต้องแยกขีดจำกัด UI, Core และ file format และบันทึกกรณีที่ต่ำกว่าเป้าหมายอย่างชัดเจน.

## หลักฐานการผ่านรอบ Final

1. มีโปรแกรม Spreadsheet ที่ติดตั้งและเปิดได้ พัฒนาด้วย Craft และ public UI API. Showcase อย่างเดียวไม่ผ่าน. ระบุ revision ของ Craft, package, แอป และชุดข้อมูลที่ใช้ตรวจ.
2. มี traceability matrix ระดับ **ทุกฟีเจอร์ในหมวด A–L** และ **ทุกกลุ่มสูตร**: owner, public API, สถานะ supported/partial/unsupported, วิธีทดสอบ, ผลและข้อจำกัด. การกล่าวว่า Final ครอบคลุมไฟล์อ้างอิงทั้งหมดต้องไม่มีรายการ partial/unsupported ที่ถูกนับเป็นผ่าน.
3. ทดสอบ workflow จริงอย่างน้อย: สร้าง/บันทึก/เปิด workbook หลายชีต; ป้อนสูตรอ้างอิงข้ามชีตแล้ว recalculate; sort/filter; เลือกและ copy/paste ช่วง; undo/redo; สร้าง chart; print preview/export; นำเข้าข้อมูล; และปิดเปิดใหม่โดยข้อมูลไม่สูญหาย.
4. ตรวจ file round-trip ด้วย workbook ที่มีค่า สูตร รูปแบบ และหลายชีต. เปรียบเทียบผลก่อน/หลังโดยแยกความถูกต้องของข้อมูลออกจากความสามารถรักษาฟีเจอร์ Excel ที่ยังไม่รองรับ; ห้ามทิ้งข้อมูลเงียบ ๆ.
5. ตรวจ scale กับ sparse workbook ที่มีพิกัดไกลและข้อมูลจำนวนมาก: node/หน่วยความจำต้องโตตาม viewport หรือข้อมูลจริง ไม่ตามจำนวนเซลล์สูงสุด. เก็บค่า frame time, input-to-draw, scroll, memory, CPU และ queue backlog บนเครื่องที่ระบุสเปก; เป้าหมายเริ่มต้นคือ scroll ใกล้ 60 FPS และแก้เซลล์ทั่วไปตอบสนองใน 16–50 ms โดยต้องปรับเกณฑ์ตามผลวัดจริงและบันทึกเหตุผล.
6. ตรวจ IME ภาษาไทย, keyboard-only, screen reader, DPI 100%/125%/หลายจอ, GDI fallback, large text, ความล้มเหลวของไฟล์/เครือข่าย และ cancellation. ผู้ใช้เป็นผู้ทดสอบ UI จริงตาม workflow ของโครงการ; ผล static check หรือภาพ mock ใช้แทนไม่ได้.

## เกณฑ์ประกาศผล

- **Craft UI Engine Final:** แอป Spreadsheet ใช้ public API สร้าง interaction, scale, accessibility และ generic components ได้ โดยไม่เพิ่มสูตร, Workbook, XLSX หรือ UI เฉพาะ Spreadsheet ลงใน Craft. ถ้าติดขัด ต้องระบุ primitive ทั่วไปที่ขาดและพิสูจน์ด้วยกรณีใช้งานอีกประเภทหนึ่งด้วย.
- **แอป Spreadsheet Final ตามไฟล์อ้างอิง:** แอป/Core/library/integration ต้องผ่าน traceability matrix ทั้ง A–L, กลุ่มสูตร, file/print และ integration ที่ประกาศไว้. ความสามารถที่ต้องใช้สิทธิ์หรือบริการภายนอกต้องมีการเชื่อมต่อจริงหรือระบุว่าไม่รองรับ; ไม่อ้างว่าทดแทน Microsoft service ได้จาก UI เพียงอย่างเดียว.
- การผ่านระดับแรก **ไม่เท่ากับ** การผ่านระดับที่สอง. ทั้งสองระดับต้องมีผลตรวจรับที่เผยแพร่ควบคู่กับเวอร์ชันที่ออกจริง.

เอกสารนี้เป็นเป้าหมายระยะ Final. แผน [Rev.32](phase1/craft-phase1-rev32.md) บันทึกช่องว่างสถาปัตยกรรมและงานใกล้ตัว; ไม่ได้หมายความว่าจะ implement ทุกฟีเจอร์ข้างต้นใน 0.1.32.
