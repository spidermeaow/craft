# Craft 0.1.45 candidate (DB package)

Craft is a statically typed language with a Go-based CLI and interpreter. It checks source before execution and currently targets Windows amd64. Projects can build CLI tools, HTTP services, and native Windows UI applications using the standard library and optional first-party packages.

**Current source version:** 0.1.45 candidate. Rev.41 added multiline syntax and explicit project modules; Rev.43 added six Windows standard-library modules; Rev.44 added `packages/api`; Rev.45 added `packages/db`. The bundled `packages/gui`, `packages/api`, and `packages/db` use the same version as Craft. See the [roadmap index](roadmap/README.md) for earlier revisions and their acceptance status.

Craft remains in [Phase 1](roadmap/craft-phases.md). The [Craft Sheet assessment](roadmap/phase1/craft-phase1-rev42.md) found that the [UI Engine Final gate](roadmap/phase1/craft-phase1-rev40-final-gate.md) has not passed; that work is paused. Rev.45 is an implemented candidate awaiting user acceptance. Only the project owner decides when Craft is ready for V1.00.

| Artifact | Current version | Source of truth |
| --- | --- | --- |
| Craft CLI and Runtime | `0.1.45` candidate | `internal/toolchain/version.go`; reported by `craft version` |
| Optional GUI package | `0.1.45` candidate | `packages/gui/craft.toml`; tied to the Craft toolchain |
| Optional API package | `0.1.45` candidate | `packages/api/craft.toml`; tied to the Craft toolchain |
| Optional DB package | `0.1.45` candidate | `packages/db/craft.toml`; tied to the Craft toolchain |
| Windows Setup | `0.1.45` candidate | `packaging/windows/craft.iss`; local build is in ignored `dist/` |

## Project map

`cmd/` starts the CLI; `internal/` implements the language, standard library, package system, and native engines. `packages/gui/`, `packages/api/`, and `packages/db/` are optional first-party Craft libraries. `examples/` contains sample applications; `docs/` contains API guides; `roadmap/` records plans and acceptance status. `packaging/` and `scripts/` build distributable files. `Archive/` holds inactive historical packages. See the [full repository map](docs/PROJECT-STRUCTURE.md).

## What Craft provides

- Static syntax, type, scope, and mutability checking with `craft check`.
- Project execution, test blocks, and source formatting through `craft run`, `craft test`, and `craft fmt`.
- Portable Craft source bundles via `craft build`.
- Standard-library support for JSON, time, filesystem paths, environment values, HTTP, databases, encoding, cryptography, URLs, processes, regex, and Windows OS integration. See the [Rev.43 standard-library guide](docs/REV43-WINDOWS-STDLIB.md) and [database contract](docs/REV5-DATABASE.md).
- GitHub package installation and dependency locking through `craft install` and `craft package lock`.
- Optional [`gui`](packages/gui/README.md), [`api`](packages/api/README.md), and [`db`](packages/db/README.md) packages, built from public `std.ui`, `std.http`, and `std.db` APIs respectively.

`craft build` produces a Craft source bundle, not a standalone application executable. Running the bundle requires Craft. The UI Runtime currently targets Windows; the [Final spreadsheet UI criterion](roadmap/craft-ui-final-excel-acceptance.md) remains open. Craft is not a sandbox for untrusted code.

## Install and get started

For published versions, download `Craft-setup.exe` or the portable `craft.exe` from the [latest GitHub Release](https://github.com/spidermeaow/craft/releases/latest). Check the release version: the latest published release may differ from the 0.1.45 source in this repository. Local build artifacts in ignored `dist/` are not GitHub Releases. The Windows x64 installer uses per-user installation, adds Craft to the user PATH, registers `.craft` files, includes matching `packages/gui`, `packages/api`, and `packages/db` sources, and provides an uninstaller. Open a new terminal after installation.

```powershell
craft version
craft new hello
cd hello
craft check
craft run
craft test
craft fmt
```

After building this source, `craft version` should report `0.1.45`. End users do not need Go: the CLI, interpreter, standard library, database drivers, and project templates are in `craft.exe`. The optional package sources are separate files, included with the installer; a portable `craft.exe` alone does not contain them.

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

## First-party packages

These packages are source libraries installed beside the Craft CLI. Add a path dependency relative to your project's `craft.toml`, run `craft package lock`, and import the package in every source file that uses it. Their versions follow the Craft toolchain; do not set a separate dependency version.

| Package | Provides | Guide and example |
| --- | --- | --- |
| `gui` | Windows UI composition and reusable controls on `std.ui` | [Guide](packages/gui/README.md), [GUI package lab](examples/gui-package-lab) |
| `api` | HTTP routes, middleware, responses, and validation on `std.http` | [Guide](packages/api/README.md), [API server example](examples/api-server-package) |
| `db` | Strict `DbRows`/`DbResult` helpers on `std.db` | [Guide](packages/db/README.md), [DB example](examples/db-package) |

For a repository checkout, `examples/db-package/craft.toml` demonstrates `db = { path = "../../packages/db" }` followed by `import package db "db"`. For an installed copy, point the path to the package source under the Craft installation directory or copy that package into your project. `packages/db` does not provide an ORM or migration/seed runner; applications own SQL and connection/transaction lifetimes.

## Documentation and examples

- [Rev.41 language modules](docs/REV41-LANGUAGE-MODULES.md) and [`examples/rev41-language`](examples/rev41-language)
- [Rev.43 Windows standard library](docs/REV43-WINDOWS-STDLIB.md) and [`examples/rev43-stdlib`](examples/rev43-stdlib)
- [Rev.44 API package](roadmap/phase1/craft-phase1-rev44-api-package.md) and [Rev.45 DB package](roadmap/phase1/craft-phase1-rev45-db-package.md)
- [Current GUI rendering](docs/REV39-UI-RENDERING.md), [window icons](docs/REV40-WINDOW-ICON.md), and [UI Final assessment](roadmap/craft-ui-final-excel-acceptance.md)

Earlier UI labs remain available: [input](docs/REV22-UI-INPUT.md), [layout](docs/REV23-UI-LAYOUT.md), and [Canvas](docs/REV24-UI-PAINT.md). Rev.13–20 guides and `Archive/` describe older APIs; check the current package guides before copying those examples.

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

Modules/imports and function values are part of the language; Rev.41 added explicit project modules and multiline formatting. `std.http` and `std.db` provide the transport and database primitives used by the optional `api` and `db` packages. Rev.43 added `std.process` and other Windows-facing library modules. ORM, migration/seed automation, standalone application binaries, other OS GUI backends, and a VM remain outside the current implementation. The [editor protocol plan](docs/EDITOR-PROTOCOL.md) records editor services that are not enabled yet.
