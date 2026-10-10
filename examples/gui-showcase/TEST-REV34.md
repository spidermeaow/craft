# UI Showcase 0.1.34 — manual test guide

This guide is for a person to run. The Showcase does not mark itself as passed.

## Setup

Use Craft 0.1.34 and the local `gui` package. Open PowerShell in `D:\02-Repository\app`, then run `craft check` and `craft run`. Open **Layouts** and scroll to **Interaction Check**. The status line reports the selected cell and whether the last selection came from mouse or keyboard. The next line reports the visible data snapshot. They should update independently.

## Grid checks

| Action | Expected result |
| --- | --- |
| Click a visible cell near the middle | One blue outline surrounds that cell. Mouse status shows the same row and column as its label. |
| Press Right, Down, Left, Up | Selection moves one cell per key and returns to the original cell. Keyboard status reports each new position. |
| Press Page Down, then Page Up | Selection and viewport move about one visible page and return toward the starting row. |
| Press Home and End | Selection moves to the first and last column of the current row. |
| Press Ctrl+End, then Ctrl+Home | Status reaches row 100,000 / column 1,000, then row 1 / column 1. The outlined cell is visible after loading. |
| Press Arrow at the first or last row/column | Selection stays within the logical 100,000 × 1,000 bounds. |
| Press Tab, then click the grid again | Tab leaves grid focus; clicking a cell restores mouse selection and arrow navigation. |
| Resize window and repeat a click | Outline stays on the clicked cell; text, border, and hit area remain aligned. |
| If available, repeat at 100% and 125% display scaling | Outline and click target remain aligned with the rendered cell. |

If a result fails, note the action, expected and observed row/column, window size, display scale, and a screenshot. Fast wheel scrolling can still briefly outpace cell drawing; that behavior was accepted for 0.1.33 and is not the target of this test.

## Component spot check

On **Components**, inspect centered Button labels, the disabled TextInput background, and the Canvas target. After the window opens, the cursor should stop showing busy state. These are spot checks because 0.1.33 was accepted without exhaustive testing.

## Excel direct editing check

1. Open **Excel** and click **A2**. It starts with `Notebook`; the value bar should show the same text.
2. Replace A2 directly with `Notebook test`. The cell and value bar should show the new value without pressing Apply.
3. Click **A10**, which starts empty. Type `New item` directly in the cell. It should remain visible after you click another cell.
4. Select **B2**, type `East` in the value bar, and press **Apply** or Enter. B2 should show `East`.
5. Switch to another page and return. The three edits should remain until you close the app.

These checks cover the Showcase's app-owned 18 × 10 sheet. They do not test in-cell editing of `VirtualGrid`.
