# Craft 0.1.5 — Database foundations

Rev.5 adds `std.db` primitives directly to the runtime. No manifest dependency or import is required. An API framework is not required and was not expanded for this revision. Database servers are installed separately; their Go drivers ship inside `craft.exe`.

## Drivers and SQL

| Craft driver | Embedded Go module | Tested local server | Parameters |
| --- | --- | --- | --- |
| `mysql` | `github.com/go-sql-driver/mysql` v1.10.1 | MySQL 8.4.11 | `?` |
| `postgres` | `github.com/jackc/pgx/v5` v5.11.0 through database/sql | PostgreSQL 18.6 | `$1`, `$2`, ... |
| `sqlserver` | `github.com/microsoft/go-mssqldb` v1.11.0 | SQL Server Express 17.0.1135.8 | `@p1`, `@p2`, ... |

Versions above are tested versions, not a promise of compatibility with every server release. Dependencies are pinned in go.mod/go.sum; licenses ship in `docs/third-party/`. MySQL uses MPL-2.0, pgx uses MIT, go-mssqldb uses BSD-3-Clause. Read upstream connection syntax at [MySQL](https://github.com/go-sql-driver/mysql#dsn-data-source-name), [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn#ParseConfig), and [SQL Server](https://github.com/microsoft/go-mssqldb#connection-parameters-and-dsn).

SQL remains database-specific. Parameters are sent separately to the driver; Craft does not interpolate or translate SQL. Identifiers such as table names cannot be parameters: use fixed application-owned SQL. No ORM, schema migrations, automatic model binding, retry policy, nested transactions or query builder.

## Connecting

```craft
func main() {
    var config: DbConfig = std.db.config()
    config.maxOpen = 8
    config.timeoutMs = 5000
    let db: DbConnection = std.db.connect("postgres", std.env.get("CRAFT_DB_DSN"), config)
    defer { std.db.close(db) }
    let result: DbRows = std.db.query(db, "SELECT CAST($1 AS TEXT)", [std.db.text("hello")])
    print(std.db.asText(result.rows[0][0]))
}
```

`connect(driver: String, dsn: String, config: DbConfig): DbConnection` validates configuration, creates a pool and pings before returning. A successful return confirms authentication/connectivity. DSNs should come from trusted runtime configuration/environment; do not accept them from HTTP input. Drivers can interpret options referencing local certificate/config files. DSNs are never included in native handle formatting or database errors. Do not hardcode them in source: source bundles contain all source text.

`ping(db): Void`, `close(db): Void`, `stats(db): DbPoolStats` operate on the pool. Stats fields `open`, `inUse`, `idle`, `waitCount` are Int. Close is idempotent and invalidates all aliases. Configuration is copied at connect; later edits to the Craft config do not change the pool.

| DbConfig field | Default | Allowed |
| --- | --- | --- |
| maxOpen | 8 | 1–128 |
| maxIdle | 2 | 0–maxOpen |
| idleTimeoutMs | 60000 | 1–86400000 |
| lifetimeMs | 300000 | 1–86400000 |
| connectTimeoutMs | 5000 | 1–300000 |
| timeoutMs | 5000 | 1–300000 |
| maxRows | 1000 | 1–10000 |
| maxBytes | 4194304 | 1–16777216 |

Each operation uses the calling Craft execution context with the configured timeout; earlier request/task cancellation still wins. `connectTimeoutMs` applies to initial ping. `timeoutMs` covers ping, execution, query materialization/pool wait, and the entire transaction lifetime starting at begin. Driver I/O and cancellation support can delay physical resource cleanup; these are cooperative limits, not forced OS-call termination. In particular, MySQL cancellation returns control and discards the affected client connection, but a server-side statement can remain visible until its bounded statement duration ends. Craft does not issue privileged `KILL` commands automatically. SQL Server Shared Memory was tested locally; SQL Server TCP/TLS remains a separate acceptance item.

## Ownership and concurrency

- A connection belongs to the function invocation that calls `connect`. On function exit, after child tasks and Craft defers, the pool closes automatically even after an exception. Returned/stored aliases are closed after the owning function returns; create the connection in the outer application scope and pass it into helpers instead of returning it from a factory.
- Native handles copy by reference. Pools support sharing with child tasks and as frozen HTTP state while their owner is alive. Each task/request has its own operation context. Closing one alias closes the shared resource for all aliases.
- A transaction belongs to the function calling `begin`. On function exit it rolls back unless committed. Closing its pool rolls back outstanding transactions first.
- Do not share a transaction for concurrent operations. Simultaneous query/execute attempts fail with `DbBusyError`; commit/rollback synchronize with an in-flight operation. Transaction handles cannot be HTTP state. Start a separate transaction per independent unit of work.
- At most 256 live owned DB handles per function. Explicitly closed connections and completed transactions are removed from the scope registry. A timed-out transaction still needs rollback/owner exit to release its registry entry.

## Parameters and result values

All parameter lists have type `DbValue[]`. Use these immutable constructors:

| Constructor / accessor | Craft value | SQL result mapping |
| --- | --- | --- |
| `int` / `asInt` | Int | Signed 64-bit integer; larger unsigned values fail |
| `float` / `asFloat` | Float | Finite floating point only |
| `bool` / `asBool` | Bool | Native booleans; MySQL BOOLEAN/TINYINT returns Int |
| `text` / `asText` | String | UTF-8 text |
| `decimal` / `asDecimal` | String | Exact plain decimal text; no Float conversion |
| `bytes` / `asBytes` | Bytes | Binary/BLOB/BYTEA, MySQL/PostgreSQL bit strings |
| `time` / `asTime` | DateTime | Driver time values; timezone behavior follows SQL type/driver |
| `nullValue` / `isNull` | SQL NULL / Bool | NULL remains distinct from empty or zero |

All functions use the `std.db.` prefix. `kind(value)` returns `int`, `float`, `bool`, `text`, `decimal`, `bytes`, `time` or `null`. Accessors are strict: calling `asText` on a decimal or NULL raises `DbTypeError`. Decimal notation is an optional sign, integer digits, and optional fraction, at most 1024 characters. Decimal parameters are bound as text: specify the destination column or an explicit SQL CAST when a type cannot be inferred. A driver returning decimal as a float is rejected to prevent silent precision loss.

MySQL date/time mapping requires `parseTime=true`; use `loc=UTC` for a defined conversion location (it does not set server timezone). Offset-free database values cannot convey an original timezone. SQL TIME/UUID/JSON or other driver-specific text representations may be returned as text; no automatic domain mapping is promised. Unsupported native values and invalid UTF-8 raise `DbTypeError`.

`query(db, sql: String, params: DbValue[]): DbRows` returns `columns: String[]` and `rows: DbValue[][]`. Indexing preserves column order and duplicate column labels. An empty result retains column metadata. There is no open cursor exposed to Craft: rows are materialized within budgets and closed before returning or throwing. Multiple result sets are rejected. Paginate explicitly in SQL; exceeding a limit raises `DbLimitError` rather than silently truncating.

SQL text is limited to 1 MiB, parameters to 2000, columns to 256. `maxBytes` separately bounds total bound parameters and materialized result data/metadata with per-value accounting. It is not a hard process-memory cap: drivers can allocate a large field or network packet before Craft checks it, and copying results follows existing Craft value semantics. Avoid unbounded SQL fields/results; server-level limits remain relevant.

`execute(db, sql, params): DbResult` exposes `affectedRows: Int?` and `lastInsertId: Int?`. Unsupported metadata is null; presence is a driver capability, not an inserted-row guarantee. PostgreSQL uses `RETURNING` with query for generated IDs; SQL Server can use `OUTPUT INSERTED...`. No automatic retry occurs after an ambiguous execution/commit outcome.

## Transactions

```craft
let tx: DbTransaction = std.db.begin(db)
defer { std.db.rollback(tx) }
std.db.executeTx(tx, "UPDATE items SET name = $1 WHERE id = $2", [std.db.text("new name"), std.db.int(1)])
std.db.commit(tx)
```

`queryTx` and `executeTx` have the same SQL/parameter/result shape as pool operations. Transactions use the server/driver's default isolation level; explicit isolation selection and read-only modes are deferred. `rollback` is idempotent, including after commit. A repeated commit or query on an ended transaction fails. A commit failure ends the handle and must be treated as potentially ambiguous: do not blindly retry writes. An operation error does not automatically undo a whole transaction until rollback or owner exit; applications should stop using a failed transaction.

## Errors

Catch exceptions with `.type` and `.message`: `DbConfigError`, `DbConnectionError`, `DbQueryError`, `DbConstraintError`, `DbTimeoutError`, `DbCancelledError`, `DbClosedError`, `DbBusyError`, `DbLimitError`, `DbTypeError`. SQL syntax/runtime failures include only sanitized vendor codes where available; credentials, SQL and row/parameter contents are omitted. Driver cancellation may surface a sanitized vendor error instead of the generic timeout type. Craft process cancellation remains governed by the interpreter's cancellation convention. Initial connection failures intentionally report generic connection errors.

## Compatibility

CLI version is 0.1.5. Source bundle format remains 3 with language 0.1.5; Rev.4 bundles still load, and Rev.4 runtimes reject bundles marked 0.1.5. New native names `DbConnection`, `DbTransaction`, `DbValue`, `DbConfig`, `DbRows`, `DbResult`, `DbPoolStats` are reserved. Rename existing declarations that collide. Database driver dependencies are compiled into the CLI, not downloaded by end users. Full earlier-revision regression, editor visuals and installer execution await user testing; see [Rev.5 status](../roadmap/phase1/output/phase1-rev5-status.md).
