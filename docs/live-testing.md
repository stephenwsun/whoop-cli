# Live WHOOP testing

The live suite is opt-in and requires a WHOOP developer application plus an
authorized account. Ordinary CI never needs credentials.

Configure client credentials using the secure store or environment variables,
then run a bounded recent-window smoke test:

```sh
WHOOP_IT=1 go test -tags=integration ./internal/integration
```

The suite reads profile, body measurements, cycles, recovery, sleep, and
workouts. Empty body-measurement results are valid. Keep the test window short
to avoid pagination churn and WHOOP rate limits. Do not print or commit profile
fields, access tokens, refresh tokens, client secrets, or raw personal records.

For an installed binary, validate non-secret behavior without an account:

```sh
whoop --version
whoop --help
whoop schema --json
whoop config diagnose --json
```

A live test failure is not evidence that a credential or token should be pasted
into an issue. Redact output and report the version, platform, command, and
sanitized error instead.
