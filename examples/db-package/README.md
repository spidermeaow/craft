# DB package example

This example reads the current database name through `std.db`, then extracts it with `db.requireOne` and `db.requiredText`. It repeats the query inside a transaction to demonstrate that the application owns the transaction. It does not write to the database.

From this directory, run `craft package lock`, `craft check`, and `craft run` with `CRAFT_DB_DRIVER` (`mysql`, `postgres`, or `sqlserver`) and `CRAFT_DB_DSN` set to a dedicated test database. The expected output is `database: craft_test` and `transaction query: passed` for the Rev.5 test databases. Keep credentials in the existing encrypted local secret store, never in source or logs.
