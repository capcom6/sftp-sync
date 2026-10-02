# Changelog

All notable changes to `sftp-sync` are documented in this file.

Entries follow the release history from newest to oldest. Dates are release (tag) dates.

## [Unreleased]

### Maintenance

- **Minimum Go version raised to 1.26.0** — building from source now requires Go 1.26.0 or newer (`go.mod` updated from 1.25.0).
- Updated dependencies: `fsnotify` 1.6.0 → 1.10.1, `urfave/cli` v3 3.7.0 → 3.13.0, `golang.org/x/crypto` 0.52.0 → 0.57.0, `jlaffaye/ftp` 0.2.0 → 0.2.4, `pkg/sftp` 1.13.10 → 1.13.11, `doublestar` 4.10.0 → 4.10.2, `samber/lo` 1.52.0 → 1.53.0.

## [1.4.0] - 2026-05-28

### New Features

#### SFTP Support

- **SFTP protocol** — `--dest` now accepts `sftp://` URLs in addition to `ftp://`, so the same watcher can sync to SSH-based servers:
  ```bash
  # password
  sftp-sync --dest=sftp://username:password@hostname:22/remote/path /local/path

  # SSH private key (supports ~ expansion and passphrases)
  sftp-sync --dest="sftp://username@hostname:22/remote/path?key=~/.ssh/id_ed25519&key_pass=secret" /local/path

  # SSH agent
  sftp-sync --dest="sftp://username@hostname:22/remote/path?agent=true" /local/path
  ```
  - Authentication methods are tried in order: SSH agent → private key → password; a default key from `~/.ssh/id_ed25519`, `~/.ssh/id_ecdsa`, or `~/.ssh/id_rsa` is auto-detected only when no SSH agent signer, explicit key, or password is available.
  - Host keys are verified against `~/.ssh/known_hosts`.
  - New `timeout` query parameter controls the connection timeout (Go duration, default `30s`), e.g. `?timeout=45s`.
  - The remote path in the URL must already exist and be a directory — an SFTP connection fails with a clear error if it does not.

## [1.3.0] - 2026-05-05

### New Features

#### Dry Run Mode

- **Dry run** — new `--dry-run` flag reports what would be created, modified, or removed without uploading or deleting anything on the remote server:
  ```bash
  sftp-sync --dry-run --dest=ftp://username:password@hostname/path /local/path
  ```
  Each detected change is logged as `Would create`, `Would modify`, or `Would remove` with the relative path.

## [1.2.0] - 2026-04-24

### New Features

#### Glob Patterns in `--exclude`

- **Glob support for `--exclude`** — exclude rules now understand `*`, `**`, and `?` (matched against paths relative to the watched folder), so entire trees can be skipped with a single pattern:
  ```bash
  sftp-sync --exclude='*.log' --exclude='node_modules/**' \
    --dest=ftp://username:password@hostname/path /local/path
  ```
  Plain path excludes keep working as before.

### Bug Fixes

- Cancelling startup no longer hangs on large directory trees — the watcher stops promptly when the process receives `Ctrl+C` while it is still adding directories recursively.

## [1.1.0] - 2026-03-27

### Maintenance

- Migrated the command-line layer to urfave/cli v3. Flags and arguments are unchanged (`--dest`, `--exclude`, `--debug`, `--version`, and the `source` positional argument).

## [1.0.3] - 2026-02-12

### Bug Fixes

- `go install github.com/capcom6/sftp-sync@latest` works again — the `main` package was moved from `cmd/sftp-sync/` to the repository root.

## [1.0.2] - 2023-10-08

### New Features

- **Interruptible recursive sync** — pressing `Ctrl+C` now stops an in-progress recursive sync instead of letting it run to completion.

### Bug Fixes

- Creating a file on Windows no longer triggers a spurious recursive re-sync of its parent directory (Windows emits an extra write event for the directory).

### Maintenance

- Improved logging output.

## [1.0.1] - 2023-09-21

### Bug Fixes

- FTP sessions now reconnect automatically after the server drops the connection, instead of failing subsequent syncs.
- Fixed handling of Windows path separators when building local and remote paths.
- Sync now resolves the watched folder to an absolute root path, fixing incorrect relative paths for some setups.

## [1.0.0] - 2023-09-17

### New Features

- **Initial release** — continuous one-way sync of a local folder to a remote FTP server:
  - Watches the source directory for file and directory changes (create, modify, delete) and mirrors them to the destination.
  - Recursive sync that creates all missing remote directories.
  - `--dest` destination URL and repeatable `--exclude` path rules.
  - Operation logging and `--version` build information.
