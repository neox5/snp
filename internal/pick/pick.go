// Package pick implements Mode 2 file selection: direct file addressing
// without traversal. Supports exact paths and glob patterns.
package pick

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neox5/snp/internal/file"
)

// Collect resolves pick paths/globs against root and loads matching files.
// Relative paths and globs are resolved against root.
// Absolute paths are resolved against the filesystem root.
// Returns: files, textCount, binaryCount, error
func Collect(root string, pickPaths []string, forceTextPatterns, forceBinaryPatterns []string) ([]*file.File, int, int, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("cannot resolve root: %w", err)
	}

	seen := make(map[string]bool)
	var files []*file.File
	var textCount, binaryCount int

	for _, pattern := range pickPaths {
		// Resolve pattern base
		var globPattern string
		if filepath.IsAbs(pattern) {
			globPattern = pattern
		} else {
			globPattern = filepath.Join(absRoot, pattern)
		}

		matches, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}

		if len(matches) == 0 {
			return nil, 0, 0, fmt.Errorf("--pick %q: no files found", pattern)
		}

		for _, match := range matches {
			absMatch, err := filepath.Abs(match)
			if err != nil {
				return nil, 0, 0, err
			}

			// Skip duplicates
			if seen[absMatch] {
				continue
			}
			seen[absMatch] = true

			info, err := os.Stat(absMatch)
			if err != nil {
				return nil, 0, 0, fmt.Errorf("--pick %q: %w", match, err)
			}
			if info.IsDir() {
				return nil, 0, 0, fmt.Errorf("--pick %q: is a directory, not a file", match)
			}

			fileSize := info.Size()

			// Determine display path
			relPath, err := filepath.Rel(absRoot, absMatch)
			if err != nil || isOutsideRoot(relPath) {
				// Absolute path outside root — show as-is
				relPath = absMatch
			}
			relPath = filepath.ToSlash(relPath)

			// Binary detection
			var isBinary bool
			isBinaryOverride, overridden := file.CheckForceOverride(relPath, forceTextPatterns, forceBinaryPatterns)
			if overridden {
				isBinary = isBinaryOverride
			} else {
				isBinary, err = file.DetectBinary(absMatch, fileSize)
				if err != nil {
					return nil, 0, 0, err
				}
			}

			f, err := file.New(relPath, absMatch, fileSize, isBinary)
			if err != nil {
				return nil, 0, 0, err
			}

			files = append(files, f)

			if isBinary {
				binaryCount++
			} else {
				textCount++
			}
		}
	}

	return files, textCount, binaryCount, nil
}

// isOutsideRoot returns true if a relative path escapes the root (starts with ..)
func isOutsideRoot(relPath string) bool {
	return len(relPath) >= 2 && relPath[:2] == ".."
}
