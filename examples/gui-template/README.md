# GUI Template example

This project shows the short `template.*` compositions from `packages/gui-template` alongside lower-level `gui.*` controls. The visual design follows `examples/gui-showcase`.

From this directory with Craft 0.1.50:

```powershell
craft package lock
craft check
craft run
```

Switch between Overview and Items, resize the window, type a name, and press Save. Navigation uses Canvas IDs (`CanvasPressed`); the app owns page state and event handling. The example uses only placeholder data and does not access a database or network.
