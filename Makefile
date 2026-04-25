BINARY   := claude
MODULE   := github.com/atom-yt/claude-code-go
CMD      := ./cmd/claude
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

PLATFORMS := \n	darwin/amd64 \n	darwin/arm64 \n	linux/amd64 \n	linux/arm64 \n	windows/amd64

.PHONY: all build install test clean release fmt vet help
.PHONY: backend-build backend-test backend-dev frontend-build frontend-test frontend-dev

all: build

# ============================================================================
# CLI Commands (original claude-code-go)
# ============================================================================

build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

install:
	go install $(LDFLAGS) $(CMD)

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf $(BINARY) dist/
	rm -rf backend/bin/
	rm -rf frontend/.next/
	rm -rf frontend/node_modules/

release: clean
	@mkdir -p dist
	@$(foreach PLATFORM,$(PLATFORMS), \n		$(eval GOOS   := $(word 1,$(subst /, ,$(PLATFORM)))) \n		$(eval GOARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \n		$(eval EXT    := $(if $(filter windows,$(GOOS)),.exe,)) \n		$(eval OUT    := dist/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT)) \n		echo "Building $(OUT)..." && \n		GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(OUT) $(CMD) && \n	) true
	@echo "Release artifacts:"
	@ls -lh dist/

# ============================================================================
# Backend Commands
# ============================================================================

backend-build:
	cd backend && go build -o bin/server cmd/server/main.go

backend-test:
	cd backend && go test ./...

backend-dev:
	cd backend && go run cmd/server/main.go

backend-deps:
	cd backend && go mod download

# ============================================================================
# Frontend Commands
# ============================================================================

frontend-build:
	cd frontend && npm run build

frontend-test:
	cd frontend && npm test

frontend-dev:
	cd frontend && npm run dev

frontend-deps:
	cd frontend && npm install

# ============================================================================
# Combined Commands
# ============================================================================

monorepo-build: build backend-build frontend-build
	@echo "All components built!"

monorepo-test: test backend-test frontend-test
	@echo "All tests passed!"

monorepo-dev:
	@echo "Starting development servers..."
	@echo "Run 'make backend-dev' and 'make frontend-dev' in separate terminals"

# ============================================================================
# Utility
# ============================================================================

help:
	@echo "Atom AI Platform - Available Commands:"
	@echo ""
	@echo "CLI (claude-code-go):"
	@echo "  make build          - Build CLI binary"
	@echo "  make test           - Run CLI tests"
	@echo "  make install        - Install CLI binary"
	@echo "  make release        - Build release binaries for all platforms"
	@echo ""
	@echo "Backend:"
	@echo "  make backend-build  - Build backend server"
	@echo "  make backend-test   - Run backend tests"
	@echo "  make backend-dev    - Start backend in development mode"
	@echo "  make backend-deps   - Install backend dependencies"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend-build - Build frontend"
	@echo "  make frontend-test  - Run frontend tests"
	@echo "  make frontend-dev   - Start frontend in development mode"
	@echo "  make frontend-deps  - Install frontend dependencies"
	@echo ""
	@echo "Monorepo:"
	@echo "  make monorepo-build - Build all components"
	@echo "  make monorepo-test  - Run all tests"
	@echo "  make clean          - Clean all build artifacts"