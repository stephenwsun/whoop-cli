# gogcli open-source release benchmark for whoop-cli

**Date:** 2026-09-22  
**Reference:** [openclaw/gogcli](https://github.com/openclaw/gogcli) (repository state on `main`; latest release metadata inspected 2026-09-22)

This is an evidence-based benchmark, not a claim that `whoop-cli` should copy gogcli's scope. gogcli is a large, mature Google Workspace client; the useful comparison for a small Go CLI is the release discipline around a narrow, documented contract.

## Executive summary

`whoop-cli` already has several strong foundations: an explicit read-only boundary, least-privilege OAuth scopes, credential redaction/storage guidance, JSON/plain output, pagination-loop protection, a Makefile, CI, and a changelog. The public-release blockers are mostly distribution and governance rather than API breadth:

1. Provide a real release pipeline that produces versioned, cross-platform archives and checksums (not CI-only build artifacts).
2. Pin and verify CI actions and add release/build verification, including a reproducible version source.
3. Publish contributor/security/support policy and a clear public installation/upgrade path.
4. Turn the existing automation contract into a tested, stable public interface (including schema and exit codes).
5. Add release-quality documentation and changelog discipline.

## What gogcli does well (observable practices)

### 1. README is a release front door

The [README](https://github.com/openclaw/gogcli/blob/main/README.md) has badges for CI, releases, Go version, license, and Homebrew; copy/paste install paths for Homebrew and `go install`; a five-minute quickstart; supported-platform/install documentation; automation examples; explicit product boundaries; and links to generated command references. It states the module path and warns about the prior module path. It also names the license and non-affiliation boundary.

**Apply to whoop-cli:** keep the existing concise product boundary, but add badges/release links, a supported-platform table, a binary installation path, upgrade/uninstall instructions, and a direct link to the automation contract. Make the module path and executable name part of the stable public contract.

### 2. Automation is treated as an API

[gogcli/docs/automation.md](https://github.com/openclaw/gogcli/blob/main/docs/automation.md) specifies stdout versus stderr, `--json` and stable TSV, `--no-input`, read-only enforcement, output-flag precedence, schema discovery (`gog schema --json`), named exit codes, cancellation/retry behavior, and machine-readable safety state. The README gives safe invocation examples using exact command allowlists and no-send/read-only guards.

**Apply to whoop-cli:** the existing [automation guide](../automation.md) is directionally excellent. Promote its schema and exit-code behavior into a versioned compatibility promise, document flag precedence and output schemas with examples, and add contract tests for stdout/stderr separation, malformed usage, no-input, and schema output. Keep the smaller three-code model if it is sufficient; do not copy gogcli's many codes without observable need.

### 3. CI checks more than compilation

[gogcli/.github/workflows/ci.yml](https://github.com/openclaw/gogcli/blob/main/.github/workflows/ci.yml) runs formatting, tests, lint, dead-code analysis, generated-doc/agent-skill checks, and Docker-version checks. It separately tests the minimum supported Go series, builds on Windows, tests on macOS, and performs a cgo/Keychain build. Actions are pinned to immutable commit SHAs. Concurrency cancels superseded PR runs and workflow permissions are read-only.

**Apply to whoop-cli:** current [CI](../../.github/workflows/ci.yml) tests Go 1.22/1.23 on Ubuntu/macOS, vets, checks formatting, and cross-builds only Linux amd64, macOS arm64, and Windows amd64. Add a minimum-supported-Go job that matches `go.mod`, an explicit build smoke test, Windows/macOS credential-path coverage where feasible, race testing for security-sensitive storage, and dependency/security scanning. Pin actions to SHAs (or document an intentional alternative), use concurrency cancellation, and retain least-privilege workflow permissions. Add arm64/amd64 coverage for the platforms you promise.

### 4. Releases are a controlled, inspectable pipeline

The [unified release workflow](https://github.com/openclaw/gogcli/blob/main/.github/workflows/release-unified.yml) delegates to a shared release pipeline. [RELEASING.md](https://github.com/openclaw/gogcli/blob/main/docs/RELEASING.md) requires a dated changelog section, a version file, green `make ci`, protected-main release authorization, immutable annotated tags, checksums, asset inventory verification, native macOS signing/notarization, non-Darwin rebuild verification, and Homebrew handoff. The [install guide](https://github.com/openclaw/gogcli/blob/main/docs/install.md) names exact archive filenames for Darwin/Linux/Windows, checksum assets, Docker/GHCR, Homebrew, source builds, and verification commands. The [latest GitHub release metadata](https://api.github.com/repos/openclaw/gogcli/releases/latest) shows bot-published `v0.41.0` assets including `checksums.txt`, `ASSET-INVENTORY.json`, and six platform archives.

**Apply to whoop-cli:** there is no release workflow, GoReleaser configuration, version injection, archive/checksum publication, signed provenance, or documented release process in the inspected repository. CI's uploaded artifacts are ephemeral workflow artifacts, not a public release channel. Add a tagged release workflow (GoReleaser or an equivalently maintained script) for Linux/macOS/Windows architectures actually supported, checksums, generated release notes from the changelog, and a post-publish smoke test (`whoop --version`, `whoop --help`, and a no-credential diagnostic). Start with checksums and reproducible metadata; signing/SLSA provenance can follow if operationally feasible.

### 5. Changelog and versioning communicate compatibility

[gogcli/CHANGELOG.md](https://github.com/openclaw/gogcli/blob/main/CHANGELOG.md) has dated SemVer sections, highlights, fixes, dependency/security notes, contributor attribution, and validation notes. `docs/RELEASING.md` couples the changelog version to a build version file and a closeout development version.

**Apply to whoop-cli:** current [CHANGELOG.md](../../CHANGELOG.md) has only `Unreleased` and no version/tag procedure. Establish SemVer (or explicitly state another scheme), add a version source injected into `whoop --version`, require a dated section before tagging, and document compatibility/migration expectations for credential files and output schemas.

### 6. Packaging and public project hygiene are explicit

The root [repository tree](https://api.github.com/repos/openclaw/gogcli/contents) includes `LICENSE`, `AGENTS.md`, `VISION.md`, `.golangci.yml`, `.goreleaser.yaml`, Docker packaging, generated docs, and multiple workflows. gogcli's MIT license is directly visible at [LICENSE](https://github.com/openclaw/gogcli/blob/main/LICENSE). Its README points to release/install/supporting docs rather than requiring users to infer operational details.

**Apply to whoop-cli:** `LICENSE`, `CONTRIBUTING.md`, `SECURITY.md`, support expectations, and a release/installation guide were not present in the inspected tree. Add a permissive license if that is the intended public model; add contribution and vulnerability-reporting instructions before inviting outside users. A small project need not add Docker, agent skills, or a vision document merely to match gogcli.

## Must-have release blockers for whoop-cli

| Priority | Blocker | Evidence/gap to inspect | Practical acceptance criterion |
| --- | --- | --- | --- |
| P0 | Public installable releases | Only `.github/workflows/ci.yml` artifacts; no release workflow/config/docs | A tagged GitHub Release publishes named archives for promised OS/architectures plus `checksums.txt`; README install command works |
| P0 | Version identity | No visible version file or documented `--version` release procedure | `whoop --version` reports tag/build version; source builds identify themselves as development builds; CI verifies it |
| P0 | Security and contribution intake | No `SECURITY.md` or `CONTRIBUTING.md` found | Private vulnerability-report path, supported versions, disclosure expectations, and PR/test guidance are public |
| P0 | Stable automation contract | Existing docs define behavior, but no evidence here of a release-gated contract-test job | Tests assert JSON/plain schemas, stderr discipline, exit codes, schema output, and no-input/read-only behavior |
| P0 | Release CI integrity | Actions use floating `@v4`/`@v5`; no release verification or dependency scan | Pin or consciously govern actions; run tests/vet/format/build and dependency vulnerability checks on PR and release candidates |
| P1 | Supported platform matrix | CI covers three target tuples, while public installation expectations are not documented | README and release workflow agree on OS/architecture support and CI builds every advertised tuple |
| P1 | Changelog/release process | Only `Unreleased` section | Dated release sections, tag procedure, generated notes, and upgrade/migration notes are required before publish |
| P1 | Artifact integrity/provenance | No checksums or attestations | Verify checksums in CI and publish provenance/signatures when the chosen distribution supports them |

## Nice-to-haves after the blockers

- Homebrew tap or another package manager, once archive names/checksums are stable.
- Docker image only if users genuinely need headless deployment; gogcli documents this because its OAuth/keyring use case warrants it.
- Man pages or generated command reference from the same command schema.
- Broader integration/live API smoke tests with opt-in credentials (never secrets in ordinary CI).
- Dependabot/Renovate policy, CodeQL, race/coverage trend reporting, and pinned tool versions.
- Signed macOS binaries/notarization if macOS distribution is a primary path.
- Reproducible-build checks and an asset inventory/provenance attestation.

## Recommended order for a small Go CLI

1. Add license, `SECURITY.md`, `CONTRIBUTING.md`, support/versioning policy, and a release checklist.
2. Define `whoop --version` and inject the tag at build time; update README with exact install and verification commands.
3. Add contract tests around the already-documented JSON/plain/schema/exit-code behavior.
4. Harden CI: action pinning, minimum Go, race/security checks, and all advertised build targets.
5. Add GoReleaser (or equivalent) for archives, checksums, changelog-driven GitHub Releases, and release smoke tests.
6. Add package-manager/container distribution only after the core release path is repeatable.

The key lesson from gogcli is not its feature count. It is that users can discover the contract, install a versioned artifact, automate it safely, inspect failures by exit code, and understand how releases are built and maintained.
