# Craft 0.1.15 — Windows GPU renderer ขั้นต้น

Rev.15 เพิ่ม hardware Direct2D render target และ DirectWrite text path ให้ UI tree ที่เริ่มใน Rev.14 เมื่อสร้าง backend นี้ไม่ได้ Windows host ใช้ GDI fallback ไม่มี WebView และผู้เขียนแอปยังใช้ Craft source ผ่าน `packages/gui`

```craft
let window: UiWindow = gui.openWindow("Example", 640, 400)
gui.mount(window, screen("Hello"))
print(gui.renderer(window))
```

`gui.renderer(window)` คืน `direct2d-hardware` เมื่อสร้าง hardware target ได้ หรือ `gdi` เมื่อ fallback หาก window ปิดแล้วจะได้ handle error ตัวอย่างจริงอยู่ใน `examples/gui-layout`

Direct2D วาดพื้นหลัง panel และปุ่มมุมโค้ง; DirectWrite วาดข้อความหนึ่งบรรทัดด้วย Segoe UI ขนาด 18 logical pixels Resource ที่ผูกกับ device จะถูกสร้างใหม่เมื่อ Direct2D แจ้ง device loss ตาม [คำแนะนำของ Microsoft](https://learn.microsoft.com/en-us/windows/win32/learnwin32/drawing-with-direct2d) และยังไม่มี frame loop ตอนอยู่นิ่ง

Input เปลี่ยนให้ปุ่มส่ง `ButtonPressed` เมื่อปล่อยเมาส์ในปุ่มที่กดไว้; ปล่อยนอกปุ่มจะยกเลิก Tab/Shift+Tab และ Enter/Space ยังใช้กับปุ่มได้ Focus ของปุ่มเดิมคงอยู่หลัง `gui.update` หาก ID นั้นยังอยู่

ข้อจำกัด: renderer นี้ยังไม่มี text measurement, wrapping, IME, clipping ตาม parent, custom style, animation หรือ virtualized list Direct2D hardware target มี GDI fallback แต่ยังต้องตรวจจริงหลายเครื่องและหลาย DPI ก่อนอ้างคุณภาพหรือ FPS ไม่มีการรับประกัน 60/120 FPS ในรุ่นนี้

GUI package ที่คัดลอกเข้ามาในโปรเจกต์เดิมไม่ถูกแก้เองโดย CLI ให้สำรองการแก้ไขของตัวเองก่อนอัปเดต source ใน `packages/gui`, dependency version และ lockfile
