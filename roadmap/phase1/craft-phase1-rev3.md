# Craft Phase 1 Rev.3

## VS Code Readability, Developer Experience, Timer System and Multi-task

สถานะ: implementation งาน A และ runtime อยู่ใน source; รอผลตรวจรับผู้ใช้ ดู [Rev.3 status](output/phase1-rev3-status.md)  
ฐาน: Craft 0.1.2 / Phase 1 Rev.2 และ Craft Forge File Icons 0.1.0  
หลักฐาน: source ใน repository และผลทดสอบที่ผู้ใช้ส่งในบทสนทนา

## 1. เป้าหมาย

ทำให้การอ่านและเขียนไฟล์ `.craft` ใน VS Code สบายตา แยก keyword, type, function, String, number และ comment ได้ชัดเจน พร้อมปรับคู่มือและข้อความผิดพลาดจากการทดลองใช้ Rev.2

ตีความคำขอเรื่องสีเป็น **สี source code ใน editor** เป็นหลัก ครอบคลุมสีพื้นหลังและ selection เมื่อเลือก Craft color theme ส่วนสี Terminal output เป็นงานต่อยอดแยก ไม่เปลี่ยนข้อความที่โปรแกรม print ในงานหลัก

ส่งมอบการปรับ editor ให้ใช้กับภาษา 0.1.2 ได้ก่อน ไม่ต้องรอฟีเจอร์ภาษาใหม่หรือ Language Server เต็มรูปแบบ

เพิ่ม **Timer System** สำหรับหน่วงเวลาและเรียกงานตามช่วงเวลา และ **Multi-task** สำหรับทำงานหลายงานร่วมกันเป็นขอบเขต runtime ของ Rev.3 โดยส่งมอบแยกจากงาน editor ทั้งสองระบบเพิ่มใน source รุ่น 0.1.3 ไม่ใช่ความสามารถที่มีแล้วใน Craft 0.1.2

## 2. สถานะที่ตรวจพบ

หัวข้อนี้บันทึก baseline ก่อนเริ่ม Rev.3; implementation ล่าสุดและข้อจำกัดดู [Rev.3 status](output/phase1-rev3-status.md) และ [runtime contracts](../../docs/REV3-LANGUAGE.md)

### มี implementation แล้ว

- Console input, program arguments, conversion, Optional/if let, Struct, Map และ deep-copy/let/var semantics
- Array/String utilities, Math/Random, File System/Path, DateTime/Duration และ structured JsonValue
- CLI check/run/test/fmt/build, portable executable และ Windows installer
- Craft Forge File Icons: SVG 6 แบบ, manifest กลาง, VS Code adapter, generic fallback และ VSIX
- ตัวอย่าง Bank, feature demo, CLI utilities และ Go/Craft test sources

### หลักฐานผู้ใช้ตรวจรับแล้ว

| รายการ | ผลที่รายงาน |
| --- | --- |
| craft version | 0.1.2 / Phase 1 Rev.2 / Windows amd64 |
| craft check | Check succeeded |
| craft run -- hello --flag | demo ทำงานจบ; arguments, copies, collections, conversion, date และ JSON ตรงกับสิ่งที่คาดหวัง |
| craft test ใน D:\02-Repository\craft-test | 16 passed, 0 failed, exit 0: Rev.2 15 tests + starter basic arithmetic 1 test |
| Interactive input | รับข้อความ/ตัวเลข, รักษาช่องว่าง และ Ctrl+Z แล้ว Enter จบด้วย EOF |
| craft fmt | จัดรูปแบบ main.craft และ full.craft |
| craft fmt --check | Formatting is up to date |
| VSIX | ผู้ใช้ติดตั้งใน VS Code สำเร็จและยืนยันว่าใช้งานได้แล้ว |

ชุด 15 tests ที่ผู้ใช้คัดลอกจากบทสนทนาไม่ใช่ชุด 12 tests ใน `examples/rev2-features` ต้องระบุแหล่งและจำนวนให้ตรงเมื่อบันทึกผล
ผล test ล่าสุดเกิดก่อน fmt จึงยังไม่ยืนยันว่า test หลัง fmt ผ่าน และการติดตั้ง VSIX สำเร็จยังไม่ครอบคลุมทุก mapping/theme/uninstall

