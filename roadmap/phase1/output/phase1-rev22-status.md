# Craft Phase 1 Rev.22 status

วันที่: 7 ตุลาคม 2026  
Source version: 0.1.22  
สถานะ: implement ใน source; รอผู้ใช้ทดสอบ UI จริง

## สิ่งที่เพิ่ม

- `std.ui.textInput(id, initialValue, placeholder)` และ `std.ui.setText(window, id, value)` สำหรับช่องข้อความบรรทัดเดียว.
- `UiEvent.text`, `TextChanged`, `TextSubmitted` และ `FocusChanged`; focus เดินด้วย Tab/Shift+Tab ระหว่าง Button กับ TextInput.
- เพิ่ม unit checks ใน source สำหรับขีดจำกัด Unicode/ความยาวของ input และลำดับ focus; ยังไม่ได้รันตามความต้องการให้ผู้ใช้ทดสอบเอง.
- Windows EDIT control ภายใน UI tree เพื่อรับ IME, caret, selection และ clipboard; engine จัดตำแหน่ง, clipping, DPI และคงค่าเดิมเมื่อ update tree ด้วย ID เดิม.
- UI Input Lab ใน `examples/gui-input` และ `D:\02-Repository\craft-test\runtime\11-ui-input`, คู่มือ `docs/REV22-UI-INPUT.md` และ `docs/TRY-REV22.md`.
- ปรับเวอร์ชัน CLI, source bundle language และ Windows installer เป็น 0.1.22. HTTP Engine และ package ที่ Archive ไม่เปลี่ยน.

## การตรวจและข้อจำกัด

Go source build สำหรับ Windows amd64 และการ compile `dist/Craft-setup.exe` ผ่าน; installer version metadata เป็น 0.1.22 และอัปเดต `dist/SHA256SUMS.txt` แล้ว. **ยังไม่ได้รัน automated tests หรือเปิด UI เพื่อทดสอบ** ตามความประสงค์ของผู้ใช้; การตรวจรับภาพ, ภาษาไทย/IME, focus, clipboard, scroll, DPI และ regression ต้องให้ผู้ใช้ทดสอบตามคู่มือ. การ build ผ่านไม่ยืนยันว่าการพิมพ์จริงถูกต้อง.

ช่องข้อความยังใช้หน้าตาของ native EDIT control และยังไม่มี Craft accessibility provider, multi-line input หรือ styling ของ input. งานเหล่านี้ยังไม่ถือว่าเสร็จใน Rev.22. อย่ารวมงาน HTTP เข้ากับการตรวจรับ UI รอบนี้.
