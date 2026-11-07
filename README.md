# News Crawler

A Go project scaffold for building a news crawler. The application is not
implemented yet: this repository currently contains the module declaration and
project configuration, but no crawler, executable, or runtime configuration.

## Requirements

- Go 1.27.3, as declared in [`go.mod`](go.mod)

## Getting started

Clone the repository and enter its directory:

```sh
git clone <repository-url>
cd news_crawler
```

The application is not runnable until Go source code and an entry point are
added. Once packages are present, run the test suite with:

```sh
go test ./...
```

## Repository layout

| Path | Purpose |
| --- | --- |
| [`go.mod`](go.mod) | Go module name and language version |
| [`AGENTS.md`](AGENTS.md) | Guidance for contributors and coding agents |
| [`Dockerfile`](Dockerfile) | Initial container build configuration |
| [`MAKEFILE`](MAKEFILE) | Make target for building the container image |
| [`.github/workflows/`](.github/workflows/) | GitHub Actions workflows |

The container build and application run workflow may need updating as the
application structure and entry point are added.

## Development

- Format Go source files with `gofmt`.
- Add tests alongside new functionality and run `go test ./...`.
- Update this README when setup, supported commands, or crawler behavior changes.
