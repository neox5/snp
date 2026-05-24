package snapshot

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/matcher"
)

func Collect(c config.Config) ([]Entry, error) {
	entries := []Entry{}

	srcDir := filepath.ToSlash(c.SourceDir) // path normalization (unix slashes)
	srcDir = filepath.Clean(srcDir)         // produce the shortest possible path
	root := strings.TrimRight(srcDir, "/")  // remove trailing slash for accurate rootDepth
	rootDepth := strings.Count(root, "/")   // count the leftover slashes to determine depth

	// pathDepth for calculating the current depth of a path
	pathDepth := func(p string) int {
		return strings.Count(p, "/") - rootDepth
	}

	m, err := matcher.New(c.BuildMatcherRules())
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, wErr error) error {
		if wErr != nil {
			if errors.Is(wErr, fs.ErrPermission) { // permission errors will be ingnored
				return nil
			}
			return wErr
		}

		path = filepath.ToSlash(path) // path normalization (unix slashes)
		depth := pathDepth(path)

		if !m.ShouldInclude(path) {
			return nil
		}

		// ### directories
		if d.IsDir() {
			if path == root { // skip root folder
				return nil
			}

			if depth >= c.Depth {
				info, _ := d.Info()
				fmt.Printf("%s %d %v %v %v\n", info.Name(), info.Size(), info.Mode(), info.ModTime(), info.IsDir())

				return filepath.SkipDir
			}
			return nil
		}

		// ### files

		return nil
	})

	return entries, err
}

func pathDepth(path string) int {
	return strings.Count(path, "/") + 1
}
