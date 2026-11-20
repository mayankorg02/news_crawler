FROM golang:1.27.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 go build -o /news-crawler ./cmd

FROM alpine:3.22
RUN apk add --no-cache ca-certificates chromium \
    && adduser -D crawler \
    && mkdir /data \
    && chown crawler:crawler /data
COPY --from=builder /news-crawler /usr/local/bin/news-crawler
ENV NEWS_CRAWLER_CHROME_PATH=/usr/bin/chromium-browser
ENV NEWS_CRAWLER_OUTPUT_DIR=/data
USER crawler
ENTRYPOINT ["/usr/local/bin/news-crawler"]
