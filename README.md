<br/>

<div align="center">
  <img src="logo.png" alt="snp logo" width="200"/>
</div>

<br/>

# snp

[![Release](https://img.shields.io/github/v/release/neox5/snp)](https://github.com/neox5/snp/releases)
![Go Version](https://img.shields.io/github/go-mod/go-version/neox5/snp)
![License](https://img.shields.io/github/license/neox5/snp)

A CLI tool that concatenates all readable files in a project into a single deterministic snapshot file for inspection, sharing, and machine processing.

## Quick Start

```bash
# Install (Linux example)
curl -LO https://github.com/neox5/snp/releases/latest/download/snp-linux-amd64
chmod +x snp-linux-amd64
sudo mv snp-linux-amd64 /usr/local/bin/snp

# Run in any project directory
cd /path/to/your/project
snp
```

Creates `./snapshot.snp` with all project files concatenated.

## Usage

```bash
snp [OPTIONS] [DIRECTORY]
```

`DIRECTORY` is optional and defaults to `.`. It sets the root for traversal and the base for resolving relative paths.

## Modes

snp operates in two mutually exclusive modes.

### Traversal (default)

Recursively walks the directory tree and filters files by ordered rules.

**Implicit default** (when no baseline flag is given):

```bash
snp  # include all files, applying default exclude patterns and .gitignore
```

**Baseline flags** — set the starting state, evaluated in position order:

```
--include-all         Include all files (overrides default excludes)
--exclude-all         Exclude all files
```

**Filter flags** — stack on top of baseline, evaluated in position order:

```
--include <pattern>   Include files matching glob pattern (repeatable)
--exclude <pattern>   Exclude files matching glob pattern (repeatable)
```

All flags are positional — **last matching rule wins**. Baseline and filter flags can be freely interleaved to express any selection logic:

```bash
# Only Go files
snp --exclude-all --include "**/*.go"

# Everything except tests
snp --exclude-all --include "**/*.go" --include "**/*.md"

# Everything except tests, but keep one specific test
snp --include-all --exclude "**/*_test.go" --include "internal/auth/auth_test.go"

# Everything, including .git/ and other normally excluded files
snp --include-all
```

**Depth control:**

```
--depth <n>           Limit traversal depth (0 = root only, -1 = unlimited, default: -1)
```

**Utility:**

```
--show-defaults       Print default exclude patterns and exit
--dry-run             List files that would be included without creating output
```

### Pick

Directly addresses files by exact path or glob. No traversal, no defaults applied.

```
--pick <path>         Include file by exact path or glob (repeatable)
```

Relative paths and globs resolve against `DIRECTORY`. Absolute paths are shown as-is in the snapshot.

```bash
# Exact files
snp --pick "cmd/main.go" --pick "README.md"

# Glob
snp --pick "internal/**/*.go"

# Cross-repo from parent directory
snp --pick "repo-a/cmd/main.go" --pick "repo-b/cmd/main.go"

# Absolute path
snp --pick "/etc/nginx/nginx.conf"
```

`--pick` cannot be combined with `--include`, `--exclude`, `--include-all`, `--exclude-all`, or `--depth`.

## Output

```
--output <path>       Set output file path (default: snapshot.snp)
--stdout              Write snapshot to stdout instead of a file
--no-summary          Omit summary section
--no-index            Omit file index section
--no-git-log          Omit git log section
--no-content          Omit file content sections
--only-summary        Include only summary section
--only-index          Include only file index section
--only-git-log        Include only git log section
--only-content        Include only file content sections
--dry-run             List files without creating output
--silent              Suppress all stdout
```

`--only-<section>` flags are the inverse of `--no-<section>` flags and can be freely combined. `--only-index --only-summary` includes both sections. Combining `--only-<section>` and `--no-<section>` flags that result in all sections being suppressed is an error.

When `--stdout` is used, all status messages are written to stderr so that stdout carries only snapshot content.

```bash
# Pipe only the file index into another tool
snp --only-index --stdout 2>/dev/null | grep "\.go"

# Save content section only to a custom path
snp --only-content --stdout > content.txt
```

## Binary and Text Overrides

Applies to both modes. Binary files are detected automatically and shown as metadata only.

```
--force-text <pattern>    Treat matched files as text (repeatable)
--force-binary <pattern>  Treat matched files as binary (repeatable)
```

`--force-binary` wins over `--force-text` on conflict.

## Configuration File

Flags can be persisted to `.snpconfig.json` in the source directory. Config file flags are merged with CLI flags; CLI flags take precedence.

```
--save-config         Save current flags to .snpconfig.json and exit
--show-config         Print current config and equivalent snp command, then exit
--no-config           Skip .snpconfig.json even if present
```

```bash
# Save current invocation as default for this directory
snp --exclude-all --include "**/*.go" --save-config

# Inspect resolved config
snp --show-config
```

## Output Format

```text
Generated: 2025-12-14 18:13:40
Total files: 24 (23 text, 1 binary)
Total size: 1.2 MB
Total lines: 2284

# File Index
.gitignore [55-59] (5 lines, 42 bytes)
README.md [63-83] (21 lines, 1.1 KB)
...

# ----------------------------------------

# Git Log (git adog)
* abc1234 (HEAD -> main) ...

# ----------------------------------------

# .gitignore
...

# ----------------------------------------

# logo.png
[Binary file - 43.7 KB - content omitted]
```

Use the file index to navigate by line number. Line ranges are omitted when content sections are excluded.

## Working with AI Tools

```
## Working with Repository Snapshots

Snapshots were generated with [snp](https://github.com/neox5/snp).

**Rules for working with snapshot files:**

- Snapshot files are READ-ONLY reference documents
- DO NOT modify snapshot files directly
- DO NOT create updated versions of snapshot files
- Changes must target actual source files in their original locations
- User will regenerate snapshot by running snp after changes

**How to use snapshot files:**

1. File index is at the top with line ranges
2. Each file section starts with # filepath
3. Binary files show size metadata instead of content
4. Git log (if present) shows recent commits
```

## Installation

### Prebuilt Binaries

**Linux (amd64)**

```bash
curl -LO https://github.com/neox5/snp/releases/latest/download/snp-linux-amd64
curl -LO https://github.com/neox5/snp/releases/latest/download/snp-linux-amd64.sha256
sha256sum -c snp-linux-amd64.sha256
chmod +x snp-linux-amd64
sudo mv snp-linux-amd64 /usr/local/bin/snp
```

**macOS (Apple Silicon)**

```bash
curl -LO https://github.com/neox5/snp/releases/latest/download/snp-darwin-arm64
curl -LO https://github.com/neox5/snp/releases/latest/download/snp-darwin-arm64.sha256
shasum -a 256 -c snp-darwin-arm64.sha256
chmod +x snp-darwin-arm64
sudo mv snp-darwin-arm64 /usr/local/bin/snp
```

**Available platforms:**

- `snp-linux-amd64` / `snp-linux-arm64`
- `snp-darwin-amd64` / `snp-darwin-arm64`
- `snp-windows-amd64.exe` / `snp-windows-arm64.exe`

### Via Go

Requires Go 1.22+

```bash
go install github.com/neox5/snp/cmd/snp@latest
```

Ensure `$HOME/go/bin` is in your `PATH`.

### From Source

```bash
git clone https://github.com/neox5/snp
cd snp
make build-local
sudo mv dist/snp /usr/local/bin/snp
```

### Verify Installation

```bash
snp --version
```

## Release

### Creating a Release

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
make release
```

### Post-Release Verification

```bash
make post-release
```

## License

MIT License — see [LICENSE](LICENSE) file for details.
