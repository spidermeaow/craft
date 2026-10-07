# Craft 0.1.12 — ความเสถียรของ GitHub packages

เอกสารนี้อธิบายสิ่งที่เปลี่ยนจาก [สัญญา Rev.11](REV11-PACKAGES.md) ระบบยังติดตั้ง **Craft source จาก public GitHub ด้วย exact tag** ไม่โหลด DLL/native library, ไม่ค้น registry และไม่อัปเดตรุ่นให้อัตโนมัติ

## การติดตั้งซ้ำและการกู้คืน

`craft install` จากโปรเจกต์ที่ manifest, lock และ cache ตรงกันจะรายงานว่าไม่มีการเปลี่ยนแปลง โดยไม่เขียน `craft.toml` หรือ `craft.lock` ใหม่ การติดตั้ง tag เดิมที่มี pin อยู่ยังใช้ commit/checksum เดิม เว้นแต่สั่ง `--refresh` พร้อม source อย่าง explicit

ถ้า cache ของ package ที่ล็อกไว้เสีย Craft จะหยุดก่อนใช้ source นั้น ให้ใช้ cache directory ใหม่แล้ว restore จาก lock:

```powershell
$env:CRAFT_PACKAGE_CACHE = Join-Path $env:LOCALAPPDATA 'Craft\packages-recovery'
craft install
craft package check
```

ใช้ชื่อ directory ใหม่ที่ว่างและคุณควบคุมได้ คำสั่งนี้ไม่ลบ cache เดิม ไม่เปลี่ยน pin ใน lock และไม่แก้ cache ของโปรเจกต์อื่น ตรวจ checksum ของ source ที่ดาวน์โหลดใหม่ก่อนใช้ หากต้องใช้ offline ต้องเตรียม verified cache ของ pin เดิมไว้ก่อน `craft install --offline` จะไม่เรียก network และไม่สามารถซ่อม cache ที่ขาดหรือเสียได้

ถ้า install ถูกขัดจังหวะระหว่างเขียน manifest/lock ให้สั่ง `craft install --recover` ก่อนใช้โปรเจกต์ Recovery จะไม่เขียนทับไฟล์ที่ถูกแก้หลังเหตุขัดจังหวะ; เมื่อข้อความระบุ manual review ให้สำรองและตรวจ `craft.toml`, `craft.lock` และ journal ก่อนแก้เอง

## Network และ retry

คำขอ GitHub มี timeout 60 วินาทีต่อครั้ง จำกัดการลองไม่เกิน 3 ครั้งสำหรับ network timeout และ HTTP 502/503/504 โดยรอ 200 และ 400 มิลลิวินาทีก่อนลองใหม่ การยกเลิกจะหยุดการรอทันที HTTP 403/429, 404, redirect ที่ไม่อนุญาต, archive ที่ไม่ถูกต้อง, checksum mismatch และ validation/security failure จะไม่ถูก retry อัตโนมัติ

ข้อความ error ระบุ package/tag และขั้นตอน resolve, download หรือ validate เมื่อมีข้อมูลเพียงพอ ไม่พิมพ์ response body จาก GitHub ลง error

## Compatibility

Rev.12 ไม่เปลี่ยน syntax ของ Craft, schema `craft.lock` หรือรูปแบบ source bundle ที่ใช้ใน Rev.11 Bundle ที่สร้างโดย Craft 0.1.12 ยังระบุ language `0.1.11` โดยตั้งใจเพื่อให้ runtime Rev.11 ที่รองรับ bundle นี้อ่านได้; นี่ไม่ใช่เลขเวอร์ชันของ CLI ที่สร้าง bundle

ดู [คู่มือทดลอง Rev.12](TRY-REV12.md) และ [สถานะการตรวจรับ](../roadmap/phase1/output/phase1-rev12-status.md) สำหรับขั้นตอนและผลที่ยืนยันแล้ว
