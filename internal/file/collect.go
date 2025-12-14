package file

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/neox5/snap/internal/ignore"
)

// Collect discovers, analyzes, and loads files to include in the snapshot
// Returns: files, textCount, binaryCount, error
func Collect(sourceDir, outputPath string, excludePatterns, includePatterns, forceTextPatterns, forceBinaryPatterns []string) ([]*File, int, int, error) {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("cannot resolve source directory: %w", err)
	}

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("cannot resolve output path: %w", err)
	}

	matchers, err := ignore.NewMatchers(absSourceDir, excludePatterns, includePatterns)
	if err != nil {
		return nil, 0, 0, err
	}

	var files []*File
	var textCount, binaryCount int

	err = filepath.WalkDir(absSourceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrPermission) {
				return nil
			}
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil
		}
		if samePath(absPath, absOutput) {
			return nil
		}

		relPath, err := filepath.Rel(absSourceDir, path)
		if err != nil {
			return nil
		}
		relUnix := filepath.ToSlash(relPath)

		if !matchers.ShouldInclude(relUnix) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		fileSize := info.Size()

		var isBinary bool

		// Check force overrides
		isBinaryOverride, overridden := CheckForceOverride(relUnix, forceTextPatterns, forceBinaryPatterns)
		if overridden {
			isBinary = isBinaryOverride
		} else {
			// Detect binary status
			isBinary, err = DetectBinary(path, fileSize)
			if err != nil {
				return nil
			}
		}

		// Create and load file immediately
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
		return nil, 0, 0, err
	}

	return files, textCount, binaryCount, nil
}

func samePath(a, b string) bool {
	ra := filepath.Clean(a)
	rb := filepath.Clean(b)
	return ra == rb
}
