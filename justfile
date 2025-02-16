set windows-shell := ["C:/Program Files/Git/bin/bash.exe", "-c"]

# Default recipe to display help information
default:
    @just --list

# Get engine version from export config
_engine-version:
    @go run scripts/extract_version.go

# Build the application with version from export config
build:
    go build -ldflags "-X github.com/robalyx/rotten/internal/version.EngineVersion=$(just _engine-version)" -o bin/rotten ./cmd/main.go

# Run tests with coverage
test:
    go test -v -race -cover ./...

# Run linter
lint:
    golangci-lint run --fix --timeout 120s

# Run the application
run:
    go run -ldflags "-X github.com/robalyx/rotten/internal/version.EngineVersion=$(just _engine-version)" ./cmd/main.go

# Clean build artifacts
clean:
    rm -rf bin/
    go clean -cache -testcache

# Download dependencies
deps:
    go mod download
    go mod tidy

# Generate mocks and other generated code
generate:
    go generate ./...
