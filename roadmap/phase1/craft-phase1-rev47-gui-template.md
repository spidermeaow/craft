# Craft Phase 1 Rev.47 (0.1.47) — GUI Template package

**Status:** source candidate; user acceptance pending.

## Intent

Provide short `template.*` calls for a consistent Showcase-style Windows UI. Developers who want a fully custom design continue to use `packages/gui` directly.

## Contract

- `packages/gui-template` depends on the matching `packages/gui` version.
- `template.navItem`, `sectionLabel`, `navigation`, `topBar`, `pageHeader`, `card`, `section`, and `shell` compose public `gui` APIs.
- Applications own page state, event routing, labels, data, validation, and persistence.
- `craft new <directory>` remains minimal. Rev.47 does not add a GUI-specific CLI template option.
- The visual design derives primarily from `examples/gui-showcase`.

## User acceptance

1. Use a 0.1.47 CLI with both matching packages.
2. In `examples/gui-template`, run `craft package lock`, `craft check`, and `craft run`.
3. Inspect Sidebar, TopBar, cards and form; switch pages, type and save, and resize narrow and wide.
4. In a separate app, import `template` and create a navigation plus card with short API calls; confirm `packages/gui` remains directly usable.

The agent builds source artifacts but does not run application or automated tests under the user's testing preference. Rev.47 remains a candidate until accepted.
