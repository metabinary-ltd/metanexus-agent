.PHONY: build build-linux-amd64 build-linux-arm64 clean test

# Build for current platform
build:
	go build -o bin/agent ./cmd/agent

# Build for Linux x86_64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/agent-linux-amd64 ./cmd/agent

# Build for Linux arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/agent-linux-arm64 ./cmd/agent

# Build for both Linux architectures
build-all: build-linux-amd64 build-linux-arm64

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

