# Craft UI Showcase

A categorized gallery of what the public `gui` package can build on Windows.
This example was copied from `D:\02-Repository\app` into the Craft repository
on 10 October 2026. It is a UI Engine demonstration, not an Excel
implementation or proof that the UI Engine has passed Final acceptance.

## Explore by capability

| Sidebar group | Pages | What to inspect |
| --- | --- | --- |
| Start here | Engine overview | Capability map and open acceptance limits |
| Controls & layout | Controls; Themes & forms; Layout & scrolling | Text, buttons, inputs, focus, themes, responsive composition and scroll views |
| Data & tables | Tables; Editable grid; Virtual grid | Mock tables, direct cell editing, clipboard, range selection and a large virtual viewport |
| Drawing & media | Canvas & paths; Images & captions | Canvas geometry, repainting, image sizes and captions |
| App integration | Form example; Window & icon | Data-entry events, local save flow and native Windows icon |

The **Editable grid** page keeps values in memory while the window is open.
The **Virtual grid** page also contains the earlier Rev.37–38 interaction lab.
The blue icon in `assets/craft-blue.ico` is applied at startup to the window
and taskbar; the Window & icon page can switch it while the app is running.
The Rev.41 language example remains in `src/language_lab.craft` as a source
example but is outside the UI capability navigation.

The `TEST-REV*.md` files are preserved as historical review notes from the
original project. Their old paths and version numbers are not the run
instructions for this repository copy.

## Run and review

Use Craft 0.1.50 and open a new PowerShell session. Run from this
directory so images and icons resolve correctly:

```powershell
cd D:\02-Repository\craft\examples\gui-showcase
craft version
craft package lock
craft check
craft run
```

If `craft check` reports a problem, send its first diagnostic. The user performs
the checks and UI testing; this repository copy has not been executed by the
agent. For the underlying capability and Final gate, see
`D:\02-Repository\craft\roadmap\phase1\craft-phase1-rev42.md`.

For rendering comparison, close the app, set `$env:CRAFT_UI_RENDERER = 'gdi'`,
and run `craft run` again. Clear it with
`Remove-Item Env:CRAFT_UI_RENDERER -ErrorAction SilentlyContinue`.
