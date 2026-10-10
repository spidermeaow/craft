# Craft 0.1.41 candidate: multiline syntax and project modules

## Multiline grammar

| Form | Meaning |
| --- | --- |
| Newline inside `()` or `[]` | Whitespace; function calls/declarations, arrays, indexing and nested expressions can span lines. |
| Newline after an infix/assignment operator, comma, dot, colon, `if`, `while`, `in`, `throw` or `assert` | The expression continues on the next line. No backslash is needed. |
| Newline before an operator outside delimiters | A statement boundary. To break before an operator, enclose the full expression in `()`. |
| Newline after a complete expression outside delimiters | A statement boundary, except one newline between a member name and its opening call parenthesis: `fields.append` then `(` continues that call. A blank line still separates statements. `return` followed by newline remains a return without a value. |
| Final comma before `)`, `]` or a type's closing `>` | Accepted in function parameters/arguments, arrays and generic function/Map types. Map literals already accepted this form. |
| Newline inside `{}` blocks | Separates statements or struct fields as before; braces do not generally suppress statement boundaries. |

The lexer keeps line and column positions. Parser, type checker and interpreter receive the same AST for equivalent single-line and multiline expressions. `craft fmt` preserves a multiline layout and checks that the formatted AST has the same meaning. This is a candidate implementation; execution and compatibility results are pending user testing.

## Project modules

Declare a module at the beginning of each participating source file. Files declaring the same module name share private declarations regardless of directory. A source file without a declaration remains in the legacy project scope.

```craft
module pages.excel
export func render(): String { return "Sheet" }
```

```craft
import module excel "pages.excel"
import package gui "gui"
```

`import gui "gui"` still imports a package for existing source. Imports are file-local. Declarations without `export` are accessible only within their declared module. Module imports are resolved inside the same project/package and cycles are rejected. The source bundle retains module declarations in its files so it can be loaded with the same semantics.

## Manual acceptance

1. With the 0.1.41 CLI, in `examples/rev41-language`, run `craft check`, `craft fmt`, `craft fmt --check`, and `craft run`. Expected output: `12`; the second formatting pass should report no changes.
2. In `D:\02-Repository\app`, run `craft check` and `craft run`. The new **Rev 41 Language** page opens first. Confirm its preview shows `60`, click **Run calculation** and expect `Module returned 60`, then click **Check predicate** and expect `Predicate passed`. Its helper and exported functions live in two files declaring `showcase.language`. The sidebar footer comes from a separate `showcase.labels` module.
3. Check the existing Excel, Contacts, Rendering Lab and Window Icon pages for regressions after visiting the language page.
4. Check existing GUI, HTTP, Database and standard library projects with `craft check`; then run the workflows you normally use. Try `craft build` and run its bundle for an app with a project module.
5. Run the Go regression suite if you want automated coverage. This implementation added regression cases but did not execute them because you perform testing personally.

Craft 0.1.41 is not accepted until the above results and any discovered compatibility issues are reviewed.
