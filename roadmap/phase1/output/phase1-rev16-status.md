# Phase 1 Rev.16 status

Date: 6 October 2026. Target: Craft 0.1.16. **Accepted by the user after personal testing; detailed performance and dense-layout acceptance remain separate.**

## Implemented in source

- Windows UI host uses an auto-reset wake event and `MsgWaitForMultipleObjectsEx` in place of the 8 ms polling sleep. Commands and cancellation wake the host; Win32 input wakes through the message queue.
- Same-shape tree updates reuse equal-share layout geometry and invalidate nodes whose text changed, falling back to full repaint after 16 changed nodes or more than half the window's area. Direct2D clips painting to the invalid region and skips non-intersecting subtrees. Hover/focus invalidation remains local. Structural updates, resize and DPI changes use full layout/paint.
- Optional `CRAFT_UI_TRACE=1` aggregate diagnostics record command queue, layout, paint-to-draw-completion, and correlated input-to-draw-completion timings. Diagnostics are off by default.
- Version, package, template, bundle compatibility, installer metadata and guides updated for 0.1.16.

## Evidence and limits

- The user reported that GUI responsiveness felt good in personal testing and explicitly accepted 0.1.16. This is a product acceptance decision, not a claim that every planned benchmark or visual scenario passed.

- Windows `go build` completed after the engine changes. This confirms compilation only.
- The Inno Setup installer compiled successfully. Automated tests, GUI execution, baseline comparison, visual inspection, actual installer installation, GDI fallback and latency measurements **were not run for this implementation task**, following the user's standing testing preference. Do not mark Rev.16 performance acceptance as passed.
- Direct2D `EndDraw` and completion of GDI draw calls do not measure display presentation. Same-shape reuse depends on the current equal-share layout; content-sized text/layout remains future work.

See [TRY-REV16](../../../docs/TRY-REV16.md) for manual acceptance.
