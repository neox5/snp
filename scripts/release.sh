#!/usr/bin/env bash
set -euo pipefail

# --- Configuration -----------------------------------------------------------

BINARY="snap"
MODULE="github.com/neox5/snap"
DIST_DIR="dist"

# --- Helpers -----------------------------------------------------------------

fail() {
  echo "ERROR: $1" >&2
  exit 1
}

info() {
  echo "==> $1"
}

# --- Preconditions -----------------------------------------------------------

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
make clean
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
  ( cd "$DIST_DIR" && sha256sum -c "$(basename "$sum")" ) \
    || fail "checksum verification failed for $sum"
done

# --- Final Instructions ------------------------------------------------------

cat <<EOF

✅ Release artifacts validated.

Next steps:

1. Push tag and branch:
   git push origin main
   git push origin $CURRENT_TAG

2. Create GitHub release:
   - Tag: $CURRENT_TAG
   - Upload all files from:
       $DIST_DIR/
   - Publish release

3. (Optional) Post-release installation check on a clean system:
   - Download a binary + .sha256 from GitHub Releases
   - Verify:
       sha256sum -c snap-<os>-<arch>.sha256
   - Run:
       ./snap-<os>-<arch> --version   # should print $CURRENT_TAG

EOF
