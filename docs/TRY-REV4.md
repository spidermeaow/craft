# ทดลอง Craft 0.1.4 / Phase 1 Rev.4

Agent เตรียม source/tests และ build artifacts แต่ไม่ได้รัน tests หรือทดลองแอปตามความต้องการของผู้ใช้ ผลด้านล่างเป็น **ผลที่คาดหวัง** จนกว่าคุณจะรันและรายงานผล

## ติดตั้ง

ติดตั้ง `dist/Craft-setup.exe` แล้วเปิด PowerShell ใหม่ หรือใช้ portable `dist/craft.exe`. ติดตั้ง Language Support VSIX **0.2.0 ตัวล่าสุดตัวเดียว** แทน language-support0.1.0; File Icons0.1.0 เป็นคนละ extension จะใช้ต่อหรือไม่ก็ได้

```powershell
craft version
```

คาดหวัง `Craft 0.1.4`, `Language: Phase 1 Rev.4`. หากยังเป็นรุ่นเก่า ใช้ `Get-Command craft -All` ตรวจ PATH

## Pure Craft tests และ HTTP app

คำสั่งนี้ใช้ checkout; ถ้าใช้ installer เปลี่ยน root เป็น `$env:LOCALAPPDATA\Programs\Craft` ซึ่งมี examples และ packages ครบ

```powershell
Set-Location D:\02-Repository\craft\packages\api-framework
craft check
craft test
Set-Location ..\..\examples\api-server
craft check
craft test
craft run
```

คาดหวัง framework6 tests และ app2 tests ผ่าน, check สำเร็จ. `craft test` ทั้งสองที่ไม่เปิด listener. `craft run` แสดง `HTTP listening on 127.0.0.1:8080` และรอ request; เปลี่ยน address ด้วย `$env:CRAFT_API_ADDRESS = '127.0.0.1:8081'` หาก port ถูกใช้

