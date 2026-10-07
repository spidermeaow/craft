# Craft 0.1.21 manual check

Build the current source before trying these commands; existing binaries in `dist/` can still be 0.1.20.

```powershell
go build -o dist/craft.exe ./cmd/craft
dist/craft.exe version
dist/craft.exe help
```

Expect `Craft 0.1.21` and `Phase 1 Rev.21`. The `craft gui init` command should explain that it is paused. `craft help` should not advertise it.

For HTTP, enter `examples/api-server`, run `..\..\dist\craft.exe check`, then `..\..\dist\craft.exe run`. In another terminal, request `http://127.0.0.1:8080/health`; expect JSON with `"language":"craft"`. The example uses `std.http` directly.

For UI, enter `examples/gui-stress`, run `..\..\dist\craft.exe check`, then `..\..\dist\craft.exe run` on Windows. Expect the existing grid and Virtual List interactions. This example uses `std.ui` directly. Close its window normally.

The archived package sources live in `Archive/packages/`. The sibling `D:\02-Repository\test-gui` keeps its own GUI package copy and is outside this change.
