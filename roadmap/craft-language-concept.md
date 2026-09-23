# Craft — Modern Statically Typed C-Family Language

## วิสัยทัศน์

Craft คือภาษาโปรแกรมแบบ Static Typing ที่มี Syntax อยู่ในตระกูล C-family โดยได้รับแรงบันดาลใจจาก C#, Swift, Go และ Rust

เป้าหมายของ Craft คือให้ผู้พัฒนาสามารถเขียนโปรแกรมและสร้างเครื่องมือของตนเองด้วยโค้ดที่อ่านง่าย ชัดเจน และมีโครงสร้างระดับระบบ โดยไม่ต้องติดตั้งเครื่องมือจำนวนมากหรือจัดการรายละเอียดที่ไม่จำเป็นด้วยตัวเอง ดู [ทิศทางการพัฒนา Craft](craft-language-direction.md)

> Code should describe the system, not the ceremony required to implement it.

## ผลลัพธ์ที่ต้องการ

ผู้ใช้ควรสามารถทำงานตามลำดับนี้ได้:

```text
ดาวน์โหลด Craft-setup.exe
        ↓
ติดตั้ง Craft
        ↓
เปิด Terminal ใหม่
        ↓
craft version
craft new my-app
cd my-app
craft run
```

หลังติดตั้งแล้ว เครื่องเป้าหมายต้องสามารถใช้ Craft CLI ได้ทันที โดยไม่จำเป็นต้องติดตั้งสิ่งเหล่านี้เพิ่ม:

- Go
- .NET
- Java
- Runtime ของ Craft
- Compiler อื่น
- Package manager อื่น

กล่าวคือ `Craft-setup.exe` ต้องติดตั้งทุกสิ่งที่จำเป็นสำหรับการพัฒนาและรันโปรแกรม Craft ให้ครบในครั้งเดียว

## นามสกุลไฟล์

```text
.craft       Source code ของ Craft
craft.toml   Project manifest
craft.lock   Dependency lock file
```

ตัวอย่างไฟล์:

```text
main.craft
user.craft
user_service.craft
001_create_users.craft
```

## ประสบการณ์เริ่มต้นใช้งาน

### ตรวจสอบการติดตั้ง

```bash
craft version
```

ผลลัพธ์ตัวอย่าง:

```text
Craft 0.1.0
Target: windows-x64
Compiler: embedded
Runtime: embedded
```

### สร้างโปรเจกต์ใหม่

```bash
craft new my-app
cd my-app
```

คำสั่งนี้สร้างโครงสร้างโปรเจกต์เริ่มต้น:

```text
my-app/
├── craft.toml
├── src/
│   └── main.craft
└── tests/
```

### โปรแกรมแรก

ไฟล์ `src/main.craft`:

```craft
func main() {
    print("hello world")
}
```

จากนั้นรันด้วย:

```bash
craft run
```

ผลลัพธ์:

```text
hello world
```

## Entry Point

ทุกโปรเจกต์ต้องมีจุดเริ่มต้นที่ชัดเจน เพื่อให้ `craft run` รู้ว่าจะเริ่มทำงานจากที่ใด แต่ภาษาไม่ควรบังคับว่าโปรเจกต์นั้นต้องเป็น `app`, API หรือโปรแกรมประเภทใดประเภทหนึ่ง

แนวทางเริ่มต้นที่แนะนำคือใช้ฟังก์ชันมาตรฐานชื่อ `main`:

```craft
func main() {
    print("hello world")
}
```

กติกาเบื้องต้น:

1. ในหนึ่ง executable project ต้องมี `main` หนึ่งจุด
2. `main` เป็นฟังก์ชันธรรมดาที่ทำหน้าที่เป็น entry point
3. `craft run` จะค้นหา `func main()` ให้อัตโนมัติ
4. ถ้าไม่พบ entry point หรือมีมากกว่าหนึ่งตัว Compiler ต้องแสดง error ที่เข้าใจง่าย

ตัวอย่างการกำหนด entry point แบบ explicit ใน `craft.toml` เมื่อโปรเจกต์มีหลายส่วน:

```toml
name = "my-service"
version = "0.1.0"
edition = "2026"
entry = "main"
```

