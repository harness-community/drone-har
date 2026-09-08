# Makefile for drone-har

# Variables
BINARY_NAME=drone-har
DOCKER_IMAGE_NAME=harness/drone-har
VERSION?=latest

# Download the pinned harness-cli build that gets embedded into the binary.
# Every build target depends on this: go:embed fails without the archive.
hc:
	./scripts/fetch-hc.sh linux/$(shell go env GOARCH)

# Download the harness-cli builds for every platform we release for
hc-all:
	./scripts/fetch-hc.sh

# Build the binary
build: hc
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_NAME) .

# Build for current platform
build-local:
	./scripts/fetch-hc.sh $(shell go env GOOS)/$(shell go env GOARCH)
	go build -o $(BINARY_NAME) .

# Run tests (unit tests only)
test:
	./scripts/fetch-hc.sh $(shell go env GOOS)/$(shell go env GOARCH)
	go test -v ./...

# Run tests with coverage
test-coverage:
	./scripts/fetch-hc.sh $(shell go env GOOS)/$(shell go env GOARCH)
	go test -cover ./...

# Run integration tests (requires HARNESS_TOKEN and HARNESS_ACCOUNT)
test-integration:
	./scripts/fetch-hc.sh $(shell go env GOOS)/$(shell go env GOARCH)
	go test -tags=integration -v ./plugin/...

# Run all tests (unit + integration)
test-all:
	./scripts/fetch-hc.sh $(shell go env GOOS)/$(shell go env GOARCH)
	go test -v ./...
	go test -tags=integration -v ./plugin/...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f plugin/packages/hcbin/*.tar.gz plugin/packages/hcbin/.version

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) .

# Run Docker container locally for testing
docker-run:
	docker run --rm \
		-e PLUGIN_REGISTRY=test-registry \
		-e PLUGIN_SOURCE=./test.txt \
		-e PLUGIN_NAME=test-artifact \
		-e PLUGIN_VERSION=1.0.0 \
		-e PLUGIN_TOKEN=test-token \
		-e PLUGIN_ACCOUNT=test-account \
		$(DOCKER_IMAGE_NAME):$(VERSION)

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run all checks
check: fmt test lint

.PHONY: hc hc-all build build-local test test-coverage test-integration test-all clean docker-build docker-run fmt lint deps check
