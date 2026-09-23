# Craft Phase 1 Rev.1

## Language Strengthening and Core Experience

เอกสารนี้กำหนดแนวทางพัฒนา Craft Phase 1 Rev.1 ซึ่งเป็นระยะต่อยอดจาก Phase 1 โดยมุ่งเน้นการทำให้ภาษาแข็งแรงขึ้น ใช้งานสนุกขึ้น และมีโครงสร้างที่ใกล้เคียงภาษาโปรแกรมสมัยใหม่มากขึ้น

Phase 1 เดิมทำให้ Craft สามารถ Lexer, Parser, AST, Static Type Checker, Interpreter และ CLI ได้จริงแล้ว

Phase 1 Rev.1 จะเน้นการพัฒนา “คุณภาพของภาษา” มากกว่าการเพิ่มความสามารถด้าน API หรือ Framework

เป้าหมายหลักคือ:

> ทำให้ Craft เป็นภาษาที่เขียนโปรแกรมจริงได้อย่างปลอดภัย อ่านง่าย และมีเอกลักษณ์ของตัวเอง

---

## 1. เป้าหมายของ Phase 1 Rev.1

Phase 1 Rev.1 มีเป้าหมายดังนี้:

- ทำให้กฎของภาษาชัดเจนและสอดคล้องกัน
- เพิ่มระบบจัดการ Runtime Error
- เพิ่มโครงสร้างข้อมูลพื้นฐาน
- เพิ่มความปลอดภัยของตัวแปรและ Scope
- เพิ่มความสามารถในการทดสอบโปรแกรม
- เพิ่มเครื่องมือช่วยจัดรูปแบบโค้ด
- ปรับปรุง Diagnostics และ Stack Trace
- วางรากฐานสำหรับ Struct, Enum, Module และ Event ในอนาคต
- ทำให้ภาษาเหมาะกับการทดลองและพัฒนาต่อได้ง่าย

Phase นี้ยังไม่มุ่งเน้นการสร้าง API Framework, Database หรือระบบ Production Runtime

---

## 2. หลักการออกแบบ

### 2.1 อ่านง่ายก่อนความซับซ้อน

Syntax ของ Craft ต้องชัดเจนและเข้าใจได้ง่าย โดยยังคงแนวทาง Modern C-Family:

```craft
func main() {
    let message: String = "Hello Craft"
    print(message)
}
```

### 2.2 ปลอดภัยตั้งแต่ระดับภาษา

Compiler และ Runtime ควรช่วยป้องกันข้อผิดพลาดที่ตรวจพบได้ตั้งแต่ก่อนรัน เช่น:

- ใช้ตัวแปรที่ไม่มีอยู่
- กำหนดค่าผิดชนิด
- แก้ไขค่าคงที่
- เรียก Function ผิดจำนวน Argument
- คืนค่าผิดชนิด
- ใช้ `break` นอก Loop
- ใช้ `return` ผิดบริบท

### 2.3 ฟีเจอร์ต้องมีเหตุผล

ไม่เพิ่มฟีเจอร์เพียงเพราะภาษาอื่นมี แต่ต้องพิจารณาว่า:

- ช่วยให้โค้ดอ่านง่ายขึ้นหรือไม่
- ลดข้อผิดพลาดหรือไม่
- ทำให้การออกแบบภาษาแข็งแรงขึ้นหรือไม่
- ทำให้ Runtime และ Compiler ซับซ้อนเกินไปหรือไม่

### 2.4 แยกภาษาออกจาก Library

สิ่งที่เป็นโครงสร้างพื้นฐานของภาษา เช่น:

```text
ตัวแปร
Function
Type
Scope
Control Flow
Exception
```

ควรอยู่ใน Language Core

ส่วนความสามารถที่เป็นพฤติกรรมของระบบ เช่น:

```text
Event Bus
File System
JSON
HTTP
Timer
```

ควรพัฒนาเป็น Standard Library ก่อน ไม่ควรฝังทุกอย่างลงใน Syntax ของภาษา

---

## 3. ขอบเขตหลัก

Phase 1 Rev.1 แบ่งงานออกเป็น 5 กลุ่ม:

```text
A. Language Safety
B. Error Handling
C. Core Data Structures
D. Developer Experience
E. Runtime Foundation
```

---

# 4. Language Safety

## 4.1 `let` และ `var`

