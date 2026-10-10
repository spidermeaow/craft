# UI Showcase 0.1.35 — manual scroll rendering check

Use Craft 0.1.35 with the local `gui` package. From `D:\02-Repository\app`, run `craft check` and `craft run`, then open **Excel**. The sheet now draws inactive cells and keeps one native TextInput in the selected cell, plus the value bar. The user performs this UI test; the source/build alone does not establish a pass.

| Action | Expected result |
| --- | --- |
| Scroll down slowly through rows 1–18, then back up | Every cell shows only its own text. No duplicated fragments, torn glyphs, or text from another row. |
| Scroll quickly and drag the vertical scrollbar | Text and cell borders settle in the same row without ghosting. |
| Compare scrolling with the previous 180-input version | Scrolling should feel more responsive; report any pause or delay that remains. |
| Click A2 after scrolling; replace `Notebook` with `Notebook test` | Caret is inside A2, the typed value stays in A2, and the value bar shows the same value. |
| Click empty A10 and type `New item`; scroll away and back | A10 still shows `New item`; no other cell changes. |
| Resize the window, switch to another page, then return | The edited values remain and native inputs stay inside their cell borders. |
| If available, repeat at 100% and 125% display scaling | Input text, caret, border, and click target remain aligned. |

If any step fails, provide a screenshot and note the row/column, scroll method, window size, display scale, and whether the issue also appears after the scroll stops. Check **Components** for one styled TextInput as a comparison. The **Layouts** VirtualGrid uses drawn text cells and is a separate case.
