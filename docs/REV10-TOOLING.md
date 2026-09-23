# Rev.10 contracts

## Logging

`std.log.setLevel(level: String): Void` accepts debug/info/warn/error. Default is info. `std.log.write(level: String, message: String, fields: Map<String,String>, sensitive: String[]): Void` writes one JSON line to stderr, with UTC time, level, message and fields. Fields are encoded in sorted key order; timestamps vary. Messages and values are limited to 4096 bytes, keys to 128 bytes, fields and sensitive keys to 64 entries. Validation runs even when filtered.

Sensitive field names are exact, case-sensitive matches. Values are replaced with `[REDACTED]` before serialization. Message text and field names are not inspected for secrets. Do not place secrets in them. Newlines are JSON-escaped. A shared execution logger serializes task/HTTP writes; HTTP children inherit the same level. Separate executions have separate configuration. Writer failure throws LogError without the writer's error text; a failed writer can still have emitted a partial record. No background queue or retries exist.

Embedders can use interpreter.RunWithWriters and Session.SetLogWriter. Configure writers before execution. CLI run forwards logs to its supplied stderr writer.

## Configuration

- `std.config.required(name: String, value: String?): String`: rejects absent or whitespace-only values; returns original text otherwise.
- `std.config.defaultValue(value: String?, fallback: String): String`: fallback only when absent; preserves empty strings.
- `std.config.intRange(name: String, value: String, min: Int, max: Int): Int`: signed decimal Int64, inclusive range; rejects whitespace, overflow and reversed bounds.
- `std.config.bool(name: String, value: String): Bool`: only exact `true`/`false`.

Failures throw ConfigError identifying the setting and constraint, never the value. Setting names must not contain secrets. Precedence is chosen by application code; no dotenv loading occurs.

## Filesystem and compatibility

Errors use Exception.type: FsNotFoundError, FsPermissionError, FsPathError, FsKindError, FsLimitError, FsEncodingError and FsIOError. Unclassified OS failures use FsIOError. Path lexical functions retain host conventions. Paths appear in diagnostics and must not contain secrets. Existing catch blocks still work, but consumers matching old RuntimeError types/messages must migrate.

Filesystem invocation checks cancellation before I/O. Atomic writes additionally check before replacement; cancellation after that check can race the successful commit. Blocking OS calls cannot be forcibly canceled. Cleanup is best effort; failure is appended to the primary error. Atomic replacement does not preserve permissions/ACLs by contract and does not guarantee power-loss durability.

New bundles retain format 3 and declare language 0.1.10, so earlier runtimes reject them. This runtime retains previously supported bundle languages 0.1.1–0.1.7 according to format. Rev.8/9 produced language 0.1.7 bundles; those remain supported.
