# Makefile for droneHarPlugin

# Variables
BINARY_NAME=droneHarPlugin
DOCKER_IMAGE_NAME=harness/droneHarPlugin
VERSION?=latest

# Build the binary
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_NAME) .

# Build for current platform
build-local:
	go build -o $(BINARY_NAME) .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

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

.PHONY: build build-local test test-coverage clean docker-build docker-run fmt lint deps check
