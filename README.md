<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Code Size][code-size-shield]][code-size-url]
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Go Report Card][go-report-card-shield]][go-report-card-url]
[![Go Version][go-version-shield]][go-version-url]
[![Issues][issues-shield]][issues-url]
[![Last Commit][last-commit-shield]][last-commit-url]
[![License][license-shield]][license-url]
[![Stargazers][stars-shield]][stars-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/capcom6/sftp-sync">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">sftp-sync</h3>

  <p align="center">
    A command-line utility for syncing a local folder with a remote FTP or SFTP server on every change of files or directories.
    <br />
    <br />
    <a href="https://github.com/capcom6/sftp-sync/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/capcom6/sftp-sync/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
- [About The Project](#about-the-project)
  - [Features](#features)
  - [Built With](#built-with)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Installation Methods](#installation-methods)
    - [Method 1: Using Go Install (Recommended)](#method-1-using-go-install-recommended)
    - [Method 2: Using Release Binaries](#method-2-using-release-binaries)
    - [Method 3: Building from Source](#method-3-building-from-source)
- [Usage](#usage)
  - [Global Options](#global-options)
  - [Options](#options)
  - [Arguments](#arguments)
  - [Error Handling](#error-handling)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [The `.env` File](#the-env-file)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)

<!-- ABOUT THE PROJECT -->
## About The Project

<!-- [![Product Name Screen Shot][product-screenshot]](https://example.com) -->

sftp-sync is a command-line utility for syncing a local folder with a remote FTP or SFTP server on every change of files or directories.

### Features

- Continuous synchronization: Automatically syncs local changes to the remote FTP or SFTP server whenever files or directories are added, modified, or deleted, including recursive changes in subdirectories.
- Exclude paths: Allows you to exclude specific paths from being synced, with support for glob patterns (`*`, `**`, `?`).
- Dry run mode: Preview what would be created, modified, or removed without uploading anything (`--dry-run`).
- Easy to use: Simple and intuitive command-line interface.
- Protocol support: Supports both FTP and SFTP (SSH File Transfer Protocol).
- Flexible SFTP authentication: Password, SSH private key (optionally passphrase-protected), or SSH agent.
- Graceful shutdown: Stops cleanly on `Ctrl+C` (SIGINT) or `SIGTERM`, including during a recursive sync.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

* [![Go][Go.dev]][Go-url]
* [![urfave/cli][urfave-cli-v3]][urfave-cli-v3-url]
* [![fsnotify][fsnotify]][fsnotify-url]
* [![jlaffaye/ftp][jlaffaye-ftp]][jlaffaye-ftp-url]
* [![pkg/sftp][pkg-sftp]][pkg-sftp-url]
* [![bmatcuk/doublestar][doublestar]][doublestar-url]
* [![joho/godotenv][godotenv]][godotenv-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- INSTALLATION -->
## Installation

### Prerequisites

- Go 1.26.0 or higher installed on your system
- Access to an (S)FTP server with valid credentials

### Installation Methods

#### Method 1: Using Go Install (Recommended)

Install the latest version directly from the repository:

```shell
go install github.com/capcom6/sftp-sync@latest
```

This will install `sftp-sync` to your `$GOBIN` directory. Make sure your `$GOBIN` is in your `$PATH`.

#### Method 2: Using Release Binaries

Download the pre-compiled binaries from the [GitHub Releases](https://github.com/capcom6/sftp-sync/releases) page:

1. Download the binary for your operating system and architecture
2. Make the binary executable:
   ```shell
   chmod +x sftp-sync
   ```
3. Move it to a directory in your `$PATH`:
   ```shell
   sudo mv sftp-sync /usr/local/bin/
   ```

#### Method 3: Building from Source

If you prefer to build from source:

```shell
git clone https://github.com/capcom6/sftp-sync.git
cd sftp-sync
make build
```

The binary will be available in the `bin/` directory.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage
Run the `sftp-sync` command with the necessary options and arguments:

**FTP:**
```shell
sftp-sync --dest=ftp://username:password@hostname:port/path/to/remote/folder \
  --exclude=.git /path/to/local/folder
```

**SFTP (password):**
```shell
sftp-sync --dest=sftp://username:password@hostname:22/path/to/remote/folder \
  --exclude=.git /path/to/local/folder
```

**SFTP (SSH key):**
```shell
sftp-sync --dest="sftp://username@hostname:22/path/to/remote/folder?key=~/.ssh/id_ed25519" \
  --exclude=.git /path/to/local/folder
```

**SFTP (SSH agent):**
```shell
sftp-sync --dest="sftp://username@hostname:22/path/to/remote/folder?agent=true" \
  --exclude=.git /path/to/local/folder
```

**Dry run (preview without uploading):**
```shell
sftp-sync --dry-run --dest=ftp://username:password@hostname:port/path/to/remote/folder \
  --exclude=.git /path/to/local/folder
```

> **Note:** SFTP uses SSH port 22 by default (vs FTP port 21).

### Global Options

- `--debug`: Enable debug mode (can also be set via `DEBUG` environment variable).
- `--version`: Print version information.

### Options

- `--dest`: The destination server URL. Supports both FTP and SFTP:
  - FTP: `ftp://username:password@hostname:port/path/to/remote/folder`
  - SFTP (password): `sftp://username:password@hostname:22/path/to/remote/folder`
  - SFTP (SSH key): `sftp://username@hostname:22/path?key=~/.ssh/id_ed25519`
  - SFTP (SSH key with passphrase): `sftp://username@hostname:22/path?key=~/.ssh/id_ed25519&key_pass=<passphrase>`
  - SFTP (SSH agent): `sftp://username@hostname:22/path?agent=true`

The URL path (e.g., `/path/to/remote/folder`) is used as the remote destination prefix. All synced files and directories are placed relative to this path. **The remote directory must already exist** before starting the sync.

SFTP URL query parameters:

| Parameter  | Description                                                                                                                                                   | Example                 |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| `key`      | Path to SSH private key file (supports `~` expansion). If omitted, auto-detects a default key only when no password is set and agent auth provides no signers | `key=~/.ssh/custom_key` |
| `key_pass` | Passphrase for encrypted private keys                                                                                                                         | `key_pass=mysecret`     |
| `agent`    | Use SSH agent for authentication when set to `true` (requires `SSH_AUTH_SOCK`)                                                                                | `agent=true`            |
| `timeout`  | Connection timeout as a Go duration (e.g., `30s`, `1m`)                                                                                                       | `timeout=45s`           |

Authentication methods are tried in order: SSH agent → private key → password. If agent setup fails, the error is returned instead of falling back to a default key. SFTP host keys are verified against `~/.ssh/known_hosts`.

> **Security note:** Avoid putting real passwords/passphrases directly in CLI arguments when possible,
> as they can be exposed via shell history and process listings.

- `--exclude`: (Optional) Specifies paths or glob patterns to exclude from synchronization. Supports `*`, `**`, and `?`. You can specify multiple `--exclude` options.
- `--dry-run`: (Optional) Log what would be created, modified, or removed without uploading or deleting anything on the remote server.

### Arguments

- `source`: The local folder path to watch for changes (required positional argument).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Error Handling

The application uses structured error handling with specific exit codes:

- `0`: Success - operation completed successfully, including an interrupted sync stopped with `Ctrl+C`
- `1`: Parameters Error - invalid positional arguments (missing or extra `source`), or an empty `--dest`
- `2`: Client Error - the destination URL could not be parsed or uses an unsupported scheme (only `ftp` and `sftp` are supported)
- `3`: Output Error - reserved for logging or output system failures
- `4`: Internal Error - unexpected internal errors, including flag errors reported by the CLI framework (e.g., a missing required `--dest` or an unknown flag)

Failures that occur **while syncing** (for example, the remote server is unreachable) are logged and the watcher keeps running; the process does not exit.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONFIGURATION -->
## Configuration

### Environment Variables

| Variable        | Purpose                                                                                                             | Format                                                                                                     | Default                       | Example                      |
| --------------- | ------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------- | ---------------------------- |
| `DEBUG`         | Enables debug logging (equivalent to the `--debug` flag).                                                           | Boolean accepted by Go's `strconv.ParseBool`: `1`, `t`, `T`, `true`, `TRUE`, `True`. `0`/`false` disables. | unset (disabled)              | `DEBUG=1`                    |
| `SSH_AUTH_SOCK` | Path to the SSH agent socket. Read-only — provided by `ssh-agent`; used when the destination URL has `?agent=true`. | Unix socket path                                                                                           | set by the system/`ssh-agent` | `/tmp/ssh-XXXXXX/agent.1234` |

> **Note:** an unparseable `DEBUG` value (e.g. `DEBUG=yes`) makes the command fail at startup with exit code `4`.

### The `.env` File

On startup the application loads environment variables from a `.env` file in the current working directory (via [godotenv](https://github.com/joho/godotenv)); a missing file is ignored. This is a convenient way to set `DEBUG` without typing it on every run:

```dotenv
DEBUG=1
```

See [.env.example](.env.example) for a documented template.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [x] Support for patterns in the `--exclude` option.
- [x] Support of Secure FTP (SFTP) protocol.
- [ ] Improved error handling and error messages.
- [ ] Integration with Git for automatic syncing on commit or branch changes.
- [ ] Integration with Git for linking branch to remote server.
- [ ] Support for other remote protocols such as S3.
- [ ] Support for syncing specific file types or file name patterns.
- [ ] Preserve attributes (if available).
- [ ] Parallel sync in multiple threads.
- [ ] Batching events for more effective sync on frequently changes.

See the [open issues](https://github.com/capcom6/sftp-sync/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open-source community a great place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion to improve this project, please fork the repository and open a pull request. You can also open an issue with the `enhancement` label.  
If this project is useful to you, consider starring it.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the Apache License 2.0. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

Project Link: [https://github.com/capcom6/sftp-sync](https://github.com/capcom6/sftp-sync)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [Best-README-Template](https://github.com/othneildrew/Best-README-Template)
* [urfave/cli](https://github.com/urfave/cli)
* [fsnotify](https://github.com/fsnotify/fsnotify)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/capcom6/sftp-sync.svg?style=for-the-badge
[contributors-url]: https://github.com/capcom6/sftp-sync/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/capcom6/sftp-sync.svg?style=for-the-badge
[forks-url]: https://github.com/capcom6/sftp-sync/network/members
[stars-shield]: https://img.shields.io/github/stars/capcom6/sftp-sync.svg?style=for-the-badge
[stars-url]: https://github.com/capcom6/sftp-sync/stargazers
[issues-shield]: https://img.shields.io/github/issues/capcom6/sftp-sync.svg?style=for-the-badge
[issues-url]: https://github.com/capcom6/sftp-sync/issues
[license-shield]: https://img.shields.io/github/license/capcom6/sftp-sync.svg?style=for-the-badge
[license-url]: https://github.com/capcom6/sftp-sync/blob/master/LICENSE
[product-screenshot]: images/screenshot.png
<!-- Shields.io badges. You can a comprehensive list with many more badges at: https://github.com/inttter/md-badges -->
[Go.dev]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://golang.org/
[urfave-cli-v3]: https://img.shields.io/badge/urfave%2Fcli-00ADD8?style=for-the-badge&logo=go&logoColor=white
[urfave-cli-v3-url]: https://github.com/urfave/cli
[fsnotify]: https://img.shields.io/badge/fsnotify-00ADD8?style=for-the-badge&logo=go&logoColor=white
[fsnotify-url]: https://github.com/fsnotify/fsnotify
[godotenv]: https://img.shields.io/badge/joho%2Fgodotenv-00ADD8?style=for-the-badge&logo=go&logoColor=white
[godotenv-url]: https://github.com/joho/godotenv
[jlaffaye-ftp]: https://img.shields.io/badge/jlaffaye%2Fftp-00ADD8?style=for-the-badge&logo=go&logoColor=white
[jlaffaye-ftp-url]: https://github.com/jlaffaye/ftp
[pkg-sftp]: https://img.shields.io/badge/pkg%2Fsftp-00ADD8?style=for-the-badge&logo=go&logoColor=white
[pkg-sftp-url]: https://github.com/pkg/sftp
[doublestar]: https://img.shields.io/badge/bmatcuk%2Fdoublestar-00ADD8?style=for-the-badge&logo=go&logoColor=white
[doublestar-url]: https://github.com/bmatcuk/doublestar
[go-report-card-shield]: https://goreportcard.com/badge/github.com/capcom6/sftp-sync
[go-report-card-url]: https://goreportcard.com/report/github.com/capcom6/sftp-sync
[go-version-shield]: https://img.shields.io/github/go-mod/go-version/capcom6/sftp-sync?style=for-the-badge
[go-version-url]: https://github.com/capcom6/sftp-sync/blob/master/go.mod
[code-size-shield]: https://img.shields.io/github/languages/code-size/capcom6/sftp-sync?style=for-the-badge
[code-size-url]: https://github.com/capcom6/sftp-sync
[last-commit-shield]: https://img.shields.io/github/last-commit/capcom6/sftp-sync?style=for-the-badge
[last-commit-url]: https://github.com/capcom6/sftp-sync/commits/master
