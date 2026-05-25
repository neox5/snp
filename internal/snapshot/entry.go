package snapshot

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/neox5/snp/internal/matcher"
)

// Entry represents an entry in the snapshot.
// Depending on IsDir it represents a directory or a file.
type Entry struct {
	IsDir bool

	// common
	Path    string
	RelPath string
	Size    int64

	// file specific
	IsBinary  bool
	Lines     []string
	StartLine int
	EndLine   int

	// dir specific
	ItemCount int
}

func NewDir(path, relPath string) *Entry {
	return &Entry{IsDir: true, Path: path, RelPath: relPath}
}

func NewFile(path, relPath string) *Entry {
	return &Entry{IsDir: false, Path: path, RelPath: relPath}
}

// LineCount implements Content.
func (e *Entry) LineCount() int { return len(e.Lines) }

// WriteTo implements Content.
func (e *Entry) WriteTo(w io.Writer) (int, error) {
	for _, line := range e.Lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

// checkForceOverride checks force-text and force-binary patterns.
// Returns (isBinary, overridden).
// Precedence: force-binary always wins (safe side).
func checkForceOverride(path string, forceBinary, forceText []string) (isBinary bool, overridden bool) {
	for _, pattern := range forceBinary {
		if ok, _ := matcher.Match(pattern, path, false); ok {
			return true, true
		}
	}
	for _, pattern := range forceText {
		if ok, _ := matcher.Match(pattern, path, false); ok {
			return false, true
		}
	}
	return false, false
}

// detectBinary reports whether a file is binary by sniffing its content.
func detectBinary(f *os.File) (bool, error) {
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return false, err
	}
	ct := http.DetectContentType(buf[:n])
	return !strings.HasPrefix(ct, "text/"), nil
}
