package snp

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/matcher"
)

func collect(c *config.Config, root string) ([]*Entry, error) {
	entries := []*Entry{}

	rootDepth := strings.Count(root, "/")

	pathDepth := func(p string) int {
		return strings.Count(p, "/") - rootDepth
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
		relPath = filepath.ToSlash(relPath)

		if !m.ShouldInclude(relPath, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		depth := pathDepth(path)

		if d.IsDir() {
			if c.Depth >= 0 && depth >= c.Depth {
				entries = append(entries, NewDir(path, relPath))
				return filepath.SkipDir
			}
			return nil
		}

		entries = append(entries, NewFile(path, relPath))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}
