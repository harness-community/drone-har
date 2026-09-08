# syntax=docker/dockerfile:1.6

########################
# Download Harness CLI
########################
FROM alpine:latest AS hc-downloader
ARG TARGETOS=linux
ARG TARGETARCH

RUN apk add --no-cache curl
WORKDIR /hc

# HC_VERSION is the single source of truth for the pinned harness-cli version.
COPY HC_VERSION .
COPY scripts/fetch-hc.sh ./scripts/
RUN sh scripts/fetch-hc.sh "${TARGETOS}/${TARGETARCH}"

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
# The plugin embeds harness-cli, so the archive has to be in place before the
# build. The binary that comes out needs no hc on disk or at runtime.
COPY --from=hc-downloader /hc/plugin/packages/hcbin/ ./plugin/packages/hcbin/
RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -o drone-har .

########################
# Runtime image
########################
FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/drone-har /bin/har

ENTRYPOINT ["/bin/har"]
