# Craft 0.1.2 — Phase 1 Rev.2 contracts

เอกสารนี้ระบุ API ที่เพิ่มใน implementation; สถานะทดสอบแยกอยู่ใน [Rev.2 status](../roadmap/phase1/output/phase1-rev2-status.md)
กฎจาก [Language reference](LANGUAGE.md) ยังใช้ร่วมกัน ยกเว้นส่วนที่ระบุการเพิ่ม/เปลี่ยนด้านล่าง

## Optional และ null

```craft
let name: String? = null
if let text = std.io.readLine() {
    print(text)
} else {
    print("EOF")
}
```

- `T?` รับ `T` หรือ `null`; ใช้ได้กับ fields, parameters, results และ containers การรับ `T` เข้า `T?` เป็นการห่อ Optional ไม่แปลง numeric types
- `if let` unwrap ได้หนึ่งชั้น สร้าง immutable binding ภายใน then block; binding ไม่อยู่ใน else หรือภายนอก
- เปรียบเทียบ Optional กับ `null` ด้วย `==`/`!=` ได้ แต่ไม่ทำ type narrowing ต้องใช้ `if let` ก่อนใช้ค่าเป็น `T`
- ไม่มี force unwrap และยังไม่รองรับ equality ระหว่าง Optional สองค่า
- รองรับ `T??` เพื่อแยก missing Map entry ออกจาก present entry ที่มีค่า null; unwrap ทีละชั้น
- `Int?[]` คือ Array ของ Optional Int; `Int[]?` คือ Optional Array
- Optional และข้อมูลที่ unwrap ใช้ deep value copy ไม่เปิด mutable aliases
- `null` นอกบริบทชนิด เช่น `print(null)` ใช้ได้ แต่ `[null]` ต้องมี declared element type เช่น `Int?[]`

## Struct

```craft
struct Account {
    let number: String
    var balanceMinor: Int
}
var account: Account = Account(number: "CRAFT-001", balanceMinor: 0)
account.balanceMinor += 100
```

Struct เป็น nominal value type, constructor ต้องใช้ named arguments ครบทุก field (ลำดับใดก็ได้) ตรวจ unknown/duplicate/missing fields
การแก้ field ต้องมี `var` ที่ root และทุก field บนเส้นทางต้องเป็น `var`
parameters, loop bindings และ if-let bindings immutable; ทำสำเนาลง local `var` หากต้องแก้
ห้ามวงจร required fields เช่น `struct Node { var next: Node }`; ใช้ Optional, Array หรือ Map เพื่อเริ่มจากค่าว่างได้
ยังไม่มี methods, inheritance, default fields, user constructors หรือ struct equality
Struct หลายไฟล์ยังอยู่ใน namespace รวมของ project; ยังไม่มี import/module

## Map

```craft
var scores: Map<String, Int> = {"math": 90}
scores.set("english", 85)
if let score = scores.get("math") {
    print(score)
}
```

| Member | Signature / พฤติกรรม |
| --- | --- |
| length | Int อ่านอย่างเดียว |
| get | `(key: String): T?`; missing คืน null |
| require | `(key: String): T`; missing throw |
| set | `(key: String, value: T): Void`; ต้อง var |
| remove | `(key: String): Bool`; ต้อง var; true เมื่อลบได้ |
| containsKey | `(key: String): Bool` |
| keys | `(): String[]`; snapshot |
| values | `(): T[]`; snapshot ลำดับสัมพันธ์กับ keys |

keys เป็น String เท่านั้น; literal keys ต้องเป็น String literals ไม่มี duplicate keys
ลำดับคือ insertion order; replace อยู่ที่เดิม, remove แล้ว set ใหม่อยู่ท้าย
empty map ต้องมีชนิดจาก variable/parameter/result context เช่น `let m: Map<String, Int> = {}`
Map, Array, Struct ใช้ deep copy เมื่อ bind/assign/pass/return/insert; `require`, `get`, `values` ไม่เปิด alias
ยังไม่มี Map index assignment, Map equality หรือ user-defined generics
การแสดง Map/Struct ด้วย print ใช้ลำดับ insertion/field declaration ที่แน่นอน

## Console, arguments และ environment

| API | Signature |
| --- | --- |
| std.io.write | `(text: String): Void` ไม่เติม newline |
| std.io.readLine | `(): String?` |
| std.env.args | `(): String[]` |
| std.env.lookup | `(name: String): String?` |

