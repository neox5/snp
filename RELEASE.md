# snap --- Release Guide

This document describes the exact procedure for creating and publishing
a new release.

------------------------------------------------------------------------

## 1. Preconditions

-   Clean working tree
-   All intended changes merged to `main`
-   Go 1.22+ installed
-   `make`, `sha256sum`, and `git` available
-   GitHub access to publish releases

Verify:

``` bash
git status
go test ./...
```

------------------------------------------------------------------------

## 2. Decide Version

Example:

``` text
v0.1.0
```

------------------------------------------------------------------------

## 3. Create Git Tag

``` bash
git tag -a v0.1.0 -m "Release v0.1.0"
```

------------------------------------------------------------------------

## 4. Build Release Artifacts

``` bash
make build
```

This produces:

``` text
dist/
  snap-linux-amd64
  snap-linux-amd64.sha256
  snap-linux-arm64
  snap-linux-arm64.sha256
  snap-darwin-amd64
  snap-darwin-amd64.sha256
  snap-darwin-arm64
  snap-darwin-arm64.sha256
  snap-windows-amd64.exe
  snap-windows-amd64.exe.sha256
  snap-windows-arm64.exe
  snap-windows-arm64.exe.sha256
```

------------------------------------------------------------------------

## 5. Local Verification

``` bash
./dist/snap-linux-amd64 --version
./dist/snap-linux-amd64 --help
```

Version must match the tag.

------------------------------------------------------------------------

## 6. Push Tag

``` bash
git push origin main
git push origin v0.1.0
```

------------------------------------------------------------------------

## 7. Create GitHub Release

1.  Go to the repository's **Releases** page.
2.  Click **Draft a new release**.
3.  Select tag `v0.1.0`.
4.  Enter a release title (e.g., `snap v0.1.0`).
5.  Add release notes.
6.  Upload all files from the `dist/` directory.
7.  Publish the release.

------------------------------------------------------------------------

## 8. Post-Release Verification

On a clean system:

``` bash
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64.sha256
sha256sum -c snap-linux-amd64.sha256
chmod +x snap-linux-amd64
./snap-linux-amd64 --version
```

The printed version must match the release tag.
