package snapshot

import (
	"context"
	"fmt"
	"io"

	"github.com/neox5/snap/internal/file"
	"github.com/neox5/snap/internal/gitlog"
	"github.com/neox5/snap/internal/writer"
)

// Snapshot represents the complete snapshot data
type Snapshot struct {
	GitLogLines GitLogLines
	Files       []*file.File
	Layout      []Content
}

// GitLogLines represents git log output
type GitLogLines []string

// Build creates a complete snapshot
func Build(ctx context.Context, cfg Config, absSourceDir string, absOutput string) (*Snapshot, error) {
	snap := &Snapshot{}

	// Collect git log if enabled
	if cfg.IncludeGitLog && gitlog.HasRepo(absSourceDir) {
		gitLogData, err := gitlog.Collect(ctx, absSourceDir)
		if err != nil {
			return nil, fmt.Errorf("failed to collect git log: %w", err)
		}
		snap.GitLogLines = gitLogData.Lines
	}

	// Collect files
	discovered, err := file.Collect(
		absSourceDir,
		absOutput,
		cfg.ExcludePatterns,
		cfg.IncludePatterns,
		cfg.ForceTextPatterns,
		cfg.ForceBinaryPatterns,
	)
	if err != nil {
		return nil, err
	}

	// Load file content
	for _, df := range discovered {
		f, err := file.New(
			df.RelPath,
			df.FullPath,
			df.Size,
			df.IsBinary,
		)
		if err != nil {
			return nil, err
		}
		snap.Files = append(snap.Files, f)
	}

	// Build layout
	layout := []Content{}

	// Git log section (if present)
	if len(snap.GitLogLines) > 0 {
		layout = append(layout,
			Header("Git Log (git adog)"),
			GitLogContent(snap.GitLogLines),
			EmptyLine(),
			Separator(),
			EmptyLine(),
		)
	}

	// File list section
	layout = append(layout,
		Header("Files"),
		FileListContent(snap.Files),
		EmptyLine(),
		Separator(),
		EmptyLine(),
	)

	// File contents sections
	for i, f := range snap.Files {
		layout = append(layout,
			Header(f.RelPath),
			FileContent(f),
		)

		// Add spacing only if not the last file
		if i < len(snap.Files)-1 {
			layout = append(layout,
				EmptyLine(),
				EmptyLine(),
			)
		}
	}

	snap.Layout = layout

	// Calculate line ranges
	snap.calculateLineRanges()

	return snap, nil
}

// calculateLineRanges calculates and sets StartLine for all files
func (s *Snapshot) calculateLineRanges() {
	currentLine := 1

	// Iterate through layout and track line positions
	for _, content := range s.Layout {
		// Check if this is a fileContent to record its position
		if fc, ok := content.(fileContent); ok {
			// The current line is where this file's content starts
			fc.File.StartLine = currentLine
		}

		// Advance current line by this content's line count
		currentLine += content.LineCount()
	}
}

// WriteTo writes the snapshot to the output
func (s *Snapshot) WriteTo(w io.Writer) error {
	if s.Layout == nil {
		return fmt.Errorf("layout not initialized")
	}

	lt := writer.NewLineTracker(w)

	for _, content := range s.Layout {
		if err := content.WriteTo(lt); err != nil {
			return err
		}
	}

	return lt.Flush()
}