readLine รับ UTF-8, รองรับ pipe และ LF/CRLF ตัดเฉพาะ line terminator; ไม่ trim spaces หรือ CR ที่ไม่มี LF ตามท้าย
blank line คืน `""`, EOF คืน null, บรรทัดสุดท้ายที่ไม่มี newline ยังคืนข้อความ
กำหนดขนาดสูงสุด 1 MiB ไม่รวม terminator; oversized line ถูกอ่านทิ้งจนจบแล้ว throw; invalid UTF-8/read errors throw
Ctrl+C ใน CLI ยกเลิก context และออกด้วย exit 1; ไม่ถูก catch เป็น Craft Exception
บน Windows console EOF ใช้ Ctrl+Z แล้ว Enter; host/terminal อาจมี keyboard shortcuts ต่างกัน

Go embedders ใช้ `interpreter.RunWithInput` หรือ `cli.RunWithInput` เพื่อ inject reader/arguments
readLine ยกเลิกการรอได้ แต่ Go ไม่สามารถยกเลิก arbitrary reader ที่กำลัง block ได้ทุกชนิด
เมื่อยกเลิก เจ้าของ custom reader ต้อง Close/unblock reader เพื่อคืน worker goroutine; CLI process จบแล้ว OS เก็บทรัพยากร
`interpreter.Run` และแต่ละ `RunTest` ใช้ empty input/arguments โดย default เพื่อไม่รอ terminal

`craft run -- one --flag` และ `craft run file.craftbundle -- one --flag` ส่งเฉพาะ arguments หลัง `--`
`lookup` แยก unset (null) จาก present-empty (`""`); `std.env.get` เดิมยังคืน empty String เมื่อ unset
ไม่มี API แก้ environment ของ parent process

## Conversion และ String

| API | Signature |
| --- | --- |
| std.convert.toInt | `(text: String): Int` |
| std.convert.tryInt | `(text: String): Int?` |
| std.convert.toFloat | `(text: String): Float` |
| std.convert.tryFloat | `(text: String): Float?` |
| std.convert.toBool | `(text: String): Bool` |
| std.convert.toString | `(value: Int/Float/Bool/String): String` builtin overload เฉพาะ primitive |
| std.string.fixed | `(value: Float, digits: Int): String` |
| std.string.join | `(parts: String[], separator: String): String` |
| String.replace | `(old: String, replacement: String): String` ทุกตำแหน่ง |
| String.substring | `(start: Int, end: Int): String` |

ไม่มี implicit Int/Float/String conversion; แปลงเลขต่างชนิดผ่านข้อความอย่างชัดเจนเมื่อต้องใช้
Int ใช้ signed decimal 64-bit รับเครื่องหมาย +/-, ไม่รับ underscores/hex/whitespace
Float รับ decimal/exponent (`1.5`, `.5`, `1e3`), ต้อง finite; ไม่รับ NaN/Infinity/hex/whitespace และไม่ละเลยข้อความท้าย
tryInt/tryFloat คืน null เมื่อแปลงไม่ได้; toInt/toFloat/toBool throw
Bool รับเฉพาะ `true`/`false` case-sensitive
toString ใช้ decimal Int, shortest general Float representation, lowercase Bool และ String เดิม ไม่ขึ้นกับ locale

fixed รับ digits 0..18 ใช้ round-to-nearest, ties-to-even บนค่า binary Float จริง
เช่น `fixed(2.5, 0) == "2"`, `fixed(3.5, 0) == "4"`; ตัวเลขฐานสิบบางค่าแทนใน Float ไม่พอดี
substring นับ Unicode code points (ไม่ใช่ grapheme clusters), end-exclusive; ต้อง `0 <= start <= end <= length`
replace กรณี old ว่างจะแทรก replacement ก่อน/หลังและระหว่าง Unicode code points ตามพฤติกรรม String library

## Array utilities

- `contains(value: T): Bool` และ `indexOf(value: T): Int?` สำหรับ primitive element types; ใช้ exact equality, คืนตำแหน่งแรก
- `slice(start: Int, end: Int): T[]` end-exclusive ตรวจ bounds เช่นเดียวกับ substring
- `reverse(): T[]` ทุก element type
- `sorted(): T[]` สำหรับ Int, Float, String; ascending, stable; String เปรียบเทียบตามลำดับ UTF-8/code points ไม่ใช้ locale
- ทั้งหมดอ่านได้จาก let และคืนสำเนา ไม่แก้ต้นฉบับ; `append`/`remove` เดิมยังต้อง var

