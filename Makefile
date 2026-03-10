BINARY        = termf1
SERVER_BINARY = termf1-server
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  = -ldflags "-s -w -X main.version=$(VERSION)"
BUILD    = go build $(LDFLAGS) -o $(BINARY) .
BUILD_SERVER = go build $(LDFLAGS) -o $(SERVER_BINARY) ./cmd/termf1-server
RUN      = ./$(BINARY)

DIST_DIR = dist
PLATFORMS = \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: all build server run run-server install clean tidy dist snapshot

all: build

## build: compile the TUI binary
build:
	$(BUILD)

## server: compile the live-timing server binary
server:
	$(BUILD_SERVER)

## run: build and launch the TUI dashboard
run: build
	$(RUN)

## run-server: build and launch the live-timing server
run-server: server
	./$(SERVER_BINARY)

## install: install to $GOPATH/bin so you can run `termf1` anywhere
install:
	go install .

## tidy: tidy go modules
tidy:
	go mod tidy

## dist: cross-compile release archives for all platforms into dist/
dist: tidy
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		OUT=$(DIST_DIR)/$(BINARY)-$(VERSION)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then EXT=.exe; else EXT=""; fi; \
		echo "→ building $$GOOS/$$GOARCH"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH \
			go build $(LDFLAGS) -o $$OUT$$EXT . ; \
		if [ "$$GOOS" = "windows" ]; then \
			(cd $(DIST_DIR) && zip $(BINARY)-$(VERSION)-$$GOOS-$$GOARCH.zip $(BINARY)-$(VERSION)-$$GOOS-$$GOARCH$$EXT README.md .env.example && rm $(BINARY)-$(VERSION)-$$GOOS-$$GOARCH$$EXT); \
		else \
			(cd $(DIST_DIR) && tar czf $(BINARY)-$(VERSION)-$$GOOS-$$GOARCH.tar.gz $(BINARY)-$(VERSION)-$$GOOS-$$GOARCH$$EXT README.md .env.example && rm $(BINARY)-$(VERSION)-$$GOOS-$$GOARCH$$EXT); \
		fi; \
	done
	@echo "\nDist archives:"; ls -lh $(DIST_DIR)/

## snapshot: quick local build for current platform (no archive)
snapshot:
	$(BUILD)
	@echo "Built $(BINARY) $(VERSION)"

## clean: remove compiled binaries and dist/
clean:
	rm -f $(BINARY) $(SERVER_BINARY)
	rm -rf $(DIST_DIR)

## vet: run go vet
vet:
	go vet ./...
