# Phase 1 Rev.4 implementation status

วันที่ 19 กันยายน 2026 — Craft 0.1.4 / Language Support0.2.0

สถานะ: **ผ่านการทดลองใช้งาน Rev.4 โดยผู้ใช้** สำหรับชุด `craft-test` บน Windows amd64 เมื่อ 19 กันยายน 2026: `craft check` สำเร็จ, `craft test` ผ่าน 7/7 (exit 0) และ HTTP server เริ่มฟังที่ `127.0.0.1:8080` แล้ว การตรวจรับเพิ่มเติมตาม roadmap ยังแยกไว้ด้านล่าง

## Delivered source

| Milestone | Implementation | Evidence/status |
| --- | --- | --- |
| M0 | Module/function/HTTP/lifecycle/compatibility contracts | [REV4-LANGUAGE](../../../docs/REV4-LANGUAGE.md), source-reviewed |
| M1 | Local versioned modules, exports, scoped identities, typed function values | User-reported passes: imported Craft framework, function values and collections; remaining package diagnostics/compatibility pending |
| M2 | Reclaim completed owned tasks, isolated request scopes, context/deadlines | User-reported passes: ownership exceeds 256 completed tasks, timer accepts a function value; request lifecycle/stress pending |
| M3 | Bounded HTTP transport, Bytes, request/response validation, drain/stop | User-reported passes: immutable Bytes/UTF-8 test and HTTP listener startup; network responses/drain pending |
| M4 | Craft router/hooks/validation/errors and API application | User-reported passes: imported framework/query values and router errors; remaining framework acceptance pending |
| M5 | Module bundles, user guides, editor support and packaging | CLI/VSIX/installer built; user CLI check/test/server-start acceptance passed; bundles/installer/editor checks pending |

Go contains no application route registry, middleware ordering or validation policy. Framework code and app handlers are Craft. New routes/hooks use the existing CLI; no Go edits are required.

## Verification ledger

- Source review and Go formatting completed during implementation.
- `go build ./...` succeeded; release script built `dist/craft.exe`, `dist/Craft-setup.exe`, `dist/craft-language-support-0.2.0.vsix` and refreshed `dist/SHA256SUMS.txt` successfully.
- Go test binaries for `./tests`, `./internal/project` and `./internal/stdlib` compiled with `go test -c` only; this does **not** execute tests.
- No automated tests, Craft runtime demos, HTTP requests, installer execution or VS Code UI tests were run by the agent, per user preference.
- Earlier user-reported Rev.3 demo completed with exit0 on0.1.3. That evidence does not validate changed runtime or Rev.4.
- User-reported Rev.4 acceptance on 19 September 2026 passed for the `craft-test` suite and HTTP listener startup; details below.
- Remaining runtime/HTTP acceptance,10,000-request stress/memory measurements, race checks, portable execution, installer and visual checks remain **pending user testing**.

## User-reported Rev.4 results — 19 September 2026

Environment: PowerShell in `D:\02-Repository\craft-test`; `craft version` reported Craft 0.1.4, Language Phase 1 Rev.4, Target windows-amd64, Compiler Go, Execution Interpreter.

| Command / test | Observed result |
| --- | --- |
| `craft check` (latest run) | `Check succeeded.` |
| REV4 function values and collections | PASS |
| REV4 immutable bytes and UTF-8 validation | PASS |
| REV4 imported Craft framework and query values | PASS |
| REV4 router errors | PASS |
| REV4 task ownership exceeds 256 completed tasks | PASS |
| REV4 timer accepts a function value | PASS |
| basic arithmetic | PASS |
| `craft test` summary | `7 passed, 0 failed, 7 total`; `$LASTEXITCODE` = `0` |
| `craft run` | `HTTP listening on 127.0.0.1:8080` |

The first reported `craft check` emitted `E3001 invalid/duplicate import api in craft-test\@0.1.0`; the subsequent check succeeded. The transcript does not identify what changed between runs, so no compiler fix or root cause is claimed.

The user confirmed that Rev.4 passed their trial. These results establish acceptance of this seven-test suite and server startup; no HTTP response, shutdown exit code, full regression, stress or bundle result was supplied. The agent did not rerun tests, in accordance with the user's preference.

See [TRY-REV4](../../../docs/TRY-REV4.md) for additional commands and expected results. The [roadmap](../craft-phase1-rev4.md) records this completed user trial separately from its broader acceptance checklist.

## Limits and deferred scope

Local packages only, noncapturing named functions, frozen route configuration with value copies per request, plain HTTP buffered bodies, cooperative cancellation. OS filesystem calls can delay cleanup beyond drain timeout. No database/shared mutable store, dynamic routes, TLS configuration, streaming, closures, registry, LSP or native app compilation. These match the deferred roadmap scope; no production throughput claim is made.
