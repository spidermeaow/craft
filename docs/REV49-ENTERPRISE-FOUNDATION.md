# Craft 0.1.49: Enterprise UI foundation (Phase 1)

Rev.49 adds shared Enterprise Light colors, typography sizes, spacing, radii, layout compositions, native buttons, icon Canvas targets, and status badges to `packages/gui-template`. It reuses `packages/gui` and keeps the existing `navigation`, `navItem`, and `shell` signatures.

The public additions are documented in [`packages/gui-template/README.md`](../packages/gui-template/README.md). The interactive [`examples/enterprise-ui-showcase`](../examples/enterprise-ui-showcase/README.md) has Dashboard, Components, and Navigation exercises. Forms, Data, Dialogs, and Settings are named placeholders for later phases. This release does not implement Phase 2–6 components.

Native buttons supply focus, disabled behavior, and click events. Icon buttons use native Canvas strokes and enabled hit regions; callers handle `CanvasTargetPressed`. The current engine does not automatically recolor Canvas hover or pressed states, so `iconButtonState` takes app-owned hover state. `splitButton` is a compound action row; a contextual menu waits for the overlay phase. No web runtime or external UI dependency is added.
