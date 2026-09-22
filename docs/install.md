# Install and upgrade

## GitHub Releases

Download the archive for your operating system and architecture from the
[latest release](https://github.com/stephenwsun/whoop-cli/releases/latest).
Release assets use these names:

- `whoop-cli_<version>_darwin_amd64.tar.gz`
- `whoop-cli_<version>_darwin_arm64.tar.gz`
- `whoop-cli_<version>_linux_amd64.tar.gz`
- `whoop-cli_<version>_linux_arm64.tar.gz`
- `whoop-cli_<version>_windows_amd64.zip`
- `whoop-cli_<version>_windows_arm64.zip`
- `checksums.txt`

Verify the downloaded archive before extracting it:

```sh
curl -fsSLO https://github.com/stephenwsun/whoop-cli/releases/latest/download/checksums.txt
sha256sum --ignore-missing -c checksums.txt
```

On macOS, use `shasum -a 256 -c checksums.txt` instead. Put the `whoop` binary
on `PATH`, then verify it:

```sh
whoop --version
whoop --help
whoop config diagnose --json
```

## Go install

```sh
go install github.com/stephensun/whoop-cli/cmd/whoop@latest
```

Go-installed binaries identify themselves as development builds unless built
from a release workflow. Use GitHub Release archives when you need a published
version and checksum.

## Build from source

```sh
git clone https://github.com/stephenwsun/whoop-cli.git
cd whoop-cli
make ci
make build
./bin/whoop --version
```

## Upgrade and uninstall

Replace the binary with the new release archive or run `go install` again.
Refresh tokens and client credentials remain in the configured secure store;
see [authentication](auth.md) and [migration](migration.md) before changing
credential backends.

To uninstall, remove the binary from `PATH`. Remove stored credentials only
when you intend to disconnect the app; `whoop-cli` does not provide a destructive
credential-delete command so users must remove the `whoop-cli` Keychain entries
or `~/.config/whoop/credentials.json` deliberately.
