package snapshot

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// FileInfo holds metadata for a file to be included in the snapshot.
type FileInfo struct {
	RelPath     string // Relative path from source directory (Unix-style)
	FullPath    string // Absolute path for reading
	Size        int64  // File size in bytes
	IsBinary    bool   // Whether file should be treated as binary
	ContentType string // MIME type detected or forced
	LineCount   int    // Number of lines in file (0 for binary)
	StartLine   int    // Line number in snap.txt where file content begins
}

func appendFileWithMetadata(w io.Writer, file FileInfo) error {
	if _, err := fmt.Fprintf(w, "# %s\n", file.RelPath); err != nil {
		return err
	}

	if file.IsBinary {
		sizeStr := formatSize(file.Size)
		if _, err := fmt.Fprintf(w, "[Binary file - %s - content omitted]\n\n", sizeStr); err != nil {
			return err
		}
		return nil
	}

	// Text file handling
	f, err := os.Open(file.FullPath)
	if err != nil {
		return fmt.Errorf("cannot open %q: %w", file.FullPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("cannot copy %q: %w", file.FullPath, err)
	}

	if _, err := io.WriteString(w, "\n\n"); err != nil {
		return err
	}

	return nil
}

// checkForceOverride checks force-text and force-binary patterns.
// Returns (isBinary, contentType, overridden).
// Precedence: force-binary always wins (safe side).
func checkForceOverride(relPath string, cfg Config) (isBinary bool, contentType string, overridden bool) {
	relUnix := filepath.ToSlash(relPath)

	// Check force-binary first (highest precedence)
	if len(cfg.ForceBinaryPatterns) > 0 {
		matcher := gitignore.CompileIgnoreLines(cfg.ForceBinaryPatterns...)
		if matcher.MatchesPath(relUnix) {
			return true, "application/octet-stream", true
		}
	}

	// Check force-text (lower precedence)
	if len(cfg.ForceTextPatterns) > 0 {
		matcher := gitignore.CompileIgnoreLines(cfg.ForceTextPatterns...)
		if matcher.MatchesPath(relUnix) {
			return false, "text/plain", true
		}
	}

	return false, "", false // No override
}

func isFileBinary(path string, size int64) (isBinary bool, contentType string, lineCount int, err error) {
	// Empty files treated as binary
	if size == 0 {
		return true, "application/octet-stream", 1, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, "", 0, err
	}
	defer f.Close()

	// Read up to 512 bytes for detection
	buf := make([]byte, 512)
	n, readErr := f.Read(buf)
	if readErr != nil && readErr != io.EOF {
		return false, "", 0, readErr
	}

	// Use http.DetectContentType
	contentType = http.DetectContentType(buf[:n])

	// Check if text content type
	if strings.HasPrefix(contentType, "text/") {
		isBinary = false
	} else {
		// Known text application types
		textAppTypes := map[string]bool{
			"application/json":       true,
			"application/xml":        true,
			"application/javascript": true,
		}
		if textAppTypes[contentType] {
			isBinary = false
		} else {
			// Fallback: check for null bytes
			isBinary = true
			for i := 0; i < n; i++ {
				if buf[i] == 0 {
					return true, contentType, 1, nil
				}
			}
			// No null bytes found, treat as text
			isBinary = false
		}
	}

	// If binary, return now
	if isBinary {
		return true, contentType, 1, nil
	}

	// Count lines for text files
	_, err = f.Seek(0, 0) // Reset to start
	if err != nil {
		return false, contentType, 0, err
	}

	scanner := bufio.NewScanner(f)
	lineCount = 0
	for scanner.Scan() {
		lineCount++
	}

	if err = scanner.Err(); err != nil {
		return false, contentType, 0, err
	}

	return false, contentType, lineCount, nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes == 0:
		return "0 bytes"
	case bytes == 1:
		return "1 byte"
	case bytes < KB:
		return fmt.Sprintf("%d bytes", bytes)
	case bytes < MB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	case bytes < GB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	}
}
