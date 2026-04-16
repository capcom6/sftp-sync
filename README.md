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
    A command-line utility for syncing a local folder with a remote FTP server on every change of files or directories.
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
  - [Environment Variables](#environment-variables)
  - [Global Options](#global-options)
  - [Sync Command Options](#sync-command-options)
  - [Sync Command Arguments](#sync-command-arguments)
  - [Error Handling](#error-handling)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)

<!-- ABOUT THE PROJECT -->
## About The Project

<!-- [![Product Name Screen Shot][product-screenshot]](https://example.com) -->

sftp-sync is a command-line utility for syncing a local folder with a remote FTP server on every change of files or directories.

### Features

- Continuous synchronization: Automatically syncs local changes to the remote FTP server whenever files or directories are added, modified, or deleted.
- Exclude paths: Allows you to exclude specific paths from being synced.
- Easy to use: Simple and intuitive command-line interface.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

* [![Go][Go.dev]][Go-url]
* [![urfave/cli][urfave-cli-v3]][urfave-cli-v3-url]
* [![fsnotify][fsnotify]][fsnotify-url]
* [![joho/godotenv][godotenv]][godotenv-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- INSTALLATION -->
## Installation

### Prerequisites

- Go 1.24.3 or higher installed on your system
- Access to an FTP server with valid credentials

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

```shell
sftp-sync --dest=ftp://username:password@hostname:port/path/to/remote/folder \
  --exclude=.git /path/to/local/folder
```

### Environment Variables

- `DEBUG`: When set to any value, enables debug mode (equivalent to `--debug` flag).

### Global Options

- `--debug`: Enable debug mode (can also be set via `DEBUG` environment variable).
- `--version`: Print version information.

### Sync Command Options

- `--dest`: The destination FTP server URL. It should follow the format `ftp://username:password@hostname:port/path/to/remote/folder`.
- `--exclude`: (Optional) Specifies paths or glob patterns to exclude from synchronization. Supports `*`, `**`, and `?`. You can specify multiple `--exclude` options.

### Sync Command Arguments

- `source`: The local folder path to watch for changes (required positional argument).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Error Handling

The application uses structured error handling with specific exit codes:

- `0`: Success - operation completed successfully
- `1`: Parameters Error - invalid command arguments or options
- `2`: Client Error - FTP client connection or operation failed
- `3`: Output Error - logging or output system failed
- `4`: Internal Error - unexpected internal error

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [x] Support for patterns in the `--exclude` option.
- [ ] Support of Secure FTP (SFTP) protocol.
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
[go-report-card-shield]: https://goreportcard.com/badge/github.com/capcom6/sftp-sync
[go-report-card-url]: https://goreportcard.com/report/github.com/capcom6/sftp-sync
[go-version-shield]: https://img.shields.io/github/go-mod/go-version/capcom6/sftp-sync?style=for-the-badge
[go-version-url]: https://github.com/capcom6/sftp-sync/blob/master/go.mod
[code-size-shield]: https://img.shields.io/github/languages/code-size/capcom6/sftp-sync?style=for-the-badge
[code-size-url]: https://github.com/capcom6/sftp-sync
[last-commit-shield]: https://img.shields.io/github/last-commit/capcom6/sftp-sync?style=for-the-badge
[last-commit-url]: https://github.com/capcom6/sftp-sync/commits/master
