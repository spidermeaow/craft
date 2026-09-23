# ย้ายโปรเจกต์จาก Craft 0.1.0 ไป 0.1.1

## 1. อัปเดตเครื่องมือ

ติดตั้ง `dist/Craft-setup.exe` รุ่นใหม่ หรือแทน portable `craft.exe` แล้วเปิด Terminal ใหม่ ตรวจด้วย `craft version` ต้องเป็น `0.1.1` หากยังเป็นรุ่นเก่า ใช้ `Get-Command craft -All` ดูว่ามี binary หลายตำแหน่งใน PATH หรือไม่

## 2. เปลี่ยนเฉพาะตัวแปรที่ต้องแก้ค่าจาก let เป็น var

เดิม:

```craft
let count: Int = 0
count += 1
```

ใหม่:

```craft
var count: Int = 0
count += 1
```

ตัวแปรที่ไม่เปลี่ยนค่าให้ใช้ `let` ต่อไป Compiler จะไม่เปลี่ยนให้เอง การแก้ `let` จะได้ `E2003` ไม่ว่าจะใช้ `=`, `+=` หรือแก้สมาชิก/เรียก mutating methods ของ Array

สำหรับโปรแกรม Bank: `transactionNumber` ใน while และ `suspiciousTransaction` ต้องเป็น `var`; owner, accountNumber และยอดแต่ละขั้นที่กำหนดครั้งเดียวใช้ `let` ได้

Parameters และตัวแปรใน `for` เปลี่ยนค่าไม่ได้ หากต้องคำนวณต่อให้สร้าง `var` ภายใน function/body อย่าประกาศชื่อซ้ำกับ parameter ใน function scope เดียวกัน ใช้ชื่อใหม่หรือ nested block

## 3. Array ใช้ value semantics

`var copy: Int[] = original` จะได้สำเนา รวมถึงสมาชิกที่เป็น Array ซ้อนกัน การแก้ copy จึงไม่เปลี่ยน original เช่นเดียวกับ arguments และ return values ฟังก์ชันที่ต้องแก้ Array ให้รับ parameter, สร้าง local var, แก้สำเนาแล้วคืนค่า

```craft
func withExtra(values: Int[], value: Int): Int[] {
    var copy: Int[] = values
    copy.append(value)
    return copy
}
```

## 4. รับมือ error ได้แล้ว

Runtime errors เดิม เช่นหารด้วยศูนย์ จับได้ด้วย `try/catch` ส่วน syntax/type errors ต้องแก้ก่อนรัน Exception ที่ไม่ถูกจับจะแสดง source และ stack trace

```craft
try {
    print(1 / 0)
} catch error {
    print(error.message)
}
```

Exit code เดิมที่ใช้ 1 สำหรับทุกข้อผิดพลาดเปลี่ยนเป็น 1 runtime, 2 syntax, 3 type, 4 project/entry point, 5 CLI usage จึงต้องปรับ scripts ที่ตรวจ `$LASTEXITCODE` แบบเจาะจง ใช้ `-ne 0` หากต้องการตรวจความล้มเหลวทุกประเภท

## 5. ตรวจและสร้าง bundle ใหม่

```powershell
craft check
craft fmt
craft test
craft run
craft build
```

Bundle รุ่น 0.1.0 ไม่สามารถรันตรงด้วย 0.1.1 ได้ ต้องย้าย source ตามกฎใหม่แล้ว rebuild เพื่อไม่ให้ความหมายของ `let` เปลี่ยนเงียบ ๆ `craft build` เขียนทับ legacy bundle ที่รู้จักได้ และ `craft clean` ลบ legacy bundle ของโปรเจกต์ได้

`craft.toml` ของโปรเจกต์ยังใช้รูปแบบเดิมได้ ค่า `version` ใน manifest เป็นเวอร์ชันแอป ไม่จำเป็นต้องเปลี่ยนตาม compiler

`craft check` ตรวจไฟล์ใน `tests/` เพิ่มด้วย ส่วน `craft run`/`build` ใช้เฉพาะ `src/` Test blocks ใน `src/` ถูกตรวจชนิดข้อมูลแต่จะรันเมื่อใช้ `craft test` เท่านั้น

## ฟีเจอร์ที่ยังไม่รองรับ

ไม่ใช้ `module`, `import`, `struct`, `enum`, Map, Optional, lambda, default parameters, Event Bus หรือ `finally` ใน source รุ่นนี้ ดูข้อกำหนด foundation ใน `REV1-FOUNDATIONS.md`
