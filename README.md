# git-download

A powerful CLI tool that downloads and synchronizes public Git repositories using ZIP files instead of git clone. Perfect for environments where direct git clone operations are restricted or when you need a lightweight alternative.

[![Go Version](https://img.shields.io/github/go-mod/go-version/pimentel/git-download)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/pimentel/git-download)](https://hub.docker.com/r/pimentel/git-download)

## 🚀 Features

- **ZIP-based Downloads**: Download repositories without using git clone
- **Multiple Repository Management**: Track and sync multiple repositories
- **Branch & Tag Support**: Download specific branches or release tags
- **Sync Status Tracking**: Monitor last sync times and repository states
- **Cross-Platform**: Works on Windows, Linux, macOS, and Docker
- **Simple Metadata**: Easy-to-understand JSON-based tracking

## 📋 Prerequisites

- Go 1.18 or later (for building from source)
- Docker (optional, for running containerized)

## 🔧 Installation

### From Source

1. Clone this repository:
```bash
git clone https://github.com/pimentel/git-download
cd git-download
```

2. Build the binary:
```bash
go build -o git-download ./cmd/git-download
```

3. (Optional) Move to your PATH:
```bash
# Linux/macOS
sudo mv git-download /usr/local/bin/

# Windows (PowerShell as Administrator)
move git-download.exe C:\Windows\System32\
```

### Using Docker

```bash
# Pull the image
docker pull ghcr.io/pimentel/git-download:latest

# Run the tool (mounting current directory for metadata and downloads)
docker run -v $(pwd):/data ghcr.io/pimentel/git-download:latest [command]

# Example: initialize a repository
docker run -v $(pwd):/data ghcr.io/pimentel/git-download:latest init \
  --url https://github.com/user/repo \
  --ref main \
  --ref-type branch \
  --destination ./local-folder
```

For ARM64 systems (like Apple M1/M2):
```bash
docker pull ghcr.io/pimentel/git-download:latest-arm64
```

### Using Go Install

```bash
go install github.com/pimentel/git-download/cmd/git-download@latest
```

### Pre-built Binaries

Download the appropriate binary for your system from the [Releases](https://github.com/pimentel/git-download/releases) page:

- Windows (AMD64): `git-download_windows_amd64.exe`
- Linux (AMD64): `git-download_linux_amd64`
- macOS (AMD64): `git-download_darwin_amd64`
- macOS (ARM64/M1): `git-download_darwin_arm64`

## 🎯 Usage

### Initialize a Repository

Using a branch:
```bash
git-download init --url https://github.com/user/repo --ref main --ref-type branch --destination ./local-folder
```

Using a tag:
```bash
git-download init --url https://github.com/user/repo --ref v1.0.0 --ref-type tag --destination ./local-folder
```

Options:
- `--url`: Repository URL (required)
- `--ref`: Repository reference (branch name or tag, default: "main")
- `--ref-type`: Type of reference ("branch" or "tag", default: "branch")
- `--destination`: Local destination path
- `--name`: Custom name for the repository (defaults to URL basename)

### Sync Repositories

Sync all tracked repositories:
```bash
git-download sync
```

Sync a specific repository:
```bash
git-download sync --name repo-name
```

### Check Status

View the status of all tracked repositories:
```bash
git-download status
```

Example output:
```
Tracked repositories:
- example-repo:
  URL: https://github.com/user/example-repo
  branch: main
  Destination: ./repos/example
  Last Sync: 2024-02-06T10:15:00Z

- stable-repo:
  URL: https://github.com/user/stable-repo
  tag: v1.0.0
  Destination: ./repos/stable
  Last Sync: 2024-02-06T09:00:00Z
```

### Remove a Repository

Remove a repository from tracking:
```bash
git-download remove --name repo-name
```

Remove a repository and delete local files:
```bash
git-download remove --name repo-name --delete-local
```

## 📁 Metadata

The tool maintains a `.syncmeta.json` file in the current directory with the following information:

```json
{
  "repositories": [
    {
      "name": "example-repo",
      "url": "https://github.com/user/example-repo",
      "branch": "main",
      "refType": "branch",
      "lastSync": "2024-02-06T10:15:00Z",
      "destination": "./repos/example"
    }
  ]
}
```

## 🛠 Building for Multiple Platforms

### Using Docker

Build the Docker image locally:
```bash
docker build -t git-download .
```

Run the tool using Docker:
```bash
docker run -v $(pwd):/data git-download [command]
```

### Using GoReleaser

1. Install GoReleaser:
```bash
# macOS
brew install goreleaser

# Linux
curl -sfL https://goreleaser.com/static/run | bash
```

2. Create `.goreleaser.yml`:
```yaml
builds:
  - binary: git-download
    goos:
      - windows
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    replacements:
      darwin: macOS
    files:
      - README.md
      - LICENSE

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - Merge pull request
      - Merge branch
```

3. Build and release:
```bash
# Test build
goreleaser build --snapshot --rm-dist

# Release (requires GITHUB_TOKEN)
goreleaser release --rm-dist
```

### Using Make

Create a Makefile for easy cross-compilation:

```makefile
BINARY_NAME=git-download
VERSION=$(shell git describe --tags --always)
BUILD_DIR=dist

.PHONY: all windows linux darwin clean

all: windows linux darwin

windows:
	GOOS=windows GOARCH=amd64 go build -o ${BUILD_DIR}/${BINARY_NAME}_windows_amd64.exe ./cmd/git-download

linux:
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR}/${BINARY_NAME}_linux_amd64 ./cmd/git-download

darwin:
	GOOS=darwin GOARCH=amd64 go build -o ${BUILD_DIR}/${BINARY_NAME}_darwin_amd64 ./cmd/git-download
	GOOS=darwin GOARCH=arm64 go build -o ${BUILD_DIR}/${BINARY_NAME}_darwin_arm64 ./cmd/git-download

clean:
	rm -rf ${BUILD_DIR}
```

Then build for all platforms:
```bash
make all
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
