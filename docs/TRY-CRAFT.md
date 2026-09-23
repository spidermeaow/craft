> คู่มือนี้เก็บขั้นตอน Rev.1 เดิมไว้ สำหรับ CLI 0.1.2 ให้เริ่มที่ [ทดลอง Rev.2](TRY-REV2.md)

# ทดลอง Craft 0.1.1 — Phase 1 Rev.1

## 1. ติดตั้ง/อัปเกรด

เปิด `dist/Craft-setup.exe` รุ่นใหม่ ติดตั้งแล้วเปิด Terminal ใหม่ หากใช้ terminal ใน editor อาจต้องเปิด editor ใหม่เพื่อรับ PATH

```powershell
craft version
```

ต้องเป็น `Craft 0.1.1` หากยังเป็นรุ่นเดิม ให้ตรวจ `Get-Command craft -All` ใช้ portable `dist/craft.exe` ได้โดยไม่ต้องติดตั้ง ไม่ต้องมี Go บนเครื่องผู้ใช้

## 2. โปรเจกต์แรก

```powershell
craft new hello-rev1
cd hello-rev1
craft check
craft run
craft test
craft fmt
craft fmt --check
```

คาดหวัง `Check succeeded.`, `Hello World`, `1 passed, 0 failed, 1 total` และหลัง fmt แล้ว fmt --check ต้องได้ exit 0 โครงสร้างใหม่มี `tests/smoke.craft` เป็นตัวอย่าง test

## 3. let และ var

ใส่ใน `src/main.craft`:

```craft
func main() {
    var number: Int = 1
    while number <= 10 {
        print(number)
        number += 1
    }
}
```

`craft run` ควรพิมพ์ 1–10 คนละบรรทัด เปลี่ยน `var` เป็น `let` แล้ว `craft check` ต้องได้ `E2003` และ `$LASTEXITCODE` เป็น 3

## 4. ตัวอย่าง Rev.1 แบบรวม

เข้า `examples/rev1` ใน repository หรือใน installation directory แล้วรัน:

```powershell
craft check
craft run
craft test
```

ผล main ที่คาดหวัง:

```text
removed: 1
numbers: [2, 3, 4]
sum: 9
name: CRAFT
words: [safe, clear, small]
named arguments: 4
sqrt: 3
caught: cannot divide by zero
bounds: RuntimeError E3001
cleanup complete
```

ผล tests ที่คาดหวัง: **6 passed, 0 failed, 6 total** รวม array copy, catch, string operations และ named arguments

## 5. ตรวจว่า let array ไม่ถูกแก้ผ่านสำเนา

```craft
func main() {
    let original: Int[][] = [[1, 2]]
    var copy: Int[][] = original
    copy[0][0] = 99
    copy[0].append(3)
    print(original)
    print(copy)
}
```

คาดหวัง `[[1, 2]]` และ `[[99, 2, 3]]` ตามลำดับ หากเปลี่ยนเป็น `original[0].append(3)` ต้องถูกปฏิเสธด้วย E2003

## 6. Exception และ defer

```craft
func work() {
    defer { print("cleanup 1") }
    defer { print("cleanup 2") }
    throw Exception("work failed")
}

func main() {
    try {
        work()
    } catch error {
        print(error.message)
    }
    print("still running")
}
```

คาดหวัง `cleanup 2`, `cleanup 1`, `work failed`, `still running` ตามลำดับ ลอง `examples/unhandled-exception` เพื่อดู stack trace ข้ามไฟล์: ต้องแสดง cleanup ก่อนจบ, E3002, source location และ frames fail → process → main พร้อม exit 1

## 7. Error codes และ exit codes

ดู `$LASTEXITCODE` ทันทีหลังคำสั่ง เพื่อไม่ให้คำสั่งอื่นเปลี่ยนค่า:

| กรณี | ผลที่คาดหวัง |
| --- | --- |
| `craft unknown` | E5001, exit 5 |
| โปรเจกต์ไม่มี craft.toml | E4002, exit 4 |
| ไม่มี main / signature main ผิด | E4001, exit 4 |
| ปิดวงเล็บไม่ครบ | E1001, exit 2 |
| `let x: Int = "bad"` | E2002, exit 3 |
| แก้ค่าของ let | E2003, exit 3 |
| `print(1 / 0)` ไม่ catch | E3001, exit 1; check อย่างเดียวยังสำเร็จ |
| throw ไม่ catch | E3002, exit 1 |
| assert false ใน test | E3003, exit 1; tests ถัดไปยังทำงาน |
| fmt --check พบโค้ดต้องจัดรูปแบบ | exit 1 โดยไม่เขียนไฟล์ |

## 8. fmt และ test

`craft fmt` ต้องรักษาข้อความ comments และพฤติกรรมโค้ด ใช้ได้แม้โค้ดมี type error แต่ syntax ต้องถูกต้อง รันซ้ำควรไม่มีการเปลี่ยนแปลงเพิ่มเติม

สร้าง `tests/example.craft`:

```craft
test "addition" {
    assert 2 + 3 == 5
}

test "deliberate failure" {
    assert false
}
```

`craft test` ต้องแสดง PASS/FAIL พร้อม source location และสรุปจำนวน โดยไม่เรียก main แก้ `assert false` เป็น `assert true` แล้วรันใหม่ควรไม่มี test ที่ล้มเหลว

## 9. Build, bundle และ clean

จากโปรเจกต์ hello-rev1 ที่แก้ source ถูกต้องแล้ว:

```powershell
craft build
craft run dist/hello-rev1.craftbundle
craft build --release
craft clean
```

Bundle ต้องรันได้แม้ย้ายออกจาก source directory และ clean ต้องไม่ลบไฟล์อื่นใน dist Bundle 0.1.0 ต้องถูกปฏิเสธพร้อมคำแนะนำย้าย source แล้ว rebuild; build/clean ต้องแทนที่/ลบ legacy bundle ของโปรเจกต์ได้

ยังไม่มี standalone executable ของโปรแกรมผู้ใช้: bundle ต้องใช้ Craft รัน

## 10. Standard library เพิ่มเติม

```craft
func main() {
    print(std.math.sqrt(16.0))
    print(std.json.isValid("{\"ok\":true}"))
    print(std.time.now())
    try {
        print(std.fs.readText("missing-file.txt"))
    } catch error {
        print(error.type)
    }
}
```

คาดหวัง 4, true, เวลา UTC และ RuntimeError เมื่อไม่มีไฟล์ดังกล่าว Paths เป็น relative ต่อ working directory ของ terminal

## 11. ตรวจรับระดับ repository และ installer

จาก repository ที่มี Go:

```powershell
go test ./...
```

ชุดทดสอบรวมของเดิมและ Rev.1: language safety, control flow, named arguments, deep copies, exceptions/rethrow, trace, defer, formatter, project/bundle และ CLI E2E

ตรวจติดตั้งซ้ำ/อัปเกรดแล้ว PATH ไม่ซ้ำ ถอนการติดตั้งแล้วไม่ลบโปรเจกต์ผู้ใช้ และทดลองบน Windows x64 ที่ไม่มี Go ทั้ง installer และ portable CLI

**Agent ไม่ได้รัน tests หรือเปิดทดลอง CLI/installer ตามความต้องการให้ผู้ใช้ทดสอบเอง** การ build เป็นการจัดทำไฟล์ส่งมอบ ไม่ใช่ผล runtime acceptance เมื่อพบปัญหา ส่ง `craft version`, source, คำสั่ง, output เต็ม และ `$LASTEXITCODE` กลับมา
