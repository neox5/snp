<br/>

<div align="center">
  <img src="logo.png" alt="snap logo" width="200"/>
</div>

<br/>

# snap

[![Release](https://img.shields.io/github/v/release/neox5/snap)](https://github.com/neox5/snap/releases)
![Go Version](https://img.shields.io/github/go-mod/go-version/neox5/snap)
![License](https://img.shields.io/github/license/neox5/snap)

A CLI tool that concatenates all readable files in a project into a single deterministic snapshot file for inspection, sharing, and machine processing.

## Quick Start

```bash
# Install (Linux example)
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64
chmod +x snap-linux-amd64
sudo mv snap-linux-amd64 /usr/local/bin/snap

# Run in any project directory
cd /path/to/your/project
snap
```

Creates `./snap.txt` with all project files concatenated.

## Usage

### Basic Usage

```bash
snap                    # Create snap.txt in current directory
snap /path/to/project   # Create snap.txt from specified directory
```

### Output Control

```bash
snap --output custom.txt              # Custom output path
snap --exclude-git-log                # Omit Git log section
snap --dry-run                        # List files without creating output
```

### File Filtering

```bash
snap --include "src/**/*.go"                    # Include only Go files in src/
snap --exclude "**/*_test.go"                   # Exclude test files
snap --include "*.log" --exclude "secret.log"   # Combine filters
snap --exclude "*.tmp" --exclude "*.log"        # Multiple exclude patterns
snap --include "**/*.go" --include "**/*.md"    # Multiple include patterns
```

**Note:** Both `--include` and `--exclude` flags can be specified multiple times.

**Filter precedence** (highest to lowest):

1. `--exclude` patterns (final, cannot be overridden)
2. `--include` patterns (override defaults and .gitignore)
3. `.gitignore` patterns
4. Default excludes (node_modules/, .git/, dist/, etc.)

### Binary File Handling

Binary files are automatically detected and excluded from content output:

```bash
# Binary files show size metadata instead of content
# logo.png
[Binary file - 45.2 KB - content omitted]
```

**Override binary detection:**

```bash
snap --force-text "**/.env"              # Force .env files to be treated as text
snap --force-binary "**/*.dat"           # Force .dat files to be treated as binary
snap --force-text "**/*.config" --force-binary "data/secret.config"
# Multiple patterns (force-binary always wins in conflicts)
```

**Detection behavior:**

- Empty files are treated as binary
- Content-based detection using MIME types and null byte checking
- Common text formats (JSON, XML, YAML, source code) automatically detected
- `--force-binary` takes precedence over `--force-text` (safer default)

## How It Works

### What Gets Included

- All text files not matching exclude patterns
- Git log (if `.git/` exists, unless `--exclude-git-log` is used)
- Files matching `--include` patterns override .gitignore
- Files forced as text via `--force-text`

### What Gets Excluded

- Directories: `.git/`, `node_modules/`, `.venv/`, `dist/`, `build/`, `target/`, `vendor/`
- Patterns: `*.log`, `*.tmp`, `**/snap.txt`
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
cmd/snap/main.go [403-511] (109 lines, 2.7 KB)
logo.png [1746-1746] (binary, 43.7 KB)
...

# ----------------------------------------

# Git Log (git adog)
* f79aeb1 (HEAD -> main) add snapshot index and refactor layout construction
* 653b8ab refactor snapshot creation and rendering
...

# ----------------------------------------

# .gitignore
# build folder
dist

# snap file
snap.txt

# ----------------------------------------

# LICENSE
MIT License
...

# ----------------------------------------

# logo.png
[Binary file - 43.7 KB - content omitted]

# ----------------------------------------

# cmd/snap/main.go
package main
...
```

**Summary section:**

- Generation timestamp
- Total file count (text and binary breakdown)
- Total lines in the snapshot

**File index:**

- `filename [start-end]` - Line range in the snapshot for quick navigation
- `(N lines, size)` - For text files
- `(binary, size)` - For binary files

**File sections:**

- Each section starts with `# relative/path/from/root`
- Text files: full content
- Binary files: metadata with size (content omitted)

**Navigation:**
Use the file index to quickly locate files by line number in the snapshot.

### Safety Features

- Default `./snap.txt` always overwrites (safe for repeated runs)
- Custom output paths require explicit `--output` flag to overwrite existing files
- Output file automatically excluded from snapshot (prevents recursion)
- Binary files excluded by default to prevent corruption

## Use Cases

- Provide complete codebase context to LLMs with easy file navigation
- Generate documentation from source with line-level references
- Code review preparation with exact file locations
- Project snapshots for archival with metadata
- Quick project structure overview via the file index

## Working with AI Tools

Include these instructions to help AI assistants understand how to work with snapshots effectively:

```
## Working with Repository Snapshots

Snapshots were generated with [snap](https://github.com/neox5/snap).

**Rules for working with snap.txt:**

- snap.txt is a READ-ONLY reference document
- DO NOT modify snap.txt directly
- DO NOT create updated versions of snap.txt
- Changes must target actual source files in their original locations
- User will regenerate snap.txt by running snap after changes

**How to use snap.txt:**

1. File index is at the top with line ranges
2. Each file section starts with # filepath
3. Binary files show size metadata instead of content
4. Git log (if present) shows recent commits
```

## Installation

### Prebuilt Binaries

**Linux (amd64)**

```bash
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64.sha256
sha256sum -c snap-linux-amd64.sha256
chmod +x snap-linux-amd64
sudo mv snap-linux-amd64 /usr/local/bin/snap
```

**macOS (Apple Silicon)**

```bash
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-darwin-arm64
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-darwin-arm64.sha256
shasum -a 256 -c snap-darwin-arm64.sha256
chmod +x snap-darwin-arm64
sudo mv snap-darwin-arm64 /usr/local/bin/snap
```

**Available platforms:**

- `snap-linux-amd64` / `snap-linux-arm64`
- `snap-darwin-amd64` / `snap-darwin-arm64`
- `snap-windows-amd64.exe` / `snap-windows-arm64.exe`

### Via Go

Requires Go 1.22+

```bash
go install github.com/neox5/snap/cmd/snap@latest
```

Ensure `$HOME/go/bin` is in your `PATH`.

### From Source

```bash
git clone https://github.com/neox5/snap
cd snap
make build-local
sudo mv dist/snap /usr/local/bin/snap
```

### Verify Installation

```bash
snap --version
```

## Advanced Examples

### Preview files before creating snapshot

```bash
snap --dry-run                        # List all files that would be included
snap --dry-run --include "**/*.go"    # Preview with filters
```

### Include only specific file types

```bash
snap --include "**/*.{go,md,txt}"
```

### Exclude tests and generated code

```bash
snap --exclude "**/*_test.go" --exclude "**/generated/**"
```

### Custom output with specific includes

```bash
snap --output docs-snapshot.txt --include "docs/**" --include "*.md"
```

### Snapshot without version control info

```bash
snap --exclude-git-log
```

### Force specific file types

```bash
# Force .env files to be treated as text (normally detected as binary)
snap --force-text "**/.env" --force-text "**/.editorconfig"

# Force .dat files to be binary (even if they contain text)
snap --force-binary "**/*.dat"

# Combine with other filters
snap --include "config/**" --force-text "**/.env"
```

### Verify filtering before snapshot

```bash
# Check which files will be included
snap --dry-run --exclude "**/*_test.go"

# If satisfied, create the snapshot
snap --exclude "**/*_test.go"
```

### Handle edge cases with force flags

```bash
# Custom binary format that looks like text
snap --force-binary "**/*.myformat"

# Text file with unusual extension
snap --force-text "data/config.bin"

# Force takes precedence over detection
snap --force-text "**/*.log"  # Include log files as text
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
