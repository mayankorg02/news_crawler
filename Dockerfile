FROM golang:1.23-alpine AS builder
WORKDIR /app
copy go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
