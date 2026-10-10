# UI Showcase — manual check for Craft 0.1.36

Use `D:\02-Repository\craft\dist\craft.exe` with this app. The app dependency is locked to the local GUI package 0.1.36.

```powershell
cd D:\02-Repository\app
D:\02-Repository\craft\dist\craft.exe check
D:\02-Repository\craft\dist\craft.exe run
```

Excel opens first. The narrow row-number gutter and A–J headers identify each cell.

On the Excel page:

1. Click B2, type a unique value, click A10, then return to B2. B2 should retain the new value and the caret should appear inside the selected cell.
2. Change A10 and press Escape. Its previous value should return. Change it again and press Enter. The new value should remain.
3. Press Tab while editing. The editor should move to the next cell. Edit through the value bar with both Enter and Apply.
4. Scroll both axes, resize the window, leave Excel and return. Check that the editor remains aligned, values stay with the right cells, and scrolling remains responsive.
5. If available, repeat at Windows DPI 100% and 125% and with the GDI fallback. Note any focus, caret, IME, or text clipping issues.

The user performs UI testing. A successful build or package lock is not an acceptance result.
