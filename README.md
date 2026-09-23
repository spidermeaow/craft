# Craft 0.1.11 — Phase 1 Rev.11

Craft คือภาษาโปรแกรมแบบ static typing พร้อม CLI และ AST interpreter ที่เขียนด้วย Go เหมาะสำหรับทดลองเขียนโปรแกรม เครื่องมือ command-line และเรียนรู้แนวคิดของภาษาแบบมี type checking ตั้งแต่ก่อนรัน โปรแกรม Craft เขียนในไฟล์ `.craft` และสั่งงานผ่านคำสั่ง `craft` บน Windows

รีลีส Rev.11 เพิ่มการติดตั้ง source library จาก public GitHub โดยระบุ tag, แคชแยกรุ่น และ lockfile ที่ตรึง commit/checksum เพื่อให้การ build ทำซ้ำได้ อ่าน [คู่มือ GitHub packages](docs/REV11-PACKAGES.md), [เริ่มทดลอง](docs/TRY-REV11.md) และ [สถานะ Rev.11](roadmap/phase1/output/phase1-rev11-status.md)

## Craft ทำอะไรได้บ้าง

- ตรวจ syntax, type, scope และกฎการเปลี่ยนค่า (`let` / `var`) ก่อนรันด้วย `craft check`
- รันโปรเจกต์, test blocks และจัดรูปแบบ source ด้วย `craft run`, `craft test` และ `craft fmt`
- สร้าง source bundle ที่พกพาไปใช้กับ Craft CLI อื่นได้ด้วย `craft build`
- ใช้ standard library สำหรับ JSON, เวลา, filesystem, path, environment, HTTP และฐานข้อมูลตามเอกสารของแต่ละ revision
- ติดตั้ง package จาก GitHub และล็อก dependency ด้วย `craft install` และ `craft package lock`

สิ่งที่ควรทราบ: `craft build` สร้าง bundle ไม่ใช่ native executable ของแอปที่เขียนด้วย Craft; การรัน bundle ยังต้องมี Craft CLI อยู่ และ Craft ยังไม่ใช่ sandbox สำหรับรันโค้ดที่ไม่น่าเชื่อถือ

## ติดตั้งและเริ่มใช้

