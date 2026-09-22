# Automation contract

Use `--json` for structured data or `--plain` for deterministic tab-separated records. Data goes to stdout. Progress, prompts, and warnings go to stderr.

`whoop schema --json` is the versioned machine-readable command contract. The schema version changes only when the command or output contract requires a compatibility decision. `whoop --version` reports build metadata without contacting WHOOP.

For unattended jobs, require the read-only profile, prevent browser interaction, and allowlist the exact command:

```sh
whoop --help
whoop recovery --safety-profile readonly --no-input --allow-command recovery --json
```

`--allow-command` is repeatable and matches the exact command name. `--no-input` rejects `auth`, which requires a browser callback. `auth --no-browser` prints the authorization URL for environments that cannot launch a browser and waits for the localhost callback.

Exit codes:

- `0`: success
- `1`: API, OAuth, or runtime failure
- `2`: usage, safety-policy, or local configuration failure

`--json` and `--plain` are mutually exclusive. Do not pass client secrets or refresh tokens as command-line arguments. Use the OS secure store or process environment for client credentials, and authorize interactively before scheduling reads.
