# Automation contract

Use `--json` for structured data or `--plain` for deterministic tab-separated records. Data goes to stdout. Progress, prompts, and errors go to stderr.

`whoop schema --json` reports the command, flag, output, read-only, and exit-code contract.

For unattended jobs, require the read-only profile and prevent browser interaction:

```sh
whoop --help
whoop recovery --safety-profile readonly --no-input --allow-command recovery --json
```

`--allow-command` is repeatable and matches the exact command name. `--no-input` rejects `auth`, which requires a browser callback. Exit codes:

- `0`: success
- `1`: API, OAuth, or runtime failure
- `2`: usage, safety-policy, or local configuration failure

Do not pass client secrets or refresh tokens as command-line arguments. Use the OS secure store or process environment for client credentials, and authorize interactively before scheduling reads.
