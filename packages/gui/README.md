# Craft GUI package 0.1.45 candidate

`gui` เป็น Craft library สำหรับประกอบ native UI บน `std.ui` ของ Craft Runtime **0.1.45** บน Windows. แอปเขียนด้วย Craft โดยไม่ใช้ WebView หรือ Go ในโค้ดแอป. Runtime จัดการหน้าต่าง, layout, event และการวาด; package ให้ API ที่เรียกง่ายและองค์ประกอบใช้ซ้ำ.

GUI ที่มากับ Craft ใช้เวอร์ชันเดียวกับ CLI/Runtime คือ `0.1.45`. Rev.32 แก้การจัดข้อความของ Styled Button และ Canvas ใน Runtime โดยไม่เปลี่ยน API ของ package. ดู [แผน Rev.32](../../roadmap/phase1/craft-phase1-rev32.md) และ [โครงสร้างของโครงการ](../../docs/PROJECT-STRUCTURE.md).

Rev.33 เพิ่ม `gui.virtualGrid` สำหรับ viewport ตารางสองแกนแบบ prototype. ดู [รูปแบบข้อมูลและข้อจำกัด](../../docs/REV33-UI-VIRTUAL-GRID.md); ตารางนี้ยังไม่มี editor, selection หลายช่อง หรือ virtual accessibility.

Rev.34 เพิ่ม `gui.gridSelection` และ `GridSelectionMoved` สำหรับเลือกเซลล์เดียวและเดินด้วยคีย์บอร์ด. ดู [คู่มือ Rev.34](../../docs/REV34-UI-VIRTUAL-GRID-SELECTION.md).

Rev.35 ปรับการย้ายและวาด native TextInput ภายใน ScrollView ที่ Runtime โดยไม่เปลี่ยน public API ของ `gui`; ดู [แผน Rev.35](../../roadmap/phase1/craft-phase1-rev35.md).

Rev.36 เพิ่ม `gui.gridEditor` สำหรับวาง TextInput หนึ่งตัวบนเซลล์ VirtualGrid และแก้การคืน native focus; ดู [คู่มือ Rev.36](../../docs/REV36-UI-VIRTUAL-GRID-EDITOR.md). Rev.37 เพิ่ม `gridRangeSelection`, `clipboardReadText` และ `clipboardWriteText`; ดู [คู่มือ Rev.37](../../docs/REV37-UI-VIRTUAL-GRID-RANGE.md).

Rev.39 เพิ่ม Canvas path commands: `pathMoveTo`, `pathLineTo`, `pathArcTo`, `pathQuadraticTo`, `pathCubicTo`, `pathClose`, `fillPath` และ `strokePath`. `pathArcTo(cx, cy, radius, startDegrees, sweepDegrees)` ใช้มุมเป็นองศา; sweep ติดลบคือทิศย้อนกลับ. ประกอบหลาย subpath ใน array เดียวเพื่อสร้างรูตรงกลางด้วย even-odd fill เช่น Donut. สี path ในรอบนี้ต้องทึบ (`alpha=255`) เพื่อให้ Direct2D/GDI แสดงรูปทรงสอดคล้องกัน. ดู [คู่มือ Rev.39](../../docs/REV39-UI-RENDERING.md).

Rev.40 candidate เพิ่ม `setWindowIcon(window, path)` สำหรับไฟล์ `.ico` บน Windows. เรียกหลัง `openWindow` ได้ทันทีหรือเปลี่ยนระหว่างทำงาน; path อ้างจาก working directory. Runtime ตั้งไอคอนของหน้าต่างและคืน native handles เมื่อเปลี่ยนหรือปิด. ไฟล์หายหรือรูปแบบไม่ใช่ `.ico` ให้ข้อผิดพลาดโดยคงไอคอนเดิม. API นี้ไม่ฝังไอคอนใน executable. ดู [Final gate Rev.40](../../roadmap/phase1/craft-phase1-rev40-final-gate.md).

Rev.42 candidate เพิ่มให้ `surface` รับ `Text` node ได้ และให้ `align` ควบคุมการจัดข้อความแนวนอนทั้ง Direct2D/GDI. Rev.43 เปลี่ยนเฉพาะเวอร์ชัน package ให้ตรงกับ toolchain; API ของ GUI ไม่มีการเปลี่ยนในรอบนี้.

