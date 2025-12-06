#!/usr/bin/env bash
set -euo pipefail

# Post-release verification script (latest release only).
#
# - Uses latest GitHub release
# - Auto-detects OS and ARCH
# - Downloads binary + checksum into a temp directory
# - Verifies SHA256
# - Verifies the binary runs and reports a valid version
# - Automatically deletes the temp directory after execution
#
# Usage:
#   make post-release
#   or
#   ./scripts/post-release.sh

BINARY="snap"
OWNER_REPO="neox5/snap"

fail() {
  echo "ERROR: $1" >&2
  exit 1
}

info() {
  echo "==> $1"
}

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
[ "$OS" = "windows" ] && EXT=".exe"

FILE="${BINARY}-${OS}-${ARCH}${EXT}"
SUM_FILE="${FILE}.sha256"

BASE_URL="https://github.com/${OWNER_REPO}/releases/latest/download"

info "os:    $OS"
info "arch:  $ARCH"
info "file:  $FILE"

# --- Temporary workspace ----------------------------------------------------

TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

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

# --- Version check ----------------------------------------------------------

info "running downloaded binary"
chmod +x "$FILE"
RAW_VERSION="$("./$FILE" --version)"
DOWNLOADED_VERSION="${RAW_VERSION##* }"

info "downloaded version: $DOWNLOADED_VERSION"

if [[ ! "$DOWNLOADED_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  fail "downloaded binary returned invalid version: $RAW_VERSION"
fi

echo
echo "✅ Post-release verification successful"
echo "✅ Latest binary is valid: $DOWNLOADED_VERSION"
