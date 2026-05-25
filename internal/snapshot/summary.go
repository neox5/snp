package snapshot

import (
	"fmt"
	"io"
)

// Summary holds snapshot metadata and implements Content.
type Summary struct {
	Timestamp   string
	TotalFiles  int
	TextFiles   int
	BinaryFiles int
	TotalSize   int64
	TotalLines  int
}

func (s *Summary) LineCount() int { return 4 }

func (s *Summary) WriteTo(w io.Writer) (int, error) {
	lines := [4]string{
		"Generated: " + s.Timestamp,
		fmt.Sprintf("Total files: %d (%d text, %d binary)", s.TotalFiles, s.TextFiles, s.BinaryFiles),
		"Total size: " + formatSize(s.TotalSize),
		fmt.Sprintf("Total lines: %d", s.TotalLines),
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return 0, err
		}
	}
	return 0, nil
}
