FROM golang:1.22.7-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o droneHarPlugin .

FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Install Harness CLI
RUN apk add --no-cache curl && \
    curl -L https://github.com/harness/harness-cli/releases/latest/download/hc-linux-amd64 -o /usr/local/bin/hc && \
    chmod +x /usr/local/bin/hc && \
    apk del curl

WORKDIR /root/

COPY --from=builder /app/droneHarPlugin .

ENTRYPOINT ["./droneHarPlugin"]
