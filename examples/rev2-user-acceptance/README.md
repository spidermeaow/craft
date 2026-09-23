# Recovered user acceptance fixture

`tests/full.craft` contains the 15 tests recovered verbatim from
`D:\02-Repository\craft-test\tests\full.craft` on 2026-09-18, after the user's
formatting pass. `smoke.craft` adds the starter arithmetic test (16 total).
The original helper functions were absent from that project's current `src/`;
the minimal `Profile`, `optionalEcho` and `updatedProfile` helpers here were
reconstructed from test assertions. This is not evidence of a new passing run.

From this directory, set `$env:CRAFT_REV2_VALUE = 'ready'` and run `craft test`.
The filesystem test requires its three named files to be absent and cleans up
the files using defer; it leaves an empty `craft-rev2-test-data` directory.
Use a clean copy if these paths already contain data. Restore the previous
environment variable value afterwards. See `docs/TRY-REV3.md` for commands.