## Math และ Random

| API | Signature |
| --- | --- |
| std.math.round/floor/ceil | `(value: Float): Float` |
| std.math.pow | `(base: Float, exponent: Float): Float` |
| std.math.random | `(seed: Int): Random` |
| Random.nextInt | `(min: Int, max: Int): Int`; min-inclusive/max-exclusive |
| Random.nextFloat | `(): Float`; [0,1) |

round ties away from zero (`-2.5 -> -3.0`); ต่างจาก fixed ซึ่งจัดรูปแบบด้วย ties-to-even
pow throw เมื่อผล NaN/Infinity เช่น negative base กับ fractional exponent; `0^0 == 1`
Random ใช้ copyable SplitMix64 state; seed เดียวให้ลำดับเดียว, assignment เป็นสำเนา state อิสระ
nextInt/nextFloat ต้อง receiver ที่แก้ค่าได้ (`var`) และไม่ใช้ global seed
nextInt ตรวจ min < max และใช้ rejection sampling ครอบคลุมช่วง signed 64-bit ที่ถูกต้อง
ไม่ใช่ cryptographic random; ไม่มี random แบบ seed ตามเวลาซ่อนอยู่

## Filesystem และ Path

| API | Signature / contract |
| --- | --- |
| std.fs.appendText | `(path: String, text: String): Void`; สร้างไฟล์หากไม่มี ไม่สร้าง parents |
| std.fs.createDirectory | `(path: String): Void`; สร้าง parents, existing directory สำเร็จ |
| std.fs.listDirectory | `(path: String): String[]`; ชื่อ entries ไม่รวม parent, sort ตามชื่อ |
| std.fs.isDirectory | `(path: String): Bool`; missing false, other errors throw |
| std.fs.removeFile | `(path: String): Void`; regular file เท่านั้น ไม่ recursive |
| std.fs.copyFile/moveFile | `(source: String, destination: String): Void` |
| std.path.join | `(parts: String[]): String` |
| std.path.fileName/extension/parent | `(path: String): String` |
| std.path.isAbsolute | `(path: String): Bool` |

copy/move เปิด destination แบบ exclusive ปฏิเสธไฟล์ที่มีอยู่เสมอ ไม่มี overwrite option ใน revision นี้
move ใช้ copy แล้ว remove source (รองรับข้าม volume, ไม่ atomic); ถ้าลบ source ไม่ได้ แจ้งว่ามีทั้งสองไฟล์
append/remove/copy/move ปฏิเสธ symlink ที่ตัวไฟล์, directory helpers ใช้ OS behavior และอาจตาม symlink ใน parents
readText/writeText/exists เดิมคง semantics รวมถึงการตาม symlink และ writeText เขียนทับ; writeText ไม่ atomic
API นี้ไม่ใช่ filesystem sandbox หรือการป้องกัน race จาก process อื่น
paths อ้างอิง process working directory แม้ discovery หา project root เจอที่อื่น
Path helpers ใช้ native OS semantics: Windows รับ separators ตามระบบและ drive/UNC; Unicode paths ใช้ได้
extension รวมจุด เช่น `.json`; empty join คืน `""`, basename/parent ของ empty path คืน `"."`
missing/permission/wrong file type errors catch ได้พร้อม source location

## DateTime และ Duration

| API | Signature |
| --- | --- |
| std.time.parse | `(text: String): DateTime` RFC3339/RFC3339Nano |
| std.time.fromUnixSeconds | `(seconds: Int): DateTime` UTC |
| std.time.durationMilliseconds | `(milliseconds: Int): Duration` |
| DateTime.format | `(): String` RFC3339Nano |
| DateTime.unixSeconds/unixMilliseconds | `(): Int` |
| DateTime.withOffset | `(minutes: Int): DateTime` |
| DateTime.add/subtract | `(duration: Duration): DateTime` |
| DateTime.difference | `(other: DateTime): Duration` receiver minus other |
| Duration.milliseconds | `(): Int` truncate toward zero |

