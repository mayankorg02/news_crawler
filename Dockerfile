FROM golang:1.27.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /news-crawler ./cmd

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=builder /news-crawler /usr/local/bin/news-crawler
ENTRYPOINT ["/usr/local/bin/news-crawler"]
