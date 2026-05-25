package snp

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/matcher"
)

// Snapshot represents the complete snapshot data.
type Snapshot struct {
	GitLogLines GitLogLines
	Entries     []*Entry
	Layout      []Content
}

// GitLogLines represents git log output.
type GitLogLines []string

func New(c *config.Config) (*Snapshot, error) {
	snap := &Snapshot{}
	entries := []*Entry{}

	srcDir := filepath.ToSlash(c.SourceDir) // path normalization (unix slashes)
	srcDir = filepath.Clean(srcDir)         // produce the shortest possible path
	root := strings.TrimRight(srcDir, "/")  // remove trailing slash for accurate rootDepth
	rootDepth := strings.Count(root, "/")   // count the leftover slashes to determine depth

	// pathDepth for calculating the current depth of a path
	pathDepth := func(p string) int {
		d := strings.Count(p, "/") - rootDepth
		// fmt.Printf("[r/p]: %d/%d - %s\n", rootDepth, d, p)
		return d
	}

	m := matcher.New(c.BuildMatcherRules())

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, wErr error) error {
		if wErr != nil {
			if errors.Is(wErr, fs.ErrPermission) {
				return nil
			}
			return wErr
		}

		// skip root
		if path == root {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		// ### filter with matcher
		if !m.ShouldInclude(relPath, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		depth := pathDepth(path)

		// ### directories
		if d.IsDir() {
			// collapsed directory case:
			// when a directory is at the end of depth, it gets added to the entires
			if c.Depth >= 0 && depth >= c.Depth {
				entries = append(entries, NewDir(path))
				return filepath.SkipDir // stop from traversing further down
			}
			return nil // proceed with containing items
		}

		// ### files
		entries = append(entries, NewFile(path))
		return nil
	})
	if err != nil {
		return nil, err
	}

	snap.Entries = entries

	return snap, nil
}

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
