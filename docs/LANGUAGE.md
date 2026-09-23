# Craft 0.1.2 — Language reference

Rev.2 adds Optional, Struct, Map, Console input and expanded libraries. See [Rev.2 contracts](REV2-LANGUAGE.md) for complete new syntax, signatures, limits and migration notes.

## Program, files and scope

UTF-8 `.craft` files under `src/` share one function namespace. Forward function calls and cross-file calls are supported. Executable projects need exactly one `func main()` with no parameters and a `Void` result. `craft test` can operate without main and never invokes it. Top-level declarations are `func` and `test` only.

Blocks use `{}`. Simple statements are separated by newlines or semicolons. Inside `()` and `[]`, newlines are ignored. Multiple simple statements on a line require semicolons. `//` starts a comment. No type inference: variables and parameters require annotations; functions returning a value require a result type. Omitted result type means Void.

Names are visible from their declaration through the end of their block. Duplicate names in the same scope are errors. Nested blocks may shadow outer names; the function body shares its parameter scope, so it cannot redeclare a parameter. A use before a later shadowing declaration resolves to an existing outer binding, if one exists. Otherwise it is an unknown-name error. The namespace `std` is reserved. Parameters, loop variables and catch bindings are immutable.

## Types and mutability

| Type | Meaning |
| --- | --- |
| `Int` | Signed 64-bit decimal integer |
| `Float` | Finite IEEE 754 64-bit float, decimal/exponent notation |
| `Bool` | true / false |
| `String` | Text in double quotes |
| `Exception` | Immutable error value |
| `T[]` | Homogeneous array, including nested arrays |
| `Void` | No result; not a variable, parameter or array element type |

`let` is immutable; `var` permits reassignment and compound assignment without changing type. There are no implicit Int/Float conversions. Non-Void functions must return a matching value or throw on every path. Loops conservatively may execute zero times; a return only inside a loop does not prove a function always returns. Both try and catch outcomes are checked. Unreachable code is still type checked.

```craft
let owner: String = "Craft"
var count: Int = 0
count += 1
```

## Arrays and ownership

```craft
var numbers: Int[] = [1, 2, 3]
numbers.append(4)
numbers[0] += 10
let removed: Int = numbers.remove(1)
print(numbers.length)
```

Arrays use **deep value-copy semantics** at variable initialization, assignment, arguments, returns and insertion into arrays. This also applies to nested arrays. There are no mutable reference aliases in Rev.1.

```craft
let original: Int[] = [1, 2]
var copy: Int[] = original
copy[0] = 99  // original[0] is still 1
```

- A `let` array cannot be reassigned, indexed for writing, appended to or removed from, including nested elements.
- A `var` array can be mutated. A parameter is immutable: make a local `var` copy and return it to provide a changed array.
- `.append(value)` returns Void. `.remove(index)` removes and returns the indexed value.
- Indices are zero-based Int values. Negative or out-of-range indices throw a catchable RuntimeError.
- Empty literals get their element type from a declared variable, parameter, return type or assignment context. An untyped `print([])` is rejected.
- `.length` is a read-only Int. Array equality is not defined yet.
- Array iteration uses a deep snapshot of the array at loop entry, so edits do not alter the current iteration sequence.

## Strings

Strings support Go-style escapes including `\n`, `\t`, `\\`, `\"` and `\uXXXX`. No interpolation, multiline literals or string indexing in Rev.1.

| Operation | Result |
| --- | --- |
| `value.length` | Int: number of Unicode code points, not bytes or grapheme clusters |
| `value.upper()`, `.lower()`, `.trim()` | New String; original unchanged |
| `.contains(text)`, `.startsWith(text)`, `.endsWith(text)` | Bool |
| `.split(separator)` | String[]; empty separator splits into Unicode code points |
| `left + right` | Concatenated String |

String comparison is lexical UTF-8 ordering. Arrays print as `[1, 2]`; `print` accepts any number of non-Void values, adds spaces between them and ends with a newline.

