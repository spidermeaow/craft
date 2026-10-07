# Craft GUI package 0.2.8

`gui` เป็น Craft library สำหรับประกอบ native UI บน `std.ui` ของ Craft Runtime **0.1.30 ขึ้นไป** บน Windows. แอปเขียนด้วย Craft โดยไม่ใช้ WebView หรือ Go ในโค้ดแอป. Runtime จัดการหน้าต่าง, layout, event และการวาด; package ให้ API ที่เรียกง่ายและองค์ประกอบใช้ซ้ำ.

Package นี้เป็นหัวข้อพัฒนาใน **Craft Phase 1 Rev.30**. เวอร์ชัน `0.2.8` ใช้กับ dependency ของ package; `craft version` แสดงเวอร์ชัน CLI/Runtime `0.1.30`. ดู [โครงสร้างและเวอร์ชันของทั้งโครงการ](../../docs/PROJECT-STRUCTURE.md).

## ติดตั้ง

คัดลอกโฟลเดอร์นี้เป็น `<project>/packages/gui` แล้วเพิ่มใน `craft.toml`:

```toml
[dependencies]
gui = { path = "packages/gui", version = "0.2.8" }
```

รัน `craft package lock` ในโปรเจกต์ แล้วใช้ `import gui "gui"`. ตัวอย่าง `examples/gui-package-lab` อ้าง `../../packages/gui` โดยตรง. ยังไม่มีคำสั่ง `craft gui init`.

## API ที่เปิดใช้

| กลุ่ม | ฟังก์ชัน |
| --- | --- |
| Window/event | `openWindow`, `setTitle`, `setBackground`, `closeWindow`, `nextEvent`, `mount`, `update`, `renderer`, `viewportWidth`, `viewportHeight`, `measureTextWidth` |
| Layout | `row`, `column`, `panel`, `grid`, `size`, `maxSize`, `align`, `spacing`, `splitView`, `scrollView`, `virtualList` |
| Native controls | `text`, `button`, `textInput`, `setText`, `textStyle`, `textWrap`, `surface` |
| Paint/Canvas | `color`, `fillRect`, `strokeRect`, `drawLine`, `drawText`, `drawTextStyled`, `hitRegion`, `targetPolicy`, `drawImage`, `canvas` |
| Focus/command | `focus`, `focusPolicy`, `accessible`, `pushFocusScope`, `popFocusScope`, `bindShortcut`, `unbindShortcut`, `bindScopedShortcut`, `unbindScopedShortcut`, `commandInvoked` |
| Composition | `card`, `toolbar`, `navigation`, `field`, `heading`, `primaryButton`, `surfaceCard`, `appShell`, `adaptiveShell`, `targetButton` |
| Helper/tokens | `compact`, `responsiveGrid`, `autoGrid`, `viewportChanged`, `targetPressed`, `primaryColor`, `foregroundColor`, `surfaceColor`, `borderColor`, `smallSpacing`, `mediumSpacing`, `largeSpacing` |

กลุ่ม wrapper สี่แถวแรกส่งต่อ signature ของ Runtime. การเรียก `std.ui` ภายใน implementation ของ `gui` เป็นส่วนของ package; app source เรียก `gui` ได้โดยตรง. ชนิด `UiWindow`, `UiNode`, `UiDraw`, `UiEvent` ยังเป็นชนิดที่ Runtime เปิดให้ Craft ใช้. Token สีเป็นค่าเริ่มต้น; แอปส่งสีของตนเข้า `surfaceCard`, `textStyle`, `surface` และ `targetButton` ได้. Token ระยะห่างคืนค่า 4/8/12 logical pixels; app override โดยเรียก `spacing` ด้วยค่าที่ต้องการ.

`field(id, ...)` สร้าง ID `<id>-label` และ `<id>-input`. `appShell(id, header, sidebar, content, sidebarWidth)` สร้าง `<id>-body`; caller ต้องใช้ ID ที่ไม่ซ้ำ. `targetButton` คืน `UiDraw[]` (พื้นหลัง, ข้อความ, hit region) ให้ app append เข้า Canvas; กำหนด width มากกว่า 16 และ height มากกว่า 12. คำสั่งสุดท้ายเป็น hit region เพื่อให้ target อยู่บนสุด. `gui.targetPressed(event, id)` ตรวจ `CanvasTargetPressed` กับ ID นั้น; pointer/Enter/Space ของ target ใช้ ID เดียวกันตาม event ที่ Runtime ส่ง. Canvas target ที่ระบุชื่อผ่าน `targetPolicy` ถูก expose ผ่าน UI Automation provider ขั้นแรก.

`gui.compact(gui.viewportWidth(window), breakpoint)` ช่วยเลือก layout. เมื่อได้ `Resized` หรือ `DpiChanged` ให้ app อ่าน viewport ใหม่ แล้วเรียก `gui.update` เฉพาะเมื่อขนาดหรือ layout ที่ต้องใช้เปลี่ยน. ถ้า content กว้างกว่าหน้าต่าง ให้จัดใหม่หรือห่อ `scrollView`; Canvas ขนาดคงที่ไม่ reflow เอง. App ยังเป็นเจ้าของ state และ routing ของ event. `nextEvent(-1)` รอ, `nextEvent(0)` poll, ค่าบวกเป็น timeout.

