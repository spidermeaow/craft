# Phase 1 Rev.11 status

Date: 23 September 2026. Version: 0.1.11. Implementation and automated acceptance passed. Windows distribution built; installer execution remains a user acceptance step.

## Delivered

- GitHub exact-tag source install, manifest editing with aliases, lock format 2 and immutable repository/commit/checksum cache.
- Locked restore, offline restore, explicit tag refresh, transitive/local-in-remote dependencies, source boundaries and version/identity conflict diagnostics.
- Source archive validation, resource limits, cancellation, cache integrity, concurrent promotion and exclusive project updates.
- Journal/recovery for interrupted manifest/lock replacement, with refusal to overwrite later user edits.
- Namespace/function/struct isolation and public-export checking; source bundles portable without cache/network.

## Evidence

- Rev.11 fixture tests with race detector pass (build/rev11-unit.log): moved tags, restored commits, corruption, offline miss, three versions across projects, transitive graphs, cycles/conflicts, escaping paths, malicious/truncated archives, transport limits, concurrency and recovery.
- Full regression, race and HTTP stress 10,000 requests passed via scripts/test-rev11.ps1 -Race -Stress -Live (build/rev11-regression.log).
- Live public GitHub install/restore/check/test/run/build and cache-independent bundle execution passed against spidermeaow/craft-test@v1.0.0, commit 9785e1a1019a02f64c543d564ca38a11d4261a91. Tree checksum 21839f8254a44a30dce9b25437f2c0f55c736e23b6c527d2340f5993065df300.

## Scope

Distribution acceptance: the release binary dist/craft.exe reports 0.1.11 and successfully installed the live GitHub dependency in examples/rev11-github, checked/tested/ran the consumer (1 passed), restored offline, checked the locked graph, built and ran its source bundle. Windows Setup and both VSIX artifacts were packaged successfully. Setup was not executed and the user's installed Craft was not upgraded automatically. Run dist/Craft-setup.exe and follow docs/TRY-REV11.md for installed acceptance.

Real GitHub acceptance uses the user's existing v1.0.0 tag. Versions 1.2.0/2.0.0 and moved-tag scenarios use isolated fixtures; no extra tags were published and no remote source was changed. Private repositories, semver ranges, registry, binary libraries and multi-version graphs remain excluded. No guarantee against arbitrary hostile concurrent filesystem modification or power-loss/storage corruption.
