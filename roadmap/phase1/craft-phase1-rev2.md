# Craft Phase 1 Rev.2

## Interactive Programs and Standard Library Expansion

สถานะเอกสาร: ขอบเขตแผน Rev.2 — ดูผล implementation และรายการรอผู้ใช้ทดสอบใน [Rev.2 status](output/phase1-rev2-status.md); API ที่ส่งมอบระบุใน [Rev.2 contracts](../../docs/REV2-LANGUAGE.md)  
วันที่จัดทำ: 18 กันยายน 2026  
ฐานการพัฒนา: Craft 0.1.1 — Phase 1 Rev.1

## 1. เป้าหมาย

ทำให้ Craft เขียนโปรแกรม Console ที่รับข้อมูล ประมวลผล จัดเก็บ และอ่านข้อมูลกลับมาใช้ได้จริง พร้อมขยาย Standard Library และโครงสร้างข้อมูลอย่างสอดคล้องกับ Static Typing

ตัวอย่างเป้าหมายคือโปรแกรม Bank ที่ผู้ใช้เลือกเมนู ฝาก/ถอนเงิน ดูประวัติ และบันทึกบัญชีลงไฟล์ได้ แทนการกำหนดข้อมูลทุกอย่างใน source

Rev.2 ยังคงใช้ Interpreter ไม่รวม API Framework, Database, Native Backend หรือ Bytecode VM

## 2. สถานะเริ่มต้น

Rev.1 มี let/var, scope, functions, named arguments, Array แบบ deep-copy, String operations, Exception, try/catch/throw, stack trace, defer, craft fmt และ craft test แล้ว

Standard Library มี print, String/Array methods และ math/fs/time/json/env ขั้นต้น แต่ยังไม่มี Console input, การแปลงชนิดข้อมูลทั่วไป, Map, JSON แบบมีโครงสร้าง หรือ Module/import

ผลที่ผู้ใช้รายงานแล้ว:

- โปรแกรมสาธิต craft run ทำงานจนจบ
- craft test ผ่าน 39 tests และ exit code 0
- craft fmt และ fmt --check ทำงาน โดย check ได้ exit code 0
- หลัง fmt ยังผ่าน 39 tests
- craft build สร้าง source bundle สำเร็จ

หลักฐานนี้ยังไม่ครอบคลุมการรัน bundle, filesystem test, Go tests ทั้ง repository หรือ installer lifecycle ทุกกรณี

## 3. ลำดับความสำคัญและขอบเขต

| ระดับ | งาน | เงื่อนไข |
| --- | --- | --- |
| A — งานหลัก | Console input, conversion, CLI arguments | ทำก่อน เพื่อให้โปรแกรมโต้ตอบได้ |
| A — งานหลัก | Optional ขั้นต้น, Struct, Map | เป็นพื้นฐานข้อมูลสำหรับโปรแกรมและ JSON |
| A — งานหลัก | Array utilities, Math, File System/Path, Date/Time | ขยายจากของเดิมโดยกำหนด semantics ให้ชัด |
| A — งานหลัก | JSON parse/stringify แบบมีโครงสร้าง | ทำหลังมีตัวแทน JSON value ที่ตรวจชนิดได้ |
| A — งานหลัก | Documentation, diagnostics, fmt/test, installer | ต้องรองรับฟีเจอร์ใหม่ครบทุกชั้น |
| A — เครื่องมือพัฒนา | Craft Forge File Icons สำหรับ VS Code | ใช้ manifest และ SVG ที่เตรียมไว้ ส่งมอบเป็น VSIX |
| B — งานต่อยอด | Enum, Module/import และ visibility | ทำหลัง type/value model เสถียร |
| B — งานต่อยอด | Function values, Lambda, Callback, Event library | ต้องทำตามลำดับ dependency |
| C — แยกแผนภายหลัง | HTTP client, process execution, REPL | ไม่บังคับใน Definition of Done ของ Rev.2 หลัก |

งานระดับ B เป็น backlog ที่รับต่อจาก Rev.1 ไม่ถือว่าทำแล้วเพียงเพราะมีเอกสารออกแบบ หากเลื่อนไป revision ถัดไปให้ระบุในสถานะส่งมอบอย่างชัดเจน

