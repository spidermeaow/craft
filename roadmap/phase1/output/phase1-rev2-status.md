# Craft Phase 1 Rev.2 — Implementation status

วันที่: 18 กันยายน 2026  
Toolchain: Craft 0.1.2  
Icon theme: Craft Forge File Icons 0.1.0

**สถานะ: implementation ขอบเขต A และการทดลองใช้งานหลักผ่านตามผลผู้ใช้; acceptance ทั้งหมดยังไม่ครบ**

## ผลล่าสุดที่ผู้ใช้รายงาน

- CLI 0.1.2: `craft check` สำเร็จ และ demo `craft run -- hello --flag` ทำงานจบ
- โปรเจกต์ `D:\02-Repository\craft-test`: **16 passed, 0 failed, exit 0** (Rev.2 จากบทสนทนา 15 tests + starter 1 test)
- Interactive input: ข้อความ/ตัวเลข/ช่องว่างทำงาน และ Ctrl+Z แล้ว Enter จบด้วย EOF
- `craft fmt` จัดรูปแบบ 2 ไฟล์ และ `craft fmt --check` แสดง Formatting is up to date
- Craft Forge File Icons 0.1.0 ติดตั้งใน VS Code สำเร็จ และผู้ใช้ยืนยันว่าใช้งานได้แล้ว

ผล 16 tests เกิดก่อน fmt ยังไม่มีผล test หลัง fmt หรือ build/run bundle ล่าสุด
ผลนี้ไม่ใช่ผล `go test ./...` และไม่ได้ยืนยันทุก icon mapping, ทุก color theme หรือ installer lifecycle

## Implemented

| กลุ่ม | สิ่งที่เพิ่ม |
| --- | --- |
| Console / CLI | write/readLine, UTF-8, EOF/blank line/CRLF, line limit, injected reader, cancellation, arguments หลัง -- ทั้ง project และ bundle |
| Conversion | toInt/toFloat/toBool/toString, tryInt/tryFloat, fixed formatting |
| Optional | T?, null, if let, null comparison, nested optionals สำหรับ Map ของ Optional |
| Struct | named constructors, typed let/var fields, deep copy, immutable path checks, recursive-required-field rejection |
| Map | String keys, get/require/set/remove/containsKey/keys/values/length, insertion order และ snapshots |
| Collections | contains/indexOf/slice/reverse/sorted, Unicode substring, replace และ join |
| Math / Random | round/floor/ceil/pow, copyable seeded Random พร้อม nextInt/nextFloat |
| Files / Path | append, directories/listing, regular-file remove/copy/move แบบ no-overwrite, native path helpers |
| Date / Time | typed DateTime/Duration, RFC3339 parse/format, fixed offsets, Unix units, arithmetic และ overflow checks |
| JSON | JsonValue, parse/stringify, constructors/accessors, Int64 precision, duplicate-key rejection, limits |
| Tooling | formatter map/struct/optional support, project struct merging, CLI help/version, bundle language 0.1.2 และอ่าน Rev.1 format 2 |
| Icon theme | VS Code manifest, generator จาก theme.json, SVG เดิม 6 ไฟล์, generic fallbacks, mapping report, fixtures และ local VSIX |
| Examples | rev2-bank, rev2-features, rev2-cli พร้อม Craft tests |
| Docs | API contracts, migration/limitations, user acceptance guide, README และ installer guide |

กฎ API ที่สรุปแล้วอยู่ใน [Rev.2 contracts](../../../docs/REV2-LANGUAGE.md)
การออกแบบที่ต้องเลือกจากข้อเสนอ เช่น DateTime.format แบบ RFC3339Nano, Random value state, nested Optional และ filesystem no-overwrite ระบุไว้ในเอกสารนี้ครบ

## Build artifacts

ส่งมอบใน `dist/`:

- `craft.exe` — portable Windows x64 CLI 0.1.2
- `Craft-setup.exe` — per-user installer; มี CLI, docs, examples และ VSIX ใน editor/ (ไม่ติดตั้ง extension ให้อัตโนมัติ)
- `craft-forge-file-icons-0.1.0.vsix` — local VS Code extension ใช้ MIT ตามผู้ใช้ยืนยัน; ไม่มีการเผยแพร่ Marketplace
- `SHA256SUMS.txt` — checksums ของ artifacts

คำสั่ง build: `go build ./...` และ `powershell -ExecutionPolicy Bypass -File scripts/build.ps1 -Installer -Icons`
การ build/packaging ไม่เท่ากับการรัน tests หรือตรวจพฤติกรรมจริง

