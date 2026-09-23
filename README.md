# Craft 0.1.11 — Phase 1 Rev.11

Craft is a statically typed programming language with a Go-based command-line interface and AST interpreter. It is designed for building small programs and command-line tools while making type and scope errors visible before execution. Craft programs use the `.craft` extension and run through the `craft` command on Windows.

Rev.11 adds public-GitHub source dependencies pinned by tag, version-specific caches, and lockfiles that record the resolved commit and checksum for reproducible builds. See the [GitHub packages guide](docs/REV11-PACKAGES.md), the [Rev.11 walkthrough](docs/TRY-REV11.md), and the [Rev.11 status](roadmap/phase1/output/phase1-rev11-status.md).

## What Craft provides

- Static syntax, type, scope, and mutability checking with `craft check`.
- Project execution, test blocks, and source formatting through `craft run`, `craft test`, and `craft fmt`.
- Portable Craft source bundles via `craft build`.
- Standard-library support for JSON, time, filesystem paths, environment values, HTTP, and databases, subject to the revision-specific documentation.
- GitHub package installation and dependency locking through `craft install` and `craft package lock`.

`craft build` produces a Craft source bundle, not a native executable for the application being built. Running a bundle still requires the Craft CLI. Craft is also not a sandbox for untrusted code.

## Install and get started

Download `Craft-setup.exe` or the portable `craft.exe` from the [latest GitHub Release](https://github.com/spidermeaow/craft/releases/latest). The installer supports Windows x64 per-user installation, adds the user PATH entry, registers `.craft` files, and includes an uninstaller. Open a new terminal after installation.

```powershell
craft version
craft new hello
cd hello
craft check
craft run
craft test
craft fmt
```

You should see version `0.1.11`, `Hello World`, and one passing starter test. End users do not need Go: the compiler, interpreter, standard library, database drivers, and project templates are embedded in `craft.exe`.

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

- [GitHub package reference](docs/REV11-PACKAGES.md), [Rev.11 walkthrough](docs/TRY-REV11.md), and `examples/rev11-github`
- [Logging, configuration, and errors (Rev.10)](docs/REV10-TOOLING.md)
- [Package manifests, lockfiles, and commands (Rev.8)](docs/REV8-PACKAGES.md) and `examples/rev8-packages`
- [Safe filesystem and path tooling (Rev.9)](docs/REV9-TOOLING.md) and `examples/rev9-tools`
- [Database API and constraints (Rev.5)](docs/REV5-DATABASE.md) and `examples/rev5-database`
- [Modules, function values, and HTTP contracts (Rev.4)](docs/REV4-LANGUAGE.md) and `examples/api-server`
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
