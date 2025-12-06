BINARY      := snap
CMD_PKG     := ./cmd/snap
MODULE_PATH := github.com/neox5/snap

DIST_DIR    := dist

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

# Version from git; falls back to "dev" if describe fails.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

LDFLAGS := -X '$(MODULE_PATH)/internal/version.Version=$(VERSION)'

.PHONY: all build build-local clean print-version release post-release

all: build

# ---------------------------------------------------------------------
# Multi-platform release build + checksums
# ---------------------------------------------------------------------

build: clean
	@mkdir -p "$(DIST_DIR)"
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		ext=""; \
		[ "$$GOOS" = "windows" ] && ext=".exe"; \
		out="$(DIST_DIR)/$(BINARY)-$${GOOS}-$${GOARCH}$${ext}"; \
		file=$$(basename "$$out"); \
		echo "building $$out (VERSION=$(VERSION))"; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "$(LDFLAGS)" -o "$$out" $(CMD_PKG); \
		echo "generating $$file.sha256"; \
		( cd "$(DIST_DIR)" && sha256sum "$$file" > "$$file.sha256" ); \
	done

# ---------------------------------------------------------------------
# Local development build
# ---------------------------------------------------------------------

build-local:
	@mkdir -p "$(DIST_DIR)"
	@echo "building $(DIST_DIR)/$(BINARY) (VERSION=$(VERSION))"
	go build -ldflags "$(LDFLAGS)" -o "$(DIST_DIR)/$(BINARY)" $(CMD_PKG)

# ---------------------------------------------------------------------
# Full release verification (tag, tests, binaries, checksums)
# ---------------------------------------------------------------------

release:
	@./scripts/release.sh

# ---------------------------------------------------------------------
# Post-release verification against GitHub latest
# ---------------------------------------------------------------------

post-release:
	@./scripts/post-release.sh

# ---------------------------------------------------------------------

print-version:
	@echo $(VERSION)

clean:
	rm -rf "$(DIST_DIR)"