Craft Rev.1 จะแยกตัวแปรค่าคงที่และตัวแปรที่เปลี่ยนค่าได้อย่างชัดเจน

```craft
let name: String = "Craft"
var count: Int = 0

count += 1
```

กฎ:

- `let` ไม่สามารถเปลี่ยนค่าได้หลังจากประกาศ
- `var` สามารถเปลี่ยนค่าได้
- ทั้งสองแบบต้องระบุ Type ในช่วง Rev.1
- ไม่สามารถกำหนดค่าใหม่ให้ `let`
- Compiler ต้องแจ้ง Error หากพยายามแก้ไข `let`

ตัวอย่าง Error:

```text
cannot assign to immutable variable 'name'
```

## 4.2 Scope

ตัวแปรต้องมีขอบเขตการใช้งานที่ชัดเจน:

```craft
func main() {
    let message: String = "Hello"

    if true {
        let value: Int = 10
        print(value)
    }

    // value ไม่สามารถใช้งานที่นี่ได้
}
```

กฎ:

- ตัวแปรใน Block ใช้ได้เฉพาะภายใน Block
- ตัวแปรใน Function ใช้ได้เฉพาะภายใน Function
- ห้ามประกาศชื่อซ้ำใน Scope เดียวกัน
- การ Shadowing ต้องกำหนดกฎให้ชัดเจน
- ไม่สามารถใช้ตัวแปรก่อนประกาศ

## 4.3 Function Return Safety

Compiler ต้องตรวจสอบว่า Function คืนค่าถูกต้องทุกเส้นทาง:

```craft
func getValue(flag: Bool): Int {
    if flag {
        return 10
    }

    return 0
}
```

กรณีผิด:

```craft
func getValue(flag: Bool): Int {
    if flag {
        return 10
    }
}
```

ต้องแจ้ง:

```text
missing return value on some execution paths
```

## 4.4 Control Flow Safety

ต้องตรวจสอบการใช้คำสั่งให้ถูกบริบท:

```craft
break       // ใช้ได้เฉพาะใน Loop
continue    // ใช้ได้เฉพาะใน Loop
return      // ใช้ได้เฉพาะใน Function
throw       // ใช้เพื่อหยุดการทำงานและส่ง Error
```

---

# 5. Error Handling

## 5.1 เป้าหมาย

Runtime Error ต้องไม่ทำให้โปรแกรมหยุดโดยไม่มีข้อมูลที่เป็นประโยชน์

ทุก Error ควรมี:

- Error Type
- Error Message
- Source File
- Line
- Column
- Stack Trace เมื่อมี Function Call หลายระดับ

## 5.2 `Exception`

Craft Rev.1 จะเพิ่ม Exception พื้นฐาน:

```craft
Exception("something went wrong")
```

## 5.3 `throw`

ใช้ส่ง Error จาก Function หรือ Block:

```craft
func divide(a: Int, b: Int): Int {
    if b == 0 {
        throw Exception("cannot divide by zero")
    }

    return a / b
}
```

## 5.4 `try / catch`

ใช้จัดการ Exception:

```craft
func main() {
    try {
        let result: Int = divide(10, 0)
        print(result)
    } catch error {
        print(error.message)
    }
}
```

กฎเบื้องต้น:

- `try` ต้องมี `catch`
- `catch` รับ Error ที่ถูกส่งออกมา
- Exception สามารถส่งผ่านหลาย Function ได้
- หากไม่มี `catch` ที่รองรับ Exception โปรแกรมจะจบด้วย Runtime Error
- Error ที่ไม่ได้จัดการต้องแสดง Stack Trace

## 5.5 Stack Trace

ตัวอย่างผลลัพธ์:

```text
Exception: cannot divide by zero

at divide (src/math.craft:4)
at main (src/main.craft:10)
```

## 5.6 สิ่งที่ยังไม่รวมใน Rev.1

ยังไม่รองรับ:

- Typed Exception
- Custom Exception หลายระดับ
- Exception Filter
- `throws` ใน Function Signature
- Exception Chaining
- `finally`
- ระบบ Resource Cleanup ขั้นสูง

ความสามารถเหล่านี้อาจเพิ่มใน Phase ถัดไปหลังจากระบบพื้นฐานเสถียรแล้ว

---

# 6. Core Data Structures

## 6.1 Array

เพิ่ม Array สำหรับเก็บข้อมูลหลายค่า:

```craft
let numbers: Int[] = [1, 2, 3, 4]

for number in numbers {
    print(number)
}
```