ความแตกต่างของโปรแกรมแต่ละประเภทจะอยู่ที่สิ่งที่เรียกใช้ภายใน `main` ไม่ใช่ชื่อของ entry point:

```craft
func main() {
    server.start()
}

func main() {
    worker.run()
}
```

ตัวอย่างโปรแกรมรูปแบบต่าง ๆ:

CLI:

```craft
func main() {
    print("hello world")
}
```

API Server:

```craft
func main() {
    server.start()
}
```

Application:

```craft
func main() {
    app.start()
}
```

Background Worker:

```craft
func main() {
    worker.run()
}
```

หลักการนี้ทำให้ทุกโปรเจกต์มีรูปแบบการเริ่มต้นที่เหมือนกันและคาดเดาได้ คือ `func main()` เสมอ ส่วนโปรเจกต์ที่เป็น API, Worker, CLI หรือ Service สามารถมีพฤติกรรมแตกต่างกันได้จาก implementation ภายใน `main` โดยไม่ต้องประกาศประเภทโปรเจกต์ด้วยคำพิเศษของภาษา

แต่สำหรับภาษาเวอร์ชันแรก `app { start {} }` เหมาะกว่า เพราะสื่อถึงโครงสร้างของ Application ได้ชัดเจน และต่อยอดไปสู่ configuration, dependency injection และ lifecycle ได้ง่าย

## คำสั่งหลักของ Craft CLI

```bash
craft version              # แสดงเวอร์ชัน
craft upgrade              # อัปเกรดตัวติดตั้งหรือเครื่องมือ
craft new my-app           # สร้างโปรเจกต์ใหม่
craft init                 # สร้าง Craft project ในโฟลเดอร์ปัจจุบัน
craft run                  # Compile และรันโปรแกรม
craft build                # Compile เป็น executable
craft test                 # รันชุดทดสอบ
craft check                # ตรวจ syntax และ type โดยไม่รัน
craft fmt                  # จัดรูปแบบโค้ด
craft add postgres         # เพิ่ม dependency
craft remove postgres      # ลบ dependency
craft migrate              # จัดการ database migration
craft generate             # สร้างโค้ดจาก declaration
craft clean                # ลบ build artifacts
```

## การติดตั้งบน Windows

ไฟล์ติดตั้งหลักคือ:

```text
Craft-setup.exe
```

Installer ต้องทำหน้าที่ดังนี้:

1. ติดตั้ง `craft.exe`
2. เพิ่มโฟลเดอร์ของ Craft ลงใน `PATH`
3. ติดตั้ง standard library และ compiler assets
4. ลงทะเบียน file association สำหรับ `.craft`
5. สร้างคำสั่งใน Start Menu หรือ shortcut ตามความเหมาะสม
6. ตรวจสอบ installation หลังติดตั้งเสร็จ

เป้าหมายคือผู้ใช้เปิด Terminal ใหม่แล้วใช้คำสั่งนี้ได้ทันที:

```bash
craft version
```

ตัวโปรแกรมที่ผู้ใช้พัฒนาด้วย Craft ควรถูก build เป็น executable ที่นำไปเปิดบนเครื่องเป้าหมายได้โดยตรง:

```bash
craft build --release
```

ผลลัพธ์ตัวอย่าง:

```text
dist/my-app.exe
```

ในโหมด release executable ต้องพึ่งพาเฉพาะ system libraries ที่มีอยู่ตามปกติของระบบปฏิบัติการ และไม่ควรต้องติดตั้ง Craft เพิ่มเพื่อรันโปรแกรมที่ build เสร็จแล้ว

## Compiler และการทำงานภายใน

สถาปัตยกรรมระยะแรก:

```text
.craft source code
        ↓
Lexer
        ↓
Parser
        ↓
AST
        ↓
Name Resolver
        ↓
Type Checker
        ↓
Code Generator
        ↓
Embedded Backend
        ↓
Native Executable
```

Compiler และ CLI ควรพัฒนาด้วย Go เพื่อให้สามารถสร้าง binary แบบ static และแจกจ่ายเป็นไฟล์เดียวได้ง่าย โดยผู้ใช้ปลายทางไม่จำเป็นต้องติดตั้ง Go

