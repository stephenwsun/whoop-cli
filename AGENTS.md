# Project agent memory

This file is the project's committed home for project-intrinsic agent knowledge: build, test, release, architecture, and sharp-edge notes that should travel with the code.

- Add durable project-specific notes here as they are discovered through real work.

## Maintaining this file

Keep this file for knowledge useful to almost every future agent session in this project.
Do not repeat what the codebase already shows; point to the authoritative file or command instead.
Prefer rewriting or pruning existing entries over appending new ones.
When updating this file, preserve this bar for all agents and keep entries concise.

## Architecture

- `cmd/whoop/main.go` is intentionally thin; CLI parsing and command policy live in `internal/cli`.
- The reusable client and API models are the public `whoop` package. CLI-only auth, runtime loading, output, and summaries live under `internal/`.
- `whoop schema --json` is the machine-readable command contract; update `internal/cli/schema.go` with command or flag changes.
- The default safety profile is read-only. `--no-input` rejects browser authorization, and `--allow-command` provides an exact command allowlist.
- Default checks are `go test ./...`, `go vet ./...`, and `make ci`; live API tests require the `integration` build tag and `WHOOP_IT=1`.