เปิด PowerShell อีกหน้าต่าง:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/health
Invoke-RestMethod http://127.0.0.1:8080/items/1
Invoke-RestMethod -Method Post -Uri http://127.0.0.1:8080/items -ContentType application/json -Body '{"name":"Craft"}'
curl.exe -i http://127.0.0.1:8080/items/2
curl.exe -i -X DELETE http://127.0.0.1:8080/health
curl.exe -i -X POST -H 'Content-Type: application/json' --data '{}' http://127.0.0.1:8080/items
```

คาดหวัง status200/200/201/404/405/422 ตามลำดับ; health มี `status=ok`, `language=craft`; response header `x-framework: Craft`;405 มี Allow. POST เป็น echo demo ไม่ได้บันทึกข้อมูล. กด Ctrl+C ในหน้าต่าง server เพื่อ drain/close; CLI exit1 ตาม cancellation convention. เริ่มใหม่ได้โดย port ไม่ค้าง

## เพิ่ม endpoint/middleware โดยไม่แก้ Go

อ่านตัวอย่าง `hello` และ `stamp` ใน [framework README](../packages/api-framework/README.md). เพิ่ม named handler และ `app = api.add(...)`/`api.useAfter(...)` ใน application แล้ว check/run ใหม่ด้วย CLI เดิม. คาดหวัง endpoint ใหม่และ header จาก hook โดยไม่ rebuild `craft.exe`. แต่ละไฟล์ที่ใช้ api ต้องมี `import api "api"`

## Bundle ย้ายออกจาก source tree

หยุด server ก่อนใช้ port เดิม แล้วรันจาก examples/api-server:

```powershell
craft build
$rev4Bundle = Join-Path $env:TEMP ('craft-rev4-' + [guid]::NewGuid().ToString() + '.craftbundle')
Copy-Item -LiteralPath .\dist\api-server.craftbundle -Destination $rev4Bundle
Push-Location $env:TEMP
craft run $rev4Bundle
Pop-Location
```

เรียก `/health` จากอีก terminal อีกครั้ง คาดหวัง200 โดยไม่ต้องมี packages ข้าง bundle. กด Ctrl+C ก่อน Pop-Location. Dependency source ถูกฝังใน bundle; เปลี่ยน source หลัง build ไม่เปลี่ยน bundle เดิม

## Regression และ network tests สำหรับผู้พัฒนา

ต้องมี Go ตาม go.mod. คำสั่งที่ระบุใช้ test timeouts เพื่อจำกัดกรณี lifecycle ผิดพลาด:

```powershell
Set-Location D:\02-Repository\craft
go test ./... -timeout 180s
go test ./tests -run 'TestRev4' -count=1 -v -timeout 180s
go test ./internal/stdlib -run TestHTTP -count=1 -v -timeout 60s
$env:CRAFT_STRESS = '1'
go test ./tests -run TestRev4HTTPStress -count=1 -v -timeout 300s
Remove-Item Env:CRAFT_STRESS
```

ชุด Rev.4 เตรียมตรวจ function signatures/callback collections, package visibility/identity/cycles, formatter, portable bundles,10,000 joined tasks, HTTP errors/body/binary/query/Set-Cookie, overload/timeout/disconnect/expired context, request-state/task isolation และ real API integration. Network tests ใช้ loopback ephemeral ports มี client/startup/shutdown deadlines; pure Craft tests ไม่เปิด network. Barriers/channels ใช้ใน transport lifecycle และ Rev.3 fake-clock tests ยังคงอยู่

Stress ส่ง10,000 HTTP requests หลัง warmup100 ใน process เดียวผ่าน Craft framework; payload GET /health, config ตาม defaults, client keepalive เดียว. บันทึก stdout ทั้งหมด, OS/CPU/RAM, Go version, config และเวลารัน; output แสดง heap หลัง GC ก่อน/หลัง workload และ goroutines. ทำซ้ำอย่างน้อย3รอบ แยกจาก active-concurrency test. ชุด lifecycle ตรวจ listener ปิดหลัง cancel; จำนวน task/goroutine/OS handles และ memory plateau ต้องประเมินร่วมกัน ไม่สรุปจาก request count อย่างเดียว

ระหว่าง manual server/stress บน Windows เก็บ process measurements เมื่อเริ่ม หลัง warmup ระหว่าง/หลัง requests และก่อนหยุด (test process ชื่อ tests.test อาจต้องเลือก PID จาก terminal):

```powershell
Get-Process craft -ErrorAction SilentlyContinue | Select-Object Id,CPU,WorkingSet64,PrivateMemorySize64,HandleCount,@{Name='Threads';Expression={$_.Threads.Count}}
```

หลัง shutdown process ต้องจบและ port กลับมาใช้ได้; repeated runs ไม่ควรมี requests/tasks/listeners ค้าง. ตัวเลข heap/OS memory ขึ้นกับ GC/allocator ไม่ใช้ threshold ที่ยังไม่เคยวัดเป็นคำรับรอง. ผล throughput, memory plateau และ race acceptance ยัง pending

Race tests ใช้ environment ที่รองรับ Go race detector และ C compiler (release script ตั้ง CGO0 เฉพาะระหว่าง build แล้วคืนค่า):

```powershell
$env:CGO_ENABLED = '1'
go test -race ./... -timeout 300s
```

ตรวจ Rev.3 ตัวอย่างเดิมด้วย `craft check`, `craft test`, `craft run` ใน `examples/rev3-tasks` และชุดที่คุณใช้ก่อนหน้า รวมถึง Rev.2 tests/bundles ตาม [TRY-REV3](TRY-REV3.md). สำหรับ editor รัน `npm install --ignore-scripts` และ `npm test` ใน craft-vscode แล้วตรวจสี/snippets ใน VS Code ด้วยตนเอง

ส่ง command, stdout/stderr, `$LASTEXITCODE`, version และ repro source เมื่อพบปัญหา; ช่อง acceptance ใน roadmap ยังไม่ถูกทำเครื่องหมายผ่านเพียงเพราะ build สำเร็จ
