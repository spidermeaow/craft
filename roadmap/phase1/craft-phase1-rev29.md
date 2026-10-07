# Craft Phase 1 Rev.29 (0.1.29) — Focus, Commands และ Accessibility ขั้นพื้นฐาน

วันที่: 7 ตุลาคม 2026  
สถานะ: **implement บางส่วน; รอผู้ใช้ทดสอบ**  
ฐาน: Craft UI Runtime 0.1.28 และ `packages/gui` 0.2.6

## เป้าหมาย

ให้แอป Windows ที่เขียนด้วย Craft ใช้คีย์บอร์ดได้ต่อเนื่อง: มองเห็นจุด focus, เดิน focus ไปข้างหน้า/ย้อนกลับได้ถูกต้อง, เรียกคำสั่งของแอปด้วย shortcut โดยไม่แย่งคีย์ของ TextInput และเปิดข้อมูลพื้นฐานให้เครื่องมือ accessibility. Package ต้องให้วิธีประกาศ command และผูกกับปุ่ม/เมนูที่อ่านง่าย โดยแอปยังเป็นเจ้าของ state และการทำงานเมื่อเกิด command.

Rev.29 เป็น **รอบฐานของ desktop interaction** ไม่เพิ่ม `State<T>`, DataGrid หรือชุด form controls ใหม่. งานเหล่านั้นใช้สัญญา focus/command/semantic ที่กำหนดในรอบนี้เป็นฐาน.

## สถานะก่อน Rev.29

- Runtime มี `FocusChanged`, `KeyPressed`/`KeyReleased`, Tab/Shift+Tab และ Enter/Space สำหรับ Button กับ Canvas target บางกรณีแล้ว. `TextInput` เป็น Windows EDIT และมีการส่ง key event กลับแอป.
- ลำดับ focus ยังอิงลำดับ tree; ยังไม่มี API ประกาศ focus scope, skip/disabled, focus เริ่มต้น หรือการย้าย focus โดยโปรแกรมที่มีสัญญาครบ.
- ยังไม่มี command/shortcut registry ที่ระบุ priority เมื่อ TextInput, menu, dialog และ app ใช้คีย์เดียวกัน. Custom nodes ยังไม่มี accessibility tree ที่ screen reader อ่านได้ครบ.

## งาน Runtime 0.1.29

### 1. Focus ที่คาดเดาได้

- กำหนดสัญญา focusable, focus order, Tab/Shift+Tab, focus เมื่อ mount/update/remove node, focus เมื่อหน้าต่างกลับมา active และการเลื่อน ScrollView ให้ target ที่ focus มองเห็นได้. ลำดับเริ่มต้นใช้ลำดับ tree; การ override ต้องระบุชัดและไม่ทำให้ target ที่ซ่อนอยู่รับ focus.
- เพิ่ม primitive สำหรับขอ focus ไปยัง node ID และสำหรับตัด node ออกจาก traversal หรือกำหนด disabled state. คืน error ที่ระบุเหตุผลเมื่อ ID ไม่อยู่ใน tree, มองไม่เห็น, disabled หรือไม่ focusable.
- กำหนด focus ring ที่มองเห็นได้บน Button และ Canvas target ทั้ง Direct2D และ GDI; native TextInput ใช้สัญญาณ focus ของ Windows. Ring ต้องไม่ถูก parent clipping หรือ scroll overlay บังอย่างผิดพลาด.
- กำหนด focus scope ขั้นต่ำสำหรับ composition อย่าง Dialog/Popup ในอนาคต: entry, restore เมื่อปิด และการไม่ปล่อย Tab ไปยังพื้นหลัง. Rev.29 อาจเปิดเฉพาะ primitive/scope metadata โดยยังไม่มี Dialog widget.
- แยกการนำทางด้วย arrow keys ภายใน composite control ออกจาก Tab traversal ระหว่าง control; ยังไม่ทำ roving focus ของ DataGrid/TreeView เต็มรูปแบบ.

### 2. Keyboard shortcut และ command dispatch

