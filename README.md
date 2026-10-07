# Craft 0.1.30

Craft is a general-purpose, statically typed programming language with a Go-based command-line interface and AST interpreter. Its language, runtime, standard library, and package system provide a foundation for applications and reusable libraries rather than prescribing one application framework. Today, Craft runs `.craft` programs on Windows and makes type and scope errors visible before execution.

Rev.30 adds focus scopes, Canvas target policies, scoped shortcuts, and an initial Windows UI Automation provider. Rev.29 added focus policies, keyboard commands, and focus rings. Rev.28 added opt-in Text wrapping and a Grid that uses the width assigned by its parent. Rev.26 added native text measurement and responsive grid composition. Rev.25 added viewport queries, a horizontal scrollbar, basic text/surface styling and Canvas hit regions to the native UI Runtime. Rev.24 added retained Canvas painting. `packages/gui` 0.2.8 is available as an optional Craft library; the API framework remains archived. Craft still cannot build standalone application executables. See the [project direction](roadmap/craft-language-direction.md) and [GUI direction](roadmap/craft-gui-direction.md) for the shared definition and future goals.

See the [Rev.28 UI guide](docs/REV28-UI-REFLOW.md), [plan](roadmap/phase1/craft-phase1-rev28.md), and [status](roadmap/phase1/output/phase1-rev28-status.md). The public-GitHub package system from Rev.12 and native UI bridge from Rev.13 remain available.

The [Craft 0.1.30 plan](roadmap/phase1/craft-phase1-rev30.md) records the scope and acceptance criteria.

Craft is currently in [Phase 1](roadmap/craft-phases.md), which includes Runtime and package development in the same revision. [Rev.24](roadmap/phase1/craft-phase1-rev24.md) covers Runtime painting and GUI package 0.2.1; [Rev.25](roadmap/phase1/craft-phase1-rev25.md) covers Runtime interaction and GUI package 0.2.2; [Rev.26](roadmap/phase1/craft-phase1-rev26.md) covers responsive composition and GUI package 0.2.3. The [repository map](docs/PROJECT-STRUCTURE.md) shows where each part lives. Only the project owner decides when Craft is ready for V1.00.

The planned **Phase 2** focuses on real-world readiness: stability, compatibility, diagnostics, performance, accessibility, and distribution, validated with real applications. It has not started. A phase change or package version does not announce Craft V1.00.

| Artifact | Current version | Source of truth |
| --- | --- | --- |
| Craft CLI and Runtime | `0.1.30` | `internal/cli/cli.go`; reported by `craft version` |
| Optional GUI package in Phase 1 Rev.30 | `0.2.8` | `packages/gui/craft.toml`; recorded in project dependencies |
| Windows Setup | `0.1.30` | `dist/Craft-setup.exe`, built from `packaging/windows/craft.iss` |

## Project map

`cmd/` starts the CLI; `internal/` implements the language, standard library, package system, and native engines. `packages/gui/` is the optional Craft GUI library. `examples/` contains sample applications, `docs/` contains API guides and manual checks, and `roadmap/` contains plans and status. `packaging/` and `scripts/` build the distributable files. `Archive/` holds inactive historical packages. See the [full map](docs/PROJECT-STRUCTURE.md) and [roadmap index](roadmap/README.md).

## What Craft provides

- Static syntax, type, scope, and mutability checking with `craft check`.
- Project execution, test blocks, and source formatting through `craft run`, `craft test`, and `craft fmt`.
- Portable Craft source bundles via `craft build`.
- Standard-library support for JSON, time, filesystem paths, environment values, HTTP, and databases, subject to the revision-specific documentation.
- GitHub package installation and dependency locking through `craft install` and `craft package lock`.

`craft build` produces a Craft source bundle, not a native executable for the application being built. Running a bundle still requires the Craft CLI. Craft is also not a sandbox for untrusted code.

## Install and get started

