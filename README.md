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

### Basic Usage

```bash
snp                    # Create snapshot.snp in current directory
snp /path/to/project   # Create snapshot.snp from specified directory
```

### Output Control

```bash
snp --output custom.snp              # Custom output path
snp --exclude-git-log                # Omit Git log section
snp --dry-run                        # List files without creating output
```

## Modes

snp operates in two mutually exclusive modes.

### Mode 1 — Traversal (default)

Recursively walks the directory tree and filters files by pattern.

```bash
snp --include "src/**/*.go"                    # Include only Go files in src/
snp --exclude "**/*_test.go"                   # Exclude test files
snp --include "**/*.go" --exclude "**/*_test.go" --include "internal/auth/auth_test.go"
snp --no-defaults --include "**/*.md"          # Disable defaults, include only markdown
snp --show-defaults                            # Print default exclude patterns and exit
```

**Filter evaluation:**

Defaults (gitignore + default excludes) are applied as implicit first rules.
`--include` and `--exclude` flags are then evaluated in the order they appear.
The last matching rule wins.

```
defaults → --include/--exclude (in order, last match wins)
```

`--no-defaults` removes the implicit first rules entirely, leaving only
explicit flags.

Both `--include` and `--exclude` are repeatable and can be interleaved:

```bash
# All Go except tests, but keep one specific test
snp --include "**/*.go" --exclude "**/*_test.go" --include "internal/auth/auth_test.go"
```

### Mode 2 — Pick

Directly addresses specific files by exact path or glob pattern.
No traversal, no defaults, no gitignore applied.

```bash
snp --pick "cmd/main.go"
snp --pick "Anthropic/claude-code.md" --pick "OpenAI/o3.md"
snp --pick "Anthropic/**/*.md"
snp --pick "/etc/nginx/nginx.conf"
```

`--pick` is repeatable. Relative paths and globs resolve against the
directory argument (default `.`). Absolute paths are shown as-is in
the snapshot.

`--pick` cannot be combined with `--include`, `--exclude`, or `--no-defaults`.

**Cross-repo example:**

```bash
# From a parent directory containing multiple repos
snp --pick "repo-a/cmd/main.go" --pick "repo-b/cmd/main.go" --pick "repo-c/cmd/main.go"
```

## File Filtering (Mode 1)

### Default Excludes

Run `snp --show-defaults` to see the full list. Includes:

- Directories: `.git/`, `node_modules/`, `.venv/`, `dist/`, `build/`, `target/`, `vendor/`
- Patterns: `*.log`, `*.tmp`, `**/*.snp`
- Files in your `.gitignore`

Use `--no-defaults` to disable all of the above.

### Binary File Handling

Binary files are automatically detected and excluded from content output:

```bash
# Binary files show size metadata instead of content
# logo.png
[Binary file - 45.2 KB - content omitted]
```

**Override binary detection:**

```bash
snp --force-text "**/.env"              # Force .env files to be treated as text
snp --force-binary "**/*.dat"           # Force .dat files to be treated as binary
snp --force-text "**/*.config" --force-binary "data/secret.config"
# Multiple patterns (force-binary always wins in conflicts)
```

**Detection behavior:**

- Empty files are treated as binary
- Content-based detection using MIME types and null byte checking
- Common text formats (JSON, XML, YAML, source code) automatically detected
- `--force-binary` takes precedence over `--force-text` (safer default)

## How It Works

### What Gets Included

- All text files not excluded by defaults or filter rules
- Git log (if `.git/` exists, unless `--exclude-git-log` is used)
- Files matched by `--pick` (Mode 2)

### What Gets Excluded (Mode 1 defaults)

- Directories: `.git/`, `node_modules/`, `.venv/`, `dist/`, `build/`, `target/`, `vendor/`
- Patterns: `*.log`, `*.tmp`, `**/*.snp`
- Files in your `.gitignore`
- Binary files (detected automatically or via `--force-binary`)
- Empty files (treated as binary)

### Output Format

