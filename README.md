# News Crawler

A concurrent Go crawler that fetches configured seed pages, discovers links on
the same host, and follows them up to each seed's configured depth.

## Requirements

- Go 1.27.3, as declared in [`go.mod`](go.mod)

## Getting started

Clone the repository and enter its directory:

```sh
git clone <repository-url>
cd news_crawler
```

Run the crawler:

```sh
go run ./cmd
```

Each configured seed has a maximum link depth. Depth `0` fetches only the seed;
depth `1` also fetches pages linked directly from it. The crawler stays on the
seed's hostname, removes URL fragments, and uses up to eight concurrent
workers. It prints each page's depth, HTTP status, discovered-link count, and
URL. Fetch failures are printed and returned as a non-zero process exit.

Seed URLs and their depths are configured in `internal/config.go`. Run tests
with:

```sh
go test ./...
```

## Repository layout

| Path | Purpose |
| --- | --- |
| [`go.mod`](go.mod) | Go module name and language version |
| [`AGENTS.md`](AGENTS.md) | Guidance for contributors and coding agents |
| [`Dockerfile`](Dockerfile) | Multi-stage container build configuration |
| [`MAKEFILE`](MAKEFILE) | Make target for building the container image |
| [`.github/workflows/`](.github/workflows/) | GitHub Actions workflows |

Build and run the container with:

```sh
docker build -t news_crawler:latest .
docker run --rm news_crawler:latest
```

## Publishing to Docker Hub

Create a **private** Docker Hub repository named `news_crawler`, then add these
repository secrets under **Settings → Secrets and variables → Actions**:

- `DOCKERHUB_USERNAME`: the Docker Hub account that owns the repository.
- `DOCKERHUB_TOKEN`: a Docker Hub access token with permission to push images.

The `Publish Docker image` workflow publishes on pushes to `main` and tags
starting with `v`. Main-branch pushes update the `main`, `latest`, and
commit-SHA tags; version-tag pushes publish the version and commit-SHA tags.
Pull requests continue to build the image without publishing it.

## Development

- Format Go source files with `gofmt`.
- Run `go test ./...` after Go code changes.
- Update this README when setup, supported commands, or crawler behavior changes.