### ยังไม่มีผลตรวจรับล่าสุด

- test หลัง fmt, build และรัน bundle ของ Rev.2
- piped input, blank input, Unicode input และ Ctrl+C ใน interactive mode (Unicode ใน demo ผ่านแล้ว)
- Go tests ทั้ง repository และชุด Rev.1 เดิม 39 tests บน CLI 0.1.2
- Bank save/load รอบใช้งานจริง, installer upgrade/uninstall บนเครื่องที่ไม่มี Go
- icon theme ครบทุก fallback, Light/Dark/High Contrast และถอน extension

### ช่องว่างที่พบจาก source และการใช้งาน

1. `craft-file-icon-theme/package.json` มี `contributes.iconThemes` แต่ไม่มี `languages`, `grammars` หรือ color themes จึงยังไม่ได้จัดสี source code ของ Craft
2. `craft fmt check` ถูกตีความเป็น path ชื่อ `check` และแสดง filesystem error ผู้ใช้แก้เป็น `craft fmt --check` แล้วใช้งานได้ ควรมี hint ที่ตรงกับความตั้งใจ
3. `phase1-rev2-status.md` เดิมยังไม่มีผล 16 tests และ VSIX ล่าสุด ต้องอัปเดตโดยแยกผลผู้ใช้จากงานที่ Agent build
4. เอกสาร foundation ของ Rev.1 เป็นประวัติการออกแบบ มีรายการ Struct/Map/Optional/JSON ที่ implement แล้วใน Rev.2 ต้องติดป้ายว่าเป็นเอกสารประวัติและชี้ไป reference ปัจจุบัน
5. มี DateTime/Duration สำหรับจัดการค่าเวลา แต่ยังไม่มี Timer System หรือ API สำหรับเริ่มและควบคุมหลาย task; worker goroutine ภายใน console input ไม่ถือเป็น Multi-task ของภาษา

## 3. ลำดับงานและขอบเขต

| ระดับ | งาน | เกณฑ์ |
| --- | --- | --- |
| A — งานหลัก | Craft Language Support + syntax highlighting | เปิด .craft แล้วรู้จักภาษาและแยก token ได้ทันที |
| A — งานหลัก | Craft Dark / Craft Light color themes | ผู้ใช้เลือกเอง อ่านง่าย และมีตัวอย่างเปรียบเทียบ |
| A — งานหลัก | Comment, bracket, indentation, folding และ snippets | ลดงานพิมพ์โดยตรงกับ syntax 0.1.2 |
| A — งานหลัก | VSIX, คู่มือ, CLI usage hints และอัปเดตสถานะ | ติดตั้งถูกโปรแกรมและตรวจรับได้ชัดเจน |
| B — ต่อจากงานสี | Format Document และ Problems integration | ต้องมี CLI/editor protocol ที่ปลอดภัยก่อน |
| B — ต่อจากงานสี | Semantic tokens, hover, completion, definition | อาศัย parser/resolver/type checker; ไม่ใช้ regex เดาความหมาย |
| R — Runtime Rev.3 | Multi-task และ task lifecycle | เริ่มงานหลายงาน รอผล ยกเลิก และส่งต่อข้อผิดพลาดได้ตาม contract ในหัวข้อ 10 |
| R — Runtime Rev.3 | Timer System | หน่วงเวลา ตั้งงานครั้งเดียว/วนซ้ำ และยกเลิกได้โดยไม่หยุดทุก task |
| C — แผนภาษาแยก | Enum, Module/import, Lambda/Callback/Event | ยังไม่บังคับใน Rev.3 งาน editor |
| C — แยกภายหลัง | HTTP/process/REPL, debugger, Marketplace publishing | ไม่รวมในงานหลัก |

## 4. Craft Language Support extension