The snapshot begins with a summary, file index, optional git log, and then the file contents:

```text
Generated: 2025-12-14 18:13:40
Total files: 24 (23 text, 1 binary)
Total lines: 2284

# File Index
.gitignore [55-59] (5 lines, 42 bytes)
LICENSE [63-83] (21 lines, 1.1 KB)
README.md [87-399] (313 lines, 7.6 KB)
cmd/snp/main.go [403-511] (109 lines, 2.7 KB)
logo.png [1746-1746] (binary, 43.7 KB)
...

# ----------------------------------------

# Git Log (git adog)
* f79aeb1 (HEAD -> main) add snapshot index and refactor layout construction
...

# ----------------------------------------

# .gitignore
# build folder
dist
...

# ----------------------------------------

# logo.png
[Binary file - 43.7 KB - content omitted]

# ----------------------------------------

# cmd/snp/main.go
package main
...
```

### Safety Features

- Default `./snapshot.snp` overwrites without warning (standard Unix behavior)
- Custom output paths also overwrite without warning
- Output file automatically excluded from snapshot (prevents recursion)
- Binary files excluded by default to prevent corruption

## Use Cases

- Provide complete codebase context to LLMs with easy file navigation
- Generate documentation from source with line-level references
- Code review preparation with exact file locations
- Project snapshots for archival with metadata
- Quick project structure overview via the file index
- Cherry-pick files from multiple repos into one snapshot

## Working with AI Tools

Include these instructions to help AI assistants understand how to work with snapshots effectively:

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

## Advanced Examples

### Preview files before creating snapshot

```bash
snp --dry-run                        # List all files that would be included
snp --dry-run --include "**/*.go"    # Preview with filters
```

### Include only specific file types

```bash
snp --include "**/*.{go,md,txt}"
```

### Exclude tests and generated code

```bash
snp --exclude "**/*_test.go" --exclude "**/generated/**"
```

### Custom output with specific includes

```bash
snp --output docs-snapshot.snp --include "docs/**" --include "*.md"
```

### Snapshot without version control info

```bash
snp --exclude-git-log
```

### Disable defaults entirely

```bash
# Only what you explicitly include
snp --no-defaults --include "**/*.go" --include "**/*.md"
```

### Ordered rules — rescue a file from exclusion

```bash
# Include all Go, exclude tests, but keep one specific test
snp --include "**/*.go" --exclude "**/*_test.go" --include "internal/auth/auth_test.go"
```

### Force specific file types

```bash
# Force .env files to be treated as text (normally detected as binary)
snp --force-text "**/.env" --force-text "**/.editorconfig"

# Force .dat files to be binary (even if they contain text)
snp --force-binary "**/*.dat"
```

### Cherry-pick files across repos

```bash
# From parent directory — no traversal, exact files only
snp --pick "repo-a/cmd/main.go" --pick "repo-b/cmd/main.go"

# Using globs
snp --pick "repo-a/**/*.go" --pick "repo-b/README.md"

# Mix relative and absolute
snp --pick "repo-a/README.md" --pick "/etc/nginx/nginx.conf"
```

### Multiple snapshots in one project

```bash
# Full project snapshot
snp --output full.snp

# Documentation only
snp --output docs.snp --include "docs/**" --include "*.md"

# Source code only
snp --output src.snp --include "src/**" --include "cmd/**"
```

## Release

### Creating a Release

Ensure all changes are merged to `main` and the working tree is clean:

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
make release
```

The `release` script will:

- Verify clean git state and exact tag match
- Run all tests
- Build release artifacts for all platforms
- Verify checksums and binary version

Follow the printed instructions to push the tag and create the GitHub release.

### Post-Release Verification

Verify the published release on a clean system:

```bash
make post-release
```

The `post-release` script will:

- Auto-detect your OS and architecture
- Download the latest release binary and checksum
- Verify the SHA256 checksum
- Verify the binary runs and reports correct version

## License

MIT License — see [LICENSE](LICENSE) file for details.
