BINARY   := claude
MODULE   := github.com/atom-yt/claude-code-go
CMD      := ./cmd/claude
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: all build install test clean release fmt vet

all: build

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

release: clean
	@mkdir -p dist
	@$(foreach PLATFORM,$(PLATFORMS), \
		$(eval GOOS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
		$(eval GOARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
		$(eval EXT    := $(if $(filter windows,$(GOOS)),.exe,)) \
		$(eval OUT    := dist/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT)) \
		echo "Building $(OUT)..." && \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(OUT) $(CMD) && \
	) true
	@echo "Release artifacts:"
	@ls -lh dist/