## ติดตั้ง

คัดลอกโฟลเดอร์นี้เป็น `<project>/packages/gui` แล้วเพิ่มใน `craft.toml`:

```toml
[dependencies]
gui = { path = "packages/gui" }
```

Craft เติมเวอร์ชัน GUI จาก toolchain ให้อัตโนมัติและตรวจว่า package ที่ path นั้นตรงกัน. รัน `craft package lock` ในโปรเจกต์ แล้วใช้ `import gui "gui"`. ตัวอย่าง `examples/gui-package-lab` อ้าง `../../packages/gui` โดยตรง. ยังไม่มีคำสั่ง `craft gui init`.

## API ที่เปิดใช้

| กลุ่ม | ฟังก์ชัน |
| --- | --- |
| Window/event | `openWindow`, `maximizeWindow`, `setTitle`, `setWindowIcon`, `setBackground`, `closeWindow`, `nextEvent`, `mount`, `update`, `renderer`, `viewportWidth`, `viewportHeight`, `measureTextWidth` |
| Layout/grid | `row`, `column`, `panel`, `grid`, `size`, `maxSize`, `align`, `spacing`, `splitView`, `scrollView`, `virtualList`, `virtualGrid`, `gridSelection`, `gridRangeSelection`, `gridRangeStyle`, `gridEditor` |
| Clipboard | `clipboardReadText`, `clipboardWriteText` |
| Native controls | `text`, `button`, `textInput`, `inputStyle`, `setText`, `textStyle`, `textWrap`, `surface` |

| Paint/Canvas | `color`, `fillRect`, `strokeRect`, `drawLine`, `drawText`, `drawTextStyled`, `hitRegion`, `targetPolicy`, `drawImage`, `canvas` |
| Focus/command | `focus`, `focusPolicy`, `accessible`, `pushFocusScope`, `popFocusScope`, `bindShortcut`, `unbindShortcut`, `bindScopedShortcut`, `unbindScopedShortcut`, `commandInvoked` |
| Composition | `card`, `toolbar`, `navigation`, `field`, `heading`, `primaryButton`, `secondaryButton`, `surfaceCard`, `appShell`, `adaptiveShell`, `targetButton` |
| Helper/tokens | `compact`, `responsiveGrid`, `autoGrid`, `viewportChanged`, `targetPressed`, `primaryColor`, `foregroundColor`, `surfaceColor`, `borderColor`, `smallSpacing`, `mediumSpacing`, `largeSpacing` |
| Theme/form | `Theme`, `lightTheme`, `darkTheme`, `themedHeading`, `themedButton`, `themedField`, `fieldMessage` |

## Theme และ Form ที่ประกอบใช้ได้

`lightTheme()` และ `darkTheme()` คืนค่า `Theme` ที่แอปเก็บไว้และส่งให้ component แต่ละตัว ไม่มี global theme ที่เปลี่ยน UI ของแอปเดิมโดยอัตโนมัติ. แอปสร้าง `Theme(...)` เองได้เพื่อกำหนดสี, มุมโค้ง, ความสูง control และระยะห่าง. ค่าสีของ `input`, `surface` และ `border` ต้องทึบตามข้อกำหนดของ native TextInput; `radius` ของ input ต้องอยู่ในช่วง 0–24.

```craft
let theme = gui.lightTheme()
let title: UiNode = gui.themedHeading("title", "Contact", theme)
let name: UiNode = gui.themedField("name", "Full name", "", "Enter a name", 360, true, theme)
let hint: UiNode = gui.fieldMessage("name-hint", "Required field", true, theme)
let save: UiNode = gui.themedButton("save", "Save", true, true, theme)
```

`themedButton` รับ `primary` และ `enabled`; เมื่อปิดใช้งานจะใช้สีข้อความรองและไม่รับ focus. `themedField` รับความกว้างจากแอป แทนการล็อก 640 pixels แบบ `field` เดิม. Input มี ID `<id>-input` สำหรับ `TextChanged` และ `setText`; label มี ID `<id>-label`. `fieldMessage` ใช้กับข้อความช่วยเหลือหรือ validation error โดยแอปเป็นผู้กำหนดข้อความและสถานะ. ฟังก์ชันเดิม `primaryButton`, `secondaryButton`, `heading` และ `field` ยังใช้งานเหมือนเดิม.

