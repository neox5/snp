package file

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/filter"
)

// DirEntry represents a collapsed directory when depth limit is reached.
type DirEntry struct {
	RelPath   string
	ItemCount int
	Size      int64
}

// Collect discovers, analyzes, and loads files to include in the snapshot.
// depth -1 means full traversal. depth >= 0 limits traversal depth.
// Returns: files, collapsedDirs, textCount, binaryCount, error
func Collect(sourceDir, outputPath string, filterRules []filter.Rule, forceTextPatterns, forceBinaryPatterns []string, depth int) ([]*File, []DirEntry, int, int, error) {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("cannot resolve source directory: %w", err)
	}

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("cannot resolve output path: %w", err)
	}

	matcher, err := filter.New(absSourceDir, filterRules)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	var files []*File
	var collapsedDirs []DirEntry
	var textCount, binaryCount int

	err = filepath.WalkDir(absSourceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrPermission) {
				return nil
			}
			return walkErr
		}

		relPath, err := filepath.Rel(absSourceDir, path)
		if err != nil {
			return nil
		}
		relUnix := filepath.ToSlash(relPath)

		// calculate current depth (root = 0)
		currentDepth := 0
		if relUnix != "." {
			currentDepth = strings.Count(relUnix, "/") + 1
		}

		if d.IsDir() {
			if relUnix == "." {
				return nil
			}

			// check if dir should be excluded entirely
			if !matcher.ShouldInclude(relUnix + "/") {
				return filepath.SkipDir
			}

			// depth limit reached — collapse this dir
			if depth >= 0 && currentDepth > depth {
				itemCount, size, err := countDir(path)
				if err != nil {
					return err
				}
				collapsedDirs = append(collapsedDirs, DirEntry{
					RelPath:   relUnix + "/",
					ItemCount: itemCount,
					Size:      size,
				})
				return filepath.SkipDir
			}

			return nil
		}

		// file handling
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil
		}
		if samePath(absPath, absOutput) {
			return nil
		}

		if !matcher.ShouldInclude(relUnix) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		fileSize := info.Size()

		var isBinary bool
		isBinaryOverride, overridden := CheckForceOverride(relUnix, forceTextPatterns, forceBinaryPatterns)
		if overridden {
			isBinary = isBinaryOverride
		} else {
			isBinary, err = DetectBinary(path, fileSize)
			if err != nil {
				return nil
			}
		}

		f, err := New(relUnix, path, fileSize, isBinary)
		if err != nil {
			return err
		}

		files = append(files, f)

		if isBinary {
			binaryCount++
		} else {
			textCount++
		}

		return nil
	})
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return files, collapsedDirs, textCount, binaryCount, nil
}

// countDir counts all items and total size under a directory recursively.
func countDir(path string) (itemCount int, totalSize int64, err error) {
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrPermission) {
				return nil
			}
			return walkErr
		}
		if p == path {
			return nil
		}
		itemCount++
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})
	return itemCount, totalSize, err
}

func samePath(a, b string) bool {
	ra := filepath.Clean(a)
	rb := filepath.Clean(b)
	return ra == rb
}
