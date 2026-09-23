# Craft Phase 1

## Language Foundation

เอกสารนี้กำหนดแผนพัฒนา Craft ใน Phase 1 โดยมีเป้าหมายเพื่อสร้างภาษาโปรแกรมที่สามารถเขียน ตรวจสอบ และรันโปรแกรมพื้นฐานได้จริง

Phase 1 จะโฟกัสที่แกนกลางของภาษา ได้แก่ Lexer, Parser, AST, Static Type Checker, Interpreter และ Craft CLI

ยังไม่รวมระบบ API, Database Migration, OpenAPI, Package Registry หรือ Framework ระดับสูง

---

## 1. เป้าหมายของ Phase 1

ผู้ใช้ต้องสามารถติดตั้ง Craft แล้วสร้างและรันโปรเจกต์ได้ด้วยคำสั่งดังนี้:

```bash
craft version
craft new hello
cd hello
craft run
```

ไฟล์ `src/main.craft`:

```craft
func main() {
    print("Hello World")
}
```

ผลลัพธ์:

```text
Hello World
```

---

## 2. เครื่องมือที่ใช้พัฒนา

### 2.1 ภาษาในการพัฒนา Compiler

ใช้ **Go** ในการพัฒนา:

- Craft Compiler
- Craft CLI
- Lexer
- Parser
- AST
- Type Checker
- Interpreter
- Bytecode VM ในอนาคต
- Project Generator
- Build Tool

Go จะเป็นเพียงภาษาที่ใช้สร้างเครื่องมือของ Craft ผู้ใช้ Craft ไม่จำเป็นต้องเขียน Go หรือติดตั้ง Go เพื่อใช้งานโปรแกรม Craft

### 2.2 เหตุผลที่เลือก Go

- Compile เป็น executable เดียวได้ง่าย
- Compile เร็ว เหมาะกับการพัฒนา Compiler
- รองรับ Windows, Linux และ macOS ได้ดี
- มี Standard Library เพียงพอสำหรับสร้าง CLI และ Compiler
- จัดการ File System, Process, Path และ Network ได้สะดวก
- แจกจ่าย `craft.exe` ได้ง่าย
- มีความซับซ้อนน้อยกว่า C++ และ Rust สำหรับช่วงเริ่มต้น
- เหมาะกับการสร้าง Interpreter และ Bytecode VM

### 2.3 เครื่องมือหลัก

| ส่วน | เครื่องมือ |
|---|---|
| Compiler และ CLI | Go |
| Lexer | เขียนเองด้วย Go |
| Parser | Recursive Descent + Pratt Parser |
| AST | โครงสร้างข้อมูลที่ออกแบบเอง |
| Type Checker | เขียนเองด้วย Go |
| Execution | Interpreter ใน Phase 1 |
| Testing | Go Testing |
| Source Format | `.craft` |
| Project Manifest | `craft.toml` |
| Windows Installer | Inno Setup |
| Source Control | Git |

---

## 3. Architecture

```text
Craft Source
    ↓
Lexer
    ↓
Parser
    ↓
AST
    ↓
Static Type Checker
    ↓
Interpreter
    ↓
Program Output
```

โครงสร้างนี้ต้องแยก Frontend ออกจาก Execution Layer เพื่อให้สามารถเปลี่ยนจาก Interpreter ไปเป็น Bytecode VM หรือ Native Backend ได้ในอนาคต

---

## 4. ขอบเขตภาษาที่ต้องรองรับ

### 4.1 Primitive Types

Phase 1 รองรับชนิดข้อมูลพื้นฐาน:

```text
Int
Float
Bool
String
Void
```

ตัวอย่าง:

```craft
let age: Int = 25
let price: Float = 99.5
let active: Bool = true
let name: String = "Craft"
```

### 4.2 Literals

```craft
10
3.14
true
false
"Hello World"
```

### 4.3 Variables

```craft
let name: String = "Craft"
let count: Int = 1

count += 1
```

Phase 1 ใช้ `let` เป็นรูปแบบประกาศตัวแปรหลัก โดยอนุญาตให้เปลี่ยนค่าได้ในช่วงแรกเพื่อให้ควบคุมการพัฒนาได้ง่าย

### 4.4 Operators

Arithmetic:

```text
+-*/%
```

Comparison:

```text
== != > < >= <=
```

Logical:

```text
&& || !
```

Assignment:

```text
= -= *= /=
```

### 4.5 Output

```craft
print("Hello World")
print(100)
```

ฟังก์ชัน `print` เป็น Built-in Function ของ Standard Library ใน Phase 1

---

## 5. Control Flow

### 5.1 if / else

```craft
func main() {
    let age: Int = 20

    if age >= 18 {
        print("Adult")
    } else {
        print("Minor")
    }
}
```

### 5.2 else if

```craft
if score >= 80 {
    print("A")
} else if score >= 70 {
    print("B")
} else {
    print("C")
}
```

### 5.3 while

```craft
let count: Int = 0

while count < 5 {
    print(count)
    count += 1
}
```

### 5.4 for

