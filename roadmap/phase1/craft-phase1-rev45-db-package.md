# Craft Phase 1 Rev.45 (0.1.45) — `packages/db`

**สถานะ:** Implemented candidate; รอผู้ใช้ตรวจรับ  
**ฐาน:** `std.db` ของ Craft 0.1.44 candidate  
**ขอบเขต:** first-party Craft package สำหรับช่วยอ่านผลลัพธ์ SQL และตรวจผลของคำสั่งเขียน; ไม่เปลี่ยน DB Engine หรือสร้าง ORM

## เป้าหมาย

ให้แอป Craft ที่ใช้ MySQL, PostgreSQL หรือ SQL Server เขียนโค้ดซ้ำเพื่ออ่าน `DbRows`, ตรวจ NULL/ชนิดข้อมูล และตรวจจำนวนแถวที่ถูกแก้น้อยลง. `packages/db` เขียนด้วย `.craft` บน Public API ของ `std.db`. SQL, schema, DSN และ repository ของแต่ละแอปยังเป็นความรับผิดชอบของแอป.

Package มากับ Craft และใช้เวอร์ชันตรงกับ toolchain เช่น `packages/gui` และ `packages/api`. โปรเจกต์เพิ่ม path dependency, รัน `craft package lock` และ `import package db "db"`; ไม่ต้องเลือกเวอร์ชัน package แยก.

## เงื่อนไขสำคัญของ DB Engine ปัจจุบัน

`DbConnection` เป็นของ function ที่เรียก `std.db.connect`; `DbTransaction` เป็นของ function ที่เรียก `std.db.begin`. Handle ที่คืนออกจาก helper ซึ่งเป็นผู้สร้างจะถูกปิด/rollback เมื่อ helper จบ. ดังนั้น 0.1.45 **ไม่เปิด API `db.open(): DbConnection` หรือ `db.begin(): DbTransaction`** ที่เพียงครอบฟังก์ชันเดิม. แอปต้องสร้างและปิด handle ใน scope เจ้าของ แล้วส่ง handle หรือผล `DbRows` ให้ helper ของ package. หากจะเพิ่ม wrapper แบบสร้างและคืน handle ภายหลัง ต้องออกแบบ ownership ใหม่ใน Runtime พร้อม tests ก่อน.

```craft
import package db "db"

func main() {
    let connection: DbConnection = std.db.connect(
        "postgres",
        std.env.get("CRAFT_DB_DSN"),
        std.db.config(),
    )
    defer { std.db.close(connection) }

    let rows: DbRows = std.db.query(
        connection,
        "SELECT id, name FROM items WHERE id = $1",
        [std.db.int(1)],
    )
    if let row = db.one(rows) {
        print(db.requiredText(rows, row, "name"))
    }
}
```

ตัวอย่างด้านบนเป็น API ที่เสนอ; ต้องยืนยันชื่อและชนิดด้วย `craft check` เมื่อ implement.

## Public API รอบแรกที่เสนอ

| กลุ่ม | API / พฤติกรรม | กฎสำคัญ |
| --- | --- | --- |
| จำนวนแถว | `one(rows): DbValue[]?`, `requireOne(rows): DbValue[]` | 0 แถวคืน null หรือ error ตามชื่อ; มากกว่า 1 แถวต้อง error ไม่เลือกแถวแรกเงียบ ๆ. ทำงานบนผลที่ materialize แล้ว; แอปต้องจำกัดผลด้วย SQL เอง |
| คอลัมน์ | `columnIndex(rows, name): Int` | ชื่อคอลัมน์ตรงตามที่ driver คืนมา; ไม่พบหรือชื่อซ้ำต้อง error เพื่อป้องกันอ่านผิดช่อง |
| อ่านค่าตามชื่อ | `requiredText`, `requiredInt`, `requiredDecimal`, `requiredBytes`, `requiredTime`, และ optional variants ที่จำเป็น | ใช้ `std.db.as*` และ `isNull`; แยก NULL, ชนิดไม่ตรง, คอลัมน์หาย และ index เกินขอบเขต; ไม่แปลงชนิดแบบเงียบ |
| ผลการเขียน | `expectAffected(result, expected): Void` | ถ้า `affectedRows` เป็น null หรือไม่เท่าค่าที่คาด ต้อง error; ไม่ตีความ `lastInsertId` ว่าใช้ได้ทุก driver |
| คู่มือรูปแบบใช้งาน | ตัวอย่าง query/execute/transaction ที่รับ `DbConnection` หรือ `DbTransaction` จากแอป | ไม่สร้าง SQL จากชื่อ table/column ที่มาจากผู้ใช้; ส่ง values เป็น `DbValue[]` ผ่าน `std.db` |

