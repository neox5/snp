package file

import (
	"bufio"
	"fmt"
	"os"
)

// File represents a file in the snapshot
type File struct {
	RelPath   string
	FullPath  string
	Size      int64
	IsBinary  bool
	Lines     []string
	StartLine int
}

// New creates a new File and loads its content
func New(relPath, fullPath string, size int64, isBinary bool) (*File, error) {
	f := &File{
		RelPath:  relPath,
		FullPath: fullPath,
		Size:     size,
		IsBinary: isBinary,
	}

	// Load content
	if err := f.LoadContent(); err != nil {
		return nil, err
	}

	return f, nil
}

// LoadContent loads the file's content into Lines
func (f *File) LoadContent() error {
	if f.IsBinary {
		// Binary: create placeholder line
		sizeStr := FormatSize(f.Size)
		f.Lines = []string{
			fmt.Sprintf("[Binary file - %s - content omitted]", sizeStr),
		}
		return nil
	}

	// Text: load actual lines
	lines, err := loadFileLines(f.FullPath)
	if err != nil {
		return fmt.Errorf("failed to load content: %w", err)
	}
	f.Lines = lines
	return nil
}

// loadFileLines reads a file into a slice of lines
func loadFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// FormatSize formats byte size in human-readable format
func FormatSize(bytes int64) string {
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