สร้าง extension แยกจาก icon theme เพื่อให้ผู้ใช้ใช้สีโค้ดกับ file icons ที่ตนเลือกได้อย่างอิสระ

โครงสร้างที่เสนอ:

```text
craft-vscode/
  package.json
  language-configuration.json
  syntaxes/craft.tmLanguage.json
  themes/craft-dark-color-theme.json
  themes/craft-light-color-theme.json
  snippets/craft.code-snippets
  fixtures/highlighting.craft
  README.md
  CHANGELOG.md
  LICENSE
```

- ประกาศ language ID `craft`, alias `Craft`, extension `.craft`, scope `source.craft`
- ประกาศ `contributes.languages`, `grammars`, `themes` และ `snippets`
- `.craftbundle` เป็น JSON bundle ไม่ผูกกับ Craft source grammar; ไม่เปลี่ยน TOML/JSON associations เดิม
- รุ่นแรกเป็น declarative extension ทำสีได้โดยไม่ต้องมี Craft CLI/Go หรือรันโปรแกรมใน workspace
- ไม่เพิ่ม telemetry, activation command หรือแก้ settings ของผู้ใช้อัตโนมัติ
- กำหนด extension ID/version/minimum VS Code version แยกจาก CLI และ icon theme พร้อมระบุ local publisher ให้ชัด
- กำหนด license/asset notices อย่างชัดเจน ใช้สีที่ออกแบบเองและอ้าง provenance หากนำ theme อื่นมาปรับ

