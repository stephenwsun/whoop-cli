# whoop-cli

![CI](https://img.shields.io/github/actions/workflow/status/stephenwsun/whoop-cli/ci.yml?branch=main&style=flat-square&label=ci)
[![Release](https://img.shields.io/github/v/release/stephenwsun/whoop-cli?style=flat-square)](https://github.com/stephenwsun/whoop-cli/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/stephenwsun/whoop-cli?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/github/license/stephenwsun/whoop-cli?style=flat-square)](LICENSE)

`whoop-cli` is open source under the [MIT license](LICENSE).

## Install

Download a verified archive from the [latest release](https://github.com/stephenwsun/whoop-cli/releases/latest), install with Go, or build from source:

```sh
go install github.com/stephensun/whoop-cli/cmd/whoop@latest
whoop --version
```

See [Install and upgrade](docs/install.md) for release archives, checksums, supported platforms, and uninstall guidance.

`whoop` is a read-only Go CLI and reusable client for the public WHOOP account-data API. It does not control devices, start/stop workouts, or access undocumented endpoints.

## Setup

1. Create a WHOOP developer application and register `http://localhost:8400/callback` as its redirect URI.
2. Configure credentials using environment variables for a one-off session:

   ```sh
   export WHOOP_CLIENT_ID='...'
   export WHOOP_CLIENT_SECRET='...'
   ```

   On macOS, production credentials belong in Keychain service `whoop-cli` under accounts `client_id`, `client_secret`, and `refresh_token`. On other platforms the fallback is `~/.config/whoop/credentials.json` with mode `0600`.
3. Run `whoop config diagnose`, then `whoop auth`. The authorization-code flow binds a localhost callback, verifies a cryptographically random CSRF state, requests offline access, and stores the refresh token. WHOOP rotates refresh tokens; every API command persists the replacement before continuing.

The minimum least-privilege scopes are `read:profile read:body_measurement read:cycles read:recovery read:sleep read:workout offline`. Do not grant write or unrelated scopes.

## Commands

```sh
whoop --version
whoop profile
whoop recovery --start 2026-09-14T00:00:00Z --json
whoop workouts --plain
whoop brief
whoop week
whoop weekly --json
```

`--json` emits stable JSON data on stdout. `--plain` emits deterministic tab-separated `key=value` records. Human summaries and resource records are emitted on stdout; progress and errors are emitted on stderr. List commands follow every `next_token` page and reject a repeated token instead of looping forever. `--limit`, `--start`, and `--end` are passed to the API. Use `whoop auth --no-browser` when the local environment cannot open a browser.

Detailed operational guides:

- [Install and upgrade](docs/install.md)
- [Quickstart](docs/quickstart.md)
- [Authentication and credential storage](docs/auth.md)
- [Credential migration](docs/migration.md)
- [Automation contract](docs/automation.md)
- [Live testing](docs/live-testing.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Command index](docs/commands/README.md)
- [Support](docs/support.md)
- [Release procedure](docs/RELEASING.md)

## API limitations

The public API exposes account data and scored activities, not direct device control. This project intentionally has no write commands. Availability, score state, timestamp precision, and sport naming follow the official API. A missing or pending score is represented as a null score; it is not synthesized. Pagination and rate limits are handled by the client, but callers must still respect WHOOP quotas.

References: [WHOOP Developer Portal](https://developer.whoop.com/), [WHOOP API documentation](https://developer.whoop.com/api/), [OAuth 2.0](https://developer.whoop.com/docs/developing/oauth/).

## Athena integration contract

Athena should invoke the executable, never import its internal packages:

- `whoop brief` returns one line each for latest recovery, latest sleep, and seven-day workout adherence.
- `whoop week` returns `YYYY-MM-DD | sport` workout lines followed by the adherence line.
- `whoop weekly` returns recovery, HRV/resting-heart-rate, sleep, and adherence rollups.
- For machine-readable integration, pass `--json`; parse the documented `{"lines":[...]}` summary shape or resource JSON rather than scraping human labels.
- Exit status `2` means local configuration or usage is missing; exit status `1` means an API, OAuth, or runtime failure.

The adherence compatibility rule counts only `weightlifting` as `lift` and `muay-thai` as `Muay Thai`; all other sports remain visible in `week` but do not count toward the `2 lift + 2 Muay Thai + 1 flex` target. The CLI does not modify Athena.

## Development

```sh
make ci
go test -race ./...
```

See [Contributing](CONTRIBUTING.md) before opening a pull request. The project uses the [MIT license](LICENSE) and reports vulnerabilities through [Security](SECURITY.md).

The package under `whoop/` is the reusable client and model API. CLI-only authentication, output, summary, and runtime wiring live under `internal/`; `cmd/whoop` contains only process entrypoint wiring.