ชื่อ API และตัวอย่างด้านล่างเป็นข้อเสนอ ต้องยืนยัน signature และกฎก่อน implementation โดยไม่เปลี่ยนขอบเขตงานเงียบ ๆ

## 4. Console Input

API ที่เสนอ:

```craft
std.io.write("Name: ")
let name: String? = std.io.readLine()
```

- write(String) แสดงข้อความโดยไม่ขึ้นบรรทัดใหม่ ส่วน print เดิมยังใช้ได้
- readLine() คืน String? โดย null หมายถึง EOF; บรรทัดว่างคืน "" ไม่ใช่ null
- ตัดเฉพาะตัวจบบรรทัด LF/CRLF ไม่ trim ช่องว่างของข้อมูลอัตโนมัติ
- อ่าน Unicode และรับ piped input ได้
- EOF ต้องทำให้โปรแกรมออกจากลูปได้ ไม่วนซ้ำไม่สิ้นสุด
- การอ่านล้มเหลวส่ง Exception พร้อม source location
- กำหนดขนาดบรรทัดสูงสุดและพฤติกรรมเมื่อเกินขนาด
- Interpreter/CLI ต้องรับ input reader แทนการอ่าน os.Stdin โดยตรงทุกจุด เพื่อให้เขียน tests ได้
- Ctrl+C ต้องหยุดการรอ input ได้ตามข้อจำกัดของระบบ โดยระบุพฤติกรรม Windows ในคู่มือ

## 5. Type Conversion และการแสดงผล

| API ที่เสนอ | ผลลัพธ์ |
| --- | --- |
| std.convert.toInt(text: String) | Int; แปลงไม่ได้ให้ throw |
| std.convert.toFloat(text: String) | Float; แปลงไม่ได้ให้ throw |
| std.convert.toBool(text: String) | Bool; รับ true/false ตามกฎที่กำหนด |
| std.convert.toString(value) | String; รองรับ primitive types ก่อน |
| std.convert.tryInt(text: String) | Int?; null เมื่อแปลงไม่ได้ |
| std.convert.tryFloat(text: String) | Float?; null เมื่อแปลงไม่ได้ |
| std.string.fixed(value: Float, digits: Int) | String แสดงทศนิยมตามจำนวนหลัก |

กฎ:

- ไม่มี implicit conversion ระหว่าง Int, Float และ String
- ตรวจข้อความทั้งค่า ไม่ยอมรับส่วนท้ายที่ไม่ใช่ตัวเลข เช่น "12abc"
- ระบุชัดว่าจะ trim ช่องว่างหรือไม่; แนวทางเริ่มต้นคือให้ผู้ใช้เรียก trim เอง
- Int ใช้ฐานสิบและตรวจ signed 64-bit overflow
- Float ไม่รับ NaN/Infinity หรือค่าที่เกินช่วง
- Bool ใช้ true/false แบบ case-sensitive ไม่เดาความหมายจาก "yes" หรือ "1"
- Conversion APIs เป็น built-ins ที่มี signatures จำกัด ไม่ใช่การเปิดใช้ user-defined overloading/generics
- fixed ต้องกำหนดช่วง digits และ rounding rule พร้อมตัวอย่าง boundary
- แยก numeric conversion จากการจัดรูปแบบข้อความ

ตัวอย่าง:

```craft
try {
    let amount: Int = std.convert.toInt("2500")
    print(amount)
} catch error {
    print(error.message)
}
```

## 6. CLI Arguments และ Environment

- std.env.args(): String[] คืน arguments ของโปรแกรม Craft ไม่รวม CLI command หรือ executable path
- รองรับ `craft run -- arg1 arg2` และ `craft run file.craftbundle -- arg1 arg2`
- แยก CLI flags จาก program arguments ด้วย `--`
- เก็บ std.env.get(name): String แบบเดิมเพื่อ compatibility
- เพิ่ม std.env.lookup(name): String? เพื่อแยก unset ออกจากค่าที่เป็น empty String
- ยังไม่เพิ่มการแก้ environment ของเครื่องหรือ parent process

## 7. Optional และ Null Safety ขั้นต้น

```craft
let nickname: String? = null
```

