# Craft DB package 0.1.45 candidate

`db` provides strict helpers for materialized `DbRows` and `DbResult` on the public `std.db` API. The application creates and owns `DbConnection` and `DbTransaction`; package helpers do not retain either handle. MySQL, PostgreSQL, and SQL Server remain the supported drivers of `std.db`.

Add `db = { path = "packages/db" }` under `[dependencies]`, run `craft package lock`, then use `import package db "db"` in each file that calls these helpers. The bundled package version follows the installed Craft toolchain; do not specify a version in the dependency.

| API | Behavior |
| --- | --- |
| `one`, `requireOne` | Require at most one or exactly one materialized row. Multiple rows always fail. |
| `columnIndex` | Exact column label lookup; missing or duplicate labels fail. |
| `requiredText/Int/Decimal/Bytes/Time` | Fail on NULL, missing column, short row, or wrong native type. |
| `optionalText/Int/Decimal/Bytes/Time` | Return null only for SQL NULL. Other failures still throw. |
| `expectAffected` | Require a known `affectedRows` value equal to the expected count. |

Errors are `Exception` messages with stable prefixes (`DbNoRows`, `DbMultipleRows`, `DbMissingColumn`, `DbDuplicateColumn`, `DbRowShapeError`, `DbNullValue`, `DbTypeMismatch`, `DbAffectedRows`). They contain no SQL, DSN, parameters, or row values. `std.db` transport errors still come from the Runtime unchanged. Query results are materialized; use SQL limits for large datasets.

See [`examples/db-package`](../../examples/db-package) for a read-only connection and transaction example. Migration and seed runners are not part of this release. For now, applications should review and explicitly run versioned SQL changes and idempotent reference-data inserts; demo or test seed data must be opt-in. Do not run schema changes or seed data implicitly on app startup.

For example, an application may keep a reviewed `001_reference_status.sql` migration:

```sql
CREATE TABLE reference_status (
    code VARCHAR(32) PRIMARY KEY,
    label VARCHAR(100) NOT NULL
);
```

Then an explicitly invoked reference-data seed can use parameterized `std.db.execute` operations after checking whether each code already exists. Track whether `001_reference_status.sql` was applied outside this package until a migration runner can record history and checksums, coordinate concurrent processes, and define recovery separately for each database driver. In particular, MySQL DDL can commit implicitly; do not assume a failed multi-statement migration rolls back as one unit. Demo/test rows should use a separate explicit command or environment, never normal application startup.
