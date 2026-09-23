# Rev.9 safe tooling example

This tool validates a JSON object and replaces the output file with a normalized JSON representation using `std.fs.writeTextAtomic`.

```powershell
cd examples/rev9-tools
craft run -- input.json output.json
```

If writing the temporary file or replacing `output.json` fails, a pre-existing output remains unchanged. This example does not load `.env` files or accept secrets.
