# ทดลอง Craft 0.1.12 — GitHub packages

ใช้ Windows x64 และเปิด terminal ใหม่หลังติดตั้ง `Craft-setup.exe` 0.1.12 หรือเรียก `dist\craft.exe` แบบ portable จาก repository

```powershell
craft version
craft new github-demo
cd github-demo
craft install github.com/spidermeaow/craft-test@v1.0.0 --alias greeting
craft package list
```

คาดหวัง version `0.1.12`, `craft.toml` มี `greeting` และ `craft.lock` มี repository, tag, commit และ checksum ของ package โดยไม่มี absolute cache path

แทนที่ `src/main.craft` ด้วย:

```craft
import greeting "greeting"

func main() {
    print(greeting.greet("Craft"))
}
```

```powershell
craft check
craft test
craft run
craft package check
craft install
craft install --offline
craft build
craft run dist/github-demo.craftbundle
```

คาดหวัง `craft run` และ bundle พิมพ์ `Hello, Craft!` การติดตั้งซ้ำรายงานว่า manifest/lock ไม่เปลี่ยน และ `--offline` ใช้ cache เท่านั้น Bundle ยังต้องใช้ Craft CLI ในการรัน

## ตรวจ restore ด้วย cache ใหม่

เก็บ `craft.lock` เดิมไว้ แล้วตั้ง cache directory ว่าง:

```powershell
$env:CRAFT_PACKAGE_CACHE = Join-Path $env:TEMP 'craft-rev12-fresh-cache'
craft install
craft package check
```

คาดหวัง commit/checksum ใน lock ไม่เปลี่ยน แม้ดาวน์โหลด source ใหม่ ใช้ directory ที่ว่างจริงก่อนเริ่ม; หากเคยใช้ชื่อนี้แล้ว ให้เลือกชื่ออื่น เมื่อทดสอบเสร็จให้เปิด terminal ใหม่หรือสั่ง `Remove-Item Env:CRAFT_PACKAGE_CACHE` เพื่อกลับไปใช้ cache ปกติ

## กรณีผิดพลาดที่ควรตรวจ

| กรณี | ผลที่คาดหวัง |
| --- | --- |
| `craft install --offline` กับ cache ว่าง | ล้มเหลว พร้อมแนะนำ online install; ไม่มี network และ manifest/lock ไม่เปลี่ยน |
| tag หรือ repository ไม่มีจริง | ล้มเหลว ระบุขั้นตอน resolve และ HTTP status; ไม่สร้าง dependency ใน manifest |
| lock ไม่ตรงกับ manifest | คำสั่งตรวจปฏิเสธ ไม่เลือกเวอร์ชันใหม่เอง |
| cache checksum ไม่ตรง | ปฏิเสธ source พร้อมวิธีใช้ cache directory ใหม่; lock ไม่เปลี่ยน |
| journal ค้าง | คำสั่งโปรเจกต์หยุดและแนะนำ `craft install --recover` |

อย่าแก้ source ใน shared cache เพื่อสร้าง library ให้ใช้ local path dependency แทน ดูวิธีกู้ใน [คู่มือ Rev.12](REV12-PACKAGES.md)

## ตรวจจาก source repository

```powershell
go test ./...
go test -race ./...
$env:CRAFT_REV11_LIVE = '1'
go test ./tests -run '^TestRev11LiveGitHub$' -count=1 -v -timeout 300s
```

Live test ใช้ public `spidermeaow/craft-test@v1.0.0` และ temporary project/cache โดยไม่ push หรือแก้ repository ปลายทาง ผลแต่ละขั้นต้องบันทึกตามจริงใน [สถานะ Rev.12](../roadmap/phase1/output/phase1-rev12-status.md)
