# Contributing to whoop-cli

Thanks for helping improve `whoop-cli`. Keep changes narrow, testable, and
compatible with the documented read-only CLI contract.

## Development setup

1. Install the Go version declared in [`go.mod`](go.mod).
2. Clone the repository and create a feature branch.
3. Run `make ci` before opening a pull request.
4. Run `go test -race ./...` for changes to authentication, storage, or transport.

The reusable API client is under `whoop/`; CLI-only wiring belongs under
`internal/`; `cmd/whoop` should remain a thin entrypoint. Update
`whoop schema --json` and the automation documentation when changing commands,
flags, output, or exit codes.

## Live WHOOP testing

Live tests are opt-in and must never run with credentials in ordinary CI:

```sh
WHOOP_IT=1 go test -tags=integration ./internal/integration
```

Use a bounded recent time window, avoid printing tokens or personal profile
fields, and do not commit live response data. See [live testing](docs/live-testing.md).

## Pull requests

- Explain the user-visible behavior and compatibility impact.
- Include focused tests for behavior and failure modes.
- Update the changelog for user-visible changes.
- Keep secrets, refresh tokens, and personal WHOOP data out of commits, logs,
  screenshots, and issue comments.
- Do not add undocumented or write-capable WHOOP endpoints.

The maintainer reviews public API, output, credential, and release changes with
extra scrutiny because scripts depend on their stability.
