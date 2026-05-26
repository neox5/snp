package snapshot

import (
	"fmt"
	"io"
)

// Content represents anything that can be written to the snapshot output.
type Content interface {
	LineCount() int
	WriteTo(io.Writer) (int64, error)
}

// header represents a section header line.
type header struct {
	Text string
}

func (h *header) LineCount() int { return 1 }

func (h *header) WriteTo(w io.Writer) (int64, error) {
	_, err := fmt.Fprintln(w, "# "+h.Text)
	return 0, err
}

// divider represents the blank + separator + blank lines between sections.
type divider struct{}

func (d *divider) LineCount() int { return 1 }

func (d *divider) WriteTo(w io.Writer) (int64, error) {
	_, err := fmt.Fprintln(w, "# ---")
	return 0, err
}

// formatSize formats byte size in human-readable format.
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
