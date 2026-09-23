# Craft 0.1.11 — GitHub source packages

Rev.11 downloads **source**, not DLLs/native binaries. Local path dependencies remain supported. Public GitHub repositories must contain a root craft.toml with `entry = "library"`, exported APIs in src/, and an exact tag `vMAJOR.MINOR.PATCH` matching the manifest version. GitHub Releases/assets and a central registry are not required.

## Install and import

```powershell
craft install github.com/spidermeaow/craft-test@v1.0.0 --alias greeting
craft package list
```

This adds/updates the project manifest and lock automatically:

```toml
[dependencies]
greeting = { github = "spidermeaow/craft-test", tag = "v1.0.0", version = "1.0.0" }
```

```craft
import greeting "greeting"

func main() {
    print(greeting.greet("Craft"))
}
```

Without --alias, the package name is used with hyphens replaced by underscores (`craft-test` → `craft_test`). Alias collisions with another source fail. An explicit install of a new tag for the same alias/source changes that project's version. Existing comments and unrelated manifest entries are retained. Package identity uses repository/commit plus manifest name/version; aliases are not identities.

Only exported symbols can be used across packages. Imported names remain module-scoped; identically named structs/functions in different packages do not merge. One graph rejects conflicting versions or identical package identities from different sources. Different projects can select different versions independently.

## Restore and updates

```powershell
craft install
craft install --offline
craft install github.com/spidermeaow/craft-test@v1.0.0 --alias greeting --refresh
```

- No-argument install restores the exact locked commits/checksums. Without a lock it resolves exact tags and creates one. If an existing lock disagrees with the graph, it fails rather than silently updating it.
- Offline requires existing pins and verified cached source. It makes no network requests.
- Reinstalling a pinned tag preserves its commit. Only explicit online --refresh re-resolves the same tag, reporting a changed commit. Prefer a new version tag rather than moving published tags.
- check/run/test/build/package check/list never fetch. Missing cache directs the user to install.
- package lock is still the explicit way to accept local source/graph changes; it preserves remote pins and never fetches or re-resolves tags.

## Cache and lock

Default cache: OS user cache directory / Craft / packages (Windows: normally LOCALAPPDATA/Craft/packages). `CRAFT_PACKAGE_CACHE` overrides the root, useful for testing or restoring to a fresh cache. Layout is repository / commit / tree checksum, so versions can coexist without activate/venv. Application source and package cache are separate; packages/ remains a user-managed convention for local libraries.

Lock format 1 remains for local-only graphs. Format 2 adds remote repository/tag/version/commit/tree checksum records. No absolute cache path or environment value is stored. Keep the application's craft.lock in source control. A library consumed as a dependency uses the application's resolved graph, not its own lockfile.

Source checksum is SHA-256 of compact Go JSON encoding of a lexicographically path-sorted array of `{path,sha256}` records: relative forward-slash file paths and SHA-256 of each file's exact bytes, including manifest/docs/tests. Archive timestamps, modes and directories are excluded. Module checksum separately hashes serialized module identity, dependency bindings and Craft sources. Cache content is checked before use; tampering fails closed. To restore corrupted content, select a fresh CRAFT_PACKAGE_CACHE and run install. No automatic deletion/garbage collection is provided.

Downloads use staging directories and atomic directory promotion; competing identical downloads verify the winner. No package hooks, builds or tests execute during install. Static checking occurs before changing project files.

## Recovery

Project installs/package lock operations use an OS-backed exclusive lock. Install journals the original/new manifest and lock before replacement. Project loads refuse to proceed while that journal exists. On a normal write failure, rollback is attempted automatically. After interruption:

```powershell
craft install --recover
craft install
```

Recovery restores the original pair only if files still match the journal's old/new versions. If someone edited them meanwhile, it fails and asks for manual review to preserve those edits. A damaged journal also requires manual review. This is process-interruption recovery, not a guarantee against storage corruption/power failure. Ignore `.craft-install.busy`, `.craft-install-journal.json` and `.craft-write-*` in source control; never remove a journal until reviewed/recovered.

## Limits and trust

- Public GitHub only; HTTPS endpoints api.github.com and codeload.github.com; redirects limited to those hosts, at most 3 hops. Each request has a 60-second timeout. No tokens/private repository support.
- 32 MiB compressed download, 64 MiB extracted content, 4096 archive entries, 16 MiB per archive file; existing module graph caps remain 64 modules/16 MiB Craft source/1024 source files. Metadata/decompression is bounded too.
- Reject archive traversal, duplicate/case-colliding names, symlinks/hardlinks/special files, reserved Windows names, submodules and LFS pointers. Local dependencies inside remote trees cannot escape their repository tree.
- Exact release tags only: no branches, prereleases, ranges, registry, monorepo package selection or automatic upgrades.
- Checksum pins content; it does not certify code quality or trust. Running/building imported code remains the user's decision. Cache validation is not a sandbox against hostile concurrent filesystem modification.

New bundles declare format 3 / language 0.1.11 and contain dependency source. They run without the original network/cache. Older runtimes reject them; prior supported bundle languages remain readable by 0.1.11.
