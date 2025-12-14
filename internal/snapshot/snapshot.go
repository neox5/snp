package snapshot

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/neox5/snp/internal/file"
	"github.com/neox5/snp/internal/gitlog"
	"github.com/neox5/snp/internal/writer"
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

	// Collect and load files
	files, textFiles, binaryFiles, err := file.Collect(
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
	snap.Files = files

	// Prepare summary metadata
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	totalFiles := len(files)
	totalLines := 0 // Will be set after layout construction

	// Build layout (single pass)
	var layout []Content

	// Summary section (with mutable totalLines pointer)
	layout = append(layout,
		newSummary(timestamp, totalFiles, textFiles, binaryFiles, &totalLines),
		newEmptyLine(),
	)

	// Index section
	layout = append(layout,
		newHeader("File Index"),
		newIndex(snap.Files),
		newEmptyLine(),
		newSeparator(),
		newEmptyLine(),
	)

	// Git log section (if present)
	if len(snap.GitLogLines) > 0 {
		layout = append(layout,
			newHeader("Git Log (git adog)"),
			newGitLog(snap.GitLogLines),
			newEmptyLine(),
			newSeparator(),
			newEmptyLine(),
		)
	}

	// File contents sections
	for i, f := range snap.Files {
		layout = append(layout,
			newHeader(f.RelPath),
			newFileContent(f),
		)

		// Add spacing only if not the last file
		if i < len(snap.Files)-1 {
			layout = append(layout,
				newEmptyLine(),
				newEmptyLine(),
			)
		}
	}

	// Assign file StartLine and calculate totalLines
	currentLine := 1
	for _, content := range layout {
		if fc, ok := content.(fileContent); ok {
			fc.File.StartLine = currentLine
		}
		currentLine += content.LineCount()
	}

	totalLines = currentLine - 1 // -1 because we started at 1

	snap.Layout = layout

	return snap, nil
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
