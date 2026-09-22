# Releasing whoop-cli

Releases are tag-driven and run through
[`.github/workflows/release.yml`](../.github/workflows/release.yml). Do not
create release tags from a dirty worktree or publish local binaries as official
artifacts.

## Before tagging

1. Land the release queue on `main`.
2. Add a dated `## X.Y.Z - YYYY-MM-DD` section to `CHANGELOG.md`.
3. Run `make ci` and `go test -race ./...`.
4. Confirm the README, install guide, and supported target matrix match the
   release configuration.
5. Create and push an annotated SemVer tag such as `v0.1.0` from the validated
   `main` commit.

The workflow builds Linux, macOS, and Windows amd64/arm64 archives, injects the
version, commit, and date, and publishes `checksums.txt` to the GitHub Release.

## After publication

Verify the release page contains every expected archive and checksum. Download
at least one native archive and run:

```sh
whoop --version
whoop --help
whoop schema --json
whoop config diagnose --json
```

If a release is defective, publish a corrected patch release rather than moving
a published tag. Add upgrade or migration notes to the changelog when the
machine-readable schema, output, credentials, or supported API contract changes.
