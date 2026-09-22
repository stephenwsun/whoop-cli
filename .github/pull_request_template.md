## What changed?

<!-- Describe the user-visible behavior and why it is needed. -->

## Verification

- [ ] `make ci`
- [ ] `go test -race ./...` when auth, storage, or transport changed
- [ ] Live testing was opt-in and secrets/personal data were not committed

## Compatibility and safety

- [ ] CLI schema, output, exit codes, and docs updated if changed
- [ ] No tokens, client secrets, or private WHOOP data appear in the diff
- [ ] Changelog updated for user-visible behavior
