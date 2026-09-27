BINARY      := snp
CMD_PKG     := ./cmd/snp
MODULE_PATH := github.com/neox5/snp

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

# LDFLAGS: Set version + strip debug symbols for smaller binaries
# -s: omit symbol table
# -w: omit DWARF debug info
# Result: ~40-50% size reduction
LDFLAGS := -s -w -X '$(MODULE_PATH)/internal/version.Version=$(VERSION)'

.PHONY: all build build-local clean test lint print-version release post-release install

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

build-local: clean
	@mkdir -p "$(DIST_DIR)"
	@echo "building $(DIST_DIR)/$(BINARY) (VERSION=$(VERSION))"
	go build -ldflags "$(LDFLAGS)" -o "$(DIST_DIR)/$(BINARY)" $(CMD_PKG)

# ---------------------------------------------------------------------
# Test and lint
# ---------------------------------------------------------------------

test:
	go test ./...

lint:
	go vet ./...
	go fmt ./...

# ---------------------------------------------------------------------
# Release — verify locally, push tag, GitHub Actions publishes
# ---------------------------------------------------------------------

release:
	@./scripts/release.sh

# ---------------------------------------------------------------------
# Post-release verification against GitHub latest
# ---------------------------------------------------------------------

post-release:
	@./scripts/post-release.sh

# ---------------------------------------------------------------------
# Install published release binary to /usr/local/bin
# ---------------------------------------------------------------------

install:
	@OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	ARCH=$$(uname -m); \
	case "$$ARCH" in \
		x86_64|amd64) ARCH=amd64 ;; \
		arm64|aarch64) ARCH=arm64 ;; \
		*) echo "unsupported ARCH: $$ARCH" >&2; exit 1 ;; \
	esac; \
	EXT=""; \
	[ "$$OS" = "windows" ] && EXT=".exe"; \
	FILE="$(BINARY)-$${OS}-$${ARCH}$${EXT}"; \
	URL="https://github.com/neox5/snp/releases/latest/download/$$FILE"; \
	TMP=$$(mktemp -d); \
	trap 'rm -rf "$$TMP"' EXIT; \
	echo "downloading $$URL"; \
	curl -fL -o "$$TMP/$$FILE" "$$URL"; \
	chmod +x "$$TMP/$$FILE"; \
	sudo mv "$$TMP/$$FILE" /usr/local/bin/$(BINARY); \
	$(BINARY) --version

# ---------------------------------------------------------------------

print-version:
	@echo $(VERSION)

clean:
	rm -rf "$(DIST_DIR)"
