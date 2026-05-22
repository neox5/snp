package snapshot

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/file"
	"github.com/neox5/snp/internal/gitlog"
	"github.com/neox5/snp/internal/pick"
	"github.com/neox5/snp/internal/writer"
)

// Snapshot represents the complete snapshot data.
type Snapshot struct {
	GitLogLines   GitLogLines
	Files         []*file.File
	CollapsedDirs []file.DirEntry
	Layout        []Content
}

// GitLogLines represents git log output.
type GitLogLines []string

// Build creates a complete snapshot.
func Build(ctx context.Context, cfg config.Config, absSourceDir string, absOutput string) (*Snapshot, error) {
	snap := &Snapshot{}

	var files []*file.File
	var collapsedDirs []file.DirEntry
	var textFiles, binaryFiles int
	var err error

	if cfg.IsPick() {
		files, textFiles, binaryFiles, err = pick.Collect(
			absSourceDir,
			cfg.PickPaths,
			cfg.ForceTextPatterns,
			cfg.ForceBinaryPatterns,
		)
		if err != nil {
			return nil, err
		}
	} else {
		if !cfg.NoGitLog && gitlog.HasRepo(absSourceDir) {
			gitLogData, err := gitlog.Collect(ctx, absSourceDir)
			if err != nil {
				return nil, fmt.Errorf("failed to collect git log: %w", err)
			}
			snap.GitLogLines = gitLogData.Lines
		}

		files, collapsedDirs, textFiles, binaryFiles, err = file.Collect(
			absSourceDir,
			absOutput,
			cfg.FilterRules,
			cfg.ForceTextPatterns,
			cfg.ForceBinaryPatterns,
			cfg.Depth,
		)
		if err != nil {
			return nil, err
		}
	}

	snap.Files = files
	snap.CollapsedDirs = collapsedDirs

	// calculate total size
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	for _, d := range collapsedDirs {
		totalSize += d.Size
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	totalFiles := len(files)
	totalLines := 0

	var layout []Content
	needsDivider := false

	section := func(contents ...Content) {
		if needsDivider {
			layout = append(layout, newDivider())
		}
		layout = append(layout, contents...)
		needsDivider = true
	}

	if !cfg.NoSummary {
		section(
			newSummary(timestamp, totalFiles, textFiles, binaryFiles, totalSize, &totalLines),
		)
	}

	if !cfg.NoIndex {
		section(
			newHeader("File Index"),
			newIndex(snap.Files, snap.CollapsedDirs),
		)
	}

	if !cfg.NoGitLog && len(snap.GitLogLines) > 0 {
		section(
			newHeader("Git Log (git adog)"),
			newGitLog(snap.GitLogLines),
		)
	}

	if !cfg.NoContent {
		for _, f := range snap.Files {
			section(
				newHeader(f.RelPath),
				newFileContent(f),
			)
		}
	}

	// assign start lines
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

// WriteTo writes the snapshot to the writer.
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