- เพิ่มการลงทะเบียน shortcut เป็นข้อมูลที่ผูกกับ window/Ui tree และ command ID. Runtime ส่ง event `CommandInvoked` ที่มี ID เดียว ไม่ส่งซ้ำเพราะ key repeat โดยไม่ตั้งใจ; ชื่อ API/signature สุดท้ายต้องตรวจ Craft checker ก่อน implement.
- รองรับคีย์เดี่ยวที่เหมาะกับคำสั่ง เช่น F5 และชุด modifiers อย่าง Ctrl+S, Ctrl+Shift+S. จัดลำดับเมื่อ shortcut ซ้ำกัน, scope ซ้อนกัน หรือ target ถูก disabled/hidden และรายงานความขัดแย้งที่ตรวจได้.
- ให้ TextInput/IME เป็นเจ้าของคีย์แก้ไขข้อความและ composition ก่อน app command: Ctrl+C/X/V/A/Z, caret movement, Enter ตามโหมด input และคีย์ที่ IME ใช้ต้องไม่ถูก shortcut ทั่วไปแย่ง. ระบุวิธี opt-in สำหรับ override เฉพาะที่จำเป็นโดยไม่เปลี่ยนค่าเริ่มต้นนี้.
- คง `KeyPressed`/`KeyReleased` เดิมเพื่อ compatibility แต่บันทึกลำดับ event เมื่อ command ถูกจับและเหตุใด key จึงถูก consume. คีย์ลัดทำงานเฉพาะหน้าต่าง active; ไม่ทำ global hotkey ของ Windows.

### 3. Accessibility ขั้นพื้นฐาน

- เพิ่ม semantic metadata ขั้นต่ำ: accessible name, role, value/state และคำอธิบายที่จำเป็นให้ Text, Button, TextInput และ Canvas target. ค่าเริ่มต้นใช้ label ที่มีอยู่เมื่อเหมาะสม; custom target ต้องระบุชื่อที่อ่านได้.
- เปิดข้อมูลของ control ที่ Runtime วาดเองผ่าน Windows UI Automation provider ขั้นแรก: ลำดับ tree, bounding rectangle, focus, enabled state และ Invoke pattern สำหรับ Button/Canvas target. TextInput native EDIT ใช้ความสามารถ accessibility ของ Windows และต้องเชื่อมความสัมพันธ์กับ label ของ Craft เมื่อกำหนดไว้.
- เมื่อ tree เปลี่ยน ให้ UI Automation เห็น name/state/focus ใหม่อย่างถูกต้องโดยไม่คืน node ที่ถูกลบ. ตรวจว่าการขอข้อมูลจาก assistive technology ไม่ค้าง UI thread หรือเปิด native handle ค้าง.
- กำหนดขอบเขตชัด: Rev.29 ยังไม่รับรอง Text pattern เต็มรูปแบบสำหรับ custom text, selection/table patterns, high-contrast theme ทั้งระบบ หรือ accessibility ครบทุกชนิด control.

## งาน GUI Package 0.2.7

- ห่อ focus/semantic primitives ของ Runtime ด้วย API ที่ตั้งชื่อสม่ำเสมอ และเพิ่ม helper ประกาศ command/shortcut ด้วย command ID. Package ไม่เรียก callback ของ Craft บน UI thread; แอปรับ `CommandInvoked` ใน event loop แล้วจัดการผลเอง.
- ให้ `primaryButton`, `navigation`, `appShell` และ `adaptiveShell` ตั้งชื่อ/role เริ่มต้นที่ถูกต้อง และแสดง disabled/focus state ตาม Runtime. หาก composition สร้าง child IDs เพิ่ม ต้องอธิบายลำดับ focus ที่เกิดขึ้น.
- เพิ่มตัวอย่าง File/Edit command เช่น Save, Refresh และ Undo โดยทำให้ Ctrl+S/F5 ใช้งานได้ และแสดงว่า Ctrl+Z ใน TextInput ยังเป็นคำสั่งแก้ข้อความของ control.
- คง wrapper และ component ของ package 0.2.6; สำเนา package ใน `gui-package-dashboard` และ `responsive-tester`, manifest และ lockfile ต้องตรงกับ source เมื่อเริ่ม implement.

