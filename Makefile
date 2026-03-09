BINARY=proxysh
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
BUILD_DIR=bin

.PHONY: all build install uninstall clean test lint release

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .

install: build
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed $(BINARY) to /usr/local/bin"

uninstall:
	-$(BINARY) stop 2>/dev/null || true
	sudo rm -f /usr/local/bin/$(BINARY)
	@echo "Uninstalled $(BINARY)"

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./... -v

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	golangci-lint run ./...

# Cross-platform release builds
release:
	@mkdir -p dist
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)_darwin_amd64  .
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)_darwin_arm64  .
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)_linux_amd64   .
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)_linux_arm64   .
	@for f in dist/$(BINARY)_*; do \
		tar -czf "$$f.tar.gz" -C dist "$$(basename $$f)"; \
		rm "$$f"; \
		echo "  $$f.tar.gz"; \
	done
	@echo "Release builds ready in dist/"

dev:
	go run . doctor

deps:
	go mod download
	go mod tidy
