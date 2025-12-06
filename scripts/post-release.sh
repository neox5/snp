#!/usr/bin/env bash
set -euo pipefail

# Post-release verification script (latest release only).
#
# - Auto-detects OS and ARCH
# - Downloads latest binary + checksum
# - Verifies SHA256
# - Compares version with locally installed snap
# - Fails if versions differ
#
# Usage:
#   scripts/post-release.sh

BINARY="snap"
OWNER_REPO="neox5/snap"

fail() {
  echo "ERROR: $1" >&2
  exit 1
}

info() {
  echo "==> $1"
}

# --- Ensure local snap exists -----------------------------------------------

command -v snap >/dev/null 2>&1 || fail "local 'snap' not found in PATH"

LOCAL_RAW="$(snap --version)"
LOCAL_VERSION="${LOCAL_RAW##* }"

info "local version:  $LOCAL_VERSION"

# --- Detect OS ---------------------------------------------------------------

uname_s="$(uname -s)"
case "$uname_s" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*|Windows_NT)
    OS="windows"
    ;;
  *)
    fail "unsupported OS from uname -s: $uname_s"
    ;;
esac

# --- Detect ARCH -------------------------------------------------------------

uname_m="$(uname -m)"
case "$uname_m" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    fail "unsupported ARCH from uname -m: $uname_m"
    ;;
esac

EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

FILE="${BINARY}-${OS}-${ARCH}${EXT}"
SUM_FILE="${FILE}.sha256"

BASE_URL="https://github.com/${OWNER_REPO}/releases/latest/download"

info "os:    $OS"
info "arch:  $ARCH"
info "file:  $FILE"

TEMP_DIR="$(mktemp -d)"
info "working directory: $TEMP_DIR"
cd "$TEMP_DIR"

# --- Download ---------------------------------------------------------------

info "downloading latest binary"
curl -fL -o "$FILE" "${BASE_URL}/${FILE}"

info "downloading checksum"
curl -fL -o "$SUM_FILE" "${BASE_URL}/${SUM_FILE}"

# --- Verify checksum --------------------------------------------------------

info "verifying checksum"
sha256sum -c "$SUM_FILE"

# --- Version compare --------------------------------------------------------

info "running downloaded binary"
chmod +x "$FILE"
DOWNLOADED_RAW="./$FILE --version"
DOWNLOADED_RAW="$("./$FILE" --version)"
DOWNLOADED_VERSION="${DOWNLOADED_RAW##* }"

info "downloaded version: $DOWNLOADED_VERSION"

if [ "$DOWNLOADED_VERSION" != "$LOCAL_VERSION" ]; then
  fail "version mismatch: local=$LOCAL_VERSION downloaded=$DOWNLOADED_VERSION"
fi

echo
echo "✅ Post-release verification successful"
echo "✅ Local and latest versions match: $LOCAL_VERSION"
