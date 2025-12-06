BINARY      := snap
MODULE_PATH := ./cmd/snap

DIST_DIR    := dist

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: all build build-local clean

all: build

build: clean
	@mkdir -p "$(DIST_DIR)"
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		ext=""; \
		[ $$GOOS = "windows" ] && ext=".exe"; \
		out="$(DIST_DIR)/$(BINARY)-$${GOOS}-$${GOARCH}$${ext}"; \
		echo "building $$out"; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -o "$$out" $(MODULE_PATH); \
	done

# Local build for current platform (useful during development)
build-local:
	@mkdir -p "$(DIST_DIR)"
	go build -o "$(DIST_DIR)/$(BINARY)" $(MODULE_PATH)

clean:
	rm -rf "$(DIST_DIR)"