## ย้ายจาก 0.2.2

เปลี่ยน dependency เป็น `0.2.8` และล็อกใหม่. API 0.2.2 ยังอยู่. ถ้า app เรียก `std.ui.viewportWidth/Height`, `textStyle`, `surface`, `drawTextStyled` หรือ `hitRegion` โดยตรง ให้เปลี่ยน prefix เป็น `gui.` โดยไม่เปลี่ยน arguments. ดู `craft-test/gui-package-dashboard` สำหรับตัวอย่างย้ายจาก direct Runtime calls. `craft-test` เป็น test app แยก; สำเนา package ในโปรเจกต์นั้นต้องตรงกับ source รุ่นที่ใช้.

## ขอบเขต

- Windows หนึ่งหน้าต่างต่อ interpreter run; `craft build` สร้าง source bundle ที่ยังต้องมี Craft Runtime.
- TextInput ใช้ native Windows EDIT เพื่อรองรับ IME/caret/selection แต่ปรับรูปลักษณ์ได้จำกัด.
- รูป PNG/JPEG ใช้ path จาก working directory; source bundle ไม่ฝัง asset.
- ยังไม่มี reactive state อัตโนมัติ, animation, popup/dialog, table/grid, accessibility ครบ, shadow, gradient หรือ CSS styling.
- Source รุ่น 0.2.8 ผ่าน static `craft package check` และ `craft check` ในตัวอย่างแล้ว; GUI/IME/UIA จริงยังรอผู้ใช้ทดสอบ.

## Focus และ shortcut ใน Rev.30

`gui.focusPolicy(buttonOrInput, tabStop, enabled)` กำหนดการเดิน Tab และ disabled state. `gui.focus(window, id)` ขอ focus ไปยัง Button, TextInput หรือ Canvas hit region ที่มีอยู่. Tab/Shift+Tab เดินตามลำดับ tree; Button และ Canvas target มี focus ring. `gui.accessible(node, name, role)` เก็บ semantic name/role ใน tree โดย role รองรับ `text`, `button`, `textbox`, `group`. Runtime มี UI Automation provider ขั้นแรกสำหรับ Button, Text, กลุ่มที่มีชื่อ และ Canvas target ที่มีชื่อ; TextInput ใช้ Windows EDIT. พฤติกรรมจริงกับ screen reader ยังรอผู้ใช้ทดสอบ.

`gui.bindShortcut(window, commandId, virtualKey, modifiers)` ผูกคำสั่งระดับหน้าต่าง. Modifiers คือ Shift=1, Ctrl=2, Alt=4; Ctrl+S คือ key 83/mod 2 และ F5 คือ key 116/mod 0. Runtime ส่ง `CommandInvoked` ที่ `event.nodeId` เป็น commandId และไม่ยิงซ้ำจาก key repeat. `gui.commandInvoked` ช่วยตรวจ event; `gui.unbindShortcut` ลบด้วย commandId. เมื่อ TextInput มี focus, Ctrl+A/C/V/X/Z/Y และคีย์ระหว่าง IME composition ยังเป็นของ native EDIT. App จัดการ state และเรียก `update` เอง.

## เพิ่มใน Rev.30

`gui.pushFocusScope(window, subtreeId)` จำกัด focus ไว้ใน subtree และ `gui.popFocusScope` คืน focus เดิมเมื่อยังใช้ได้. `gui.targetPolicy(region, tabStop, enabled, name)` กำหนด policy ให้ Canvas target แต่ละรายการ. `gui.bindScopedShortcut(window, subtreeId, commandId, key, modifiers)` ผูกคำสั่งกับ scope; scope ที่ active มาก่อน window binding. `gui.unbindScopedShortcut` ลบคำสั่งจาก scope.

UIA provider รุ่นแรกยังไม่มี Text/Selection/Table patterns และ hit testing จากพิกัดยังคืน root element. Narrator, Accessibility Insights, DPI และ IME ยังต้องตรวจจริงก่อนถือว่าผ่านเกณฑ์ Rev.30.

## Responsive layout ใน Rev.28

ใช้ `gui.textWrap(gui.text(...))` เพื่อให้ Text เพิ่มความสูงตามความกว้างที่ parent จัดให้. `gui.autoGrid(id, items, minCellWidth, maxColumns, gap)` คำนวณคอลัมน์จากความกว้างที่ Runtime จัดให้ โดยไม่ต้องหัก sidebar/padding ในแอป. `gui.responsiveGrid` เดิมยังใช้ได้. `gui.adaptiveShell` สลับระหว่าง Sidebar กับเมนูบนตาม breakpoint ที่แอปส่งให้; app ยังเป็นเจ้าของ state/event และเรียก `gui.update` หลัง resize/DPI change. Canvas พิกัดคงที่ไม่ reflow เอง. ดู [คู่มือ Rev.28](../../docs/REV28-UI-REFLOW.md).




