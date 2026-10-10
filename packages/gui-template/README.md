# Craft GUI Template package 0.1.50

`gui-template` is a Showcase-inspired composition package built on `packages/gui`. It provides ready-designed navigation, TopBar, page headers, cards, sections, and a responsive shell. It does not own application state, event handling, data, or database access. Use `packages/gui` directly when you want full control over the design.

Add both packages to `craft.toml`:

```toml
[dependencies]
gui = { path = "../craft/packages/gui" }
template = { path = "../craft/packages/gui-template", version = "0.1.50" }
```

Adjust paths for your project, run `craft package lock`, then import:

```craft
import gui "gui"
import template "template"

let menu: UiNode[] = [
    template.sectionLabel("group", "WORKSPACE"),
    template.navItem("home", "Home", selectedPage),
    template.navItem("items", "Items", selectedPage),
]
let sidebar: UiNode = template.navigation("sidebar", "CRAFT / APP", "Desktop app", menu, "Craft 0.1.50")
let topbar: UiNode = template.topBar("topbar", "My App", [gui.secondaryButton("refresh", "Refresh")])
let content: UiNode = template.card("welcome", "WELCOME", "Ready", "Add your own content here.")
let root: UiNode = template.shell("root", sidebar, topbar, content, windowWidth)
```

`navItem` and `navigation` retain their Rev.47 signatures. `navItemIcon(id, label, icon, selected)` adds a leading icon; `navItemState(id, label, icon, selected, hovered)` accepts app-owned hover state. Supported icon names: `layout-dashboard`, `package`, `arrow-down-to-line`, `arrow-up-from-line`, `history`. Unknown names render without an icon. The returned Canvas keeps the supplied ID for `CanvasPressed` and has a full-row target named `<id>-target`. Handle `CanvasTargetEntered` and `CanvasTargetLeft` to update the hovered ID, then call `gui.update` only when it changes. The current engine does not suppress `CanvasPressed` for a disabled Canvas, so this package does not expose a disabled item variant.

`navigationWithStatus(id, brand, caption, items, status, runtime)` adds a fixed connection/status footer. Navigation items scroll separately when needed. `navigationFooter` and the internal brand header are distinct compositions. `navigation` accepts any `UiNode[]`, so section labels and custom GUI nodes can be mixed. `shell` switches to a stacked layout below 760 logical px. `card` and `section` are styled panels; `pageHeader` provides the Showcase typography. Icons use native Canvas strokes; there is no icon font, WebView, or additional dependency.

See [`examples/gui-template`](../../examples/gui-template) for a complete application. The visuals derive primarily from [`examples/gui-showcase`](../../examples/gui-showcase), especially its `shell.craft`, `app_view.craft`, and `components.craft`.

## Rev.49 Phase 1 foundation

Shared Enterprise Light tokens live in `src/theme.craft`; layout in `src/layout.craft`; actions in `src/actions.craft`; badges in `src/badges.craft`. Existing functions remain available. New layout functions are `appHeader(id, title, subtitle, actions)`, `pageContainer(id, children)`, `contentSection(id, title, description, actions, content)`, `breadcrumb(id, parts)`, `tabs(id, labels, selected)`, `toolbar(id, actions)`, and `splitPanel(id, first, second, percent)`.

New actions are `actionButton(id, label, variant, enabled)` with `primary`, `secondary`, `outline`, and `danger` variants; `iconButton(id, icon, label, enabled)`; `iconButtonState(id, icon, label, enabled, hovered)`; `buttonGroup(id, buttons)`; `splitButton(id, primaryLabel, secondary, enabled)`; and `toggleSwitch(id, label, on, enabled)`. `statusBadge(id, label, state)` supports `success`, `warning`, `danger`, and neutral. Native buttons emit `ButtonPressed`. Icon buttons emit `CanvasTargetPressed` with `<id>-target`; handle that event instead of `CanvasPressed` so disabled icon targets remain inactive. Tabs emit `ButtonPressed` IDs `<id>-<index>`; toggle state belongs to the app.

`splitButton` presents adjacent actions, without a popup menu. Engine-level pressed styling applies to native buttons; Canvas icon buttons need app-owned hover state and do not yet have a separate pressed visual. See [Craft Enterprise UI Showcase](../../examples/enterprise-ui-showcase/README.md) for interactive Phase 1 usage. Later phase pages there are labeled placeholders.

## Rev.50 responsive compositions

Rev.50 adds width-aware helpers without changing the original `navigation`, `navItem`, `topBar`, or `shell` signatures:

| API | Purpose |
| --- | --- |
| `layoutMode(width: Int): String` | Returns `narrow` below 650, `compact` below 900, `standard` below 1200, otherwise `expanded`. |
| `pagePadding(width: Int): Int` | Returns 14, 20, or 28 logical px at the corresponding narrow/compact/desktop widths. |
| `appHeaderAdaptive(id, title, subtitle, leading, center, actions, compactActions, width)` | White toolbar with title, optional subtitle, action slots, and a bottom divider. The app supplies compact actions. |
| `shellAdaptive(id, sidebar, compactSidebar, topbar, content, width, preference, narrowOpen)` | Expanded, collapsed, and narrow shell. The caller retains `preference` (`auto`, `expanded`, or `collapsed`) and `narrowOpen`. |
| `estimatedContentWidth(width, preference)` | Estimate for grid column caps after sidebar and page padding. |
| `navigationCompact(id, brand, items, connected)` and `navItemCompact(id, label, icon, selected, hovered)` | 72 px sidebar and icon-only navigation tiles. |
| `pageContainerResponsive(id, children, width)` | Page composition with adaptive padding. |
| `formGrid(id, fields, availableWidth, maxColumns)` | One to three columns using `gui.autoGrid`; fields come from `packages/gui` or the application. |
| `statCard(id, label, value, detail, icon, status)` | Metric card with 30 px value and optional supported icon. |
| `metricGrid(id, cards, availableWidth)` and `metricGridResponsive(id, cards, availableWidth, viewportWidth)` | Content-width card layout with a narrow-window one-column variant. |
| `tableViewport(id, table)` | Scroll view around an application-provided table node. |

The app should refresh its width on `Resized` and `DpiChanged`, then rebuild the affected UI tree while retaining form values, selection, and navigation state. `shellAdaptive`'s narrow sidebar is stacked above content; it is not a floating overlay. `tableViewport` is not a data table and does not add virtualization or sticky headers. `formGrid` does not provide input/validation components. The package does not provide a general icon renderer beyond its five Canvas icon names.

The separate [Craft UI Showcase](https://github.com/spidermeaow/craft-ui) (currently private) inventories all 61 public functions (36 UiNode components and 25 design/layout helpers) and marks unavailable Feedback/Overlay components as planned. Its repository uses a sibling Craft checkout as a path dependency. See the [Rev.50 guide](../../docs/REV50-RESPONSIVE-GUI.md) for acceptance limits.
