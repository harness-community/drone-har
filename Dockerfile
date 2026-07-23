# syntax=docker/dockerfile:1.6

########################
# Drone HAR build
########################
FROM golang:1.24-alpine AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -o drone-har .

########################
# Download Harness CLI
########################
FROM alpine:latest AS hc-downloader
ARG TARGETOS=linux
ARG TARGETARCH

RUN apk add --no-cache curl tar
WORKDIR /hc

RUN ARCH=$([ "$TARGETARCH" = "amd64" ] && echo "x86_64" || echo "$TARGETARCH") && \
    curl -fsSL "https://github.com/harness/harness-cli/releases/download/v1.3.36/hc_1.3.36_${TARGETOS}_${ARCH}.tar.gz" | tar -xz

########################
# Runtime image
########################
FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=hc-downloader /hc/hc /usr/local/bin/hc
COPY --from=builder /app/drone-har /bin/har

ENTRYPOINT ["/bin/har"]
