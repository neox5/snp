// Package snapshot implements the core snapshot pipeline: collection, processing, layout, and output.
package snapshot

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neox5/snp/internal/config"
)

// Snapshot represents the complete snapshot data.
type Snapshot struct {
	Config  *config.Config
	Root    string
	Summary *Summary
	GitData *GitData
	Entries []*Entry
	Layout  []Content
}

// New creates a snapshot instance and sets config and root directory.
func New(c *config.Config) *Snapshot {
	srcDir := filepath.ToSlash(c.SourceDir)
	srcDir = filepath.Clean(srcDir)
	root := strings.TrimRight(srcDir, "/")

	return &Snapshot{
		Config:  c,
		Root:    root,
		Summary: &Summary{},
	}
}

// Collect collects git data (if present) and all snapshot entries (files + dirs).
func (s *Snapshot) Collect(ctx context.Context) error {
	// ### git
	if !s.Config.NoGitLog && isGitRepo(s.Root) {
		data, err := collectGitData(ctx, s.Root)
		if err != nil {
			return err
		}
		s.GitData = data
	}

	// ### entries
	if len(s.Config.PickPaths) > 0 {
		entries, err := collectPick(s.Config, s.Root)
		if err != nil {
			return err
		}
		s.Entries = entries
		return nil
	}

	entries, err := collect(s.Config, s.Root)
	if err != nil {
		return err
	}
	s.Entries = entries

	return nil
}

// PrintRawEntries prints only entry paths (dry-run) to w.
func (s Snapshot) PrintRawEntries(w io.Writer) {
	for _, e := range s.Entries {
		p := e.Path
		if e.IsDir {
			p = p + "/"
		}
		fmt.Fprintf(w, "%s\n", p)
	}
}

// Build processes all entry contents, calculates stats, and builds the layout.
func (s *Snapshot) Build() error {
	// ### process entry contents
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

	// ### calculate stats
	var totalSize int64
	var textFiles, binaryFiles int
	for _, e := range s.Entries {
		if e.IsDir {
			continue
		}
		totalSize += e.Size
		if e.IsBinary {
			binaryFiles++
		} else {
			textFiles++
		}
	}

	s.Summary.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	s.Summary.TotalFiles = textFiles + binaryFiles
	s.Summary.TextFiles = textFiles
	s.Summary.BinaryFiles = binaryFiles
	s.Summary.TotalSize = totalSize

	s.buildLayout()

	return nil
}

func (s *Snapshot) Write() error {
	f, err := os.Create(s.Config.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriter(f)
	if _, err = s.WriteTo(bw); err != nil {
		return err
	}
	return bw.Flush()
}

func (s *Snapshot) WriteTo(w io.Writer) (int64, error) {
	for _, c := range s.Layout {
		if _, err := c.WriteTo(w); err != nil {
			return 0, err
		}
	}
	return 0, nil
}
