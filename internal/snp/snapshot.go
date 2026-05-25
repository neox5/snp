package snp

import (
	"context"
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

func New(c *config.Config) *Snapshot {
	srcDir := filepath.ToSlash(c.SourceDir) // path normalization (unix slashes)
	srcDir = filepath.Clean(srcDir)         // produce the shortest possible path
	root := strings.TrimRight(srcDir, "/")  // remove trailing slash for accurate rootDepth

	return &Snapshot{Config: c, Root: root}
}

func (s *Snapshot) Collect(ctc context.Context) error {
	// ### git
	if isGitRepo(s.Root) {
		data, err := collectGitData(ctc, s.Root)
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
