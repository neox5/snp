package snp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/matcher"
)

func Collect(c *config.Config) ([]Entry, error) {
	entries := []Entry{}

	srcDir := filepath.ToSlash(c.SourceDir) // path normalization (unix slashes)
	srcDir = filepath.Clean(srcDir)         // produce the shortest possible path
	root := strings.TrimRight(srcDir, "/")  // remove trailing slash for accurate rootDepth
	rootDepth := strings.Count(root, "/")   // count the leftover slashes to determine depth

	// pathDepth for calculating the current depth of a path
	pathDepth := func(p string) int {
		return strings.Count(p, "/") - rootDepth
	}

	m := matcher.New(c.BuildMatcherRules())

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				return nil
			}
			return err
		}

		if path == root {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		if !m.ShouldInclude(relPath, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		depth := pathDepth(relPath)

		// ### directories
		if d.IsDir() {
			if depth >= c.Depth {
				info, err := dirInfo(path) // using the original path from WalkDir
				if err != nil {
					return err
				}
				fmt.Printf("[%d] %s (%d items, %s)\n", depth, relPath, info.ItemCount, formatBytes(info.TotalSize))
				return filepath.SkipDir
			}
			return nil // proceed with containing items
		}

		// ### files
		fmt.Printf("[%d] %s\n", depth, relPath)
		return nil
	})

	return entries, err
}

type DirInfo struct {
	ItemCount int
	TotalSize int64
}

func dirInfo(path string) (DirInfo, error) {
	var info DirInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return info, err
	}
	info.ItemCount = len(entries)

	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}
			info.TotalSize += fileInfo.Size()
		}
		return nil
	})

	return info, err
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
