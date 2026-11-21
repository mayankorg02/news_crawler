# News Crawler

A Go crawler that renders configured seed pages in headless Chromium, saves
their HTML, discovers links on the same host, and follows them up to each
seed's configured depth.

## Requirements

- Docker

The crawler and its Chromium browser run inside Docker. Building the image
downloads Go dependencies only into the build container. Chromium runs in the
image, and downloaded pages are kept in a Docker-managed volume.

## Getting started

```sh
docker build -t news_crawler:latest .
docker volume create news-crawler-data
docker run --rm -v news-crawler-data:/data news_crawler:latest
```

The named Docker volume keeps downloaded page HTML in Docker-managed storage.
The image build runs the Go test suite before compiling the crawler.

Each configured seed has a maximum link depth. Depth `0` fetches only the seed;
depth `1` also fetches pages linked directly from it. The crawler stays on the
seed's hostname, removes URL fragments, renders pages in Chromium, and spaces
requests to each host by at least one second. It saves rendered HTML files in
`/data`, named by a hash of each page URL, and prints each file path with the
page's depth, HTTP status, discovered-link count, and URL. Fetch failures are
printed and returned as a non-zero process exit. The crawler identifies itself
as `news-crawler/1.0`; it does not attempt to bypass access controls or
anti-automation challenges.

Seed URLs and their depths are configured in `internal/config.go`. Browser
settings are in the same file. The Docker image includes Chromium and sets
`NEWS_CRAWLER_OUTPUT_DIR=/data`; no browser or Go dependencies need to be
installed on the host. For a different browser path or output directory, pass
`NEWS_CRAWLER_CHROME_PATH` or `NEWS_CRAWLER_OUTPUT_DIR` to `docker run`.

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
docker volume create news-crawler-data
docker run --rm -v news-crawler-data:/data news_crawler:latest
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
- The Docker build runs `go test ./...` after Go code changes.
- Update this README when setup, supported commands, or crawler behavior changes.