DateTime/Duration เป็น immutable built-in value types ไม่ใช่ user Struct, ไม่ใช้ Int สลับหน่วยโดยปริยาย
DateTime รองรับปี 1..9999, offset -1439..1439 minutes, format คง fixed offset; ไม่มี named timezones/DST หรือ locale-dependent format
parse ต้องระบุ timezone; fractional seconds เก็บละเอียดถึง nanoseconds ส่วนที่เกินถูกตัดตาม parser
Unix seconds/milliseconds ปัดลงตามตำแหน่ง instant; Duration รองรับ signed 64-bit nanoseconds (~292 ปี) และตรวจ overflow
add/subtract/difference ตรวจช่วงและ throw แทน saturating result
`std.time.now(): String` และ `sleep(milliseconds)` เดิมยังใช้ได้

## Structured JSON

| API | Signature |
| --- | --- |
| std.json.parse | `(text: String): JsonValue` |
| std.json.stringify | `(value: JsonValue): String` |
| std.json.nullValue | `(): JsonValue` |
| std.json.fromInt/fromFloat/fromBool/fromString | `(value: corresponding primitive): JsonValue` |
| std.json.fromArray | `(values: JsonValue[]): JsonValue` |
| std.json.fromObject | `(values: Map<String, JsonValue>): JsonValue` |
| JsonValue.kind | `(): String` — null/bool/number/string/array/object |
| JsonValue.asInt/asFloat/asBool/asString | `(): corresponding primitive` |
| JsonValue.asArray | `(): JsonValue[]` |
| JsonValue.asObject | `(): Map<String, JsonValue>` |

JsonValue เก็บข้อมูลภายใน immutable; accessors คืน Craft values ที่แก้โดยไม่กระทบต้นฉบับ
ผิดชนิดหรือเลขเกินช่วง throw; asInt รับ integer spelling เท่านั้น (`1.0`/`1e0` ไม่ถือเป็น Int)
parse เก็บ exact numeric spelling ทำให้ Int64 ไม่สูญเสีย precision; asFloat คือการแปลงที่ผู้ใช้เลือกเองและอาจปัดค่า
object serialization และ asObject keys เรียงชื่อแน่นอน ไม่เก็บลำดับจาก JSON text
ปฏิเสธ duplicate keys (รวม keys ที่ escape แล้วชื่อเดียวกัน), malformed/trailing data, invalid UTF-8, input/output >16 MiB และ nesting >128
Unicode escapes รองรับ surrogate pairs; unpaired surrogates ถูกแทนด้วย U+FFFD ตาม JSON decoder ของ Go
stringify ใช้ compact JSON, escaping ตาม Go encoding/json รวม HTML-sensitive characters
ไม่มี implicit Struct serialization, Any หรือ dynamic properties; ต้อง map fields ชัดเจนดังตัวอย่าง Bank
`isValid` และ `quote` เดิมยังคง behavior; isValid ตรวจ JSON syntax จึงอาจรับ duplicate keys ที่ parse ปฏิเสธ

## Diagnostics และ compatibility

- E1001 / exit 2: syntax เช่น Map key type ไม่ใช่ String
- E2001 / exit 3: unknown symbol; E2002 / exit 3: type/constructor/Optional misuse/required recursive fields
- E2003 / exit 3: แก้ immutable variable/field หรือเรียก mutating method ผ่าน immutable receiver
- E3001 / exit 1: catchable runtime/library failures; E3002 explicit throw; E3003 assertion; รักษา source location และ stack trace เดิม
- E4001/E4002 / exit 4: entry/project/bundle errors; E5001 / exit 5: CLI usage
- CLI 0.1.2 สร้าง source bundle format 2, language 0.1.2 และอ่าน format 2 language 0.1.1 ได้ด้วย
- Toolchain 0.1.1 อ่าน bundle ภาษา 0.1.2 ไม่ได้ ให้ใช้ CLI ใหม่; format 1 ต้อง migrate/rebuild เหมือนเดิม
- `struct` และ `null` เป็น reserved keywords ใหม่: หาก Rev.1 ใช้เป็นชื่อ variable/function ให้เปลี่ยนชื่อ; built-in type names ใหม่ห้ามใช้เป็นชื่อ Struct
- กฎ let/var, exceptions, defer, Array copy และ named arguments ของ Rev.1 คงเดิม
- Module/import, Enum, Lambda/Callback/Event เป็น deferred ระดับ B; HTTP/process/REPL เป็นระดับ C
