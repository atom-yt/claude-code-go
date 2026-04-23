# Development Guide

## Build

```bash
# Build the CLI binary
go build -o claude ./cmd/claude

# Build all packages
go build ./...
```

## Test

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/tools/...

# Run verbose tests
go test -v ./...
```

## Lint

```bash
# Format code
gofmt -s -w .

# Run golangci-lint (recommended)
golangci-lint run

# Run go vet
go vet ./...

# Check for unused dependencies
go mod tidy
go mod verify
```

## Harness Validation

```bash
# Check architecture compliance
go run .harness/scripts/lint-deps.go

# Check code quality rules
go run .harness/scripts/lint-quality.go

# Run full validation pipeline
go run .harness/scripts/validate.go
```

## Development Workflow

1. Create a feature branch from main
2. Make changes
3. Run `make test` to verify tests pass
4. Run `gofmt` to format code
5. Commit with conventional commit format
6. Push and create PR

## Quick Start for New Features

1. Implement feature
2. Add tests (if applicable)
3. Update documentation
4. Run validation: `go run .harness/scripts/validate.go`
5. Commit

## Adding a New Tool

1. Create package in `internal/tools/mytool/`
2. Implement `Tool` interface
3. Add tests in `mytool_test.go`
4. Register in `internal/tools/registry.go`
5. Update this documentation

## Adding a New Provider

1. Add provider info to `internal/tui/model.go`
2. (Optional) Add capabilities to `internal/api/capabilities.go`
3. Update README.md

## Debugging

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o claude ./cmd/claude

# Run tests with race detection
go test -race ./...

# Profile memory
go test -memprofile=mem.prof ./...
```