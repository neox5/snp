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

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			if c.Depth >= 0 && depth >= c.Depth {
				e := NewDir(path, relPath)
				e.Size = info.Size()
				entries = append(entries, e)
				return filepath.SkipDir
			}
			return nil
		}

		e := NewFile(path, relPath)
		e.Size = info.Size()
		entries = append(entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func collectPick(c *config.Config, root string) ([]*Entry, error) {
	seen := make(map[string]bool)
	var entries []*Entry

	for _, pattern := range c.PickPaths {
		var globPattern string
		if filepath.IsAbs(pattern) {
			globPattern = pattern
		} else {
			globPattern = filepath.Join(root, pattern)
		}

		matches, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, fmt.Errorf("--pick %q: invalid pattern: %w", pattern, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("--pick %q: no files found", pattern)
		}

		for _, match := range matches {
			absMatch, err := filepath.Abs(match)
			if err != nil {
				return nil, err
			}

			if seen[absMatch] {
				continue
			}
			seen[absMatch] = true

			info, err := os.Stat(absMatch)
			if err != nil {
				return nil, fmt.Errorf("--pick %q: %w", match, err)
			}
			if info.IsDir() {
				return nil, fmt.Errorf("--pick %q: is a directory, not a file", match)
			}

			relPath, err := filepath.Rel(root, absMatch)
			if err != nil {
				relPath = absMatch
			}
			relPath = filepath.ToSlash(relPath)

			e := NewFile(absMatch, relPath)
			e.Size = info.Size()
			entries = append(entries, e)
		}
	}

	return entries, nil
}
