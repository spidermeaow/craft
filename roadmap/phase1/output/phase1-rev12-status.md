# Phase 1 Rev.12 status

Date: 6 October 2026. Target version: 0.1.12. This file records observed results separately from planned acceptance.

## Implemented

- Repeat `craft install` leaves unchanged manifest/lock files in place and reports the no-change result.
- GitHub transport retries network timeouts and HTTP 502/503/504 at most twice after the first attempt, with bounded delays and cancellation; permanent/validation failures are not retried.
- Package resolve/download/archive validation errors carry operation and package context. Cache integrity errors give a non-destructive recovery procedure.
- Package failures with a manifest position now use the project error code `E4002` and exit code 4, including offline cache misses.
- CLI and Windows Setup metadata target 0.1.12. Lock format 2 and bundle language 0.1.11 remain compatible.

## Evidence collected

- `go test ./...` — passed.
- `go test -race ./...` — passed. Rev.12 fixtures cover no-op install, targeted error code, bounded transport retry/cancellation and recovery in a fresh cache without changing the lock.
- `scripts/test-rev11.ps1 -Race -Stress -Live` — passed, including database-backed regression, 10,000 HTTP requests and live GitHub acceptance.
- Live public GitHub test against `spidermeaow/craft-test@v1.0.0` — passed: commit `9785e1a1019a02f64c543d564ca38a11d4261a91`, source checksum `21839f8254a44a30dce9b25437f2c0f55c736e23b6c527d2340f5993065df300`. It installed, checked, tested, ran, restored the locked commit into a fresh cache and ran a bundle without source cache.
- Windows portable `dist/craft.exe` — reported 0.1.12; in a temporary project/cache it rejected offline restore from an empty cache, installed from the existing lock, checked, passed one Craft test, ran, built and ran a bundle, then reported an unchanged repeat install.
- `scripts/build.ps1 -Installer` — built Windows amd64 CLI, language VSIX and Setup 0.1.12. Setup product version reported 0.1.12; SHA-256 values in `dist/SHA256SUMS.txt` matched the built CLI and Setup at the time of that build.
- Windows Setup acceptance — upgraded an existing 0.1.11 per-user installation to 0.1.12 with exit 0. Installed CLI passed package check, one Craft test, run, build and bundle run against the locked `spidermeaow/craft-test@v1.0.0` package. Rev.12 docs and both VSIX files were present; user PATH contained exactly one Craft entry.
- Uninstall/reinstall acceptance — uninstaller exited 0, removed `craft.exe` and the user PATH entry; reinstall exited 0, restored Craft 0.1.12 and exactly one PATH entry. A fresh process using current machine/user PATH resolved `craft` to the installed executable and reported 0.1.12. The installed Craft remains 0.1.12.

## Failure matrix

| Case | Expected result | Evidence |
| --- | --- | --- |
| Repeat install with valid lock/cache | No manifest/lock replacement | Rev.12 file-identity test and portable CLI |
| Offline with empty cache | Exit 4 / E4002, no network or project edits | Rev.12 CLI test and portable CLI |
| Moved tag versus pinned commit | Use pinned commit until explicit refresh | Rev.11 fixture and live restore |
| Corrupted cache/source checksum | Reject; recover using fresh cache directory without changing lock | Rev.11 integrity fixture and Rev.12 fresh-cache recovery fixture |
| 502/503/504 or timeout | Bounded retry, cancellation honored | Rev.12 transport fixtures |
| 403/429/404 or invalid archive | No retry of permanent/validation failure | Rev.11 archive/transport fixtures; Rev.12 404 fixture |
| Manifest/lock interruption or later user edit | Recovery or explicit refusal without overwriting edits | Rev.11 injected-failure/recovery fixtures |
| Transitive graph and identity/version conflict | Deterministic graph or clear rejection | Rev.11 graph fixtures and full regression |
| Bundle without original cache/network | Runs with Craft CLI | Live GitHub test and portable CLI |

## Remaining scope

- Linux/macOS release acceptance, private GitHub, registry, binary libraries and native application executables remain outside Rev.12.
- No public GitHub release was created by this work; the Windows artifacts are local in `dist/`.

The project remains in Phase 1; only the user decides when Craft is ready for version 1.00.
