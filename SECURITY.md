# Security policy

## Supported versions

Security fixes target the latest release and the default branch. Older releases
may not receive fixes; upgrade before reporting a suspected issue.

## Reporting a vulnerability

Do not open a public issue for a vulnerability or include credentials, refresh
tokens, access tokens, client secrets, or private WHOOP data in any report.
Use GitHub's **Report a vulnerability** action on the repository's Security tab.
If private reporting is unavailable, contact the repository maintainer through
the GitHub profile before disclosing details publicly.

Include the affected version or commit, operating system, reproducible steps
with secrets removed, impact, and any proposed mitigation. We will acknowledge
reports when practicable, investigate privately, and coordinate disclosure
once a fix or mitigation is available.

## Credential safety

`whoop-cli` stores refresh tokens in the macOS Keychain when available and uses
a mode-0600 file fallback on other platforms. Treat the credential store as
sensitive. Never pass tokens or client secrets as command-line arguments, and
redact them from logs, bug reports, and test fixtures.