รูปแบบแรกที่รองรับ:

```craft
for i in 0..5 {
    print(i)
}
```

### 5.5 break และ continue

```craft
while true {
    if shouldStop {
        break
    }

    continue
}
```

---

## 6. Functions

```craft
func add(a: Int, b: Int): Int {
    return a + b
}
```

การเรียกใช้:

```craft
func main() {
    let result: Int = add(10, 20)
    print(result)
}
```

กฎของ Functions ใน Phase 1:

- ใช้ keyword `func`
- Parameter ต้องระบุชนิดข้อมูล
- Function ที่คืนค่าต้องระบุ Return Type
- ใช้ `return` เพื่อคืนค่า
- Function ที่ไม่คืนค่าจะเป็น `Void`
- ไม่รองรับ Generic Function
- ไม่รองรับ Overloading
- ไม่รองรับ Default Parameter

---

## 7. Entry Point

ทุก Executable Project ใช้ Entry Point รูปแบบเดียวกัน:

```craft
func main() {
    print("Hello World")
}
```

กฎ:

1. ต้องมี `func main()` หนึ่งฟังก์ชัน
2. `main` ต้องไม่มี Parameter ใน Phase 1
3. `main` ต้องไม่มี Return Value
4. หากไม่พบ `main` ต้องแสดง Error
5. หากพบ `main` มากกว่าหนึ่งตัวต้องแสดง Error
6. ประเภทโปรแกรมไม่ได้กำหนดจากชื่อ Entry Point

ตัวอย่าง:

```craft
func main() {
    server.start()
}
```

```craft
func main() {
    worker.run()
}
```

ความแตกต่างของโปรแกรมอยู่ที่โค้ดภายใน `main` ไม่ใช่ชื่อของ Entry Point

---

## 8. Syntax Rules

Craft ใช้ Syntax ในกลุ่ม Modern C-Family:

- ใช้ `{}` กำหนด Block
- ใช้ `func` สำหรับ Function
- ใช้ `let` สำหรับตัวแปร
- ใช้ `name: Type` สำหรับ Type Annotation
- ไม่บังคับใช้ Semicolon
- อนุญาต Semicolon แบบ Optional ในช่วงแรก
- ใช้ Comment แบบ `//`

ตัวอย่าง:

```craft
// This is a comment
func main() {
    print("Hello World")
}
```

---

## 9. Compiler Components

### 9.1 Lexer

รับ Source Code แล้วแปลงเป็น Tokens:

```text
Identifier
Keyword
Integer
Float
String
Operator
Delimiter
EOF
```

### 9.2 Parser

แปลง Tokens เป็น Abstract Syntax Tree โดยใช้:

- Recursive Descent Parser สำหรับ Statements
- Pratt Parser สำหรับ Expressions

### 9.3 AST

AST ขั้นต้นต้องรองรับ:

```text
Program
FunctionDeclaration
VariableDeclaration
BlockStatement
IfStatement
WhileStatement
ForStatement
ReturnStatement
BreakStatement
ContinueStatement
ExpressionStatement
BinaryExpression
UnaryExpression
CallExpression
LiteralExpression
IdentifierExpression
AssignmentExpression
```

### 9.4 Type Checker

ตรวจสอบ:

- การกำหนดค่าผิดชนิด
- การใช้ Operator ผิดชนิด
- การเรียก Function ด้วยจำนวน Argument ผิด
- การเรียก Function ด้วยชนิด Argument ผิด
- การคืนค่าผิดชนิด
- การใช้ Identifier ที่ไม่มีอยู่
- การประกาศ Function ซ้ำ
- การมี `main` ไม่ถูกต้อง

ตัวอย่าง Error:

```text
Type error: cannot assign String to Int
```

### 9.5 Interpreter

Interpreter จะรัน AST โดยตรงใน Phase 1 เพื่อให้พัฒนาภาษาได้เร็วและ Debug ได้ง่าย

ลำดับการพัฒนา Interpreter:

1. Literals
2. Expressions
3. Variables
4. Assignment
5. Built-in Functions
6. User Functions
7. if / else
8. while
9. for
10. return
11. break / continue

---

## 10. Craft CLI

คำสั่งที่ต้องมีใน Phase 1:

```bash
craft version
craft new my-app
craft init
craft run
craft check
craft build
craft clean
```

### `craft version`

```text
Craft 0.1.0
Target: windows-x64
Compiler: Go
Execution: Interpreter
```

### `craft new`

สร้างโครงสร้าง:

```text
my-app/
├── craft.toml
├── src/
│   └── main.craft
└── tests/
```

### `craft run`

ทำงานตามลำดับ:

```text
ค้นหา craft.toml
ค้นหาไฟล์ .craft
Lexer
Parser
Type Checker
ค้นหา func main()
Interpreter
```

### `craft check`

ตรวจสอบ Source Code โดยไม่รันโปรแกรม:

```bash
craft check
```

### `craft build`

ในช่วงแรกสามารถสร้าง Build Artifact ระดับทดลองได้ หลังจาก Interpreter เสถียรแล้วจึงพัฒนา Bytecode VM และ Standalone Executable ต่อ

