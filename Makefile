# Makefile for mahin CLI

BINARY_NAME=mahin
GO=go
GOFLAGS=-ldflags="-s -w"

# Default: build for current OS
.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# Build for Windows
.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME).exe .

# Build for macOS (Apple Silicon)
.PHONY: build-mac-arm
build-mac-arm:
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# Build for macOS (Intel)
.PHONY: build-mac-intel
build-mac-intel:
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# Build for Linux
.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# Run
.PHONY: run
run: build
	./$(BINARY_NAME)

# Clean
.PHONY: clean
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

# Install locally (copies to /usr/local/bin on Mac/Linux)
.PHONY: install
install: build
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Tidy modules
.PHONY: tidy
tidy:
	$(GO) mod tidy
