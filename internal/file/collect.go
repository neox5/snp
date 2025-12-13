package file

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/neox5/snap/internal/ignore"
)

// DiscoveredFile holds file metadata before content is loaded
type DiscoveredFile struct {
	RelPath  string
	FullPath string
	Size     int64
	IsBinary bool
}

// Collect discovers and analyzes files to include in the snapshot
func Collect(sourceDir, outputPath string, excludePatterns, includePatterns, forceTextPatterns, forceBinaryPatterns []string) ([]DiscoveredFile, error) {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve source directory: %w", err)
	}

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve output path: %w", err)
	}

	matchers, err := ignore.NewMatchers(absSourceDir, excludePatterns, includePatterns)
	if err != nil {
		return nil, err
	}

	var discovered []DiscoveredFile

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

		discovered = append(discovered, DiscoveredFile{
			RelPath:  relUnix,
			FullPath: path,
			Size:     fileSize,
			IsBinary: isBinary,
		})

		return nil
	})

	return discovered, err
}

// CountLinesInFile counts lines in a file by path
func CountLinesInFile(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

func samePath(a, b string) bool {
	ra := filepath.Clean(a)
	rb := filepath.Clean(b)
	return ra == rb
}