---

## 11. Project Manifest

ไฟล์ `craft.toml` ขั้นต้น:

```toml
name = "hello"
version = "0.1.0"
edition = "2026"
entry = "main"
```

ใน Phase 1 ยังไม่ต้องมีระบบ Dependency ที่ซับซ้อน

---

## 12. โครงสร้าง Repository

```text
craft/
├── cmd/
│   └── craft/
│       └── main.go
├── internal/
│   ├── lexer/
│   ├── parser/
│   ├── ast/
│   ├── types/
│   ├── interpreter/
│   ├── diagnostics/
│   └── project/
├── templates/
│   └── project/
├── tests/
├── go.mod
└── README.md
```

---

## 13. ลำดับการพัฒนา

### Milestone 1: CLI Skeleton

- สร้าง Go Module
- ทำ `craft version`
- ทำ `craft new`
- สร้าง Project Template

### Milestone 2: Lexer

- อ่านไฟล์ `.craft`
- แยก Token
- แสดงตำแหน่งบรรทัดและคอลัมน์
- เพิ่ม Lexer Tests

### Milestone 3: Parser และ AST

- Parse Function
- Parse Block
- Parse Variable
- Parse Expression
- Parse Function Call
- เพิ่ม Parser Tests

### Milestone 4: Interpreter

- รัน `print("Hello World")`
- รองรับ Variables
- รองรับ Operators
- รองรับ Functions

### Milestone 5: Control Flow

- `if`
- `else`
- `while`
- `for`
- `break`
- `continue`

### Milestone 6: Static Type Checker

- ตรวจ Primitive Types
- ตรวจ Expression
- ตรวจ Function Call
- ตรวจ Return
- ตรวจ Entry Point

### Milestone 7: Diagnostics

- Error Message
- File Path
- Line
- Column
- Source Code Context
- Caret ชี้ตำแหน่ง Error

### Milestone 8: Build และ Packaging

- `craft build`
- Build สำหรับ Windows
- สร้าง `craft.exe`
- สร้าง `Craft-setup.exe`
- ทดสอบบนเครื่องที่ไม่มี Go

---

## 14. Testing Strategy

ต้องทดสอบทั้งระดับ Component และ End-to-End

### Lexer Tests

```craft
let name: String = "Craft"
```

ต้องแยก Token ได้ถูกต้อง

### Parser Tests

```craft
func main() {
    print("Hello")
}
```

ต้องสร้าง AST ที่ถูกต้อง

### Interpreter Tests

```craft
func main() {
    let result: Int = 10 + 20
    print(result)
}
```

ผลลัพธ์ต้องเป็น:

```text
30
```

### Error Tests

```craft
let value: Int = "wrong"
```

ต้องถูกปฏิเสธโดย Type Checker

---

## 15. Definition of Done

Craft Phase 1 ถือว่าสำเร็จเมื่อ:

- สามารถติดตั้ง Craft บน Windows ได้
- ใช้ `craft version` ได้
- ใช้ `craft new` สร้างโปรเจกต์ได้
- ใช้ `craft run` รันโปรแกรมได้
- ใช้ `craft check` ตรวจ Source Code ได้
- รองรับ `func main()`
- รองรับ `print`
- รองรับตัวแปร
- รองรับ `Int`, `Float`, `Bool`, `String`
- รองรับ Operators พื้นฐาน
- รองรับ `if / else`
- รองรับ `while`
- รองรับ `for`
- รองรับ `break / continue`
- รองรับ Functions
- มี Static Type Checking
- มี Error Message พร้อมตำแหน่งบรรทัด
- มี Unit Tests และ End-to-End Tests
- สร้าง `craft.exe` ได้ด้วย Go
- ผู้ใช้ปลายทางไม่จำเป็นต้องติดตั้ง Go

---

## 16. สิ่งที่ยังไม่ทำใน Phase 1

สิ่งต่อไปนี้จะยังไม่รวมอยู่ใน Phase 1:

- Generics
- Struct แบบเต็มรูปแบบ
- Enum
- Protocol หรือ Interface ขั้นสูง
- Package Manager เต็มรูปแบบ
- Async/Await
- Concurrency
- HTTP Server
- Database
- Migration
- OpenAPI
- Dependency Injection
- Reflection
- Garbage Collector แบบกำหนดเอง
- LLVM Backend
- Optimizer ขั้นสูง

---

## สรุป

Craft Phase 1 ใช้ Go เป็นภาษาสำหรับพัฒนา Compiler และ CLI โดยเริ่มจาก Interpreter เพื่อให้ภาษาใช้งานได้เร็วและทดสอบได้ง่าย

เป้าหมายหลักคือ:

```text
Craft Source
→ Lexer
→ Parser
→ AST
→ Type Checker
→ Interpreter
→ craft run
```

เมื่อแกนกลางเสถียรแล้ว จึงต่อยอดไปสู่ Bytecode VM, Standalone Executable, API Framework, Migration และ OpenAPI ใน Phase ถัดไป