For published versions, download `Craft-setup.exe` or the portable `craft.exe` from the [latest GitHub Release](https://github.com/spidermeaow/craft/releases/latest). Local build artifacts are not a published release. The installer supports Windows x64 per-user installation, adds the user PATH entry, registers `.craft` files, includes `packages/gui` 0.2.8 and an uninstaller. Open a new terminal after installation.

```powershell
craft version
craft new hello
cd hello
craft check
craft run
craft test
craft fmt
```

After building this source, you should see version `0.1.30`, `Hello World`, and one passing starter test. End users do not need Go: the compiler, interpreter, standard library, database drivers, and project templates are embedded in `craft.exe`.

Release binaries are intentionally not committed to the source repository. To build the CLI yourself, see [Developing Craft](#developing-craft).

> Migration from 0.1.0: `let` is immutable. Use `var` for values that must change. See the [migration note](docs/REV1-MIGRATION.md).

## Example

```craft
func main() {
    defer {
        print("done")
    }

    var numbers: Int[] = [1, 2, 3]
    numbers.append(4)
    for number in numbers {
        print(number)
    }

    try {
        print(numbers[99])
    } catch error {
        print(error.message)
    }
}
```

## CLI commands

| Command | Purpose |
| --- | --- |
| `craft version` | Print the Craft version and execution engine. |
| `craft new <directory>` / `craft init` | Create a project and starter test without overwriting existing files. |
| `craft run [bundle] [-- args...]` | Check and run a project or bundle. |
| `craft check` | Check `src/` and `tests/` without executing them. |
| `craft build [--release]` | Create a `dist/<name>.craftbundle` with dependency sources. |
| `craft install [github.com/owner/repo@v1.2.3]` | Add or restore a GitHub dependency. Supports `--alias`, `--offline`, `--refresh`, and `--recover`. |
| `craft package check` | Validate the dependency graph, imports, cache, and lockfile without downloading. |
| `craft package list` | List resolved packages in deterministic order. |
| `craft package lock` | Write `craft.lock` with paths, versions, and checksums. |
| `craft clean` | Remove only the current project's bundle. |
| `craft fmt [--check] [file.craft]` | Format a project or a single file while retaining comments and AST semantics. |
| `craft test` | Run test blocks without invoking `main`. |
| `craft help` | Show command help. |

## Documentation and examples

[`packages/gui`](packages/gui) รุ่น 0.2.8 ใช้กับ UI Runtime 0.1.30; ดู [คู่มือ package](packages/gui/README.md) และ `examples/gui-package-lab`. Installer รวม package ไว้ที่ `{app}\packages\gui`; โปรเจกต์ภายนอกต้องคัดลอก package หรืออ้าง path ใน dependency แล้วรัน `craft package lock`. `craft gui init` ยังพักอยู่.

Rev.22 UI Input Lab is in `examples/gui-input` and `D:\02-Repository\craft-test\runtime\11-ui-input`; it uses `std.ui` directly. The first text input uses a native Windows EDIT control and has limited styling; see the [guide](docs/REV22-UI-INPUT.md).

Rev.23 UI Layout Lab is in `examples/gui-layout-lab`; it demonstrates sizing, nested split panes, independently scrolling panes and native text inputs. See the [layout guide](docs/REV23-UI-LAYOUT.md) and [manual walkthrough](docs/TRY-REV23.md).

Rev.24 UI Paint Lab is in `examples/gui-paint-lab`; it demonstrates retained Canvas commands for shapes, text, PNG images, and local click events. See the [Canvas guide](docs/REV24-UI-PAINT.md) and [manual walkthrough](docs/TRY-REV24.md).

Rev.13–20 guides describe the earlier package-based examples. In Rev.21, the current GUI examples call `std.ui` directly, and `examples/api-server` calls `std.http` directly. The original package and API server sources are under `Archive/`.

- [GitHub package reference](docs/REV11-PACKAGES.md), [Rev.12 package guide](docs/REV12-PACKAGES.md), [Rev.12 walkthrough](docs/TRY-REV12.md), and `examples/rev11-github`
- [Windows GUI bridge](docs/REV13-GUI-BRIDGE.md), [Rev.13 walkthrough](docs/TRY-REV13.md), and `examples/gui-bridge` (now using `std.ui` directly)
- [Rev.14 UI tree](docs/REV14-UI-ENGINE.md), [walkthrough](docs/TRY-REV14.md), and `examples/gui-layout`
- [Rev.16 responsive UI host](docs/REV16-UI-ENGINE.md), [walkthrough](docs/TRY-REV16.md), and `examples/gui-layout`
- [Rev.17 content-aware UI layout](docs/REV17-UI-LAYOUT.md), [walkthrough](docs/TRY-REV17.md), and `examples/gui-stress`
- [Rev.18 ScrollView](docs/REV18-SCROLLVIEW.md), [walkthrough](docs/TRY-REV18.md), and `examples/gui-stress`
- [Rev.19 ScrollView sizing](docs/REV19-SCROLLVIEW-SIZING.md), [walkthrough](docs/TRY-REV19.md), and `examples/gui-stress`
- [Rev.20 Virtualized List](docs/REV20-VIRTUAL-LIST.md), [walkthrough](docs/TRY-REV20.md), and `examples/gui-stress`
- [Rev.15 renderer and text](docs/REV15-UI-ENGINE.md)
- [Logging, configuration, and errors (Rev.10)](docs/REV10-TOOLING.md)
- [Package manifests, lockfiles, and commands (Rev.8)](docs/REV8-PACKAGES.md) and `examples/rev8-packages`
- [Safe filesystem and path tooling (Rev.9)](docs/REV9-TOOLING.md) and `examples/rev9-tools`
- [Database API and constraints (Rev.5)](docs/REV5-DATABASE.md) and `examples/rev5-database`
- [Modules, function values, and HTTP contracts (Rev.4)](docs/REV4-LANGUAGE.md) and `examples/api-server` (now using `std.http` directly)
- [Timers and multi-task contracts (Rev.3)](docs/REV3-LANGUAGE.md) and `examples/rev3-tasks`
- [Language and standard-library contracts (Rev.2)](docs/REV2-LANGUAGE.md), `examples/rev2-bank`, `examples/rev2-features`, and `examples/rev2-cli`
- [Core language reference](docs/LANGUAGE.md), [Rev.1 walkthrough](docs/TRY-CRAFT.md), and `examples/rev1`
- [VS Code language support](craft-vscode/README.md)
- [Third-party notices](docs/third-party/README.md)

The repository also includes intentional error examples: `examples/type-error`, `examples/runtime-error`, `examples/immutable-error`, and `examples/unhandled-exception`.

## Developing Craft

Developing the CLI requires the Go version declared in `go.mod` (Go 1.26.6 or newer). Building the Windows installer also requires Inno Setup. Node.js and npm are needed only to package the VS Code extensions; end users do not need any of these tools.

```powershell
go build -o dist/craft.exe ./cmd/craft
powershell -ExecutionPolicy Bypass -File scripts/build.ps1 -Installer -Icons -Language
```

The release script disables CGO and produces the Windows amd64 CLI, installer, and `dist/SHA256SUMS.txt`. Pass `-IsccPath 'C:\...\ISCC.exe'` when Inno Setup must be specified explicitly.

For repository verification, run:

```powershell
go test ./...
pwsh -NoProfile -File .\scripts\test-rev8.ps1
go test ./internal/lexer -fuzz=FuzzScan -fuzztime=10s
go test ./internal/parser -fuzz=FuzzParse -fuzztime=10s
```

Build success alone does not certify runtime behavior or the installer. Revision-specific acceptance results and limitations are recorded under `roadmap/phase1/output/`.

## Architecture

`lexer` → `parser/ast` → `resolver/types` → validated call argument ordering → `interpreter` → `runtime/stdlib`

- `internal/resolver`: lexical symbols and mutability.
- `internal/types`: type checking, control-flow outcomes, return safety, and named-argument validation.
- `internal/runtime`: array value semantics and exception values.
- `internal/stdlib`: built-in signatures and implementations separate from the interpreter.
- `internal/interpreter`: function frames, exception propagation, stack traces, and `defer` unwinding.
- `internal/formatter`: comment-preserving formatting with AST comparison.
- `internal/project`, `internal/cli`: project discovery, bundles, packages, and developer commands.

Modules/imports, function values, and a buffered HTTP server were added in Rev.4; database primitives arrived in Rev.5. ORM/migrations, enums, lambdas/closures, event libraries, process/REPL support, a native backend, and a VM remain backlog items. The [editor protocol plan](docs/EDITOR-PROTOCOL.md) documents planned Format Document, Problems, and semantic services that are not enabled yet.