สำหรับ Windows x64 ให้ใช้ `Craft-setup.exe` หรือ `craft.exe` จากไฟล์ release ของเวอร์ชันที่ต้องการ แล้วเปิด Terminal ใหม่; ไฟล์ build ใน `dist/` ไม่ถูก commit ลง source repository เพื่อไม่ให้ประวัติ Git มี binary artifacts หากต้องการสร้าง CLI จาก source ให้ดูหัวข้อ [การพัฒนา Craft](#การพัฒนา-craft)

```powershell
craft version
craft new hello
cd hello
craft check
craft run
craft test
craft fmt
```

ผลที่คาดหวัง: เวอร์ชัน `0.1.11`, `Hello World` และ starter test ผ่าน 1 รายการ ผู้ใช้ไม่ต้องติดตั้ง Go ตัว compiler, interpreter, standard library, database drivers และ templates ฝังใน `craft.exe`

ใช้ `dist/craft.exe` แบบ portable ได้เช่นกัน Installer รองรับ Windows x64 แบบ per-user มี user PATH, Start Menu, Open With สำหรับ `.craft` และ uninstaller

**การเปลี่ยนจาก 0.1.0:** `let` เปลี่ยนค่าไม่ได้แล้ว ตัวนับหรือค่าที่ต้องแก้ไขต้องใช้ `var` ดู [Migration Note](docs/REV1-MIGRATION.md)

## ตัวอย่าง

```craft
func main() {
    defer {
        print("done")
    }

    var numbers: Int[] = [1, 2, 3]
    numbers.append(4)
    for number in numbers {
        print(number)
    }

    try {
        print(numbers[99])
    } catch error {
        print(error.message)
    }
}
```

## คำสั่ง

| คำสั่ง | การทำงาน |
| --- | --- |
| `craft version` | เวอร์ชันและ execution engine |
| `craft new <directory>` / `craft init` | สร้างโปรเจกต์และ starter test โดยไม่เขียนทับไฟล์เดิม |
| `craft run [bundle] [-- args...]` | ตรวจและรันโปรเจกต์หรือ bundle |
| `craft check` | ตรวจ `src/` และ `tests/` โดยไม่รัน |
| `craft build [--release]` | สร้าง `dist/<name>.craftbundle` format3/language0.1.11 พร้อม dependency sources |
| `craft install [github.com/owner/repo@v1.2.3]` | เพิ่ม GitHub dependency หรือ restore ตาม lock; รองรับ --alias, --offline, --refresh และ --recover ตามคู่มือ |
| `craft package check` | ตรวจ dependency graph, imports, cache และ lockfile โดยไม่ดาวน์โหลด |
| `craft package list` | แสดง package ที่ resolve แล้วตามลำดับคงที่ |
| `craft package lock` | เขียน `craft.lock` พร้อม path, version และ checksum |
| `craft clean` | ลบเฉพาะ bundle ของโปรเจกต์ |
| `craft fmt [--check] [file.craft]` | จัดรูปแบบทั้งโปรเจกต์หรือไฟล์เดียว รักษา comments และ AST |
| `craft test` | รัน test blocks โดยไม่เรียก `main` พร้อมผลและตำแหน่งข้อผิดพลาด |
| `craft help` | วิธีใช้ |

`craft build` ยังสร้าง source bundle ที่ต้องใช้ Craft รัน ไม่ใช่ native executable ของโปรแกรมผู้ใช้ ส่วน CLI เองเป็น executable ที่ไม่ต้องติดตั้ง runtime เพิ่ม

## เอกสารและตัวอย่าง

- [GitHub packages Rev.11](docs/REV11-PACKAGES.md), [วิธีทดลอง](docs/TRY-REV11.md) และ `examples/rev11-github`
- [Logging/config/errors Rev.10](docs/REV10-TOOLING.md)
- [ทดลอง Rev.8 และ local package](docs/TRY-REV8.md)
- [Package manifest, lockfile และคำสั่ง](docs/REV8-PACKAGES.md)
- [แผน Rev.9: safe tooling primitives](roadmap/phase1/craft-phase1-rev9.md) และ [คู่มือ filesystem/path](docs/REV9-TOOLING.md)
- [ทดลอง Rev.9](docs/TRY-REV9.md)
- `examples/rev9-tools`: ตรวจ JSON และเขียน output แบบ atomic
- `examples/rev8-packages`: application และ library package แบบ local
- [สถานะ Rev.8](roadmap/phase1/output/phase1-rev8-status.md)
- [ทดลอง Rev.5: เชื่อมต่อฐานข้อมูล](docs/TRY-REV5.md)
- [Database API, ownership, types และข้อจำกัด](docs/REV5-DATABASE.md)
- [สถานะ Rev.5 และผลทดสอบฐานข้อมูลจริง](roadmap/phase1/output/phase1-rev5-status.md)
- `examples/rev5-database`: ตัวอย่าง SELECT/parameters ที่ไม่แก้ข้อมูล พร้อม Craft tests 3 รายการ
- [ใบอนุญาต dependencies ที่รวมใน CLI](docs/third-party/README.md)
- [เริ่มทดลอง Rev.4/HTTP และ package reference](docs/TRY-REV4.md)
- [Modules, function values และ HTTP contracts](docs/REV4-LANGUAGE.md)
- [สถานะ Rev.4: ผู้ใช้ทดลองผ่าน 7/7 และเริ่ม HTTP server ได้แล้ว พร้อมรายการตรวจรับเพิ่มเติม](roadmap/phase1/output/phase1-rev4-status.md)
- [Reference package ที่เขียนด้วย Craft](packages/api-framework/README.md) และ `examples/api-server` (ไม่ใช่ release gate ของภาษา)
- [เริ่มทดลอง Rev.3 และคำสั่งตรวจรับ](docs/TRY-REV3.md)
- [Timer/Multi-task contracts](docs/REV3-LANGUAGE.md)
- [สถานะ Rev.3](roadmap/phase1/output/phase1-rev3-status.md)
- [Craft Language Support และ Dark/Light themes](craft-vscode/README.md)
- `examples/rev3-tasks`: ตัวอย่างงานพร้อมกัน, timers, cancellation และ Craft tests 4 รายการ
- `examples/rev2-user-acceptance`: กู้ชุดผู้ใช้ 15 tests + starter 1 พร้อมข้อมูล provenance
- [เริ่มทดลอง Rev.2 และคำสั่งตรวจรับ](docs/TRY-REV2.md)
- [Rev.2 language/Standard Library contracts](docs/REV2-LANGUAGE.md)
- [สถานะ Rev.2: implementation และผลตรวจรับแยกกัน](roadmap/phase1/output/phase1-rev2-status.md)
- `examples/rev2-bank`: Console Bank ฝาก/ถอน/ประวัติ/บันทึกและโหลด JSON
- `examples/rev2-features`: feature demo และ Craft tests 12 รายการ
- `examples/rev2-cli`: word count, config JSON และวันเวลา
- [Craft file icon](craft-file-icon-theme/README.md): VSIX 0.2.0 เพิ่มไอคอน .craft โดยไม่แทน file icon theme เดิม; Language Support 0.2.2 มีไอคอนนี้ด้วย


- [คู่มือทดลองใช้ Rev.1](docs/TRY-CRAFT.md)
- [กติกาภาษา](docs/LANGUAGE.md)
- [การย้ายโค้ดจาก 0.1.0](docs/REV1-MIGRATION.md)
- [รากฐานและฟีเจอร์ที่วางไว้สำหรับ Rev.2](docs/REV1-FOUNDATIONS.md)
- [สถานะการส่งมอบ Rev.1](roadmap/phase1/output/phase1-rev1-status.md)
- ตัวอย่าง `examples/rev1` มี exception, array, defer, named arguments และ Craft tests 6 รายการ
- ตัวอย่างเดิม `examples/hello`, `examples/language-tour` อัปเดตตามกฎภาษาใหม่แล้ว
- ตัวอย่างที่ตั้งใจให้ error: `examples/type-error`, `examples/runtime-error`, `examples/immutable-error`, `examples/unhandled-exception`

## การพัฒนา Craft

ใช้ Go ตาม `go.mod` (1.26.6 ขึ้นไป) และ Inno Setup สำหรับสร้าง installer; Node.js/npm สำหรับ build VSIX เท่านั้น (ผู้ใช้ปลายทางไม่ต้องติดตั้ง):

```powershell
go build -o dist/craft.exe ./cmd/craft
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 -Installer -Icons -Language
```

สคริปต์ release ปิด CGO สร้าง Windows amd64 CLI/installer และ `dist/SHA256SUMS.txt` โดยไม่รัน tests ใช้ `-IsccPath 'C:\...\ISCC.exe'` หากต้องระบุ compiler เอง

รัน tests จาก repository:

```powershell
go test ./...
pwsh -NoProfile -File .\scripts\test-rev8.ps1
go test ./internal/lexer -fuzz=FuzzScan -fuzztime=10s
go test ./internal/parser -fuzz=FuzzParse -fuzztime=10s
```

ผลตรวจรับและข้อจำกัดของแต่ละ revision บันทึกแยกใน `roadmap/phase1/output/` การ build สำเร็จอย่างเดียวไม่ใช่ผลรับรอง runtime หรือ installer

## สถาปัตยกรรม

`lexer` → `parser/ast` → `resolver/types` → validated call argument ordering → `interpreter` → `runtime/stdlib`

- `internal/resolver`: lexical symbols และ mutability
- `internal/types`: type checking, control-flow outcomes, return safety, named argument validation
- `internal/runtime`: Array value semantics และ Exception values
- `internal/stdlib`: built-in signatures/implementations แยกจาก Interpreter
- `internal/interpreter`: function frames, exception propagation, stack traces และ defer unwinding
- `internal/formatter`: comment-preserving formatting พร้อม AST comparison
- `internal/project`, `internal/cli`: discovery, bundles และ developer commands

Module/import, function values และ buffered HTTP server อยู่ใน Rev.4 แล้ว; database primitives อยู่ใน Rev.5 ส่วน ORM/migrations, Enum, lambda/closure, Event library, process/REPL, native backend และ VM ยังเป็น backlog. Format Document/Problems/semantic services มี [แผน protocol ระดับ B](docs/EDITOR-PROTOCOL.md) ที่ยังไม่เปิดใช้