ความสามารถขั้นต่ำ:

- สร้าง Array
- อ่านค่าด้วย Index
- เปลี่ยนค่าใน Array ที่เป็น `var`
- อ่านจำนวนสมาชิก
- เพิ่มสมาชิก
- ลบสมาชิก
- ตรวจสอบ Index

ตัวอย่าง:

```craft
var numbers: Int[] = [1, 2, 3]

numbers.append(4)
print(numbers[0])
print(numbers.length)
```

กรณี Index ไม่ถูกต้องต้องแสดง Runtime Error ที่ชัดเจน

## 6.2 String Operations

เพิ่มความสามารถพื้นฐานสำหรับ String:

```craft
let name: String = "Craft"

print(name.length)
print(name.upper())
print(name.lower())
```

ความสามารถเบื้องต้น:

- `length`
- `contains`
- `startsWith`
- `endsWith`
- `trim`
- `upper`
- `lower`
- `split`
- การต่อ String

## 6.3 Map

Map อาจเริ่มพัฒนาใน Rev.1 ได้ แต่ไม่จำเป็นต้องเสร็จในช่วงแรก:

```craft
let scores: Map<String, Int> = {
    "math": 90,
    "english": 85
}
```

หากยังไม่พร้อม ให้เลื่อนไป Rev.2 หลังจาก Array และ Type System เสถียร

---

# 7. `defer`

`defer` ใช้กำหนดคำสั่งที่ต้องทำงานก่อนออกจาก Function:

```craft
func readFile(): String {
    let file = open("data.txt")

    defer {
        file.close()
    }

    return file.read()
}
```

คำสั่งใน `defer` ต้องทำงานเมื่อ Function จบจาก:

- `return`
- `throw`
- การทำงานปกติ
- Exception

เป้าหมายของ `defer` คือช่วยจัดการ Resource เช่น:

- File
- Lock
- Connection
- Temporary Resource

`defer` ควรพัฒนาหลังจากระบบ `try / catch / throw` ทำงานได้แล้ว

---

# 8. Struct และ Enum

## 8.1 Struct

Struct เป็นโครงสร้างข้อมูลที่ผู้ใช้กำหนดเอง:

```craft
struct User {
    let id: Int
    let name: String
    let email: String
}
```

การสร้างและใช้งาน:

```craft
let user = User(
    id: 1,
    name: "Kampanat",
    email: "user@example.com"
)

print(user.name)
```

เป้าหมายของ Struct ในระยะแรก:

- ประกาศ Field
- ระบุ Type ของ Field
- สร้าง Instance
- อ่าน Field
- ตรวจสอบ Type
- รองรับ `let` และ `var` Field

## 8.2 Enum

Enum ใช้สร้างกลุ่มค่าที่กำหนดไว้ล่วงหน้า:

```craft
enum Status {
    active
    inactive
    pending
}
```

การใช้งาน:

```craft
let status: Status = Status.active
```

Enum ควรเพิ่มหลังจากระบบ Type และ Struct พื้นฐานเสถียรแล้ว

---

# 9. Optional และ Null Safety

Optional ใช้ระบุค่าที่อาจไม่มีค่า:

```craft
let nickname: String? = null
```

การตรวจสอบ:

```craft
if nickname != null {
    print(nickname)
}
```

เป้าหมาย:

- แยกค่าปกติออกจากค่าที่ไม่มี
- ป้องกัน Null Runtime Error
- บังคับให้ผู้ใช้ตรวจสอบค่าก่อนนำไปใช้งาน
- รองรับ Optional ใน Function Return

ฟีเจอร์นี้ควรเริ่มออกแบบใน Rev.1 แต่สามารถพัฒนาเต็มรูปแบบใน Rev.2 ได้

---

# 10. Function Improvements

## 10.1 Named Arguments

```craft
func greet(name: String, message: String) {
    print(message + " " + name)
}

greet(
    name: "Craft",
    message: "Welcome"
)
```

## 10.2 Default Parameters

```craft
func greet(name: String, message: String = "Hello") {
    print(message + " " + name)
}
```

Default Parameters สามารถเลื่อนไป Rev.2 หากทำให้ Parser และ Type Checker ซับซ้อนเกินไป

## 10.3 Lambda

```craft
let double = (value: Int): Int => value * 2
```

Lambda จะเป็นพื้นฐานสำหรับ:

