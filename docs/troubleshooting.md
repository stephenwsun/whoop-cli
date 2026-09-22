# Troubleshooting

Use `whoop config diagnose --json` first. It reports only whether each credential is present; it never prints a client secret or token.

## Keychain and file-store checks

On macOS, the default store is Keychain service `whoop-cli` with accounts `client_id`, `client_secret`, and `refresh_token`. Prefer the Keychain Access application when checking or moving entries so secrets do not enter shell history or process listings. `whoop config diagnose` is the safe command-line presence check.

On non-macOS systems, the fallback file is `~/.config/whoop/credentials.json`. The directory is created with mode `0700` and the file with mode `0600`. This fallback is protected by filesystem permissions, not an encrypted keychain; use an OS secret manager or an encrypted environment/volume when that distinction matters.

If a credential file is malformed, restore it from a protected backup or remove it and re-authorize. Never paste a token into an issue, CI log, or command-line argument.

## OAuth callback failures

- The registered redirect URI must exactly match `http://localhost:8400/callback`, including scheme, port, and path.
- Port 8400 must be free before `whoop auth` starts. Stop the process using it, then retry.
- A callback with a missing or changed state is rejected as a CSRF failure; restart authorization rather than bypassing the check.
- If the browser does not open, or the platform does not support automatic browser launch, rerun `whoop auth --no-browser`, copy the printed authorization URL only from the terminal, and open it manually. Do not share it publicly.
- OAuth client credentials must belong to the WHOOP developer application whose redirect URI is registered.

## Refresh-token rotation

WHOOP rotates refresh tokens. The CLI writes a returned replacement before completing an API command. If the process or credential store fails during that write, run `whoop auth` again rather than repeatedly retrying an old token. A refresh-token error means authorization state may need to be re-established.

## Rate limits and API errors

The client retries transient `5xx` and `429` responses, honoring `Retry-After` when present and using bounded exponential backoff otherwise. A final error includes the HTTP status, safe API message, and request ID when available; bearer tokens, access tokens, refresh tokens, and client secrets are redacted. Reduce request volume or wait for the rate-limit window before retrying manually.

## Scope and capability limits

Use only the documented read scopes. The public WHOOP API used here exposes account data and scored activities; it does not expose direct device control or the write operations needed to start/stop activities. A pending or unavailable score remains null and is not synthesized.
