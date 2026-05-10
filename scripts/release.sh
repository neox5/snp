#!/usr/bin/env bash
set -euo pipefail

export MAKEFLAGS="--no-print-directory"

# --- Load configuration ------------------------------------------------------

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ ! -f "$SCRIPT_DIR/release.env" ]; then
  echo "ERROR: release.env not found" >&2
  exit 1
fi

source "$SCRIPT_DIR/release.env"

# --- Helpers -----------------------------------------------------------------

fail() {
  echo "ERROR: $1" >&2
  exit 1
}

info() {
  echo "==> $1"
}

# --- Preconditions -----------------------------------------------------------

cd "$PROJECT_ROOT"

info "checking git state"

git diff --quiet || fail "working tree is dirty"
git diff --cached --quiet || fail "index is dirty"

CURRENT_TAG="$(git describe --tags --exact-match 2>/dev/null || true)"
[ -z "$CURRENT_TAG" ] && fail "no exact git tag found on HEAD"

info "found release tag: $CURRENT_TAG"

# --- Tests -------------------------------------------------------------------

info "running tests"
go test ./...

# --- Build -------------------------------------------------------------------

info "building release artifacts"
make build

# --- Verify host-native binary version ---------------------------------------

info "verifying host-native binary contains correct version"

HOST_GOOS="$(go env GOOS)"
HOST_GOARCH="$(go env GOARCH)"
HOST_EXT=""

if [ "$HOST_GOOS" = "windows" ]; then
  HOST_EXT=".exe"
fi

HOST_BIN="${DIST_DIR}/${BINARY}-${HOST_GOOS}-${HOST_GOARCH}${HOST_EXT}"

if [ ! -x "$HOST_BIN" ]; then
  fail "host-native binary not found: $HOST_BIN"
fi

RAW_VERSION="$("$HOST_BIN" --version)"
HOST_VERSION="${RAW_VERSION##* }"

if [ "$HOST_VERSION" != "$CURRENT_TAG" ]; then
  fail "version mismatch in host binary (expected $CURRENT_TAG, got $RAW_VERSION)"
fi

# --- Verify sha256 files -----------------------------------------------------

info "verifying sha256 checksums"

for sum in "$DIST_DIR"/*.sha256; do
  (cd "$DIST_DIR" && sha256sum -c "$(basename "$sum")") ||
    fail "checksum verification failed for $sum"
done

# --- Push --------------------------------------------------------------------

info "pushing branch and tag"
git push origin main
git push origin "$CURRENT_TAG"

echo
echo "✅ Release artifacts validated and tag pushed."
echo "✅ GitHub Actions will build and publish the release automatically."
echo
echo "Monitor the release at: https://github.com/${OWNER_REPO}/actions"
