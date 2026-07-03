# Build stage
FROM golang:1.25-alpine AS builder


WORKDIR /app


COPY go.mod go.sum ./


RUN go mod download


COPY . .


RUN go build \
    -o url-shortener \
    ./cmd/server


# Runtime stage
FROM alpine:latest


WORKDIR /app


COPY --from=builder \
    /app/url-shortener .


EXPOSE 8080


CMD ["./url-shortener"]