- T? หมายถึง T หรือ null; T ปกติรับ null ไม่ได้
- รองรับตัวแปร parameters และ function results
- ต้องตรวจและ unwrap ก่อนใช้เป็น T หรือเรียก methods ของ T
- ออกแบบ `if let` เป็นรูปแบบเริ่มต้นสำหรับ immutable binding ภายใน block
- การตรวจ `x != null` อย่างเดียวไม่ให้แปลงชนิดโดยปริยาย จนกว่าจะมี flow-sensitive narrowing ที่ถูกต้อง
- ไม่เพิ่ม force unwrap ที่เสี่ยงเกิด null error โดยไม่มีความจำเป็น
- Optional ของ Array/Struct/Map ต้องรักษา deep-copy และ let/var semantics

ตัวอย่าง syntax ที่เสนอ:

```craft
if let name = std.io.readLine() {
    print("Hello " + name)
} else {
    print("End of input")
}
```

การละ type annotation ใน if let เป็นข้อยกเว้นเฉพาะ binding จาก Optional ที่ทราบชนิดแล้ว ไม่เปิด type inference สำหรับ let/var ทั่วไปในทันที

## 8. Struct ขั้นต้น

```craft
struct Account {
    let number: String
    var balanceMinor: Int
}

func main() {
    var account: Account = Account(number: "CRAFT-001", balanceMinor: 1000000)
    account.balanceMinor += 250000
    print(account.number, account.balanceMinor)
}
```

- ประกาศ fields ที่มีชนิดชัดเจนและ let/var
- สร้าง instance ด้วย named arguments ตรวจ missing/unknown/duplicate fields
- อ่านและแก้ fields ตาม mutability
- let ของ instance ต้อง immutable ลึกถึงข้อมูลภายใน
- การแก้ field ต้องมี mutable root และ mutable fields ตามเส้นทางที่เข้าถึง
- ใช้ value semantics เช่นเดียวกับ Array
- ตรวจ recursive value types ที่ไม่สามารถสร้างได้ และรายงานข้อผิดพลาดชัดเจน
- ยังไม่มี inheritance, methods, custom constructors หรือ default fields

## 9. Map ขั้นต้น

เริ่มจาก Map<String, T> แบบ built-in container โดยยังไม่รองรับ generic type declarations ของผู้ใช้

```craft
var scores: Map<String, Int> = {
    "math": 90,
    "english": 85
}
scores.set("science", 95)
```

API ที่เสนอ:

- get(key): T? — null เมื่อไม่พบ
- require(key): T — throw เมื่อไม่พบ
- set(key, value): Void — เพิ่มหรือแทนค่า ต้องใช้ var
- remove(key): Bool — คืนว่าลบได้หรือไม่ ต้องใช้ var
- containsKey(key): Bool
- keys(): String[] และ values(): T[] — เป็น snapshots ที่ลำดับสัมพันธ์กัน
- length: Int — อ่านอย่างเดียว

กำหนดลำดับ iteration เป็น insertion order; การแทนค่าคงตำแหน่งเดิม การลบแล้วเพิ่มใหม่อยู่ท้าย ใช้ deep-copy เช่น Array และห้ามแก้ Map ของ let ผ่าน alias

## 10. Array และ String Utilities

เพิ่ม Array operations ที่ยังไม่ต้องใช้ callbacks:

- contains(value): Bool
- indexOf(value): Int? — null เมื่อไม่พบ
- slice(start, end): T[] — end-exclusive และตรวจ bounds
- reverse(): T[] — คืนสำเนา ไม่แก้ต้นฉบับ
- sorted(): T[] — เริ่มจาก Int, Float และ String; ระบุลำดับ comparison

จำกัด contains/indexOf ให้ชนิดที่มีกฎ equality ชัดเจนก่อน ไม่สมมติว่า Array/Struct เปรียบเทียบกันได้แล้ว

เพิ่ม String operations ตามความจำเป็น:

- replace(old, replacement): String โดยกำหนดว่าจะเปลี่ยนทุกตำแหน่ง
- substring(start, end): String — index เป็น Unicode code points, end-exclusive
- join(parts: String[], separator: String): String ผ่าน std.string

