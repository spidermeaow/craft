# Craft Rev.8 local packages

เอกสารนี้บันทึก contract ของ Rev.8; ตั้งแต่ Rev.11 มี GitHub dependencies เพิ่มแล้ว ดู [คู่มือปัจจุบัน](REV11-PACKAGES.md)

Rev.8 รับรอง package ที่อยู่ในเครื่องหรือ workspace เดียวกัน โดย resolve path จากตำแหน่งของ `craft.toml` ไม่ขึ้นกับ directory ที่เรียก CLI

## Manifest

Application ระบุ dependency ในตาราง `[dependencies]` โดยหนึ่ง package ใช้หนึ่งบรรทัด:

```toml
name = "my-app"
version = "0.1.0"
edition = "2026"
entry = "main"

[dependencies]
greeting = { path = "packages/greeting", version = "1.0.0" }
```

Package ที่ถูก import ต้องมี manifest ของตนเอง สำหรับ package ใหม่แนะนำให้ใช้ `entry = "library"`:

```toml
name = "greeting"
version = "1.0.0"
edition = "2026"
entry = "library"
```

รูปแบบเดิม `dependency.<alias>` และ `dependency-version.<alias>` รวมถึง dependency รุ่นเก่าที่ใช้ค่าเริ่มต้น `entry = "main"` ยังอ่านได้ใน Rev.8 เพื่อให้โปรเจ็คเดิมทำงานต่อ แต่รูปแบบ `[dependencies]` และ `entry = "library"` เป็นรูปแบบที่แนะนำสำหรับโค้ดใหม่

## คำสั่ง

```powershell
craft package check
craft package list
craft package lock
```

- `check` ตรวจ manifest, paths, versions, graph, imports และ lockfile ถ้ามี
- `list` แสดง root และ package ทุกชั้นตามลำดับคงที่ พร้อม source และ checksum แบบย่อ
- `lock` resolve graph ใหม่และเขียน `craft.lock` แบบ atomic

เมื่อมี `craft.lock`, `craft check`, `run`, `test`, `build` และคำสั่ง package จะปฏิเสธ graph ที่ source, version, path หรือ dependency เปลี่ยน ให้ตรวจการเปลี่ยนแปลงแล้วรัน `craft package lock` เพื่อยอมรับ graph ใหม่

ควร commit `craft.lock` ของ application เข้าระบบ version control ส่วน library ไม่จำเป็นต้องมี lock จนกว่าจะนำไปรันเป็น root project

`craft.lock` เก็บ package identity, relative source path และ SHA-256 ของ module graph ไม่เก็บ environment variables หรือ database credentials อย่างไรก็ตาม source code ของ package ยังถูกนำไปสร้าง source bundle ตามพฤติกรรม `craft build` ปัจจุบัน จึงห้ามเขียน secret ลง source หรือ manifest

## ขอบเขต

Rev.8 ยังไม่มี registry, download, publish, update, semantic version range หรือ GitHub dependency ทุก dependency ต้องมีอยู่ใน local filesystem ก่อนเรียก CLI
