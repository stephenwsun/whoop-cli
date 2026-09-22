# Changelog

## Unreleased

## 0.1.0 - 2026-09-22

- Add public governance, contribution, security, support, and issue-template guidance.
- Add build version metadata, `whoop --version`, and automation contract coverage.
- Add cross-platform install/release documentation and tagged GoReleaser publishing.
- Harden credential-file failure handling and headless OAuth authorization.
- Fix pagination to send the WHOOP v2 `nextToken` query parameter so multi-page reads advance instead of repeating the first page.
- Decode `user_id` as a number in `--json` output and the public `whoop` models to match the WHOOP v2 API.