map/filter/reduce และ sort ด้วย comparator อยู่ระดับ B หลัง Function values/Callback ไม่เพิ่ม API ที่ยังไม่มีชนิด callback รองรับ

## 11. Math และ Random

เพิ่ม std.math.round, floor, ceil และ pow พร้อม signatures ที่ชัดเจน โดยคง numeric types ไม่แปลง Int/Float อัตโนมัติ

- กำหนด round สำหรับค่ากึ่งกลางและค่าติดลบ
- ตรวจ domain errors และผลลัพธ์ที่ไม่ finite
- แยก rounding เป็นตัวเลขออกจาก fixed ที่คืนข้อความ
- เพิ่ม random Int แบบช่วง [min, max) และ random Float แบบ [0, 1)
- ออกแบบ generator ที่กำหนด seed ได้สำหรับ reproducible tests ไม่ใช้ global seed ที่ทำให้ tests กระทบกัน
- ระบุว่า random ชุดนี้ไม่ใช่ cryptographic random

ชื่อและรูปแบบ generator ต้องสรุปหลัง Struct/runtime object model พร้อม ก่อนเผยแพร่ API

## 12. File System และ Path

รักษา readText/writeText/exists เดิม และเพิ่ม:

| กลุ่ม | API ที่เสนอ |
| --- | --- |
| ข้อความ | appendText(path, text) |
| โฟลเดอร์ | createDirectory(path), listDirectory(path), isDirectory(path) |
| ไฟล์ | removeFile(path), copyFile(source, destination), moveFile(source, destination) |
| Path | std.path.join(parts: String[]), fileName(path), extension(path), parent(path), isAbsolute(path) |

- กำหนด createDirectory ว่าสร้าง parents ด้วยหรือไม่ และผลเมื่อมีอยู่แล้ว
- listDirectory ต้องมีลำดับที่แน่นอนสำหรับ tests
- copy/move ปฏิเสธ destination ที่มีอยู่โดยค่าเริ่มต้น; overwrite ต้องเลือกอย่างชัดเจน
- removeFile ไม่ลบโฟลเดอร์ recursive
- รักษา relative-path behavior ที่อ้างอิง process working directory และอธิบายให้ชัด
- กำหนด Windows separators, Unicode paths และการจัดการ symlink
- ข้อผิดพลาด permission, missing path และ wrong file type ต้อง catch ได้ พร้อม source location
- เพิ่ม test fixture directory ของแต่ละ test และ cleanup โดยไม่แตะไฟล์ส่วนตัวของผู้ใช้

## 13. Date/Time

คง now() เดิมที่คืน UTC String และ sleep(milliseconds) แล้วเพิ่ม typed time value/API สำหรับ:

- Parse RFC3339
- Format ด้วยรูปแบบที่กำหนด
- Unix timestamp โดยระบุหน่วยวินาทีหรือมิลลิวินาทีในชื่อ API
- เพิ่ม/ลด duration และหาความต่างของเวลา
- ระบุ timezone/offset อย่างชัดเจน

เริ่มจาก UTC และ fixed offsets ก่อน ฐานข้อมูล named time zones และ daylight-saving rules เป็นงานต่อยอด ไม่ผูกพฤติกรรมกับ locale ของเครื่องโดยไม่ประกาศ

ต้องสรุปตัวแทนเวลาและ duration รวมถึง rounding/overflow ก่อนเพิ่ม signatures หลีกเลี่ยงการเพิ่มหลาย API ที่ใช้ Int แต่มีหน่วยต่างกันโดยดูไม่ออก

## 14. JSON แบบมีโครงสร้าง

คง isValid และ quote เดิม แล้วเพิ่ม parse/stringify หลังออกแบบตัวแทน JSON ครบทั้ง null, Bool, number, String, Array และ object

แนวทางเริ่มต้น:

- ใช้ built-in JsonValue แบบ tagged value พร้อม typed accessors
- parse(text): JsonValue และ stringify(value): String
- Accessor ของชนิดผิดต้องให้ผลที่กำหนดชัดเจน เช่น throw ไม่คืนศูนย์หรือข้อความว่างแทน
- มี constructors จาก primitive/Array/object ตาม signatures ที่ตรวจสอบได้
- ไม่ต้องเปิด Any หรือ dynamic property access ให้ภาษาโดยรวม
- กำหนดตัวแทน JSON number ให้ไม่สูญเสีย Int 64-bit จากการแปลงผ่าน Float โดยไม่ตั้งใจ
- กำหนดพฤติกรรม duplicate keys, malformed input, nesting limit และ Unicode escapes
- object serialization ต้องมีลำดับที่แน่นอนสำหรับ tests
- ไม่รวม automatic Struct serialization/deserialization จนกว่าจะมี mapping rules ชัดเจน

เป้าหมายรับงานคืออ่าน JSON ที่มีโครงสร้าง ตรวจชนิด แก้ข้อมูล แล้วเขียนกลับได้ ไม่ใช่เพียงตรวจว่า JSON valid

## 15. งานระดับ B: Language และ Module Foundation

### Enum

เริ่มจากชุด named constants ของ nominal type ตรวจไม่ให้ปนกับ String/Int โดยอัตโนมัติ ยังไม่รวม associated values หรือ pattern matching

### Module / Import

- กำหนดความสัมพันธ์ระหว่าง directory, file และ module
- public/private symbols และข้อผิดพลาดการเข้าถึง
- ตรวจ imports ที่ไม่มีอยู่ ชื่อกำกวม และ circular imports พร้อมแสดงเส้นทาง cycle
- กำหนด deterministic compile order
- ไม่เพิ่ม module-level executable initialization ในระยะแรก
- มี migration path จาก namespace รวมหลายไฟล์ของ Rev.1
- built-in std namespaces ต้องไม่ชนกับ user modules

### Function Values / Lambda / Callback / Event

ลำดับ: typed function signatures → function values → lambda/capture rules → callback → Array map/filter → Event library

ต้องกำหนด lifetime, mutable captures, argument evaluation และ exception propagation ก่อน implementation ส่วน Event library ควรเป็น synchronous library ในขั้นแรก พร้อม on/off/emit และ subscription lifecycle; ไม่เพิ่ม event keyword หรือ async runtime ในเฟสนี้

Default parameters และ REPL เป็นทางเลือกภายหลัง ไม่จำเป็นต่อการทำโปรแกรม Console เป้าหมาย

## 16. HTTP และ Process — แผนภายหลัง

HTTP client เป็นช่องว่างที่มีประโยชน์ แต่ต้องมี request/response types, timeout, redirects, headers, body limits และ error policy ก่อนเพิ่ม ไม่รวม HTTP server/framework ใน Rev.2

Process execution ต้องออกแบบ argument array ที่ไม่ต้องผ่าน shell โดยอัตโนมัติ, working directory, environment, stdout/stderr, exit status และ cancellation ก่อนเพิ่ม

สองกลุ่มนี้ไม่ใช่เกณฑ์บังคับของ Rev.2 หลัก เพื่อให้ส่งมอบแกน Console/Data/Library ได้ครบก่อน

## 17. Architecture และ Compatibility

- แยก built-in signatures/implementations ใน Standard Library ต่อไป
- Type Checker ต้องรู้ชนิดและ mutability ของ Optional/Struct/Map/JsonValue
- Runtime values ใหม่ต้องมีกฎ copy/equality/formatting ที่ชัดเจน
- Formatter และ test runner ต้องรองรับ syntax ใหม่พร้อมกัน
- เพิ่ม error codes อย่างมีรายการอ้างอิง รักษา exit categories ของ Rev.1
- กำหนด compiler version และ bundle language/version ใหม่ก่อน release; ไม่สมมติว่า manifest version คือ language version
- หาก bundle compatibility เปลี่ยน ต้องปฏิเสธรุ่นที่ไม่รองรับพร้อมคำแนะนำ migration/rebuild
- โปรแกรม Rev.1 ที่ถูกต้องต้องทำงานเหมือนเดิม หรือมี migration note ที่ระบุจุดเปลี่ยนทุกข้อ
- ไม่เปลี่ยน let กลับเป็น mutable และไม่ลดความเข้มของ type checking เพื่อให้เพิ่มฟีเจอร์ง่ายขึ้น

### 17.1 Craft Forge File Icons

