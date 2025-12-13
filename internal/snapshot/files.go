package snapshot

import (
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

func isFileBinary(path string, size int64) (bool, string, error) {
	// Empty files treated as binary
	if size == 0 {
		return true, "application/octet-stream", nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, "", err
	}
	defer f.Close()

	// Read up to 512 bytes for detection
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, "", err
	}

	// Use http.DetectContentType
	contentType := http.DetectContentType(buf[:n])

	// Check if text content type
	if strings.HasPrefix(contentType, "text/") {
		return false, contentType, nil
	}

	// Known text application types
	textAppTypes := map[string]bool{
		"application/json":       true,
		"application/xml":        true,
		"application/javascript": true,
	}
	if textAppTypes[contentType] {
		return false, contentType, nil
	}

	// Fallback: check for null bytes
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, contentType, nil
		}
	}

	// Default to binary when uncertain (safer)
	return true, contentType, nil
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
