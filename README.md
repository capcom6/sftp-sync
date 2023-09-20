# sftp-sync

sftp-sync is a command-line utility for syncing a local folder with a remote FTP server on every change of files or directories.

## Table of contents

- [sftp-sync](#sftp-sync)
  - [Table of contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Options](#options)
    - [Arguments](#arguments)
  - [Roadmap](#roadmap)
  - [Contributing](#contributing)
  - [License](#license)


## Features

- Continuous synchronization: Automatically syncs local changes to the remote FTP server whenever files or directories are added, modified, or deleted.
- Exclude paths: Allows you to exclude specific paths from being synced.
- Easy to use: Simple and intuitive command-line interface.

## Installation

You can download the pre-built binary for your operating system from the [Releases](https://github.com/capcom6/sftp-sync/releases) section of the GitHub repository.

If you prefer to build from source, make sure you have Go installed. If not, you can download it from the official Go website: https://golang.org/dl/

Then, follow these steps:

1. Clone the sftp-sync repository:
    ```shell
    git clone https://github.com/capcom6/sftp-sync.git
    ```

2. Build the sftp-sync binary using the following command:
    ```shell
    cd sftp-sync
    go build -o sftp-sync cmd/sftp-sync/main.go
    ```

3. The binary will be generated in the current directory. You can move it to a location in your PATH for easy access.


## Usage

Run the `sftp-sync` command followed by the necessary options and arguments:

```shell
ftp-sync --dest=ftp://username:password@hostname:port/path/to/remote/folder --exclude=.git /path/to/local/folder
```

### Options

- `--dest`: The destination FTP server URL. It should follow the format `ftp://username:password@hostname:port/path/to/remote/folder`.
- `--exclude`: (Optional) Specifies paths or patterns to exclude from the synchronization process. You can specify multiple `--exclude` options to exclude multiple paths or patterns.

### Arguments

- The local folder path: The path to the local folder you want to sync with the remote FTP server.

## Roadmap

Here are some ideas and suggestions for future releases:

- [ ] Support for patterns in the `--exclude` option.
- [ ] Support of Secure FTP (SFTP) protocol.
- [ ] Improved error handling and error messages.
- [ ] Integration with Git for automatic syncing on commit or branch changes.
- [ ] Integration with Git for linking branch to remote server.
- [ ] Support for other remote protocols such as S3.
- [ ] Support for syncing specific file types or file name patterns.
- [ ] Preserve attributes (if available).
- [ ] Parallel sync in multiple threads.
- [ ] Batching events for more effective sync on frequently changes.

Feel free to open an issue or submit a pull request if you have any other ideas or suggestions!

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
