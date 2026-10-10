# Craft 0.1.50: Responsive GUI template

**Status:** source candidate. Build and limited visual checks have been performed locally; the complete responsive/DPI and Inventory Manager acceptance matrix remains open.

Rev.50 extends `packages/gui-template` on the existing native Windows `std.ui` engine. It does not introduce React, WebView, HTML/CSS, a browser runtime, or an external UI dependency. Original `navigation`, `navItem`, `topBar`, and `shell` signatures remain available.

## Implemented in the shared package

- Enterprise Light color, typography, spacing, radius, and width helpers in `src/theme.craft`.
- `appHeaderAdaptive`, `pageContainerResponsive`, `formGrid`, and `tableViewport` in `src/layout.craft`.
- `navigationCompact`, `navItemCompact`, `shellAdaptive`, and `estimatedContentWidth` in `src/template.craft`.
- `statCard`, `metricGrid`, and `metricGridResponsive` in `src/cards.craft`.
- Native action buttons and status badges from the Rev.49 foundation. `topBar` delegates to the shared application header.

`layoutMode` uses logical-width breakpoints at 650, 900, and 1200 px. `shellAdaptive` hides its sidebar below 650 px unless the caller sets `narrowOpen`; below 900 px it defaults to a 72 px compact sidebar, and otherwise uses a 250 px expanded sidebar. The caller owns the manual preference, active page, inputs, focus-sensitive state, and data refresh. `metricGrid` caps columns from available content width; its responsive variant forces one column below a 650 px viewport. `formGrid` caps at one, two, or three columns from its supplied available width and maximum.

The API signatures and concrete usage are listed in [`packages/gui-template/README.md`](../packages/gui-template/README.md). The separate [Craft UI Showcase](https://github.com/spidermeaow/craft-ui) (currently private) provides a live component gallery and registry for the current package. It is a separate repository, not bundled with Craft Setup. In this repository, [`examples/gui-template`](../examples/gui-template) and [`examples/enterprise-ui-showcase`](../examples/enterprise-ui-showcase) remain smaller examples.

## Known limits

- The narrow sidebar is a stacked panel, not an overlay. The application must provide a reachable Menu action and retain its open state.
- `tableViewport` is a scroll wrapper; `gui-template` has no `dataTable`, sticky header, or table virtualization component.
- `formGrid` lays out caller-provided fields. Input controls and validation are supplied by `packages/gui` or the application.
- `splitButton` has adjacent actions but no popup. `toggleSwitch` is a button labeled On/Off with app-owned state.
- The built-in Canvas icon renderer recognizes only `layout-dashboard`, `package`, `arrow-down-to-line`, `arrow-up-from-line`, and `history`.
- The full requested matrix of 1600×900, 1280×720, 1024×768, 800×600, and 640×480 at 100%, 125%, and 150% DPI has not been completed. No claim is made for flicker, focus/scroll preservation, or performance across that matrix.

## Validation status

The local 0.1.50 CLI and Windows installer compiled successfully. Inventory Manager completed `craft check` during the UI work, and limited native screenshots were inspected at one display scale. The separate Craft UI Showcase was checked, tested (3 registry/preview tests), built, and launched locally; its 125% DPI captures cover expanded and narrow layouts with category navigation. These checks do not establish complete Rev.50 acceptance. Inventory Manager's database/transaction behavior and full page interaction matrix remain for the project owner to test.

The installer source is `packaging/windows/craft.iss`; generated binaries under `dist/` are ignored by Git and are not published merely by pushing this source repository. The project owner decides when the candidate is accepted or published as a release.
