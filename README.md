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

## How It Works

### What Gets Included

- All text files not matching exclude patterns
- Git log (if `.git/` exists, unless `--exclude-git-log` is used)
- Files matching `--include` patterns override .gitignore

### What Gets Excluded

- Directories: `.git/`, `node_modules/`, `.venv/`, `dist/`, `build/`, `target/`, `vendor/`
- Patterns: `*.log`, `*.tmp`, `**/snap.txt`
- Files in your `.gitignore`
- Binary files (shown as `[Binary file - content omitted]`)
- Files matching `--exclude` patterns

### Output Format

```text
# Git Log (git adog)
* ad7b219 (HEAD -> main) improve readme
* 79e96b6 improve post release check

# ----------------------------------------

# cmd/snap/main.go
package main
...

# internal/snapshot/snapshot.go
package snapshot
...
```

Each file section starts with `# relative/path/from/root` followed by the file contents.

### Safety Features

- Default `./snap.txt` always overwrites (safe for repeated runs)
- Custom output paths require explicit `--output` flag to overwrite existing files
- Output file automatically excluded from snapshot (prevents recursion)

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

## Use Cases

- Provide complete codebase context to LLMs
- Generate documentation from source
- Code review preparation
- Project archival and snapshots
- Quick sharing of entire project structure

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

### Verify filtering before snapshot

```bash
# Check which files will be included
snap --dry-run --exclude "**/*_test.go"

# If satisfied, create the snapshot
snap --exclude "**/*_test.go"
```

## License

MIT License — see [LICENSE](LICENSE) file for details.