- Callback
- Array Operations
- Event
- Async
- Higher-order Function

## 10.4 Function เป็นค่า

ในอนาคต Function ควรสามารถส่งเป็น Argument ได้:

```craft
func apply(value: Int, action: Function): Int {
    return action(value)
}
```

ความสามารถนี้ควรพัฒนาหลังจากระบบ Function ปัจจุบันเสถียรแล้ว

---

# 11. Event และ Callback

## 11.1 แนวทางการออกแบบ

Event ไม่ควรเป็น Keyword ของภาษาในทันที

ลำดับที่แนะนำคือ:

```text
Function
→ Lambda
→ Callback
→ Event Bus ใน Standard Library
→ Event Syntax หากจำเป็น
```

## 11.2 Callback

```craft
func onMessage(message: String) {
    print(message)
}
```

## 11.3 Event Bus

Event Bus ควรอยู่ใน Standard Library:

```craft
eventBus.on("user.created", onUserCreated)
eventBus.emit("user.created", user)
```

## 11.4 Event Syntax ในอนาคต

หากพบว่า Event เป็นรูปแบบที่ใช้บ่อยและเหมาะกับ Craft อาจเพิ่ม Syntax ภายหลัง:

```craft
event UserCreated {
    let userId: Int
    let name: String
}

on UserCreated event {
    print(event.name)
}
```

แต่ Phase 1 Rev.1 ยังไม่บังคับให้มี `event` เป็น Keyword

---

# 12. Module และ Import

เนื่องจาก Craft รองรับหลายไฟล์แล้ว ควรเริ่มวางรากฐาน Module System:

```craft
module users
```

การ Import:

```craft
import users
```

หรือ:

```craft
import users.User
```

ควรกำหนด:

- การ Export Function
- การ Export Struct
- Public และ Private Symbol
- การจัดการชื่อซ้ำ
- Circular Import
- ความสัมพันธ์ระหว่าง File และ Module
- ลำดับการ Compile หลายไฟล์

ตัวอย่าง:

```craft
public func createUser(name: String): User {
    ...
}

private func validateName(name: String): Bool {
    ...
}
```

Module System อาจเริ่มใน Rev.1 แต่สามารถทำให้สมบูรณ์ใน Rev.2

---

# 13. Developer Experience

## 13.1 `craft fmt`

จัดรูปแบบ Source Code:

```bash
craft fmt
```

เป้าหมาย:

- รูปแบบโค้ดเหมือนกัน
- ลดการถกเถียงเรื่อง Style
- ทำงานได้ทั้ง Project และไฟล์เดียว
- ไม่เปลี่ยนความหมายของโปรแกรม

## 13.2 `craft test`

เพิ่มระบบทดสอบพื้นฐาน:

```bash
craft test
```

ตัวอย่าง Syntax:

```craft
test "addition works" {
    assert add(2, 3) == 5
}
```

ความสามารถเบื้องต้น:

- Test Block
- `assert`
- แสดง Test ที่ผ่านและไม่ผ่าน
- Exit Code ตามผลการทดสอบ
- แสดง Source Location เมื่อ Test ล้มเหลว

## 13.3 `craft repl`

ในอนาคตอาจเพิ่ม REPL:

```bash
craft repl
```

ตัวอย่าง:

```text
> 10 + 20
30

> "Craft".length
5
```

REPL เหมาะกับการทดลองภาษาและตรวจสอบ Expression อย่างรวดเร็ว

## 13.4 Error Code

Error ควรมี Code ที่แน่นอน:

```text
E1001 Syntax error
E2001 Unknown identifier
E2002 Type mismatch
E2003 Immutable variable
E3001 Runtime error
E3002 Unhandled exception
E4001 Invalid entry point
```

ตัวอย่าง:

```text
E2002 type mismatch: cannot assign String to Int
 --> src/main.craft:4:5
```

## 13.5 Exit Code

CLI ต้องมี Exit Code ที่สม่ำเสมอ:

```text
0  สำเร็จ
1  Runtime Error
2  Syntax Error
3  Type Error
4  Project Error
5  Invalid CLI Usage
```

---

# 14. Standard Library Foundation

Phase 1 Rev.1 ยังไม่ต้องมี Framework แต่ควรเริ่มวาง Standard Library:

```text
std.string
std.array
std.map
std.math
std.io
std.fs
std.time
std.json
std.env
```

ลำดับที่แนะนำ:

