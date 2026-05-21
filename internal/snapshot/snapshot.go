package snapshot

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/neox5/snp/internal/file"
	"github.com/neox5/snp/internal/gitlog"
	"github.com/neox5/snp/internal/pick"
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

	var files []*file.File
	var textFiles, binaryFiles int
	var err error

	switch cfg.Mode {
	case ModeTraversal:
		if cfg.IncludeGitLog && gitlog.HasRepo(absSourceDir) {
			gitLogData, err := gitlog.Collect(ctx, absSourceDir)
			if err != nil {
				return nil, fmt.Errorf("failed to collect git log: %w", err)
			}
			snap.GitLogLines = gitLogData.Lines
		}

		files, textFiles, binaryFiles, err = file.Collect(
			absSourceDir,
			absOutput,
			cfg.FilterRules,
			cfg.ForceTextPatterns,
			cfg.ForceBinaryPatterns,
		)
		if err != nil {
			return nil, err
		}

	case ModePick:
		files, textFiles, binaryFiles, err = pick.Collect(
			absSourceDir,
			cfg.PickPaths,
			cfg.ForceTextPatterns,
			cfg.ForceBinaryPatterns,
		)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown mode: %d", cfg.Mode)
	}

	snap.Files = files

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	totalFiles := len(files)
	totalLines := 0

	var layout []Content

	layout = append(layout,
		newSummary(timestamp, totalFiles, textFiles, binaryFiles, &totalLines),
		newEmptyLine(),
	)

	layout = append(layout,
		newHeader("File Index"),
		newIndex(snap.Files),
		newEmptyLine(),
		newSeparator(),
		newEmptyLine(),
	)

	if len(snap.GitLogLines) > 0 {
		layout = append(layout,
			newHeader("Git Log (git adog)"),
			newGitLog(snap.GitLogLines),
			newEmptyLine(),
			newSeparator(),
			newEmptyLine(),
		)
	}

	for i, f := range snap.Files {
		layout = append(layout,
			newHeader(f.RelPath),
			newFileContent(f),
		)
		if i < len(snap.Files)-1 {
			layout = append(layout,
				newEmptyLine(),
				newEmptyLine(),
			)
		}
	}

	currentLine := 1
	for _, content := range layout {
		if fc, ok := content.(fileContent); ok {
			fc.File.StartLine = currentLine
		}
		currentLine += content.LineCount()
	}

	totalLines = currentLine - 1
	snap.Layout = layout

	return snap, nil
}

// WriteTo writes the snapshot to the output
func (s *Snapshot) WriteTo(w io.Writer) (int64, error) {
	if s.Layout == nil {
		return 0, fmt.Errorf("layout not initialized")
	}

	lt := writer.NewLineTracker(w)

	for _, content := range s.Layout {
		if err := content.WriteTo(lt); err != nil {
			return 0, err
		}
	}

	return 0, lt.Flush()
}
