# Rev.1 foundation decisions and Rev.2 boundary

> Historical design for 0.1.1, not the current feature inventory. Struct, Map,
> Optional and structured JSON are implemented in [Rev.2 contracts](REV2-LANGUAGE.md).
> Named task/timer callbacks are added in [Rev.3 contracts](REV3-LANGUAGE.md).
> General function values, lambdas and events below remain future designs.

The roadmap describes both required Rev.1 behavior and future-facing design. Its Definition of Done requires safety, arrays, strings, exceptions, diagnostics and developer tools. Its final sequence places Map/Optional/Module/Lambda/Callback/Event Library in Rev.2. This implementation follows that boundary, implements defer and named arguments now, and records the remaining foundations below. These drafts are **not executable syntax in 0.1.1**.

## Current architecture

- AST is independent from runtime values. Symbols and mutability live in `internal/resolver`.
- Type checking computes control-flow outcomes (fallthrough, return, throw, break, continue). Named arguments are validated and mapped to parameter indices while preserving source evaluation order.
- `internal/runtime` owns array and exception values. The interpreter owns function frames and cleanup stacks.
- `internal/stdlib` owns built-in signatures and implementations. String and Array methods use this layer; `std.math`, `std.fs`, `std.time`, `std.json`, `std.env` are reserved built-in namespaces.
- These namespaces are not a user module loader. There is no filesystem import or dependency resolution yet.

## Struct foundation (design)

Initial proposal: nominal types identified by their declaring module and type name; fields have explicit types and `let`/`var` mutability; construction uses named arguments with complete field coverage; unknown/duplicate/missing fields are compile errors. No inheritance, user methods, custom constructors or defaults in the first implementation.

Structs should retain the value-copy semantics used for arrays. A `let` binding must be deeply immutable. A field write requires both a mutable root binding and a `var` field at every traversed struct boundary. A `let` field remains immutable even when its parent binding is `var`.

Work before execution support: add declaration/type tables, constructor and field resolution, recursive-value layout checks, field-specific diagnostics, runtime structured values, formatter support and tests. Field syntax is deliberately not accepted until all layers agree on these rules.

## Map foundation (design)

Begin with `Map<String, T>`, rather than exposing a general generics implementation. Keys use exact String equality. Maps should be value types with deep copying and let/var rules matching arrays. Deterministic insertion-order iteration is proposed. Missing-key behavior must be settled alongside Optional: prefer a safe lookup returning `T?`, with a separate required lookup that throws. Avoid pretending an absent value is zero.

## Module/import foundation (design)

Proposed file relationship: a directory under `src/` is one module; files in it must agree on the module declaration. A qualified module name maps to that directory, and the project root module contains main. Same-module private functions are available across its files. Only explicitly public symbols may cross module boundaries. Imported names must not silently replace local declarations.

Compilation plan:

1. Discover files and parse declarations without running code.
2. Build a table of modules, public symbols and import edges.
3. Resolve exact imports inside the project; reject missing/ambiguous names and duplicate exports.
4. Detect circular imports with a DFS visiting stack and report the cycle paths and import locations.
5. Process modules in deterministic topological order, then type-check bodies and lower calls.

There will be no module-level executable initialization in the first version. This avoids order-dependent side effects. Rev.2 must add compile/runtime/CLI cases for cross-module calls, private access, cycles and multiple-file modules before advertising import support. Current Rev.1 files still share a single project-wide namespace.

## Other deferred features

| Feature | Boundary |
| --- | --- |
| Enum | Add after nominal Struct/type resolution is stable |
| Optional/null safety | Design alongside flow-sensitive narrowing and Map lookup |
| Default parameters | Add with evaluation-order and dependency rules; not silently filled now |
| Lambda / first-class functions | Require concrete function signatures, capture lifetime and mutability rules |
| Callback / std.events | Depend on first-class functions; no event keyword or partial Event Bus now |
| REPL | Optional roadmap item; deferred until incremental symbol/type state is designed |
| Full std.json parse/stringify | Requires structured values/Map; Rev.1 provides isValid and quote |

The filesystem foundation currently operates on complete text files and does not expose open file handles. Defer is available for ordinary cleanup actions now; advanced resource ownership, exception chaining and finally remain out of scope.

## Scope status

Rev.1 executable core: implemented. Module/Struct/Map foundations: documented designs. Full module loading, circular-import checking and Event Bus execution: Rev.2 work, not marked as implemented or tested. No API, database, package manager, native backend or bytecode VM is added by this revision.
