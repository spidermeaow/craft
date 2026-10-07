# Craft UI Engine 0.1.20 — Virtualized List

`gui.virtualList(id, itemCount, rowHeight, firstIndex, items)` สร้างรายการแนวตั้งที่ทุกแถวสูงเท่ากัน. `itemCount` คือจำนวนแถวทั้งหมด; `firstIndex` คือ index ของแถวแรกใน `items`. Engine คำนวณความสูง scroll จากจำนวนแถวทั้งหมด แต่เก็บและวัดเฉพาะ `items` ที่แอปส่งมา.

เริ่มด้วย `items` ว่าง แล้วรอ `VisibleRangeChanged` event. `event.nodeId` คือ ID ของ list, `event.firstIndex` และ `event.endIndex` เป็นช่วง `[first, end)` ที่ engine ต้องการ ซึ่งรวม overscan สูงสุด 3 แถวแต่ละด้าน. แอปสร้างแถวด้วย Craft แล้ว `gui.update` โดยส่ง `firstIndex` และ `items` ชุดใหม่. ตัวอย่างเต็มอยู่ใน `examples/gui-stress` และ `D:\02-Repository\test-gui`.

```craft
if event.kind == "VisibleRangeChanged" && event.nodeId == "list" {
    first = event.firstIndex
    end = event.endIndex
    var rows: UiNode[] = []
    for index in first..end {
        rows.append(gui.button("item-" + std.convert.toString(index), "Item " + std.convert.toString(index)))
    }
    gui.update(window, gui.virtualList("list", 100000, 40, first, rows))
}
```

ใช้ ID ของแถวที่คงที่ตามข้อมูล ไม่ใช้ตำแหน่งบนจอ. เมื่อจำนวนแถวลดลง engine clamp offset และร้องขอช่วงใหม่ถ้าจำเป็น. ช่วงที่ยังไม่ได้ส่งอาจว่างชั่วคราวหลัง scroll หรือ jump ไกล; อย่าเอาข้อมูลของ index อื่นมาวาดแทน. Event ที่มีช่วงเก่าอาจค้างใน queue เมื่อสลับหน้าจอ จึงควรตรวจ `nodeId` และ state ปัจจุบันก่อน update.

ขอบเขต: `itemCount` 0–100,000,000, `rowHeight` 16–8192 logical pixels, ความสูงรวมไม่เกิน 2^40 logical pixels และส่งสูงสุด 1024 row nodes ต่อครั้ง. UI tree ทั้งหมดจำกัด 4096 nodes. รุ่นนี้ยังไม่มี API กระโดดไป index โดยตรง, แถวสูงต่างกัน, Virtualized Grid, scrollbar แบบลาก, TextBox/IME หรือการแสดง placeholder ระหว่างรอข้อมูล. `CRAFT_UI_TRACE=1` เพิ่ม `virtual_rows_last_layout` และ `visible_range_requests`; เวลา draw completion ไม่ใช่เวลาภาพขึ้นจริงบนจอ.
