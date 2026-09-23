# Phase 1 handoff

Historical record for 0.1.0. Current Rev.1 behavior and acceptance status are documented in [phase1-rev1-status.md](../roadmap/phase1/output/phase1-rev1-status.md) and [LANGUAGE.md](LANGUAGE.md).

## Implemented

1. CLI skeleton, embedded starter template, project creation and initialization.
2. Lexer with Unicode source positions, comments, literals and range/operator tokens.
3. Recursive descent statements, Pratt expressions and backend-independent AST.
4. AST Interpreter with variables, functions, recursive calls and print.
5. If/else, while, exclusive-end ranges, return, break and continue.
6. Static checking of names, primitive types, operators, arguments, return paths and main.
7. Located syntax/type/runtime diagnostics with source context and caret.
8. Experimental build bundles, conservative clean, Windows release build script and Inno Setup packaging.
9. Unit, integration, fuzz seeds and executable end-to-end tests for the user to run.
10. Manual acceptance guide and runnable/invalid examples.

## Decisions resolving gaps in the roadmap

- `func main()` is the entry point. The older concept's `app { start {} }` is not implemented.
- Phase 1 uses an Interpreter, not the generated-Go/native backend suggested by the broader concept.
- `craft build` emits a versioned source bundle; standalone user-program executables remain future work.
- `+=` is supported as shown in roadmap examples, alongside the listed assignments and `%=`.
- `let` is mutable, type annotations are required, and Int/Float do not coerce implicitly.
- `0..5` is exclusive of 5. Loops run forwards only; bounds evaluate once.
- `print` accepts multiple values; parameters and ordinary calls remain strictly typed.
- Manifest parsing intentionally supports a limited documented scalar subset.
- The CLI has no external Go module dependencies. Runtime/compiler assets and template are compiled into one binary.

## Validation status

Release compilation is performed to produce the handoff artifacts. Automated tests and runtime/installer acceptance are **not run by the agent**, in accordance with the user's testing preference. The implementation is ready for user acceptance testing; full Phase 1 acceptance remains subject to the checks in `TRY-CRAFT.md`, including a Windows machine without Go.

## Not included

Native user-program executables, bytecode VM, optimization, package registry, HTTP/API framework, database/migrations, DI, generics/structs/interfaces, async/concurrency and custom GC. `craft test`, `fmt`, `add`, `upgrade` and other commands from the broader concept are not Phase 1 CLI commands. Installer and CLI are not code-signed.