สิ่งสำคัญคือ Go เป็นเพียงเครื่องมือที่ใช้สร้าง Craft ไม่ใช่สิ่งที่ผู้ใช้ Craft ต้องรู้หรือติดตั้ง

## กลยุทธ์ Backend ระยะแรก

เพื่อให้พัฒนาได้เร็ว Craft สามารถใช้แนวทางสองระยะ:

### ระยะที่หนึ่ง: Compile ผ่าน Backend ที่มีอยู่

```text
Craft → Generated Code → Backend Compiler → Executable
```

Backend ภายในอาจใช้ Go หรือระบบอื่นเป็นขั้นตอนเบื้องหลัง แต่ต้องถูกฝังหรือจัดการโดย `craft.exe` ผู้ใช้จึงเห็นเพียงคำสั่ง Craft เท่านั้น

### ระยะที่สอง: Native Backend

เมื่อภาษาและ Type System เสถียรแล้ว จึงพัฒนา Native Backend ของ Craft เอง:

```text
Craft → Intermediate Representation → Native Code → Executable
```

การออกแบบ CLI และ project structure ควรซ่อนรายละเอียดของ Backend ตั้งแต่ต้น เพื่อให้สามารถเปลี่ยน Backend ในอนาคตโดยไม่กระทบผู้ใช้

## โครงสร้าง Syntax พื้นฐาน

### Struct

```craft
struct User {
    let id: Int
    let name: String
    let email: String
}
```

### Service

```craft
service UserService {
    inject UserRepository repository
    inject Logger logger
}
```

### Migration

```craft
migration CreateUsers {
    up {
        table users {
            id: Int primaryKey autoIncrement
            name: String required
            email: String unique required
        }
    }

    down {
        drop users
    }
}
```

### API Controller

```craft
controller UserController {
    inject UserService service

    get "/users/{id}" -> User? {
        return await service.find(id)
    }
}
```

## โครงสร้างโปรเจกต์ API

```text
my-api/
├── craft.toml
├── src/
│   ├── main.craft
│   ├── users.craft
│   ├── user_service.craft
│   └── user_controller.craft
├── migrations/
│   └── 001_create_users.craft
├── tests/
│   └── user_test.craft
└── dist/
```

## หลักการสำคัญ

- `.craft` ต้องเป็น source file หลักของภาษา
- `craft.exe` ต้องเป็น CLI หลักสำหรับทุก workflow
- `Craft-setup.exe` ต้องติดตั้งเครื่องมือที่จำเป็นให้ครบ
- `craft run` ต้องทำงานได้ทันทีหลังสร้างโปรเจกต์
- ผู้ใช้ไม่ควรต้องรู้ว่า Compiler ภายในใช้เทคโนโลยีอะไร
- โปรแกรมที่ build แล้วต้องนำไปรันบนเครื่องเป้าหมายได้โดยไม่ต้องติดตั้ง Craft เพิ่ม
- Entry point ต้องชัดเจนและตรวจสอบได้โดย Compiler
- Error message ต้องบอกไฟล์ บรรทัด ตำแหน่ง และวิธีแก้ไข

## สรุปคอนเซปต์

Craft ไม่ควรเป็นเพียงภาษา Syntax ใหม่ แต่ควรเป็นชุดเครื่องมือที่ให้ประสบการณ์ครบตั้งแต่ติดตั้ง สร้างโปรเจกต์ เขียนโค้ด ตรวจสอบ ทดสอบ Compile และนำไปใช้งานจริง

ประสบการณ์ที่ต้องการคือ:

```text
ติดตั้งครั้งเดียว
        ↓
craft new my-app
        ↓
เขียนไฟล์ .craft
        ↓
craft run
        ↓
ได้โปรแกรมที่ทำงานจริง
```

> Craft คือภาษาที่ทำให้การสร้าง Backend เริ่มต้นได้ง่ายเหมือนการติดตั้งโปรแกรมหนึ่งตัว แต่ยังคงความชัดเจนและความสามารถของภาษาแบบ Static Typing เอาไว้