## Operators and control flow

Precedence from low to high: assignment `= += -= *= /= %=`, `||`, `&&`, equality `== !=`, comparison `< > <= >=`, addition `+ -`, multiplication `* / %`, unary `! - +`, calls/member/index access. Assignments are right associative. Other binary operators are left associative. `&&` and `||` short circuit. Integer division truncates toward zero; remainder has the dividend's sign. Overflow and division by zero throw RuntimeError.

```craft
if condition { ... } else if other { ... } else { ... }
while condition { ... }
for number in 0..5 { ... }
for number in numbers { ... }
```

Conditions are Bool. Range bounds are Int, evaluate once, and exclude the upper bound. Ranges advance by one; start >= end means no iterations. `break` and `continue` target the nearest loop. `return` exits the current function.

## Functions and named arguments

```craft
func subtract(a: Int, b: Int): Int {
    return a - b
}
```

Both `subtract(10, 2)` and `subtract(b: 2, a: 10)` are supported. All arguments must be positional or all named. Missing, duplicate or unknown names are errors. Arguments evaluate in the order written, then bind to parameters. Built-ins and methods currently accept positional arguments only. No overloads, defaults, lambdas or first-class function values yet.

## Exceptions

```craft
func divide(a: Int, b: Int): Int {
    if b == 0 {
        throw Exception("cannot divide by zero")
    }
    return a / b
}

func main() {
    try {
        print(divide(1, 0))
    } catch error {
        print(error.message, error.code)
    }
}
```

`try` requires a catch binding and block. Catch handles explicit Exception and runtime faults, including array bounds, numeric overflow, standard-library failures and assertions. Syntax/type/project errors happen before execution and cannot be caught. Ctrl+C cancellation is not catchable.

Error properties are immutable: `.message`, `.type`, `.code`, `.file` are String; `.line`, `.column` are Int. A new unthrown Exception has an empty file and zero location. Throwing it attaches an origin and stack. Rethrowing `throw error` preserves that origin and trace. An uncaught error prints its source context and stack frames, then exits with code 1.

No typed catch, custom exception hierarchy, filter, throws signature, chaining or finally.

## defer

`defer { ... }` registers a block on the current function/test frame. Blocks run in LIFO order when that frame finishes normally, returns or unwinds through an exception. They do not run at the end of the lexical block where they were registered.

Deferred blocks capture the bindings visible at registration time, observe later changes to existing `var` cells, and cannot see later declarations. Return values are evaluated and copied **before** deferred blocks run. A defer may call functions, throw or use its own loops and try/catch. It cannot return from its enclosing function, break/continue an outer loop, or register another defer directly inside its body.

All registered cleanups are attempted even if one fails. An existing propagating exception takes precedence; otherwise the first cleanup exception in execution order replaces a normal return. This is a minimal policy without exception chaining. Deferred cleanup is not guaranteed on process termination or Ctrl+C cancellation.

## Tests and formatting

```craft
test "addition works" {
    assert 2 + 3 == 5
}
```

`craft test` discovers `.craft` files in `src/` and `tests/`, checks all definitions, then executes each test in a fresh interpreter in sorted file/source order. It does not call main. Names must be unique project-wide. Failures report source context and traces, and subsequent tests continue. Zero tests reports `0 passed, 0 failed, 0 total` with success. `return` cannot exit a test body; defer is supported. External file/environment effects are not isolated between tests.

`assert BoolExpression` throws AssertionError (`E3003`) if false and may also be used in functions.

`craft fmt` formats both directories; `craft fmt path.craft` formats one file without requiring a project. Comments are retained, semicolons become line breaks, and indentation uses four spaces. Before writing, formatter parses the output and compares ASTs excluding source locations. All input files are formatted/validated before any write. Each changed file is replaced through a temporary file. `--check` never writes; exit 1 means formatting is needed. A later filesystem write failure may leave earlier files formatted.

## Standard library foundation

These reserved built-in names are directly available without imports; they are not user modules:

| Function | Signature / behavior |
| --- | --- |
| `print(...)` | Variadic non-Void values → Void |
| `Exception(message)` | String → Exception |
| `std.math.abs(value)`, `.sqrt(value)` | Float → Float; negative sqrt throws |
| `std.math.min(a, b)`, `.max(a, b)` | Float, Float → Float |
| `std.fs.readText(path)` | String → String; regular UTF-8 text file, 16 MiB limit |
| `std.fs.writeText(path, text)` | String, String → Void; creates or overwrites a file |
| `std.fs.writeTextAtomic(path, text)` | String, String → Void; writes a same-directory temporary file then replaces the destination after close; best-effort atomicity of the target filesystem |
| `std.fs.readBytes(path)` | String → Bytes; regular file, 16 MiB limit; does not decode text |
| `std.fs.writeBytesAtomic(path, bytes)` | String, Bytes → Void; same replacement contract as `writeTextAtomic` |
| `std.fs.exists(path)` | String → Bool; access errors throw |
| `std.path.clean(path)` | String → String; lexical normalization only; does not resolve symlinks or require the path to exist |
| `std.path.absolute(path)` | String → String; resolves relative to the process working directory |
| `std.path.relative(base, target)` | String, String → String; fails when the OS cannot express a relative path, such as incompatible Windows volumes |
| `std.time.now()` | UTC timestamp String in RFC3339 format |
| `std.time.sleep(milliseconds)` | Int → Void; accepts 0–86400000, cancellation aware |
| `std.json.isValid(text)` | String → Bool |
| `std.json.quote(text)` | String → JSON-escaped String |
| `std.env.get(name)` | String → String; absent value is empty String |

String and array operations live in the same standard-library layer through instance methods. Filesystem paths are relative to the process working directory, not automatically the project root. Structured JSON, Map and Struct are available in Rev.2 (see the linked contracts); events still await callbacks.

Filesystem errors identify the operation and requested path, but do not include arbitrary operating-system error text. `readText` requires regular UTF-8 files; `readBytes` permits any byte sequence. All complete-file reads are limited to 16 MiB. `writeText` remains a compatibility API and may overwrite directly; tools that must preserve a previous destination on a write failure should use `writeTextAtomic`. Atomic replacement is best effort: it depends on the destination filesystem and does not claim multi-file transaction or crash-durability guarantees. `removeFile` remains non-recursive and accepts regular files only. See [Rev.9 tooling primitives](REV9-TOOLING.md).

## Diagnostics and limits

| Error code | Meaning |
| --- | --- |
| E1001 | Syntax |
| E2001 | Unknown variable/function |
| E2002 | Type/signature/scope/control-flow error |
| E2003 | Immutable mutation |
| E3001 | Runtime fault |
| E3002 | Explicit Exception |
| E3003 | Assertion failure |
| E4001 | Invalid/missing entry point |
| E4002 | Project/manifest/bundle/filesystem error |
| E5001 | Invalid CLI usage |

Exit codes: 0 success; 1 runtime/test failure (also formatting needed under fmt --check); 2 syntax; 3 type; 4 project, including invalid entry point; 5 usage.

Limits: 256 function frames, 256 parser nesting levels, 512 checked expression levels, 2 MiB per source file, 16 MiB and 1024 files per project. Source symlinks are rejected. This interpreter is not a sandbox and does not enforce a total heap or execution-time budget.

## Manifest and bundles

The manifest still uses `name`, `version`, `edition = "2026"`, `entry = "main"`. Only these scalar string keys are supported; arrays/tables and dependencies are not. Entry defaults to main. The manifest `version` is the application's version, not the compiler version.

Rev.2 builds bundle format 2 with `language = "0.1.2"` and can also read format 2 language 0.1.1 in the JSON bundle metadata. Legacy format 1 is rejected on execution with a migration hint, but can be replaced by build or removed by clean. Migrate original source and rebuild. Bundle execution still requires Craft and does static checking again.