VS Code ใช้ TextMate grammar เพื่อจัดประเภท token แล้วให้ theme กำหนดสี ควรใช้ scope มาตรฐานเพื่อรองรับ theme ที่ผู้ใช้มีอยู่ด้วย อ้างอิง [Syntax Highlight Guide](https://code.visualstudio.com/api/language-extensions/syntax-highlight-guide)

## 5. Syntax highlighting ที่ต้องครอบคลุม

| กลุ่ม | ตัวอย่าง Craft | Scope แนวทาง |
| --- | --- | --- |
| Declaration | func, struct, let, var | storage.type / storage.modifier |
| Control flow | if, else, while, for, in, return, break, continue | keyword.control |
| Errors / tests | try, catch, throw, defer, test, assert | keyword.control |
| Built-in types | Int, Float, Bool, String, Void, Exception, Map, JsonValue, DateTime, Duration, Random | support.type |
| User type declaration / annotation | struct Account, value: Account | entity.name.type / storage.type |
| Functions / calls | func deposit, deposit(...), append(...) | entity.name.function / support.function |
| Library namespace | std.io, std.json, std.fs | support.namespace |
| Literals | true, false, null, 42, 12.5, 1e3 | constant.language / constant.numeric |
| String / escapes | "Hello โลก", \n, \uXXXX | string.quoted.double / constant.character.escape |
| Comment | // คำอธิบาย | comment.line.double-slash |
| Operators / punctuation | +=, ==, &&, .., ?, [], {}, : | keyword.operator / punctuation |
| Identifiers / fields / arguments | balance, account.number, number: | variable / variable.other.property ตามบริบทที่ระบุได้ |

ข้อกำหนด:

- ยึด tokens และกฎใน `internal/lexer/lexer.go`, parser และ Rev.2 reference รวม Unicode identifiers
- keyword ต้องมีขอบเขตที่ถูกต้อง เช่น `if` ไม่ทำให้ `gift` หรือชื่อที่ติดอักษรไทยเปลี่ยนสีบางส่วน
- String/comment มีลำดับเหนือ keyword/call rules; `"if // var"` ต้องเป็น String ทั้งก้อน
- แยก `1..3` ออกจาก Float, รองรับ exponent/signed exponent และ operators หลายตัวอักษร
- รองรับ `Map<String, Int?>`, nested Maps, optional arrays และ if let โดยไม่สับสน `<`/`>` กับ comparison
- รองรับ named arguments, chained calls, member access และหลายบรรทัด; ไม่เดาว่าตัวแปรชื่อขึ้นต้นตัวใหญ่เป็น type เสมอ
- String ที่ยังพิมพ์ไม่จบต้องไม่ทำให้สีไหลผิดไปทั้งไฟล์ เพราะ Craft ไม่รองรับ multiline String
- ไม่ทำให้ unsupported syntax เช่น block comments, interpolation, enum/import/lambda ดูเสมือนรองรับแล้ว
- Grammar ทำหน้าที่แสดงสี ไม่ใช่ type checker; ชื่อที่กำกวมให้ใช้สีทั่วไปและรอ semantic analysis

## 6. สีที่อ่านง่าย

ส่งมอบ Craft Dark และ Craft Light เป็นตัวเลือก ผู้ใช้ยังใช้ Dark+, Light+ หรือ theme เดิมได้ ไม่บังคับสลับ theme เมื่อติดตั้ง

| กลุ่ม | แนวทางสี Dark | แนวทางสี Light |
| --- | --- | --- |
| Keywords | ม่วงอ่อน / น้ำเงินอ่อน | ม่วงเข้ม / น้ำเงินเข้ม |
| Types | เขียวอมฟ้าสว่าง | เขียวอมฟ้าเข้ม |
| Functions | เหลืองนวล | น้ำตาลทองเข้ม |
| Strings | ส้มอ่อน | แดงน้ำตาล |
| Numbers / Bool / null | เขียวอ่อน | เขียวเข้ม |
| Variables / fields | สีตัวอักษรหลักหรือฟ้าอ่อน | สีตัวอักษรหลักหรือน้ำเงินเข้ม |
| Comments | เทาอมเขียวที่ยังอ่านได้ | เทาเขียวเข้ม |

สีในตารางเป็นทิศทางออกแบบ ยังไม่ใช่ palette ที่ผ่านตรวจรับ ต้องกำหนด hex และตรวจ contrast กับ background/selection จริงก่อนส่งมอบ

- ให้ความสำคัญกับตัวอักษรโค้ดปกติ comments และข้อความบน selection; ไม่ใช้สีจางจนหายไป
- ใช้สีจำนวนจำกัด ไม่ทำให้ทุกเครื่องหมายโดดเด่นเท่ากัน และไม่ใช้ italic/bold กับโค้ดส่วนใหญ่
- ออกแบบ editor background/foreground, cursor, line numbers, current line, selection, find matches และ bracket pairs ให้กลมกลืน
- Error/warning ใช้ underline/marker ร่วมกับสี ไม่ใช้สีเพียงอย่างเดียวสื่อความหมาย
- ใช้ High Contrast ของ VS Code เป็นทางเลือกและตรวจว่ากฎ Craft ไม่ทำให้ข้อความหาย ไม่อ้างว่ามี Craft High Contrast theme จนกว่าจะสร้างและทดสอบจริง
- ตรวจภาพที่ zoom 100%, 125%, 150% รวมภาษาไทยและการเลือกข้อความ โดยผู้ใช้เป็นผู้ทดลอง

## 7. ความสะดวกในการพิมพ์

กำหนดผ่าน [Language Configuration](https://code.visualstudio.com/api/language-extensions/language-configuration-guide):

- Toggle line comment ด้วย `//`
- Bracket matching/auto closing สำหรับ `{}`, `[]`, `()` และ quotes โดยไม่รบกวน String/comment
- ไม่ auto-close `< >` ทั่วไป เพราะใช้เป็น comparison ด้วย
- Indentation หลังเปิด block/map และลดก่อนปิด brace; ตรวจ map literal ใน function arguments ด้วย
- Folding สำหรับ blocks และการตั้ง word pattern ที่รองรับชื่อ Craft/Unicode
- Snippets: main, func, struct, let/var แบบมี type, if/else, if let, while, for range/array, try/catch, defer และ test
- Snippets ต้องเป็น syntax ที่ compiler ปัจจุบันรับจริง ไม่ใส่ฟีเจอร์ระดับ C ล่วงหน้า

## 8. CLI และเอกสารจากปัญหาที่พบจริง

- เมื่อ `craft fmt check` หา path ไม่พบ ให้เพิ่ม hint ว่าอาจต้องการ `craft fmt --check` โดยยังรายงาน error และไม่ format/write อัตโนมัติ
- อย่าแปลง positional argument ทุกตัวเป็น option; รักษาคำสั่ง format ไฟล์ที่ถูกต้องและ exit categories เดิม
- แสดงตัวอย่างการใช้ทั้ง `fmt --check` และ `fmt path/to/file.craft` ใน help
- คู่มือแยกการติดตั้ง VSIX ผ่าน VS Code จาก Visual Studio Installer และแยกคำสั่งเลือก File Icon Theme กับ Color Theme
- นำชุด 15 tests จากบทสนทนามาเก็บใน repository เป็น acceptance fixture พร้อม environment setup/cleanup และคำอธิบาย starter test ที่เพิ่มจำนวนเป็น 16
- อัปเดต status เมื่อมีหลักฐานจริง ระบุ user-reported / source-reviewed / built / not tested แยกกัน
- เอกสาร Rev.1 foundation เก็บเป็นประวัติ พร้อมลิงก์มาที่ Rev.2 contracts แทนการปล่อยข้อความเก่าดูเป็นสถานะปัจจุบัน

## 9. งานต่อยอด editor หลังงานสี

### Format Document / diagnostics

- ออกแบบ formatter interface ที่รับข้อความจาก buffer และคืนข้อความ/edits โดยไม่เขียนไฟล์บนดิสก์ เพื่อรองรับ unsaved files และไม่ทับการแก้ไขระหว่างรอผล
- diagnostics ควรมี machine-readable output: code, severity, file, range, message; ไม่ parse ข้อความสำหรับมนุษย์ด้วย regex อย่างเดียว
- แปลงตำแหน่ง rune-based ของ Craft เป็น UTF-16 positions ของ VS Code ให้ถูกต้อง โดยเฉพาะภาษาไทย/emoji และ CRLF
- กำหนด executable path, project root, debounce/cancellation, multi-root workspace และจัดการ CLI ไม่อยู่ใน PATH
- การตรวจโค้ดต้องไม่เรียก main/tests; การเรียก executable เคารพ Workspace Trust และไม่ใช้ shell command interpolation

### Semantic tokens / language services

- ใช้ resolver/type checker เพื่อจำแนก parameter, field, function, Struct และ immutable binding อย่างถูกต้องทุกจุดอ้างอิง
- Semantic tokens เป็นชั้นเพิ่มเติมบน TextMate grammar; ต้องยังมีสีพื้นฐานหาก analysis ยังไม่พร้อม
- เริ่มจาก semantic tokens/hover แล้วจึง completion/go-to-definition; ไม่ผูกงานหลักกับการสร้าง LSP เต็มระบบ

อ้างอิง [Semantic Highlight Guide](https://code.visualstudio.com/api/language-extensions/semantic-highlight-guide)

## 10. Timer System และ Multi-task

### ขอบเขตและลำดับพัฒนา

- พัฒนา task runtime และ cancellation ก่อน แล้วให้ Timer System ใช้ runtime เดียวกัน
- รุ่นแรกเน้น concurrency: งานที่รอเวลาหรือ I/O ต้องเปิดโอกาสให้งานอื่นเดินต่อ ไม่รับประกันการประมวลผล CPU หลาย core พร้อมกัน
- รองรับการอ้างถึง named function เป็น task/timer entry point ก่อน ไม่บังคับให้มี Lambda, closure หรือ Event library เต็มระบบ; resolver/type checker ต้องตรวจ signature ได้
- ชื่อ API ในหัวข้อนี้สรุปเป็น [language contract 0.1.3](../../docs/REV3-LANGUAGE.md) แล้ว รวม ownership ระดับ function invocation, limits และการคง Int sleep เดิม; ไม่ใช้ตัวอย่างนี้กับ CLI 0.1.2

### Multi-task

| API ที่เสนอ | พฤติกรรม |
| --- | --- |
| `std.task.spawn(function, arguments...)` | เริ่มงานและคืน `Task` handle ทันที; entry point รุ่นแรกคืน `Void` |
| `task.join()` | พักเฉพาะ task ผู้เรียกจนงานเป้าหมายจบ และส่งต่อ exception หากงานล้มเหลว |
| `task.cancel()` | ขอให้ยกเลิกงานแบบ cooperative เรียกซ้ำได้ |
| `task.status()` | คืนสถานะ `pending`, `running`, `completed`, `failed` หรือ `cancelled` |

- ประเมิน arguments ครั้งเดียวใน task ผู้เรียกก่อนเริ่มงาน และส่งค่าแบบ deep copy ตาม semantics เดิม; ไม่เปิด shared mutable variables หรือ mutable captures ในรุ่นแรก
- `Task`/`Timer` เป็น opaque runtime handles: การส่งต่อหรือกำหนดค่าอ้างงานเดิม ไม่ clone งาน และไม่ serialize เป็น JSON; ต้องกำหนดข้อยกเว้นจาก value-copy semantics ให้ชัดเจน
- Runtime ตรวจคำขอยกเลิกที่จุดพักและขอบเขต statement/loop; I/O ที่รองรับต้องปลุกหรือยกเลิกการรอได้ พร้อมระบุข้อจำกัดของ blocking operation ที่ยังยกเลิกไม่ได้
- งานที่ถูกยกเลิกต้องรัน `defer` เพื่อคืนทรัพยากร; `join()` ของงานที่ยกเลิกคืน cancellation exception และการรอซ้ำต้องได้ผลสิ้นสุดเดิม
- ห้าม join ตัวเองและต้องตรวจวงจรการรอระหว่าง task พร้อม diagnostic แทนการค้างเงียบ
- ใช้ structured lifetime: งานลูกอยู่ในขอบเขตงานแม่ เมื่อออกจากขอบเขตให้ยกเลิกและรอ cleanup ของงานลูก; ไม่ปล่อย detached task หลัง main หรือ test จบ
- Exception ที่ไม่มีผู้รับผ่าน join ต้องส่งกลับเจ้าของขอบเขตและทำให้ run/test ล้มเหลว; Ctrl+C ต้องยกเลิกงานทั้งหมดและใช้ exit convention เดิมของ CLI
- กำหนดขีดจำกัดจำนวน task และการจัดคิวอย่างเป็นธรรม พร้อม error เมื่อเกินขีดจำกัด; ไม่สร้าง OS thread ต่อ task โดยไม่มีขอบเขต
- stdout ต้องป้องกันข้อความจาก print หนึ่งครั้งปะปนกัน โดยไม่รับประกันลำดับระหว่าง task; console input อนุญาตผู้รออ่านครั้งละหนึ่ง task และแจ้ง error เมื่อมีผู้รอซ้อน
- การ deep copy ไม่แยก external resources เช่นไฟล์เดียวกัน ผู้ใช้ต้องจัดลำดับการเข้าถึงเอง; รุ่นแรกยังไม่เพิ่ม mutex/channel หรือ API ส่งข้อความระหว่าง task

### Timer System

| API ที่เสนอ | พฤติกรรม |
| --- | --- |
| `std.time.sleep(duration)` | พักเฉพาะ task ปัจจุบันและตื่นเมื่อครบเวลา หรือถูกยกเลิก |
| `std.timer.after(duration, function, arguments...)` | คืน `Timer` handle และเริ่ม callback ครั้งเดียวหลังครบเวลา |
| `std.timer.every(interval, function, arguments...)` | คืน `Timer` handle และเรียก callback ซ้ำตามนโยบายด้านล่าง |
| `timer.cancel()` | หยุดการนัดครั้งถัดไปและขอยกเลิก callback ที่กำลังทำงาน เรียกซ้ำได้ |
| `timer.join()` / `timer.status()` | รอสิ้นสุด/ตรวจสถานะด้วยหลักเดียวกับ Task; timer แบบซ้ำสิ้นสุดเมื่อยกเลิกหรือล้มเหลว |

- รับ `Duration` เดิมของ Rev.2; ค่าติดลบเป็น error, `sleep(0)` คืนโอกาสให้ scheduler, `after(0, ...)` เข้าคิวรอบถัดไป และ `every` ต้องมี interval มากกว่า 0
- ใช้ monotonic clock วัดเวลาที่ผ่านไป ไม่ให้การเปลี่ยนเวลาของเครื่องกระทบช่วงรอ; เวลาที่ระบุเป็นเวลารอขั้นต่ำ ไม่รับประกัน real-time precision
- Timer แบบซ้ำใช้ fixed delay: เริ่มนับ interval ถัดไปหลัง callback ก่อนหน้าจบ จึงไม่มี callback ของ timer เดียวกันซ้อนกันหรือคิวชดเชย tick สะสม; ครั้งแรกเริ่มหลังครบ interval
- Callback เป็น named function ที่คืน `Void`; ประเมินและเก็บ arguments แบบ deep copy ตอนสร้าง timer และให้สำเนาใหม่แก่ callback แต่ละรอบ
- Callback exception ทำให้ timer หยุดในสถานะ failed และส่งต่อผ่าน join/เจ้าของขอบเขต; ไม่กลืน error หรือ retry อัตโนมัติ
- การยกเลิกที่แข่งกับเวลาครบกำหนดต้องมีจุดตัดสินใน scheduler: หากเริ่ม callback แล้วให้ใช้ cooperative cancellation หากยังไม่เริ่มต้องไม่เรียก callback
- Timer อยู่ภายใต้ structured lifetime เดียวกับ task; เมื่อ main/test/เจ้าของขอบเขตจบต้องไม่มี timer หรือ callback ค้างข้ามการรัน
- รุ่นแรกเป็น timer ภายใน process ยังไม่รวม cron, calendar schedule, การบันทึกตารางถาวร หรือการปลุกโปรแกรมที่ปิดแล้ว

### งานประกอบและเกณฑ์ตรวจรับ runtime

- อัปเดต resolver, type checker, interpreter, standard library และ formatter ตาม syntax ที่อนุมัติ พร้อมตรวจ resource ownership และไม่เปลี่ยนพฤติกรรมโปรแกรม Rev.2 ที่ไม่ได้ใช้ task/timer
- เมื่อ runtime พร้อมแล้วจึงเพิ่ม highlighting/snippets ของ Task/Timer/API ใหม่ โดยระบุ CLI version ขั้นต่ำแยกจาก extension สำหรับ 0.1.2
- จัดทำตัวอย่างงานสอง task ที่รอเวลาคนละช่วง, one-shot timer, repeating timer และ cancellation พร้อมคำสั่งให้ผู้ใช้รัน
- เตรียม scheduler tests ด้วย fake clock สำหรับผู้ใช้รัน ลดการอาศัยเวลาจริงและไม่ assert ลำดับที่ contract ไม่รับประกัน

- [ ] Task หนึ่ง sleep หรือรอ I/O ที่รองรับแล้ว task อื่นยังเดินต่อได้
- [ ] spawn/join/status, exception propagation, cancellation และ defer เป็นไปตาม contract
- [ ] แก้ไข arguments ในงานลูกแล้วไม่เปลี่ยนค่าของงานแม่ และ handles ยังคงอ้างงานเดิม
- [ ] self-join, วงจร join และ task limit ให้ error ที่อธิบายได้
- [ ] after ทำงานครั้งเดียว และ every ไม่ซ้อน callback หรือสะสม tick เมื่องานช้า
- [ ] ค่าเวลา 0/ติดลบ, cancel ก่อนเริ่ม/ระหว่าง callback และ callback exception ได้ผลตามที่ระบุ
- [ ] main/test จบหรือ Ctrl+C แล้วไม่มี task/timer ค้างหรือรบกวน test ถัดไป
- [ ] มีผล regression ของ Rev.2 และ bundle ที่ใช้ task/timer จากการรันโดยผู้ใช้ก่อนประกาศรองรับ

## 11. Milestones

1. **Baseline:** บันทึกผล Rev.2 ที่มีจริง แยก pending และเก็บชุด acceptance จากบทสนทนา
2. **Language extension:** language registration, grammar, language configuration และ snippets
3. **Readability:** Craft Dark/Light, token gallery, ตรวจ edge cases ของ grammar และสี
4. **Handoff:** VSIX ใหม่, คู่มือ install/select/uninstall, usage hint ของ fmt และอัปเดต status
5. **User acceptance:** ผู้ใช้ตรวจโค้ดตัวอย่างและความอ่านง่าย แล้วรายงานผลก่อนประกาศพร้อมใช้
6. **ระดับ B:** ประเมิน formatter/diagnostics protocol และ semantic services เป็นงานส่งมอบแยก
7. **Runtime contract:** สรุป syntax/signatures, named function references, handle semantics, scheduler, ownership และ cancellation ของ Multi-task/Timer
8. **Multi-task:** ส่งมอบ task runtime, lifecycle, diagnostics และตัวอย่างให้ผู้ใช้ตรวจรับ
9. **Timer System:** เพิ่ม sleep/after/every บน task runtime พร้อม cancellation และ cleanup
10. **Runtime acceptance:** ผู้ใช้รันชุด task/timer และ regression Rev.2 แล้วบันทึกผลแยกจาก acceptance ของ editor

## 12. การทดสอบและเกณฑ์เสร็จงาน editor

เตรียม token fixtures ที่มี declaration, literals, escapes, comments, Unicode, nested types, range/comparison, chained calls, named arguments และโค้ดที่กำลังพิมพ์ไม่ครบ พร้อม expected scopes
เตรียม grammar tests และ regression ของ CLI hint ให้ผู้ใช้รัน ไม่มีการเปิด VS Code หรือรัน tests แทนผู้ใช้โดยไม่ได้รับคำสั่ง

- [ ] เปิด `.craft` แล้ว language mode เป็น Craft และมีสีพื้นฐานโดยไม่ต้องติดตั้ง CLI
- [ ] Token scopes ครอบคลุม syntax 0.1.2 และไม่รั่วจาก String/comment ไปยังบรรทัดอื่น
- [ ] เลือก Craft Dark/Light ได้เอง และใช้ theme เดิมของผู้ใช้ได้ด้วย
- [ ] ภาษาไทย/ตัวเลข/comments/selection/brackets อ่านชัดในชุดหน้าจอที่ผู้ใช้ตรวจจริง
- [ ] Comment, brackets, indentation, folding และ snippets ทำงานตามขอบเขต
- [ ] ใช้ร่วมกับ Craft Forge File Icons ได้ และถอน extension หนึ่งไม่กระทบอีกตัว
- [ ] VSIX ติดตั้งผ่าน VS Code ได้ มีคู่มือที่แยก icon/color theme และวิธีเปลี่ยนกลับ
- [ ] CLI fmt typo มี hint โดยไม่เปลี่ยนข้อมูล และมีผล regression ตามที่ผู้ใช้รัน
- [ ] มีผลทดสอบหลัง fmt และ bundle ของ Rev.2 หรือระบุว่าเป็น pending ที่ยังไม่ปิด
- [ ] Status ระบุ implemented, user-tested, not tested และ deferred ตรงกับหลักฐาน

Rev.3 งาน editor ไม่ใช้เป็นเหตุประกาศว่า acceptance ของ runtime/installer Rev.2 ผ่านทั้งหมด งานที่ยังไม่มีผลต้องติดตามแยกต่อไป

Rev.3 ทั้งฉบับถือว่าเสร็จเมื่อผ่านทั้งเกณฑ์ editor ในหัวข้อนี้และ runtime ในหัวข้อ 10; การส่งมอบ editor ก่อนต้องระบุ Timer System/Multi-task เป็น pending จนกว่าจะมี implementation และผลตรวจรับจริง

อ้างอิงโครงการ: [Rev.2 roadmap](craft-phase1-rev2.md), [Rev.2 status](output/phase1-rev2-status.md), [Language contracts](../../docs/REV2-LANGUAGE.md), [Icon theme](../../craft-file-icon-theme/README.md)