นำข้อมูลใน [craft-file-icon-theme](../../craft-file-icon-theme/README.md) มาทำ File Icon Theme ที่ติดตั้งใช้ใน VS Code ได้ โดยใช้ assets เดิมเป็นฐาน งานนี้แยกจาก runtime และทำได้โดยไม่รอระบบชนิดข้อมูลใหม่

**สิ่งที่เตรียมไว้ ณ เริ่มต้น Rev.2**

- [theme.json](../../craft-file-icon-theme/theme.json): manifest กลางชื่อ Craft Forge, display name `Craft Forge File Icons`, version `0.1.0`, license ระบุ MIT พร้อมสีและ file associations
- `icons/` มี SVG 6 ไฟล์: `craft.svg`, `config.svg`, `lock.svg`, `test.svg`, `migration.svg` และ `build.svg`
- ใช้สี Craft Blue, Forge Navy, Build Gold, Test Green, Migration Purple และ Lock Red ตาม manifest; รักษารูปทรงและสีเดิมเป็นค่าเริ่มต้น
- ณ เริ่มต้นแผนเป็นชุดข้อมูลและภาพ ยังไม่มี `package.json` สำหรับ extension หรือไฟล์ LICENSE จึงยังไม่ใช่ extension ที่ติดตั้งได้

**ขอบเขตส่งมอบ**

1. ใช้ `theme.json` เป็น source of truth แล้วสร้าง adapter สำหรับ VS Code; ไม่แก้ mapping ซ้ำแยกกันหลายไฟล์
2. เพิ่ม extension manifest ที่ประกาศ `contributes.iconThemes`, theme ID ที่คงที่ และไฟล์ `craft-forge-icon-theme.json` พร้อม `iconDefinitions` ซึ่งอ้างถึง SVG ที่บรรจุใน extension
3. เตรียม generator/validation สำหรับ manifest, icon IDs, asset paths และ mapping ที่รองรับ ให้สร้างผลลัพธ์ซ้ำได้เหมือนเดิม และรายงานกฎที่แปลงไม่ได้
4. กำหนด extension identifier, publisher และ minimum VS Code version ก่อน package; แยก version ของ icon theme ออกจาก compiler และ source bundle
5. ส่งมอบ `.vsix` พร้อม README, changelog และ LICENSE ที่สอดคล้องกับ MIT ใน manifest โดยตรวจข้อมูลเจ้าของลิขสิทธิ์ก่อนใส่ชื่อ
6. มีคู่มือติดตั้งจาก VSIX เลือก `Craft Forge File Icons` ผ่าน File Icon Theme และวิธีเปลี่ยนกลับ/ถอนการติดตั้ง โดยให้ผู้ใช้เป็นผู้ติดตั้งและเลือก theme

**การแปลง associations ที่เตรียมไว้**

| กฎใน manifest กลาง | Icon | แผนสำหรับ VS Code |
| --- | --- | --- |
| `*.craft` | `craft.svg` | `fileExtensions` key `craft` |
| `craft.toml` | `config.svg` | `fileNames` key `craft.toml` |
| `craft.lock` | `lock.svg` | `fileNames` key `craft.lock` |
| `*_test.craft` | `test.svg` | suffix wildcard นี้แปลงเป็น native mapping โดยตรงไม่ได้; เก็บเจตนาไว้ใน manifest และใช้ icon ตามกฎที่รองรับแทน |
| `tests/*.craft` | `test.svg` | `fileExtensions` key `tests/craft` สำหรับไฟล์ที่ parent โดยตรงชื่อ `tests` |
| `migrations/*.craft` | `migration.svg` | `fileExtensions` key `migrations/craft` สำหรับไฟล์ที่ parent โดยตรงชื่อ `migrations` |
| `build/`, `dist/` | `build.svg` | `folderNames` และ `folderNamesExpanded` สำหรับชื่อ `build` และ `dist` |