1. `std.string`
2. `std.array`
3. `std.map`
4. `std.math`
5. `std.fs`
6. `std.time`
7. `std.json`
8. `std.env`

Event Bus ควรอยู่ใน Standard Library เช่น:

```text
std.events
```

ไม่ควรเป็นส่วนหนึ่งของ Language Core จนกว่าจะมีรูปแบบที่ชัดเจน

---

# 15. Architecture Changes

สถาปัตยกรรมของ Craft Rev.1:

```text
Craft Source
    ↓
Lexer
    ↓
Parser
    ↓
AST
    ↓
Name Resolver
    ↓
Static Type Checker
    ↓
Lowering / Validation
    ↓
Interpreter
    ↓
Runtime
    ↓
Program Output
```

ส่วนที่เพิ่มขึ้น:

### Name Resolver

จัดการ:

- Scope
- Symbol
- Variable Binding
- Function Binding
- Module Binding
- Struct Field
- Enum Value

### Runtime Error System

จัดการ:

- Exception
- Stack Trace
- Runtime Error
- Source Location
- Error Propagation
- Exit Code

### Standard Library Layer

แยก Built-in Function ออกจาก Interpreter โดยตรง:

```text
Language Core
    ↓
Runtime
    ↓
Standard Library
```

การแยกนี้จะช่วยให้ในอนาคตสามารถเปลี่ยนจาก Interpreter ไปเป็น Bytecode VM ได้ง่ายขึ้น

---

# 16. ลำดับการพัฒนา

## Milestone 1: Language Safety

- เพิ่ม `var`
- ปรับความหมายของ `let`
- ตรวจ Immutable Variable
- ปรับ Scope
- ตรวจตัวแปรก่อนประกาศ
- ตรวจ Function Return ทุกเส้นทาง
- ตรวจ `break`, `continue`, `return`

## Milestone 2: Error Handling

- เพิ่ม `Exception`
- เพิ่ม `throw`
- เพิ่ม `try`
- เพิ่ม `catch`
- เพิ่ม Error Propagation
- เพิ่ม Stack Trace
- เพิ่ม Error Code
- เพิ่ม Exit Code

## Milestone 3: Core Data

- เพิ่ม Array
- เพิ่ม Array Index Validation
- เพิ่ม String Operations
- เพิ่ม Array Methods
- ออกแบบ Map
- เริ่มออกแบบ Struct

## Milestone 4: Runtime Foundation

- เพิ่ม `defer`
- จัดการ Function Frame
- จัดการ Exception Unwinding
- ทดสอบ Nested Function Calls
- ทดสอบ Error ในหลายไฟล์
- ทดสอบ Runtime Cleanup

## Milestone 5: Developer Tools

- เพิ่ม `craft fmt`
- เพิ่ม `craft test`
- เพิ่ม `assert`
- เพิ่ม Test Reporter
- ปรับปรุง Diagnostics
- เพิ่ม REPL หากพร้อม

## Milestone 6: Module Foundation

- ออกแบบ `module`
- ออกแบบ `import`
- Public / Private Symbol
- ตรวจ Circular Import
- ทดสอบการเรียกข้าม Module

## Milestone 7: Standard Library

- `std.string`
- `std.array`
- `std.math`
- `std.fs`
- `std.time`
- `std.json`
- `std.events`

---

# 17. Testing Strategy

ต้องทดสอบทั้ง Compiler, Runtime และ CLI

## Language Safety Tests

```craft
let value: Int = 10
value = 20
```

ต้องถูกปฏิเสธ เพราะ `let` เปลี่ยนค่าไม่ได้

## Exception Tests

```craft
func main() {
    try {
        throw Exception("test error")
    } catch error {
        print(error.message)
    }
}
```

ผลลัพธ์ต้องเป็น:

```text
test error
```

## Unhandled Exception Test

```craft
func main() {
    throw Exception("fatal error")
}
```

ต้องแสดง:

- Error Message
- Source Location
- Stack Trace
- Exit Code ที่ถูกต้อง

## Array Tests

```craft
func main() {
    var numbers: Int[] = [1, 2, 3]
    numbers.append(4)
    print(numbers.length)
}
```

ผลลัพธ์:

```text
4
```

## Function Return Tests

ต้องทดสอบ:

- Return ครบทุกเส้นทาง
- Return ผิดชนิด
- Return ใน Function แบบ `Void`
- Return ภายใน Loop
- Return ภายใน `try`

