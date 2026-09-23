# Rev.10 acceptance

Implementation/build does not constitute acceptance. From the repository root, run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\test-rev10.ps1
```

The regression script needs the existing database test credentials/fixtures. Use `-Race -Stress` for the extended suite. No credentials should appear in output.

After installing the candidate Setup, open a new terminal:

```powershell
craft version
cd examples/rev10-tools
craft check
craft test
craft run -- output.txt
craft build
craft run dist/rev10-tools.craftbundle -- output.txt
```

Expect version 0.1.10, a passing configuration test, stdout `Saved; port: 8080`, and JSON log on stderr with token `[REDACTED]`. Set PORT to an invalid value to expect ConfigError without the input value. Use a directory as output to expect FsKindError. Use a new/disposable output file: successful runs replace it.

Inspect redirected stderr separately from stdout to confirm JSON records do not contaminate command output. Restore PORT after testing. Check installer upgrade and bundle compatibility before release approval.
