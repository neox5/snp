package snp

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/matcher"
)

// Entry represents an entry in the snapshot.
// depending on IsDir it represents a directory or a file.
type Entry struct {
	IsDir bool

	// common
	Path string
	Size int64 // number of bytes

	// file specific
	IsBinary  bool // decides if it has a line or metadata representation
	Lines     []string
	StartLine int // startline in the snapshot

	// dir specific
	ItemCount int // the number of direct children in the directory
}

func NewDir(path string) *Entry {
	return &Entry{
		IsDir: true,
		Path:  path,
	}
}

func NewFile(path string) *Entry {
	return &Entry{
		IsDir: false,
		Path:  path,
	}
}

// checkForceOverride checks force-text and force-binary patterns.
// Returns (isBinary, overridden).
// Precedence: force-binary always wins (safe side).
func checkForceOverride(relPath string, forceBinaryPatterns, forceTextPatterns []string) (isBinary bool, overridden bool) {
	relUnix := filepath.ToSlash(relPath)

	// Check force-binary first (highest precedence)
	for _, pattern := range forceBinaryPatterns {
		matched, err := matcher.Match(pattern, relUnix, false)
		if err == nil && matched {
			return true, true
		}
	}

	// Check force-text (lower precedence)
	for _, pattern := range forceTextPatterns {
		matched, err := matcher.Match(pattern, relUnix, false)
		if err == nil && matched {
			return false, true
		}
	}

	return false, false
}

// detectBinary checks if a file is binary
func detectBinary(f *os.File) (bool, error) {
	stat, err := f.Stat()
	if err != nil {
		return false, err
	}
	size := stat.Size()

	if size == 0 {
		return true, nil
	}

	// Reset read position to start
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return false, err
	}

	buf := make([]byte, 512)
	n, readErr := f.Read(buf)
	if readErr != nil && readErr != io.EOF {
		return false, readErr
	}

	contentType := http.DetectContentType(buf[:n])

	if strings.HasPrefix(contentType, "text/") {
		return false, nil
	}

	textAppTypes := map[string]bool{
		"application/json":       true,
		"application/xml":        true,
		"application/javascript": true,
	}
	if textAppTypes[contentType] {
		return false, nil
	}

	for i := range n {
		if buf[i] == 0 {
			return true, nil
		}
	}

	return false, nil
}