กลุ่ม wrapper สี่แถวแรกส่งต่อ signature ของ Runtime. การเรียก `std.ui` ภายใน implementation ของ `gui` เป็นส่วนของ package; app source เรียก `gui` ได้โดยตรง. ชนิด `UiWindow`, `UiNode`, `UiDraw`, `UiEvent` ยังเป็นชนิดที่ Runtime เปิดให้ Craft ใช้. Token สีเป็นค่าเริ่มต้น; แอปส่งสีของตนเข้า `surfaceCard`, `textStyle`, `surface` และ `targetButton` ได้. Token ระยะห่างคืนค่า 4/8/12 logical pixels; app override โดยเรียก `spacing` ด้วยค่าที่ต้องการ.

`field(id, ...)` สร้าง ID `<id>-label` และ `<id>-input`. `appShell(id, header, sidebar, content, sidebarWidth)` สร้าง `<id>-body`; caller ต้องใช้ ID ที่ไม่ซ้ำ. `targetButton` คืน `UiDraw[]` (พื้นหลัง, ข้อความ, hit region) ให้ app append เข้า Canvas; กำหนด width มากกว่า 16 และ height มากกว่า 12. คำสั่งสุดท้ายเป็น hit region เพื่อให้ target อยู่บนสุด. `gui.targetPressed(event, id)` ตรวจ `CanvasTargetPressed` กับ ID นั้น; pointer/Enter/Space ของ target ใช้ ID เดียวกันตาม event ที่ Runtime ส่ง. Canvas target ที่ระบุชื่อผ่าน `targetPolicy` ถูก expose ผ่าน UI Automation provider ขั้นแรก.

`gui.compact(gui.viewportWidth(window), breakpoint)` ช่วยเลือก layout. เมื่อได้ `Resized` หรือ `DpiChanged` ให้ app อ่าน viewport ใหม่ แล้วเรียก `gui.update` เฉพาะเมื่อขนาดหรือ layout ที่ต้องใช้เปลี่ยน. ถ้า content กว้างกว่าหน้าต่าง ให้จัดใหม่หรือห่อ `scrollView`; Canvas ขนาดคงที่ไม่ reflow เอง. App ยังเป็นเจ้าของ state และ routing ของ event. `nextEvent(-1)` รอ, `nextEvent(0)` poll, ค่าบวกเป็น timeout.

## ย้ายจาก 0.2.2

ลบ `version` ออกจาก dependency แล้วล็อกใหม่ด้วย Craft 0.1.37. API 0.2.2 ยังอยู่. ถ้า app เรียก `std.ui.viewportWidth/Height`, `textStyle`, `surface`, `drawTextStyled` หรือ `hitRegion` โดยตรง ให้เปลี่ยน prefix เป็น `gui.` โดยไม่เปลี่ยน arguments. ดู `craft-test/gui-package-dashboard` สำหรับตัวอย่างย้ายจาก direct Runtime calls. `craft-test` เป็น test app แยก; สำเนา package ในโปรเจกต์นั้นต้องตรงกับ source รุ่นที่ใช้.

## ขอบเขต

- Windows หนึ่งหน้าต่างต่อ interpreter run; `craft build` สร้าง source bundle ที่ยังต้องมี Craft Runtime.
- TextInput ใช้ native Windows EDIT เพื่อรองรับ IME/caret/selection แต่ปรับรูปลักษณ์ได้จำกัด.
- รูป PNG/JPEG ใช้ path จาก working directory; source bundle ไม่ฝัง asset.
- ยังไม่มี reactive state อัตโนมัติ, animation, popup/dialog, virtualized accessibility ครบ, shadow, gradient หรือ CSS styling.
- Source รุ่น 0.1.37 ยังรอผู้ใช้รัน `craft package check`, `craft check` และตรวจ GUI/IME/UIA จริง.

## Focus และ shortcut (เพิ่มใน Rev.30)

