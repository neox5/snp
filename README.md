# snap

`snap` concatenates all readable files of a project into a single snapshot file.
It is intended for fast full-context inspection, sharing, and machine processing.

---

## Output Structure (`snap.txt`)

The snapshot file is a plain text file with a strict, repeatable structure:

1. **Optional Git log section** (if enabled and a Git repo is detected)
2. **A sequence of file blocks**, one per included file

Each file is written as:

```text
# relative/path/to/file
<file contents>

````

* The `#` header always contains the path relative to the scanned root.
* Two newline characters separate file blocks.
* File order is stable and sorted by path.

---

## Text vs. Binary Files

`snap` classifies files by **extension only**.

### Text files

* All non-binary extensions
* Printed **in full**
* No truncation or size limits

Example:

```text
# main.go
package main

func main() {}
```

### Binary files

Binary file contents are **never included**. Only a placeholder is written:

```text
# image.png
[Binary file - content omitted]
```

---

## Default Binary Extensions

These extensions are treated as binary by default:

* Executables & libraries
  `*.exe, *.dll, *.so, *.dylib, *.a`

* Archives
  `*.zip, *.tar, *.gz, *.7z, *.rar`

* Images
  `*.png, *.jpg, *.jpeg, *.gif, *.bmp, *.tiff, *.webp`

* Video
  `*.mp4, *.avi, *.mov, *.mkv`

* Audio
  `*.mp3, *.wav, *.flac`

* Documents
  `*.pdf, *.doc, *.docx, *.xls, *.xlsx, *.ppt, *.pptx`

All other extensions are treated as text and printed in full.

---

## Basic Usage

```bash
snap
```

Concatenates the current directory into `./snap.txt`.

```bash
snap src/
```

Concatenates the `src/` directory into `./snap.txt`.

```bash
snap --output project.txt
```

Writes the snapshot to `project.txt`.

---

## Include and Exclude

Exclude files or directories:

```bash
snap --exclude "**/*_test.go" --exclude "vendor/**"
```

Rescue files from excluded paths:

```bash
snap --exclude "vendor/**" --include "vendor/my-lib/**"
```

---

## Git Log Section

By default, if the directory is a Git repository, a Git log section is added at
the top of the snapshot. Disable it with:

```bash
snap --exclude-git-log
```

---

## Example Input

Directory structure:

```text
project/
  main.go
  go.mod
  internal/app/app.go
  internal/app/app_test.go
  assets/logo.png
```

Command:

```bash
snap --exclude "**/*_test.go"
```

---

## Example Output (`snap.txt`)

```text
# main.go
package main

func main() {}

# go.mod
module example.com/project

# internal/app/app.go
package app

func Run() {}

# assets/logo.png
[Binary file - content omitted]
```

---

## Build

Local build:

```bash
make build-local
```

Multi-platform build:

```bash
make build
```

Artifacts are written to `./dist/`.
