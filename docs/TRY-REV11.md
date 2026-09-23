# ทดลอง Craft Rev.11

ติดตั้ง Craft-setup.exe 0.1.11 แล้วเปิด terminal ใหม่ จาก directory ที่ต้องการสร้างโปรเจกต์:

```powershell
craft version
craft new github-demo
cd github-demo
craft install github.com/spidermeaow/craft-test@v1.0.0 --alias greeting
craft package list
```

แก้ src/main.craft เป็น:

```craft
import greeting "greeting"

func main() {
    print(greeting.greet("Craft"))
}
```

```powershell
craft check
craft run
craft install --offline
craft build
craft run dist/github-demo.craftbundle
```

คาดหวัง `Hello, Craft!` ทั้ง source และ bundle; craft.toml ระบุ repository/tag ส่วน craft.lock ระบุ commit/checksum ไม่มี cache path ของเครื่อง

ตัวอย่างพร้อมใช้ใน repository: examples/rev11-github (ต้องสั่ง craft install ก่อน run)

ทดสอบสำหรับผู้พัฒนา จาก repository Craft:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/test-rev11.ps1 -Race -Stress -Live
```

Regression ต้องใช้ database test fixtures/credentials เดิม ตัวเลือก Live เรียก GitHub จริงและใช้ temporary cache/project ส่วน tests ปกติใช้ fixtures เพื่อไม่ขึ้นกับ network/rate limit

ดู contracts และวิธี recovery ใน [REV11-PACKAGES.md](REV11-PACKAGES.md) ห้ามแก้ source ใน cache เพื่อพัฒนา library ให้ใช้ local path dependency แทน