ใช้ exception ของ package ที่แยกกรณี เช่น ไม่มีแถว, หลายแถว, คอลัมน์หาย/ซ้ำ, NULL และจำนวน affected rows ไม่ตรง. ข้อความ error ต้องไม่ใส่ DSN, SQL parameters หรือค่าข้อมูล. การออกแบบ error type สุดท้ายต้องสอดคล้องกับรูปแบบ `Exception` ที่ Craft รองรับจริง.

## ลำดับพัฒนาและเกณฑ์ทดสอบ

1. สร้าง `packages/db` พร้อม manifest ที่ใช้เวอร์ชัน Craft อัตโนมัติ; เพิ่ม regression test ของการ lock และการปฏิเสธ version field ที่ระบุเอง.
2. ลง helper สำหรับ row cardinality, column lookup และ typed extraction. ทดสอบผล 0/1/หลายแถว, ชื่อคอลัมน์ซ้ำ, NULL, type mismatch, Unicode และ binary.
3. ลง `expectAffected` โดยทดสอบ metadata ที่เป็น null รวมถึงค่า 0, 1 และมากกว่า 1.
4. ทำตัวอย่างที่ใช้ `std.db` โดยตรงสร้าง connection/transaction แล้วเรียก helper ของ package. มี tests ของ package ที่สร้าง `DbRows`/`DbResult` ผ่าน Public API หรือ fixture ที่ตรวจชนิดได้; integration กับ MySQL/PostgreSQL/SQL Server ต้องแยกตาม server ที่พร้อมใช้งาน และบันทึกชนิด transport ที่ใช้จริง.
5. ตรวจ `craft check`, `craft test`, `craft fmt --check`, `craft build`, `go test ./...` และตัวอย่างร่วมกับ HTTP request context/cancellation. ยืนยันว่า helper ไม่ยืดอายุ connection/transaction และไม่กลบ `DbTimeoutError`/`DbCancelledError`.
6. เพิ่ม package ลงตัวติดตั้งและเอกสารหลัง source ผ่าน tests; ให้ผู้ใช้ทดลองใช้งานก่อนรับ Rev.45.

## ความสัมพันธ์กับ `packages/api`

`packages/db` ไม่ import `packages/api` และ `packages/api` ไม่ import `packages/db`. Handler ของ API สามารถเรียก helper DB เมื่อได้รับ `DbConnection` ที่ยังมีชีวิต แต่ `api.App` ปัจจุบันยังไม่มีช่องส่ง shared connection เข้า handler. 0.1.45 ต้อง **ทดสอบและบันทึก** รูปแบบร่วมกันที่ไม่ละเมิด ownership; ถ้าต้องแก้ state injection ของ `packages/api` ให้แยกสัญญาและ tests ให้ชัดก่อนเพิ่ม API นั้น. ห้ามอ้างว่า API+DB แบบใช้ pool ร่วมพร้อมใช้เพียงเพราะ package ทั้งสองติดตั้งได้.

## Migration และ seed

