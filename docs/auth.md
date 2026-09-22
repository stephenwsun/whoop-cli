# Authentication and credential storage

`whoop auth` uses the WHOOP OAuth authorization-code flow. It creates a cryptographically random state value, opens the configured consent URL, binds the localhost callback, verifies the returned state, exchanges the code, and stores the offline refresh token.

Automatic browser launch is available on macOS and Linux only. On Windows and other platforms, plain `whoop auth` exits with an error; run `whoop auth --no-browser`, which prints the authorization URL on stderr and waits for the localhost callback.

WHOOP rotates refresh tokens. Every refresh persists the replacement before an API command continues. If storage fails, the command fails instead of silently continuing with a token that may be invalidated.

## Storage

- macOS: Keychain service `whoop-cli`, accounts `client_id`, `client_secret`, and `refresh_token`.
- Other platforms: `~/.config/whoop/credentials.json`, created with mode `0600`.
- Environment variables override stored client credentials for the current process. Tokens are never read from command-line arguments.

Run `whoop config diagnose --json` to inspect presence only. Values and tokens are never printed.

See [Troubleshooting](troubleshooting.md) for callback, token-rotation, and rate-limit recovery. See [Credential migration](migration.md) for moving from Athena's `athena-whoop` Keychain entries without exposing secrets.

## Least privilege

Use only: `read:profile`, `read:body_measurement`, `read:cycles`, `read:recovery`, `read:sleep`, `read:workout`, and `offline`. The public WHOOP API used by this project exposes account data; it does not expose direct device control. This CLI therefore has no write commands.
