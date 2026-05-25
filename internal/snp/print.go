package snp

import "fmt"

func (s Snapshot) PrintRawEntries() {
	for _, e := range s.Entries {
		p := e.Path
		if e.IsDir {
			p = p + "/"
		}
		fmt.Printf("%s\n", p)
	}
}

// formatSize formats byte size in human-readable format
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
