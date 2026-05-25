package snp

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/config"
)

// Snapshot represents the complete snapshot data.
type Snapshot struct {
	Config  *config.Config
	Root    string
	GitData *GitData
	Entries []*Entry
	Layout  []Content
}

// New creates a snapshot instance and sets config and root directory.
func New(c *config.Config) *Snapshot {
	srcDir := filepath.ToSlash(c.SourceDir) // path normalization (unix slashes)
	srcDir = filepath.Clean(srcDir)         // produce the shortest possible path
	root := strings.TrimRight(srcDir, "/")  // remove trailing slash for accurate rootDepth

	return &Snapshot{Config: c, Root: root}
}

// Collect collects git data (if present) and all snapshot entries (files + dirs).
func (s *Snapshot) Collect(ctx context.Context) error {
	// ### git
	if isGitRepo(s.Root) {
		data, err := collectGitData(ctx, s.Root)
		if err != nil {
			return err
		}
		s.GitData = data
	}

	// ### entries
	entries, err := collect(s.Config, s.Root)
	if err != nil {
		return err
	}
	s.Entries = entries

	return nil
}

// Process process all entry contents
func (s *Snapshot) Process() error {
	for _, e := range s.Entries {
		if e.IsDir {
			if err := s.processDir(e); err != nil {
				return err
			}
			continue
		}

		if err := s.processFile(e); err != nil {
			return err
		}
	}

	return nil
}

func (s *Snapshot) Write() error {
	f, err := os.Create(s.Config.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}
