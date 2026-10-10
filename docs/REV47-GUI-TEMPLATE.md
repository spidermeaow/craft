# Craft 0.1.47: GUI Template package

Rev.47 adds `packages/gui-template`, a Showcase-inspired package built on `packages/gui`. It provides styled navigation, TopBar, cards, sections, page headers, and a responsive shell. `craft new <directory>` remains the minimal project command; there is no GUI-specific `craft new` option.

Use both packages in an application:

```toml
[dependencies]
gui = { path = "../craft/packages/gui" }
template = { path = "../craft/packages/gui-template", version = "0.1.47" }
```

After `craft package lock`, import `gui` for windows, controls, and events and `template` for the prepared compositions. For example:

```craft
import gui "gui"
import template "template"

let items: UiNode[] = [template.navItem("home", "Home", selected)]
let sidebar: UiNode = template.navigation("nav", "MY APP", "Desktop app", items, "Craft 0.1.47")
```

The application owns the selected page and handles each navigation Canvas's `CanvasPressed` event. It also owns data, validation, and any database connection. This keeps `packages/gui-template` focused on reusable visual composition while `packages/gui` remains available for custom design.

The package uses `examples/gui-showcase` as its main visual reference: Canvas navigation from `src/shell.craft`, sidebar and page hierarchy from `src/app_view.craft`, and cards/sections from `src/components.craft`. TopBar uses the same color, border, typography, and spacing values. See the [package guide](../packages/gui-template/README.md) and [complete example](../examples/gui-template/README.md).

This is a source candidate. The user performs `craft check`, `craft run`, resizing, navigation, form interaction, and visual acceptance.
