# UI Showcase manual check — Craft 0.1.37–0.1.38

The app starts on **Rev 37–38 Lab**. This checklist is for manual testing; UI results have not been recorded yet.

## Before opening

From `D:\02-Repository\app`, use Craft 0.1.38:

```powershell
craft version
craft check
craft run
```

Expected: `craft version` reports `0.1.38`, `craft check` reports no errors, and the Lab page opens. If `craft version` is older, open a new PowerShell or invoke `D:\02-Repository\craft\dist\craft.exe` directly.

## 1. Range selection and direct editing (Rev.37)

1. Click B2. Its value should appear in the value bar, and the cell editor should sit over B2.
2. Shift+click D4, drag over another range, and try Shift+Arrow. The highlighted rectangle and active cell should follow the pointer or keyboard; no values should change from selection alone.
3. Click a single cell, type a value, then press Enter. Navigate away and back; the value should remain in that cell. Repeat with Escape; the prior value should return.
4. Resize the window and scroll the page; the editor, caret, cell text, and selection must stay aligned.

## 2. Unicode clipboard (Rev.37)

1. Select a 2×3 range and press **Copy** or Ctrl+C. Paste into a text editor; expect tabs between columns and line breaks between rows.
2. Put a small tab-separated block containing Thai text into the clipboard. Select a destination cell and press **Paste** or Ctrl+V. The text should appear in the intended cells without changing unrelated cells.
3. Paste near the bottom or right edge. Values outside A–J/row 18 should be ignored; existing unrelated values should remain.

## 3. Virtual grid accessibility (Rev.38 work in progress)

Use Accessibility Insights or a screen reader on the sheet and the generic grid farther down the Lab page. A visible VirtualGrid should report a DataGrid role. Each loaded cell should report its one-based row, one-based column, and value. Keyboard focus should be reported on the grid when it is focused. UIA Grid/Table/Selection patterns and lookup of unloaded cells are still pending; record them as open work, not passed behavior.

## 4. Generic large viewport

Scroll the 100,000-row × 1,000-column sample quickly on both axes. Click a visible cell and compare its outline, status line, and reported coordinates. Use arrows, Page Up/Down, Ctrl+End, and Ctrl+Home. The visible range label should move with the viewport; blank or stale cells that remain after scrolling are defects to report.

Scroll the outer page down until this grid becomes visible and keep scrolling to the bottom. The page must stay where you moved it when the grid loads a new range; it must not jump back to the sheet editor near the top. Clicking or focusing a cell may intentionally bring that cell into view.

## 5. Environment and timing

Repeat the visual and input checks with Direct2D and `CRAFT_UI_RENDERER=gdi`, at 100%, 125%, and 150% DPI if available. Try Thai IME and keyboard-only navigation. Record Windows version, renderer, DPI, steps, expected/actual result, and a screenshot for failures.

To collect timing in a fresh PowerShell, set `CRAFT_UI_TRACE=1` before `craft run`, then close the app normally and save the `craft-ui-engine-timing` JSON from stderr. It measures time to finish draw calls, not display presentation. Trace samples are bounded in this build.

The user performs these runtime/UI checks. Preparing this page and regenerating `craft.lock` did not run the app or its tests.