VS Code ไม่รับ glob ใน `fileNames`; mapping ที่มี parent ใช้ได้เพียง parent โดยตรง โดย filename มีลำดับเหนือ extension และ extension ที่ระบุ parent มีลำดับเหนือ extension ทั่วไป อ้างอิง [VS Code File Icon Theme API](https://code.visualstudio.com/api/extension-guides/file-icon-theme)

ข้อกำหนดของ adapter:

- ห้ามคัดลอก `fileAssociations` กลางลงเป็น VS Code theme โดยตรง เพราะ schema ต่างกัน
- `src/account_test.craft` ใช้ Craft icon เป็น fallback; `tests/account_test.craft` ใช้ Test icon จาก parent rule ส่วน `tests/unit/account.craft` ใช้ Craft icon เพราะ parent คือ `unit` ต้องระบุข้อจำกัดนี้ในคู่มือ
- กฎ parent อ้างอิงชื่อโฟลเดอร์ ไม่ได้จำกัดเฉพาะโฟลเดอร์ที่ project root; ต้องมีตัวอย่าง nested paths เพื่อไม่ทำให้ผู้ใช้เข้าใจผิด
- จัดเตรียม generic file/folder fallback สำหรับไฟล์ชนิดอื่น และระบุว่าการเลือก theme นี้ไม่ได้รวม icon mappings จาก theme เดิมให้อัตโนมัติ
- หากเพิ่ม `.craftbundle` หรือ associations ใหม่ ต้องเพิ่มใน manifest กลางพร้อมเอกสารและกรณีตรวจรับ ไม่ถือว่าอยู่ในชุดที่เตรียมไว้แล้ว
- ตรวจความชัดของ SVG ที่ขนาดใช้งานจริง 16–24 px บน light/dark/high contrast และเพิ่ม variants เฉพาะเมื่อจำเป็น
- ไอคอน lock/migration เป็นการแสดงประเภทไฟล์ ไม่ใช่หลักฐานว่าระบบ dependency locking หรือ database migration ของภาษาใช้งานได้แล้ว

JetBrains, Craft editor และการแสดงผลใน CLI เป็น adapter backlog หลัง VS Code; ไม่รวมการเปลี่ยนไอคอนใน Windows Explorer, syntax highlighting, LSP หรือการเผยแพร่ Marketplace ในขอบเขตนี้

## 18. Milestones

1. **กำหนด contracts:** signatures, Optional, value semantics, errors และ compatibility
2. **Console/Conversion:** readLine/write, parsing, formatting และ program arguments
3. **Core Data:** Optional, Struct, Map พร้อม formatter/type/runtime support
4. **Library Utilities:** Array/String, Math/Random, File System/Path และ Date/Time
5. **Structured JSON:** JsonValue, parse/stringify และ typed conversion/access
6. **Integration App:** Console Bank พร้อม input validation และ persistence
7. **File Icon Theme:** แปลง Craft Forge manifest เป็น VS Code theme, เตรียม fixtures, VSIX และคู่มือติดตั้ง/เลือก theme; เริ่มทำได้โดยไม่รอ milestones ด้านภาษา
8. **Release:** docs, examples, migration, tests, portable CLI, Windows installer และ Craft Forge VSIX พร้อมสถานะตรวจรับแยกแต่ละ artifact
9. **งานระดับ B:** ประเมินและแยกส่ง Enum/Module/Function values/Event ตาม dependency โดยไม่ปะปนสถานะกับ release หลัก

## 19. โปรแกรมตัวอย่างสำหรับตรวจรับ

### Interactive Bank

- เมนูดูยอด ฝาก ถอน ประวัติ บันทึก และออก
- รับ input จริง จัดการ invalid input, empty input และ EOF
- ใช้ Int หน่วยย่อยสำหรับจำนวนเงินและตรวจ overflow; ไม่ใช้ Float เป็นตัวเก็บยอดเงินในตัวอย่างนี้
- เก็บ Account ด้วย Struct และรายการด้วย Array
- จัดเก็บ/อ่าน JSON ผ่าน conversion ที่เขียนชัดเจน
- หากอ่านไฟล์ที่ผิดรูปแบบต้องแจ้ง error และไม่เขียนทับข้อมูลเดิมโดยอัตโนมัติ

### CLI Utilities

- โปรแกรมอ่าน program arguments
- โปรแกรมอ่านข้อความและนับคำ
- โปรแกรมอ่าน config JSON
- โปรแกรมคำนวณวันเวลา
- ตัวอย่าง Map และ Optional ที่แสดงกรณีไม่พบข้อมูล

## 20. Testing Strategy

ต้องเตรียม unit, integration, Craft tests และ CLI E2E สำหรับ:

- Input: Unicode, blank line, EOF, CRLF/LF, piped input และ cancellation
- Conversion: invalid input, overflow, negative values, boundary และ rounding
- Optional: null assignment, unwrap, scope และการเรียก member โดยยังไม่ตรวจ
- Struct/Map: fields, missing keys, mutability, nested copies และ deterministic iteration
- Collections: empty values, index bounds, sorting และ Unicode substrings
- Filesystem: temporary fixtures, permission failures, overwrite protection และ cleanup
- Time: parse/format, offsets, leap-day cases, durations และ overflow
- JSON: nested values, null, large integers, duplicate keys, malformed documents และ round-trip
- Compatibility: ชุด Rev.1 เดิมและ bundle migration
- CLI: fmt idempotence, fmt รักษา semantics, test reporting, error/exit codes และ arguments หลัง --
- Icon theme: ตรวจ manifest/schema, asset paths, generator output และรายการกฎที่แปลงไม่ได้; ตรวจว่า VSIX บรรจุ SVG และ theme definition ครบ
- Icon fixtures: main.craft, craft.toml, craft.lock, src/account_test.craft, tests/account.craft, tests/unit/account.craft, migrations/001.craft, build/dist ทั้งแบบเปิดและปิด และไฟล์ที่ไม่ใช่ Craft
- ผู้ใช้ตรวจ icon theme ใน VS Code จริง: ติดตั้ง VSIX, เลือก theme, ตรวจ mappings/fallback และ light/dark/high contrast แล้วเปลี่ยนกลับ/ถอนการติดตั้งได้

ผู้ใช้เป็นผู้รัน tests ตาม preference ปัจจุบัน Agent สามารถเขียน tests ตรวจ source และ build artifacts ได้ แต่ต้องไม่บันทึกว่า tests ผ่านจนมีผลจริง และไม่ทดลองใช้โปรแกรม/installer แทนผู้ใช้โดยไม่ได้รับคำสั่ง

## 21. Definition of Done — Rev.2 หลัก

- [ ] Console รับ input และ EOF ได้ พร้อม program arguments
- [ ] Conversion มี typed results, error handling และรูปแบบตัวเลขที่กำหนดชัดเจน
- [ ] Optional, Struct และ Map ทำงานครบทั้ง parser/type/runtime/formatter
- [ ] Array/String utilities และ Math/Random ตามขอบเขตหลักพร้อมเอกสาร
- [ ] File System/Path และ Date/Time ตาม contracts ที่สรุปแล้ว
- [ ] JSON parse/stringify ใช้ข้อมูลแบบมีโครงสร้างได้จริง
- [ ] Interactive Bank รับข้อมูลและบันทึก/โหลดกลับได้
- [ ] ชุด Rev.1 เดิมยังผ่านหรือมี migration note สำหรับทุก breaking change
- [ ] มีผล tests สำหรับฟีเจอร์ใหม่และกรณีผิดพลาด ไม่มีการอ้างผลจากการ build เพียงอย่างเดียว
- [ ] fmt/test/check/run/build รองรับ source ใหม่สอดคล้องกัน
- [ ] เอกสาร ตัวอย่าง และรายการ error codes อัปเดต
- [ ] Portable CLI และ installer build ใหม่ พร้อมตรวจรับบน Windows x64 ที่ไม่มี Go
- [ ] Craft Forge VSIX ใช้ manifest และ SVG ที่เตรียมไว้ มี mapping/fallback ตามข้อ 17.1 และเอกสารข้อจำกัด wildcard ชัดเจน
- [ ] ผู้ใช้ตรวจรับการติดตั้ง เลือก แสดงผล และถอน icon theme ใน VS Code แล้ว พร้อมบันทึก version ที่ทดสอบ
- [ ] สถานะส่งมอบแยก implemented, tested, design-only และ deferred ชัดเจน

อ้างอิง: [Rev.1 roadmap](craft-phase1-rev1.md), [Rev.1 foundations](../../docs/REV1-FOUNDATIONS.md), [กติกาภาษาปัจจุบัน](../../docs/LANGUAGE.md)
