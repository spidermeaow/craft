# Phase 1 Rev.6 status

Implementation prepared for user testing (2026-09-20); acceptance pending. No Rev.6 test has been run by the agent. Version 0.1.6 is an implementation build, not a stability certification.

## M0: frozen P0 inventory

Regression command: `go test ./... -count=1 -timeout 600s` with the three Rev.5 test DSNs loaded. This includes parser, checker, formatter, modules/bundles, CLI, Rev.1–Rev.5, tasks/defer, HTTP transport and DB ownership.

New P0: `TestRev6*` in `tests` and `internal/stdlib`: empty DSN/typed redaction, HTTP CRUD on three drivers, independent read-back, concurrent unique rows, rollback on exception, pool bounds/quiescence, timeout/cancellation, overload and shutdown cleanup. Missing scenarios remain pending, not PASS.

Configuration before execution: loopback ephemeral ports; database `craft_test` only; unique `craft_rev6_` fixture names; HTTP concurrency 8 / DB pool 2; HTTP deadline 3s / DB deadline 5s; client 10s; shutdown drain 2s. Concurrent workload: warmup then three batches of 12 requests at concurrency 4. SQL Server Shared Memory only unless separately configured. Do not modify DB server configuration.

## Evidence ledger

| Milestone | Implementation | Execution evidence |
| --- | --- | --- |
| M0 baseline/contracts | Inventory recorded above | Regression pending |
| M1 Craft application | Direct std.http/std.db example | All three drivers pending |
| M2 lifecycle | CRUD, load, pool-wait, SQL deadlines/activity barrier disconnect, pre/post-commit disconnect, child-task shutdown and restart fixtures; forced drain reports its deadline | Pending |
| M3 diagnostics | Empty DSN and typed safe connection causes | Pending |
| M4 handoff | TRY-REV6, run/test scripts, CLI/installer/checksums | Build passed; runtime/installer execution pending |

## Build evidence (not test results)

- `go build ./...`: exit 0.
- `go test -c -o build/rev6-tests.test.exe ./tests`: compile only, exit 0; no test executable launched.
- `go test -c -o build/rev6-stdlib.test.exe ./internal/stdlib`: compile only, exit 0.
- `pwsh -NoProfile -File scripts/build.ps1 -Installer`: exit 0; generated `dist/craft.exe`, `dist/Craft-setup.exe`, VSIX and `dist/SHA256SUMS.txt`. Installer was not launched.
- No Craft program, HTTP listener, database query or automated test was executed by the agent for Rev.6. Source examples and embedded fixtures still need the user's compiler/runtime acceptance.

## Compatibility and remaining evidence

- Empty DSN now raises `DbConfigError` before driver defaults can apply. Nonempty connection failures preserve `DbConnectionError`, with sanitized typed causes.
- Explicit stop that exceeds its drain deadline now returns an error after waiting for cleanup; graceful explicit stop still succeeds. Parent cancellation remains discoverable through Go `errors.Is`.
- Bundles emitted as language 0.1.6; Rev.4/Rev.5 module bundles remain accepted. Ownership and transaction-lifetime semantics unchanged.
- In-flight cancellation fixtures require visibility of own sessions in DB activity views; unavailable permission is a failed prerequisite, not PASS. No permission grants or server reconfiguration are performed.
- P0 execution remains pending for all drivers and full regression. Ambiguous commit is documented as unknown, with no automatic write retry; fault injection to force that network edge case remains pending. P1 bundle deployment, resource trend/OS handle/race checks, Ctrl+C and installer execution remain pending.
- Run `pwsh -File scripts/test-rev6.ps1`, then `pwsh -File scripts/test-rev6.ps1 -Regression`. Expected exit 0 with no credential-related skips. Manual run: `pwsh -File scripts/run-rev6.ps1 -Driver mysql` (also postgres/sqlserver).

Rev.5 results remain historical evidence only. Do not mark Rev.6 stable or complete until required scenarios pass. Record command, exit code, server versions, transport and sanitized output here after user testing. Race, OS handles, installer execution, Ctrl+C and SQL Server TCP/TLS are separately pending.
