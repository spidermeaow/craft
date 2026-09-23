# Phase 1 Rev.7 status

Targeted acceptance and the later Rev.8 full regression completed with exit code 0 on 20 September 2026. Rev.7 P0 is closed by the Rev.8 acceptance run.

## Changes prepared

- HTTP transport clears its read deadline immediately after request-body consumption. The read deadline no longer remains active while a completed request handler waits on database work.
- `TestRev7HTTPPoolWaitCancellation` uses read/write budgets longer than request/database timeout, so it does not test two deadlines at once.
- `TestRev7RequestDeadlineDoesNotUseReadDeadline`, `TestRev7BodyDeadlineSkipsDispatch` and `TestRev7CompletedBodyDoesNotExpireHandler` isolate HTTP deadline contracts.
- `TestRev7CraftDeadlineIsolation` runs a Craft application against MySQL, PostgreSQL and SQL Server in separate DB-timeout and request-timeout modes, then confirms pool quiescence and a subsequent ping. `TestRev7CraftDatabaseLifecycle` separately covers CRUD, load, transaction rollback, disconnect, in-flight SQL, restart and child-task shutdown with independent-observer checks.
- `scripts/test-rev7.ps1` loads local encrypted test credentials temporarily and restores the process environment afterwards.
- Rev.7 integration output uses `REV7 trace` lines for driver, selected deadline, HTTP result, elapsed time and pool observation; it intentionally excludes DSNs, SQL and request values.

## Acceptance result

The authorized command `pwsh -NoProfile -File .\scripts\test-rev7.ps1` completed with exit code 0 and no driver skipped. It passed `TestRev7CraftDatabaseLifecycle` for MySQL 8.4.11, PostgreSQL 18.6 and SQL Server Express 17.0.1135.8, including CRUD, concurrent batches, rollback, in-flight observer barrier, disconnect, timeout cleanup, restart and child-task shutdown. It also passed `TestRev7CraftDeadlineIsolation` in DB-timeout and request-timeout modes for all three drivers, `TestRev7HTTPPoolWaitCancellation` in all four modes for all three drivers, HTTP deadline isolation and diagnostics fixtures.

The separate full regression was run repeatedly through `scripts/test-rev8.ps1` and completed with exit code 0 after the final changes. Race detection also completed with exit code 0, and the HTTP stress fixture completed 10,000 real requests.

The in-flight SQL cancellation and child-task transaction shutdown fixtures passed in at least three lifecycle runs per driver. MySQL can keep a delayed statement visible until its configured operation timeout even after the caller disconnects; the caller, pool and subsequent operation still recover within contract. This driver limitation is documented rather than hidden by a shorter assertion.

Rev.8 rebuilt the release artifacts, upgraded the existing per-user installation with `Craft-setup.exe`, and ran the installed 0.1.8 CLI successfully. SQL Server TCP/TLS is not claimed because the local SQL Express instance exposes Shared Memory and port 1433 was closed. Ambiguous-commit fault injection and platform-specific OS-handle telemetry remain separate deployment certification scenarios; Craft performs no automatic write retry.
