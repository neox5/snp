# snap

[![Release](https://img.shields.io/github/v/release/neox5/snap)](https://github.com/neox5/snap/releases)
![Go Version](https://img.shields.io/github/go-mod/go-version/neox5/snap)
![License](https://img.shields.io/github/license/neox5/snap)

`snap` concatenates all readable files of a project into a single deterministic
snapshot file (`snap.txt`) for inspection, sharing, and machine processing.

---

## Installation

### Prebuilt Binaries (Recommended)

Download, verify, and install in one step sequence:

```bash
# download binary
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64

# download checksum
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64.sha256

# verify integrity
sha256sum -c snap-linux-amd64.sha256

# install
chmod +x snap-linux-amd64
sudo mv snap-linux-amd64 /usr/local/bin/snap
````

Replace `linux-amd64` with your platform:

* `linux-arm64`
* `darwin-amd64`
* `darwin-arm64`
* `windows-amd64.exe`
* `windows-arm64.exe`

---

### Via `go install`

Requires **Go 1.22+**.

```bash
go install github.com/neox5/snap/cmd/snap@latest
```

Ensure Go’s bin directory is in your `PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

---

### From Source

```bash
git clone https://github.com/neox5/snap
cd snap
make build-local
sudo mv dist/snap /usr/local/bin/snap
```

---

### Verify Installation

```bash
snap --version
```

---

## Usage

Run `snap` in any project directory:

```bash
snap
```

This creates:

```text
./snap.txt
```

With custom filtering:

```bash
snap --include "src/**/*.go" --exclude "**/*_test.go"
```

With Git log omitted:

```bash
snap --exclude-git-log
```

---

## Output Structure (`snap.txt`)

The snapshot file is structured as a continuous, deterministic stream:

```text
# Git Log (git adog3)
<optional git history>

# ----------------------------------------

# path/to/file.go
<file contents>

# another/file.txt
<file contents>
```

Each file is preceded by a header:

```text
# relative/path/from/project/root
```

---

## Text vs. Binary Handling

* **Readable text files**
  → Fully inlined into `snap.txt`

* **Binary / media files by extension**
  → Listed with placeholder:

  ```text
  [Binary file - content omitted]
  ```

* **Non-text, unknown formats**
  → Omitted with:

  ```text
  [Non-text file - content omitted]
  ```

---

## Default Binary Extensions

Content is omitted for common binary formats including:

```text
exe, dll, so, dylib, zip, tar, gz, 7z,
png, jpg, jpeg, gif, bmp, tiff, webp,
mp4, mp3, avi, mov, mkv, wav, flac,
pdf, doc, docx, xls, xlsx, ppt, pptx
```

---

## License

MIT License — see `LICENSE` file for full text.

