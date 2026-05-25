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

func Collect(c *config.Config) ([]*Entry, error) {
	entries := []*Entry{}

	srcDir := filepath.ToSlash(c.SourceDir) // path normalization (unix slashes)
	srcDir = filepath.Clean(srcDir)         // produce the shortest possible path
	root := strings.TrimRight(srcDir, "/")  // remove trailing slash for accurate rootDepth
	rootDepth := strings.Count(root, "/")   // count the leftover slashes to determine depth

	// pathDepth for calculating the current depth of a path
	pathDepth := func(p string) int {
		d := strings.Count(p, "/") - rootDepth
		fmt.Printf("[r/p]: %d/%d - %s\n", rootDepth, d, p)
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

		if path == root {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// absPath, err := filepath.Abs(path)
		// if err != nil {
		// 	return err
		// }
		relPath = filepath.ToSlash(relPath)
		// absPath = filepath.ToSlash(absPath)

		if !m.ShouldInclude(relPath, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		depth := pathDepth(path)

		// ### directories
		if d.IsDir() {
			if c.Depth >= 0 && depth >= c.Depth {
				info, err := dirInfo(path) // using the original path from WalkDir
				if err != nil {
					return err
				}
				entries = append(entries, NewDir(path, info.TotalSize, info.ItemCount))
				return filepath.SkipDir
			}
			return nil // proceed with containing items
		}

		// ### files
		info, err := d.Info()
		if err != nil {
			return nil
		}
		entries = append(entries, NewFile(path, info.Size(), false))
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
