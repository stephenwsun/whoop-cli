# Commands

The executable is `whoop`. Every resource command is read-only.

| Command | Purpose |
| --- | --- |
| `auth` | OAuth authorization and refresh-token setup |
| `config diagnose` | Report credential presence without secrets |
| `schema --json` | Emit the machine-readable command contract |
| `profile` | Basic WHOOP profile |
| `body` | Body measurements |
| `cycles` | Physiological cycles |
| `recovery` | Recovery scores |
| `sleep` | Sleep records |
| `workouts` | Workout records |
| `brief` | Latest recovery/sleep plus seven-day adherence |
| `week` | Seven-day workout list plus adherence |
| `weekly` | Seven-day recovery, HRV/RHR, sleep, and adherence rollup |

Resource list commands accept `--limit`, `--start`, `--end`, `--json`, and `--plain`. Pagination follows every WHOOP `next_token` and rejects repeated tokens.
