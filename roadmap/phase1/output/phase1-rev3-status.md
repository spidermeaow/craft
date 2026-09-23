# Craft Phase 1 Rev.3 — Implementation status

วันที่: 18 กันยายน 2026
CLI source: 0.1.3 / Craft Language Support: 0.1.0 / File Icons: 0.1.0

สถานะ: implementation งาน A และ runtime อยู่ใน source; user acceptance ยัง pending
Agent ไม่รัน tests, Craft examples, installer หรือ VS Code UI ตาม preference ผู้ใช้

## Implemented / source-reviewed

- Declarative language extension: registration, grammar, Unicode boundaries,
  String/comment precedence, range/exponent, types, functions, namespaces,
  user type declarations/annotations, brackets/comments/indentation และ snippets
- Craft Dark/Light original palettes, token gallery, install/select/uninstall guide
- CLI fmt typo hint เฉพาะ path check ที่ไม่มีจริง โดยยังคืน error และไม่เขียนไฟล์
- Task/Timer types, named Void callback resolution, positional signature checking,
  isolated machines, deep-copy arguments, shared handles, synchronized output
- spawn/join/cancel/status, function-owned cleanup, failure propagation, wait-cycle
  และ self/ancestor rejection, bounded task count, cooperative cancellation
- Duration sleep พร้อม Int overload เดิม, after/every fixed delay, monotonic waits,
  callback failure/cancellation และ input guard
- Source bundle language 0.1.3 พร้อมอ่าน 0.1.1/0.1.2 เดิม, CLI/installer version ใหม่
- Examples, runtime contracts, user test guide, historical Rev.1 notice
- กู้ 15 tests จากไฟล์ผู้ใช้ + starter 1; helper functions สร้างกลับจาก assertions
  เพราะไม่พบ source เดิม จึงไม่อ้างว่า fixture ที่กู้ใหม่นี้ผ่านแล้ว

## Build evidence

- `go build ./...` สำเร็จหลังเพิ่ม runtime; ไม่ใช่ผล automated tests
- `scripts/build.ps1 -Installer -Language` สำเร็จ: สร้าง `dist/craft.exe` 0.1.3,
  `dist/craft-language-support-0.1.0.vsix`, `dist/Craft-setup.exe` และ SHA256SUMS
  พร้อมแนบ File Icons VSIX เดิม ไม่ได้ติดตั้งหรือรัน artifacts เพื่อทดสอบ

## Tests prepared, not run by Agent

- `internal/interpreter/tasks_test.go`: manual clock, independent waits, fixed delay,
  cancellation/defer, zero timer, copied arguments, frame lifetime, errors/limits/cycles
- `tests/rev3_test.go`: static rejection, example tests, formatter, bundle และ fmt hint
- `internal/stdlib/rev2_test.go`: overlapping console readers และ cancelled session
- `craft-vscode/tests/grammar.test.cjs`: TextMate/Oniguruma scopes, Unicode boundaries,
  incomplete string recovery, fixture tokenization และ palette contrast assertions
- `examples/rev3-tasks/tests/tasks.craft`: 4 tests สำหรับ API/copy/failure/cancel/duration
- ชุด Rev.1/Rev.2 เดิมเก็บไว้; CLI version assertion อัปเดตเป็น 0.1.3

## Pending user acceptance

- [ ] Go tests/regression และ Craft tests ใหม่ผ่าน
- [ ] Demo, format/test-after-format และ bundle บน CLI 0.1.3 ผ่าน
- [ ] ชุดผู้ใช้ที่กู้กลับ 16 tests และ Rev.2 examples ผ่าน
- [ ] Ctrl+C, blocking input และ cleanup ตรวจจริง
- [ ] Grammar tests/contrast และ VS Code visual acceptance ผ่าน
- [ ] VSIX coexistence/uninstall, installer install/upgrade/PATH/uninstall ผ่าน

## Level B and limits

Format Document/Problems/semantic services: ประเมิน protocol แล้วใน
[EDITOR-PROTOCOL](../../../docs/EDITOR-PROTOCOL.md); implementation เป็นงานแยกหลังงานสี
Enum/Module/Lambda/Event, cron/persistent schedules และ native program compilation
ยังไม่รวม runtime มีเพดาน 256 active tasks และ 256 owned handles ต่อ function call
ยกเลิกแบบ cooperative; file I/O บางชนิดและ raw reader ต้องรอ OS/เจ้าของ resource
รายละเอียด ownership, error precedence และ cleanup budget อยู่ใน
[REV3-LANGUAGE](../../../docs/REV3-LANGUAGE.md)

ผล Rev.2 ที่ผู้ใช้รายงานก่อนหน้านี้ยังอยู่ใน [Rev.2 status](phase1-rev2-status.md)
และไม่ใช้เป็นผลรับรอง Rev.3 ดูคำสั่งตรวจรับใน [TRY-REV3](../../../docs/TRY-REV3.md)
