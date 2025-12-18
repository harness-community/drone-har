FROM golang:1.22.7-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o drone-har .

# Build Harness CLI from specific commit
FROM golang:1.24-alpine AS hc-builder

RUN apk add --no-cache git
WORKDIR /hc
RUN git clone https://github.com/harness/harness-cli.git . && \
    git checkout baa7fbd531696b91a7df0bb104806455a3924e9d
RUN CGO_ENABLED=0 go build -o hc ./cmd/hc

FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy harness-cli from hc-builder
COPY --from=hc-builder /hc/hc /usr/local/bin/hc

# Copy drone-har binary to /bin/har
COPY --from=builder /app/drone-har /bin/har

ENTRYPOINT ["/bin/har"]