ควรอยู่ในทิศทางของ `packages/db` แต่ **ยังไม่เปิด runner อัตโนมัติใน 0.1.45 รอบแรก**. รอบนี้ให้กำหนดสัญญาและทำตัวอย่าง SQL/seed ที่แอปรันอย่างชัดเจนก่อน. Migration runner ที่เปิดใช้จริงต้องมี version history และ checksum, ตรวจลำดับ/ไฟล์ที่เปลี่ยนหลัง apply, กันการรันพร้อมกันจากหลาย process, ระบุนโยบายเมื่อสำเร็จบางส่วน และรองรับความต่างของ DDL ในแต่ละฐานข้อมูล. โดยเฉพาะ MySQL DDL หลายชนิด commit transaction โดยปริยาย จึงห้ามสัญญาว่า migration ทุก driver rollback ได้ทั้งชุด.

Seed ให้แยก **reference data ที่ต้องอยู่ในระบบ** (เขียนแบบ idempotent และ versioned หรือเป็นส่วนของ migration) จาก **demo/test data** (รันเฉพาะเมื่อระบุคำสั่งหรือ environment ชัดเจน). ห้าม seed production ทุกครั้งที่แอปเริ่ม, ห้ามใช้ seed ลบข้อมูลผู้ใช้โดยปริยาย. ก่อนลง runner ให้ทดสอบ run ซ้ำ, seed ผิดกลางทาง, concurrent deploy และการกู้คืนหลัง migration ล้มเหลวกับ MySQL/PostgreSQL/SQL Server แยกกัน.

## ไม่รวมใน 0.1.45

ORM, model auto-mapping, query builder, migration/seed runner, automatic retry, cache, SQL dialect translation, cursor/streaming, nested transaction และการเปลี่ยน isolation level. ขีดจำกัดจำนวนแถว/ไบต์, timeout, cancellation, pool และ transaction ยังเป็นของ [`std.db`](../../docs/REV5-DATABASE.md). สำหรับ transaction ต้องระวังผล commit ที่ไม่แน่นอนและไม่ retry write โดยอัตโนมัติ.

## ผลการลงงานและการทดสอบ candidate

- เพิ่ม `packages/db` ด้วย Public API ของ `std.db`: row cardinality, exact column lookup, required/optional extraction สำหรับ Text, Int, Decimal, Bytes, Time และ `expectAffected`. ข้อผิดพลาดของ helper มี prefix แยกกรณีและไม่ใส่ SQL/DSN/ค่าจากแถว.
- ผูก package `db` กับเวอร์ชัน Craft 0.1.45 และปฏิเสธ version field ที่ระบุเอง. อัปเดตเวอร์ชันของ package ที่ bundled อยู่แล้ว (`gui`, `api`), lockfile ตัวอย่าง และรายการไฟล์ใน Windows installer source.
- `packages/db`: `craft check`, `craft test` 5 กรณี, `craft fmt --check` และ `craft build` ผ่าน. `examples/db-package`: `craft check`, `craft fmt --check`, `craft build` ผ่าน.
- `go test ./... -count=1 -timeout 600s` ผ่าน. ตัวอย่าง read-only ทำงานกับฐาน `craft_test` บน MySQL, PostgreSQL และ SQL Server Express ผ่านทั้ง query ปกติและ query ภายใน transaction โดยอ่าน DSN จาก encrypted local secret store; ไม่พิมพ์ credentials.
- ตัวอย่าง `packages/api` กับ `packages/db` ยังไม่มี shared connection injection ผ่าน `api.App`. ใช้ `std.db` ภายใน handler ที่เป็นเจ้าของ connection เฉพาะ request ได้ แต่การแชร์ pool ข้าม request ยังไม่รับรอง; ไม่เพิ่ม API ที่ขัดกับ ownership ในรอบนี้. Migration/seed runner ยังไม่รวมตามขอบเขตด้านบน.
- ผู้ใช้ยังต้องตรวจรับ candidate และทดลอง Windows installer ที่ build จาก source นี้; รอบนี้ยังไม่ได้สร้าง setup ใหม่.
