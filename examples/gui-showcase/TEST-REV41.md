# UI Showcase Rev.41 manual check

Use Craft 0.1.41 and `gui` 0.1.41. From `D:\02-Repository\app`:

```powershell
craft version
craft check
craft run
```

The app should open on **Rev 41 Language**. Confirm **Module preview: 60 / expected: 60**. Click **Run calculation** and expect `Module returned 60`. Click **Check predicate** and expect `Predicate passed`. These calls cross the `showcase.language` module boundary and use a private helper shared across two source files.

Then visit **Excel**, **Contacts**, **Rendering Lab**, **Window Icon** and **Theme & Forms**. Confirm navigation, existing input, and rendering still behave as before. This checks compatibility between declared modules and the older shared project files. Source formatting can be inspected with `craft fmt --check`; this command only reports differences.

The agent has not run `craft check`, opened the app, or executed automated tests. Record the first error and page where behavior differs so the candidate can be corrected.