## Repository tests prepared — Agent has not run them

Rev.3 เพิ่มสำเนาชุดผู้ใช้ที่กู้จาก `craft-test/tests/full.craft` ใน
`examples/rev2-user-acceptance` พร้อม starter รวม 16 tests และ helper ที่สร้างกลับ
จาก assertions ดู README ของ fixture; ยังไม่มีผลรันชุดที่กู้กลับนี้

- `tests/rev2_test.go`: runtime scenarios, static rejection, reader/args project+bundle, formatter semantics/idempotence และ example projects
- `internal/stdlib/rev2_test.go`: input size/EOF/cancellation, temporary filesystem/no-overwrite, JSON limits/precision, conversion/time boundaries
- `examples/rev2-features/tests/full.craft`: 12 feature tests
- `examples/rev2-bank/tests/account.craft`: 2 tests สำหรับ persistence round-trip และ invalid balance
- `examples/rev2-cli/tests/words.craft`: 1 test สำหรับ word count
- ชุด Rev.1 เดิมคงไว้ และปรับ CLI version expectation เป็น 0.1.2

**Agent ไม่รัน automated tests, Craft examples, installer หรือ VS Code UI ตาม preference ของผู้ใช้**
ผล 39 passed ของ Rev.1 ที่ผู้ใช้เคยรายงานไม่ใช่หลักฐานว่า Rev.2 ผ่าน
ผล Rev.2 ที่ผู้ใช้รายงานล่าสุดแสดงไว้ด้านบน และเป็นคนละชุดกับ feature tests 12 รายการใน repository

## Awaiting user acceptance

- [ ] `go test ./...` รวม regression/negative cases ผ่าน
- [x] ผู้ใช้รายงาน check/run และ 16 tests ก่อน fmt ผ่าน พร้อม test exit 0; fmt/--check สำเร็จ
- [ ] test หลัง fmt และ build/run bundle ของ Rev.2 ผ่าน พร้อม exit codes
- [ ] Bank ฝาก/ถอน/invalid input/overflow/EOF/Ctrl+C และ save/load ตรวจด้วยข้อมูลจริง
- [ ] 39 tests ใน project ส่วนตัวเดิมผ่านด้วย CLI 0.1.2
- [ ] Installer install/upgrade/PATH/uninstall บน Windows x64 ที่ไม่มี Go ผ่าน
- [x] ผู้ใช้ติดตั้ง VSIX ใน VS Code และยืนยันใช้งานเบื้องต้นแล้ว
- [ ] ตรวจ icons/fallback/light/dark/high-contrast/uninstall ครบ พร้อมบันทึก VS Code version

ขั้นตอนและผลที่คาดหวัง: [TRY-REV2](../../../docs/TRY-REV2.md)
รายการเหล่านี้ต้องมีผลจริงก่อนประกาศว่า Definition of Done ของ roadmap ผ่านครบ

## Deferred / limitations

- ระดับ B: Enum, Module/import/visibility, Function values, Lambda/Callback และ Event library
- ระดับ C: HTTP client, process execution และ REPL
- ไม่มี automatic Struct JSON serialization, named timezone database, atomic file persistence หรือ native program compilation
- File icon theme ใช้ immediate parent mappings; `*_test.craft` นอก tests/ fallback เป็น Craft และ nested tests/unit ไม่ match tests โดยอัตโนมัติ
- JetBrains/Craft editor/CLI icon adapters และ Marketplace publishing ไม่อยู่ใน release นี้
- ข้อมูลเจ้าของลิขสิทธิ์ไม่ได้รับชื่อจากผู้ใช้ จึงใช้ MIT โดยไม่เดาชื่อเจ้าของใน notice
- License/version ของ icon theme แยกจากภาษา; local publisher `craft-local` ไม่ได้อ้างว่าเป็น Marketplace account

## Migration

`let` ยังคง immutable ลึก, parameters immutable, compound assignments/errors/defer คงเดิม
`struct` และ `null` เป็น keywords ใหม่ หาก source เก่าใช้สองคำนี้เป็นชื่อให้เปลี่ยนชื่อ
bundle ใหม่ใช้ language 0.1.2 ต้องรันด้วย CLI ใหม่; 0.1.2 อ่าน bundle Rev.1 format 2 ได้ ส่วน format 1 ต้อง migrate/rebuild
