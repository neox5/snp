package file

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// DetectBinary checks if a file is binary
func DetectBinary(path string, size int64) (bool, error) {
	// Empty files treated as binary
	if size == 0 {
		return true, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read up to 512 bytes for detection
	buf := make([]byte, 512)
	n, readErr := f.Read(buf)
	if readErr != nil && readErr != io.EOF {
		return false, readErr
	}

	// Use http.DetectContentType
	contentType := http.DetectContentType(buf[:n])

	// Check if text content type
	if strings.HasPrefix(contentType, "text/") {
		return false, nil
	}

	// Known text application types
	textAppTypes := map[string]bool{
		"application/json":       true,
		"application/xml":        true,
		"application/javascript": true,
	}
	if textAppTypes[contentType] {
		return false, nil
	}

	// Fallback: check for null bytes
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}

	// No null bytes found, treat as text
	return false, nil
}

// CheckForceOverride checks force-text and force-binary patterns
// Returns (isBinary, overridden)
// Precedence: force-binary always wins (safe side)
func CheckForceOverride(relPath string, forceTextPatterns, forceBinaryPatterns []string) (isBinary bool, overridden bool) {
	relUnix := filepath.ToSlash(relPath)

	// Check force-binary first (highest precedence)
	if len(forceBinaryPatterns) > 0 {
		matcher := gitignore.CompileIgnoreLines(forceBinaryPatterns...)
		if matcher.MatchesPath(relUnix) {
			return true, true
		}
	}

	// Check force-text (lower precedence)
	if len(forceTextPatterns) > 0 {
		matcher := gitignore.CompileIgnoreLines(forceTextPatterns...)
		if matcher.MatchesPath(relUnix) {
			return false, true
		}
	}

	return false, false
}
