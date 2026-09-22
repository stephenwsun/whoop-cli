# Credential migration

## From Athena's WHOOP helper

The Athena helper used macOS Keychain service `athena-whoop`:

- `client-creds`: one `id:secret` value
- `refresh-token`: the offline WHOOP refresh token

`whoop-cli` intentionally uses a separate service, `whoop-cli`, and stores client ID, client secret, and refresh token as separate accounts. This prevents the new CLI from mutating Athena's credentials while both tools coexist.

The safest migration is to authorize a fresh token:

1. Retrieve the WHOOP application's client ID and secret from the existing application credential record using Keychain Access, not shell output.
2. Export them only for the current process:

   ```sh
   export WHOOP_CLIENT_ID='...'
   export WHOOP_CLIENT_SECRET='...'
   ```

3. Run `whoop auth` and approve access.
4. Remove the temporary environment variables after authorization. The new refresh token is stored under `whoop-cli`.
5. Verify presence without printing values:

   ```sh
   whoop config diagnose --json
   ```

Do not copy Athena's refresh token by putting it in a shell command or a file under source control. Re-authorizing is preferred because WHOOP rotates refresh tokens and the two tools may otherwise invalidate each other's stored token.

## Coexistence and rollback

Keep `athena-whoop` untouched until the new CLI has produced the expected read-only summaries. Removing `whoop-cli` credentials does not remove Athena credentials. If migration fails, unset the environment variables and continue using Athena; then retry `whoop auth` after correcting the developer-app or redirect configuration.

The CLI does not modify Athena and does not migrate or delete Keychain entries automatically.