`gui.focusPolicy(buttonOrInput, tabStop, enabled)` กำหนดการเดิน Tab และ disabled state. `gui.focus(window, id)` ขอ focus ไปยัง Button, TextInput หรือ Canvas hit region ที่มีอยู่. Tab/Shift+Tab เดินตามลำดับ tree; Button และ Canvas target มี focus ring. `gui.accessible(node, name, role)` เก็บ semantic name/role ใน tree โดย role รองรับ `text`, `button`, `textbox`, `group`. Runtime มี UI Automation provider ขั้นแรกสำหรับ Button, Text, กลุ่มที่มีชื่อ และ Canvas target ที่มีชื่อ; TextInput ใช้ Windows EDIT. พฤติกรรมจริงกับ screen reader ยังรอผู้ใช้ทดสอบ.

`gui.bindShortcut(window, commandId, virtualKey, modifiers)` ผูกคำสั่งระดับหน้าต่าง. Modifiers คือ Shift=1, Ctrl=2, Alt=4; Ctrl+S คือ key 83/mod 2 และ F5 คือ key 116/mod 0. Runtime ส่ง `CommandInvoked` ที่ `event.nodeId` เป็น commandId และไม่ยิงซ้ำจาก key repeat. `gui.commandInvoked` ช่วยตรวจ event; `gui.unbindShortcut` ลบด้วย commandId. เมื่อ TextInput มี focus, Ctrl+A/C/V/X/Z/Y และคีย์ระหว่าง IME composition ยังเป็นของ native EDIT. App จัดการ state และเรียก `update` เอง.

## Focus scope และ Canvas target (เพิ่มใน Rev.30)

`gui.pushFocusScope(window, subtreeId)` จำกัด focus ไว้ใน subtree และ `gui.popFocusScope` คืน focus เดิมเมื่อยังใช้ได้. `gui.targetPolicy(region, tabStop, enabled, name)` กำหนด policy ให้ Canvas target แต่ละรายการ. `gui.bindScopedShortcut(window, subtreeId, commandId, key, modifiers)` ผูกคำสั่งกับ scope; scope ที่ active มาก่อน window binding. `gui.unbindScopedShortcut` ลบคำสั่งจาก scope.

UIA provider รุ่นแรกยังไม่มี Text/Selection/Table patterns และ hit testing จากพิกัดยังคืน root element. Narrator, Accessibility Insights, DPI และ IME ยังต้องตรวจจริง.

## เพิ่มใน Rev.31

`gui.maximizeWindow(window)` ขยายหน้าต่างด้วยกลไก Windows หลัง `openWindow`; ใช้ก่อนหรือหลัง `mount` ได้. `gui.secondaryButton` ใช้สีรองในชุดเดียวกับ `primaryButton`. ปุ่มยังใช้ native focus, keyboard และ UI Automation. กรอบ focus ของ Button และ Canvas target แสดงเมื่อเดินด้วย Tab แต่ไม่ค้างหลังคลิกเมาส์.

`gui.inputStyle(input, fill, border, radius)` ใช้กับ TextInput โดยตรง; รองรับสีพื้นและสีขอบแบบทึบ และ radius 0–24. Runtime ใช้สีพื้นเดียวกันทั้งกรอบและ native EDIT รวมถึงเมื่อปิดใช้งานช่องกรอก. Native EDIT ยังเป็นตัวรับข้อความและ IME. `gui.field` ใช้ style นี้และจำกัดความกว้าง input ที่ 640 logical pixels. สีขอบเมื่อ focus กำหนดโดย Runtime; แอปส่งสีขอบเพื่อแสดงข้อผิดพลาดเมื่อไม่ได้ focus ได้. การปรับสี placeholder และ TextInput หลายบรรทัดยังไม่รองรับ.

## Responsive layout ใน Rev.28

ใช้ `gui.textWrap(gui.text(...))` เพื่อให้ Text เพิ่มความสูงตามความกว้างที่ parent จัดให้. `gui.autoGrid(id, items, minCellWidth, maxColumns, gap)` คำนวณคอลัมน์จากความกว้างที่ Runtime จัดให้ โดยไม่ต้องหัก sidebar/padding ในแอป. `gui.responsiveGrid` เดิมยังใช้ได้. `gui.adaptiveShell` สลับระหว่าง Sidebar กับเมนูบนตาม breakpoint ที่แอปส่งให้; app ยังเป็นเจ้าของ state/event และเรียก `gui.update` หลัง resize/DPI change. Canvas พิกัดคงที่ไม่ reflow เอง. ดู [คู่มือ Rev.28](../../docs/REV28-UI-REFLOW.md).




