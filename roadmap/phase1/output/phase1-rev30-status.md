# Rev.30 source status (Craft 0.1.30 / GUI 0.2.8)

วันที่ 7 ตุลาคม 2026 — ลง source แล้ว; รอผู้ใช้ทดสอบพฤติกรรมจริง

## Runtime

- Focus scope แบบ stack สำหรับ subtree ID, restore focus และจำกัด Tab/programmatic focus ตาม scope
- Canvas target policy ต่อ hit region: tab stop, enabled และ accessible name; disabled target ไม่รับการกด
- Scoped shortcuts พร้อม conflict diagnostics และ precedence เหนือ window shortcuts ขณะ scope active
- Windows UI Automation provider ขั้นแรกสำหรับ custom Text, Button, named group และ named Canvas target; Invoke ส่งเข้า event queue เดิม; label ของ native EDIT ใช้ IAccPropServices
- Focus target ที่ไม่มี, disabled, อยู่นอก scope หรือไม่ materialized คืน error แยกสาเหตุ

## Package และตัวอย่าง

- `packages/gui` 0.2.8 เพิ่ม wrapper ของ scope, target policy และ scoped shortcuts; `field` ใส่ชื่อให้ EDIT และ `targetButton` ใส่ชื่อให้ Canvas target
- อัปเดต package copy, manifest และ lockfile ของ `gui-package-dashboard` กับ `responsive-tester`; dashboard มี scope, scoped Ctrl+Shift+S และ disabled target
- CLI และ Inno Setup เป็น 0.1.30; Setup อยู่ที่ `dist/Craft-setup.exe`

## หลักฐานที่ตรวจแล้ว

- `go build` สำเร็จ
- `craft package check` และ `craft check` สำเร็จสำหรับ `examples/gui-package-lab`, `gui-package-dashboard`, `responsive-tester`
- Inno Setup compile สำเร็จ

ไม่ได้รัน automated tests หรือเปิดแอป GUI ตามความต้องการของผู้ใช้. UIA provider รุ่นแรกยังคืน root สำหรับ hit testing จากพิกัด และยังไม่มี Text/Selection/Table patterns. Narrator, IME, DPI, focus ring/scroll และ Setup หลังติดตั้งต้องให้ผู้ใช้ทดสอบจริงก่อนปิด acceptance ของ Rev.30.

