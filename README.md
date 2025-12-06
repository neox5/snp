# snap

`snap` concatenates all readable files of a project into a single
snapshot file. It is intended for fast full-context inspection, sharing,
and machine processing.

------------------------------------------------------------------------

## Installation

### Prebuilt Binaries (Recommended)

Download the binary and corresponding `.sha256` file for your platform
from the GitHub Releases page.

Example (Linux amd64):

``` bash
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64
curl -LO https://github.com/neox5/snap/releases/latest/download/snap-linux-amd64.sha256
```

Verify integrity:

``` bash
sha256sum -c snap-linux-amd64.sha256
```

Install:

``` bash
chmod +x snap-linux-amd64
sudo mv snap-linux-amd64 /usr/local/bin/snap
```

------------------------------------------------------------------------

### Via `go install`

Requires **Go 1.22+**.

``` bash
go install github.com/neox5/snap/cmd/snap@latest
```

Ensure Go's bin directory is in your `PATH`:

``` bash
export PATH="$HOME/go/bin:$PATH"
```

------------------------------------------------------------------------

### From Source

``` bash
git clone https://github.com/neox5/snap
cd snap
make build-local
sudo mv dist/snap /usr/local/bin/snap
```

------------------------------------------------------------------------

### Verify Installation

``` bash
snap --version
```

------------------------------------------------------------------------

## Output Structure (`snap.txt`)

The snapshot file is a plain text file with a strict, repeatable
structure:

1.  **Optional Git log section** (if enabled and a Git repo is detected)
2.  **A sequence of file blocks**, one per included file

Each file is written as:

``` text
# relative/path/to/file
<file contents>
```

-   The `#` header always contains the path relative to the scanned
    root.
-   Two newline characters separate file blocks.
-   File order is stable and sorted by path.

------------------------------------------------------------------------

## Text vs. Binary Files

`snap` classifies files by **extension only**.

### Text files

-   All non-binary extensions
-   Printed **in full**
-   No truncation or size limits

### Binary files

Binary file contents are **never included**. Only a placeholder is
written:

``` text
# image.png
[Binary file - content omitted]
```

------------------------------------------------------------------------

## Default Binary Extensions

-   Executables & libraries\
    `*.exe, *.dll, *.so, *.dylib, *.a`

-   Archives\
    `*.zip, *.tar, *.gz, *.7z, *.rar`

-   Images\
    `*.png, *.jpg, *.jpeg, *.gif, *.bmp, *.tiff, *.webp`

-   Video\
    `*.mp4, *.avi, *.mov, *.mkv`

-   Audio\
    `*.mp3, *.wav, *.flac`

-   Documents\
    `*.pdf, *.doc, *.docx, *.xls, *.xlsx, *.ppt, *.pptx`

All other extensions are treated as text and printed in full.

------------------------------------------------------------------------

## License

MIT License — see `LICENSE` file for full text.
