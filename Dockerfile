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
# Harness CLI build
########################
FROM golang:1.24-alpine AS hc-builder
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache git
WORKDIR /hc

RUN git clone https://github.com/harness/harness-cli.git . && \
    git checkout bdff66bc074eeff75d9a8d10f6c6f530b9bd1302

RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -o hc ./cmd/hc

########################
# Runtime image
########################
FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=hc-builder /hc/hc /usr/local/bin/hc
COPY --from=builder /app/drone-har /bin/har

ENTRYPOINT ["/bin/har"]
