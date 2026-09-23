# Rev.3 level-B editor protocol assessment

Status: design only. The declarative extension is independently deliverable.
Do not activate these features until the protocol and user acceptance exist.

## Format Document

Proposed command: `craft editor format --stdin --path <workspace-relative-file>`.
Read UTF-8 buffer text (maximum 2 MiB), call the existing pure
`formatter.Format(path, text)` function, return versioned JSON containing the
formatted text or diagnostics, and never write the file. The extension should
capture document version before launch and discard stale output. Use argv
without a shell. Support cancellation and no-change results explicitly.

## Diagnostics

Proposed `craft editor check --json --stdin` accepts a versioned document-overlay
request with root and open buffers. Load other project files through existing
project limits; do not invoke main or tests. Return diagnostic code, severity,
message, URI and a zero-based UTF-16 range. Existing diagnostics use one-based
rune positions, so the adapter must convert from original text, account for
CRLF and surrogate pairs, and preserve each original source for conversion.
Include protocol version and document versions in the response; publish only
results whose documents still match. Avoid parsing human-readable CLI output.

## Execution and ownership

Use an explicit configured CLI path or resolved executable, one project root
per multi-root folder, a 250 ms edit debounce, and cancellation on superseded
requests. Restricted workspaces retain declarative highlighting only; executable
invocation requires Workspace Trust. Missing CLI produces one actionable message
and must not repeatedly prompt or alter PATH/settings. Limit response size and
separate machine JSON on stdout from debug output on stderr.

## Semantic services

Reuse lexer/parser, resolver and type checker. Expose a separate analysis result
mapping declarations, parameter/field uses, lexical bindings and type references
to source ranges instead of letting the grammar guess names. Start with semantic
tokens and hover; then use symbol identities for completion/definition. Preserve
TextMate fallback while analysis is pending or syntax is incomplete. An LSP
server is optional after this result model is stable.

Acceptance must cover unsaved buffers, changes during formatting, Unicode/emoji,
CRLF, absent CLI, two roots, cancellation, invalid JSON, protocol mismatch and
untrusted workspaces. This design creates no new CLI commands by itself.
