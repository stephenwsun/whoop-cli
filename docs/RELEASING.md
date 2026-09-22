# Releasing whoop-cli

Releases are tag-driven and run through
[`.github/workflows/release.yml`](../.github/workflows/release.yml). Do not
create release tags from a dirty worktree or publish local binaries as official
artifacts.

## One-time repository administration

Complete this once before the first public tag:

1. Confirm `main` contains the merged release-readiness implementation and set
   `main` as the repository default branch in **Settings → General → Default
   branch**.
2. Protect `main` in **Settings → Rules → Rulesets** (or the equivalent branch
   protection page) with required pull requests, required passing CI checks,
   no force pushes, and no branch deletion. Require the branch to be up to date
   before merging when that fits the repository workflow.
3. Require the checks listed in the first-release administration issue:
   the four OS/Go test jobs, `minimum-go`, `windows`, six target builds, and
   `vuln`.
4. Confirm Actions are allowed to run workflows and that the release workflow's
   `contents: write` permission can create GitHub Releases.
5. Verify the setting from a shell:

   ```sh
   gh api repos/stephenwsun/whoop-cli --jq .default_branch
   ```

   This must print `main` before tagging.

Track the one-time setup and first-release verification in
[issue #34](https://github.com/stephenwsun/whoop-cli/issues/34).

## Before tagging

1. Complete the one-time repository administration section above.
2. Land the release queue on `main`.
3. Add a dated `## X.Y.Z - YYYY-MM-DD` section to `CHANGELOG.md`.
4. Run `make ci` and `go test -race ./...`.
5. Confirm the README, install guide, and supported target matrix match the
   release configuration.
6. Create and push an annotated SemVer tag such as `v0.1.0` from the validated
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
