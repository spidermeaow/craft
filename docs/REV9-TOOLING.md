# Rev.9 tooling primitives

Craft 0.1.9 extends the existing `std.fs`, `std.path` and `std.env` primitives for tools that must safely replace a file. It does not make Craft a filesystem sandbox and it does not add a dotenv loader, process execution, directory deletion, file watcher or configuration framework.

## Paths

`std.path.clean(path)` performs lexical cleanup only. It neither accesses the filesystem nor resolves symlinks. `std.path.absolute(path)` resolves a relative path from the **process working directory**, and `std.path.relative(base, target)` asks the operating system for a relative path. On Windows, paths on incompatible volumes can fail to produce a relative path.

`join`, `fileName`, `extension`, `parent` and `isAbsolute` retain their existing behavior. All path APIs use the host operating system's path conventions; Rev.9's supported acceptance target is Windows, including drive and UNC cases.

## Atomic writes

Atomic writes are limited to 16 MiB and reject existing non-regular destinations, including symlinks. Text writes require UTF-8. Parent directories must already exist; existing file permissions/ACLs are not promised to be preserved by replacement. Cleanup after failure is best effort. Paths themselves appear in diagnostics, so do not put secrets in path names.

Use `std.fs.writeTextAtomic` to write a whole UTF-8 text file, or `std.fs.writeBytesAtomic` for binary data:

```craft
let data: String = std.json.stringify(settings)
std.fs.writeTextAtomic(outputPath, data)
```

Craft creates a temporary file in the destination directory, writes and closes it, then asks the OS to replace the target. A write or close failure leaves an existing destination untouched. If replacement fails—for example because another process locks the destination—Craft reports an error and removes its temporary file on a best-effort basis.

This is not a cross-directory transaction, a multi-file transaction, a filesystem sandbox, or a guarantee that data has survived a machine/power crash. A different process can still race any path validation. New atomic-write files use the platform's normal `0644` creation mode where that concept applies.

`std.fs.readBytes(path)` reads a regular file as immutable `Bytes`; it does not require UTF-8. `readText` and `readBytes` both have a 16 MiB complete-file limit. Use `Bytes.text()` only when the bytes are valid UTF-8.

## Diagnostics and configuration

Filesystem diagnostics state the attempted operation and requested path while avoiding arbitrary OS error text, which can contain machine-specific configuration. They never include file contents or environment values.

Use `std.env.lookup(name)` when missing and present-but-empty need different handling. `std.env.get(name)` returns `""` for both cases. `std.env.args()` contains only the program arguments passed to Craft; Craft cannot modify its parent process environment. Do not store passwords in source, `craft.toml`, `craft.lock`, or output diagnostics, and do not rely on automatic `.env` loading.