## Tool Tests

ต้องทดสอบ:

```bash
craft fmt
craft test
craft check
craft run
craft build
```

รวมถึง:

- Project หลายไฟล์
- Syntax Error
- Type Error
- Runtime Error
- Unhandled Exception
- Invalid Project
- Invalid CLI Arguments

---

# 18. Definition of Done

Craft Phase 1 Rev.1 ถือว่าสำเร็จเมื่อ:

## Language

- รองรับ `let`
- รองรับ `var`
- แยก Immutable และ Mutable ได้
- มี Scope ที่ชัดเจน
- ตรวจการใช้ตัวแปรก่อนประกาศ
- ตรวจ Function Return ครบทุกเส้นทาง
- รองรับ Array
- รองรับ String Operations
- รองรับ `try / catch / throw`
- แสดง Stack Trace ได้
- รองรับ Runtime Error ที่มี Source Location

## Runtime

- Exception ส่งผ่าน Function ได้
- Unhandled Exception แสดง Error อย่างเหมาะสม
- Runtime ไม่หยุดโดยไม่มีข้อมูล
- มี Exit Code ที่สม่ำเสมอ
- รองรับ `defer` หรือมีข้อกำหนดชัดเจนว่าเลื่อนไป Phase ถัดไป

## Tooling

- `craft fmt` ทำงานได้
- `craft test` ทำงานได้
- มี `assert`
- Error มี Error Code
- Unit Tests ผ่าน
- Integration Tests ผ่าน
- End-to-End Tests ผ่าน

## Compatibility

- โปรแกรม Phase 1 เดิมยังทำงานได้ หรือมี Migration Note
- Syntax ที่เปลี่ยนแปลงมีเอกสาร
- ตัวอย่างทั้งหมดถูกอัปเดต
- Installer และ Portable CLI ถูก Build ใหม่
- ทดสอบบน Windows x64 ที่ไม่มี Go ได้

---

# 19. สิ่งที่ยังไม่อยู่ใน Phase 1 Rev.1

ยังไม่รวม:

- Generics
- Typed Exception
- Custom Exception ขั้นสูง
- `finally`
- Async/Await
- Concurrency
- Native Compilation
- Bytecode VM
- Package Manager
- HTTP Framework
- Database
- Migration
- OpenAPI
- Dependency Injection
- Reflection
- Garbage Collector แบบกำหนดเอง
- Language Server เต็มรูปแบบ

สิ่งเหล่านี้สามารถนำไปพิจารณาใน Phase 2 หลังจาก Language Core และ Runtime มีเสถียรภาพ

---

# 20. แนวทางต่อยอดหลัง Phase 1 Rev.1

```text
Phase 1
    Language Core + Interpreter

Phase 1 Rev.1
    Language Safety
    Exception
    Array
    Struct Foundation
    fmt
    test
    Stack Trace
    Standard Library Foundation

Phase 1 Rev.2
    Enum
    Optional
    Map
    Module / Import
    Lambda
    Callback
    Event Library

Phase 2
    Generics
    Pattern Matching
    Async / Await
    Concurrency
    Bytecode VM

Phase 3
    Package Manager
    HTTP/API Framework
    Database
    Migration
    OpenAPI
    IDE Support
```

---

## สรุป

Craft Phase 1 Rev.1 ไม่ได้มีเป้าหมายเพียงเพิ่มจำนวนคำสั่ง แต่เป็นช่วงพัฒนาคุณภาพและบุคลิกของภาษา

จุดเน้นสำคัญคือ:

```text
ภาษาอ่านง่าย
Type ปลอดภัย
จัดการ Error ได้
Runtime ควบคุมได้
โครงสร้างข้อมูลเพียงพอ
ทดสอบได้
จัดรูปแบบได้
ต่อยอดได้
```

ฟีเจอร์สำคัญที่สุดของ Rev.1 ได้แก่:

```text
let / var
Scope
Array
Struct Foundation
try / catch / throw
Stack Trace
defer
craft fmt
craft test
Error Code
Exit Code
Standard Library Foundation
```

เมื่อ Phase 1 Rev.1 เสร็จ Craft จะมีพื้นฐานที่แข็งแรงพอสำหรับการทดลองภาษาในระดับที่ลึกขึ้น และพร้อมต่อยอดไปสู่ Enum, Optional, Module, Event, Async และ Bytecode VM ในอนาคต.