## สัญญาที่ต้องสรุปก่อนเขียนโค้ด

1. นิยามชนิดข้อมูลและ signature ของ focus request, focus policy, shortcut และ semantic metadata ด้วยชนิดที่ Craft 0.1.28 รองรับจริง; อย่าตั้งชื่อ API ในแผนเป็นสัญญาที่เผยแพร่ก่อนผ่าน checker.
2. ระบุ priority ของ key event ระหว่าง Windows EDIT/IME, focus navigation, command ของ scope ปัจจุบัน, command ของ window และ raw key event. ระบุ repeat/keyup และกรณี modifier ค้างเมื่อหน้าต่างเสีย focus.
3. ระบุ ownership ของ UI Automation provider, thread ที่ตอบ query, lifetime เมื่อ mount/update/close และการแจ้ง focus/state change ก่อน implement.
4. แยก semantic role กับ visual style: การเปลี่ยนสีหรือ Canvas hit region เพียงอย่างเดียวไม่ถือว่า accessible.

## ลำดับดำเนินงาน

1. เก็บ baseline ด้วยแอปที่มี Button, TextInput, ScrollView และ Canvas target: Tab/Shift+Tab, Enter/Space, Ctrl+C/V/Z และการ resize/update ที่มี focus อยู่.
2. ลง focus contract, focus ring, focus request/disabled/hidden policy และ focus visibility ใน ScrollView.
3. ลง shortcut/command dispatch ที่ไม่แย่ง EDIT/IME พร้อม event และ conflict diagnostics.
4. ลง semantic metadata และ Windows UI Automation provider ขั้นแรก; ตรวจด้วย keyboard-only และ screen reader/เครื่องมือ UIA บน Windows.
5. ห่อ API ใน Package 0.2.7, ปรับตัวอย่าง, คู่มือ migration, version/lockfile และสร้าง Setup หลัง source พร้อม.

## เกณฑ์ตรวจรับที่ผู้ใช้ทดสอบเอง

| หัวข้อ | ผลที่ต้องเห็น |
| --- | --- |
| Focus | Tab/Shift+Tab เดินไปข้างหน้า/ย้อนกลับอย่างคงที่, ring มองเห็น, ScrollView เลื่อนให้ target ปรากฏ, update/remove/disable คืน focus ตามสัญญา |
| Commands | Ctrl+S และ F5 ส่ง command ID ครั้งเดียวต่อการกดตามสัญญา; shortcut ซ้ำหรือ disabled ไม่ทำงานผิด target |
| TextInput/IME | Ctrl+C/X/V/A/Z, caret, selection และการพิมพ์ไทยยังทำงาน; shortcut ของแอปไม่แย่ง composition หรือ undo ของ TextInput |
| Accessibility | เครื่องมือ UI Automation หรือ Narrator อ่านชื่อ/role/state ของ Button และ Canvas target ได้, focus ที่รายงานตรงกับภาพและคีย์บอร์ด |
| Rendering | focus ring และ enabled state ตรงกันบน Direct2D/GDI ที่ DPI 100%/125% โดยไม่มีภาพค้างหลัง resize/scroll |
| Compatibility | ตัวอย่าง 0.1.28 และ package 0.2.6 ยังใช้ได้ หรือมี migration เฉพาะจุดที่จำเป็น |

การเปิด GUI และทดสอบจริงให้ผู้ใช้ทำเองตามความต้องการปัจจุบัน. ผล build/check และผลตรวจหน้าจอต้องรายงานแยกกัน; แผนนี้ยังไม่ใช่หลักฐานว่า feature ผ่านแล้ว.

## นอกขอบเขต Rev.29

ยังไม่เพิ่ม reactive `State<T>`, TextInput หลายบรรทัด, DataGrid, Menu/Dialog widget, global Windows hotkeys, macro system หรือ accessibility ที่สมบูรณ์ทุก pattern. Rev.29 สร้างฐานเพื่อให้งานเหล่านั้นมี focus/command/semantic contract เดียวกัน.
