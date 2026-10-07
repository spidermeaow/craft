# Craft UI 0.1.27 / GUI package 0.2.5

## อาการจาก dashboard

ภาพหน้าจอแสดงหัวข้อ ตัวเลข และแถวตารางซ้อนกันในการ์ด; ข้อความเมนู sidebar หายหรือถูกปุ่มขนาดใหญ่ทับ; มี scrollbar แนวนอนแม้หน้าต่างกว้าง. ภาพนี้เป็นผลจาก layout ไม่ใช่ข้อมูลสถิติผิด.

## สาเหตุและการแก้

- Runtime `Panel` วาง children ทุกตัวในพื้นที่เดียวกันเพื่อการซ้อนชั้น และ `surface` รับเฉพาะ `Panel` หรือ `Button`. Package 0.2.3 ใช้ `Panel` กับ children หลายตัวจึงทำให้ข้อความซ้อนกัน; การใช้ `Column` ตรง ๆ ใน 0.2.4 ทำให้ `surface` ล้ม. Package 0.2.5 ใช้ `Panel` ชั้นนอกที่มี `Column` เพียงตัวเดียว. Dashboard ใช้โครงสร้างเดียวกันใน sidebar.
- Runtime 0.1.27 ให้เนื้อหาของ `ScrollView` เติมความกว้าง viewport เมื่อ intrinsic width เล็กกว่า viewport แต่ยังคง intrinsic height. ส่วนที่กว้างกว่า viewport ยังเลื่อนแนวนอนได้.
- Dashboard เคยคำนวณ `contentWidth` และ header จาก viewport โดยหัก padding ไม่ครบ. ปรับให้หัก root, row, sidebar และ content padding ก่อนส่งความกว้างให้ grid และ Canvas.

## ตรวจด้วยตนเอง

Build CLI จาก source 0.1.27 แล้วใช้คำสั่ง `craft package check`, `craft check`, `craft run` ใน `D:\02-Repository\gui-package-dashboard`. ตรวจที่หน้าต่างกว้างและแคบว่า stat card, ปุ่ม, ตาราง, chart และ sidebar ไม่ซ้อนกัน; ไม่มี scrollbar แนวนอนที่ความกว้างปกติ; scrollbar แนวตั้ง, TextInput และ Canvas target ยังทำงาน. ยังไม่ได้รัน automated tests หรือเปิด GUI หลังแก้ตามความต้องการของผู้ใช้.
