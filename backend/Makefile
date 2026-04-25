.PHONY: all build test clean migrate-up migrate-down migrate-create migrate-version

# Default target
all: build

# Build the backend
build:
	go build -o bin/api ./cmd/api

# Run tests
test:
	go test -v ./...

# Run all migrations
migrate-up:
	@echo "Running up migrations..."
	go run migrate.go -up

# Rollback all migrations
migrate-down:
	@echo "Running down migrations..."
	go run migrate.go -down

# Rollback specific number of migrations
migrate-steps:
	@echo "Running $(STEPS) migrations..."
	go run migrate.go -steps $(STEPS)

# Create a new migration (usage: make migrate-create NAME=add_table)
migrate-create:
	@if [ -z "$(NAME)" ]; then \n		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \n		exit 1; \n	fi
	go run migrate.go -create $(NAME)

# Show current migration version
migrate-version:
	go run migrate.go

# Force set migration version (usage: make migrate-force VERSION=1)
migrate-force:
	@if [ -z "$(VERSION)" ]; then \n		echo "Error: VERSION is required. Usage: make migrate-force VERSION=1"; \n		exit 1; \n	fi
	go run migrate.go -force $(VERSION)

# Clean build artifacts
clean:
	rm -rf bin/

# Run with docker-compose (if available)
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Generate Go types from SQL schema (requires sqlc)
generate:
	sqlc generate

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Development server
dev:
	go run ./cmd/api --dev