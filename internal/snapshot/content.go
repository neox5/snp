package snapshot

import (
	"fmt"

	"github.com/neox5/snp/internal/file"
	"github.com/neox5/snp/internal/writer"
)

// Content represents anything that can be written to the output
type Content interface {
	LineCount() int
	WriteTo(lt *writer.LineTracker) error
}

// ===== Summary Content =====

// summary represents the metadata header
type summary struct {
	Timestamp   string
	TotalFiles  int
	TextFiles   int
	BinaryFiles int
	TotalLines  *int
}

func (s summary) LineCount() int {
	return 3
}

func (s summary) WriteTo(lt *writer.LineTracker) error {
	if err := lt.WriteLine("Generated: " + s.Timestamp); err != nil {
		return err
	}

	summary := fmt.Sprintf("Total files: %d (%d text, %d binary)",
		s.TotalFiles, s.TextFiles, s.BinaryFiles)
	if err := lt.WriteLine(summary); err != nil {
		return err
	}

	totalLinesStr := fmt.Sprintf("Total lines: %d", *s.TotalLines)
	return lt.WriteLine(totalLinesStr)
}

func newSummary(timestamp string, totalFiles, textFiles, binaryFiles int, totalLines *int) Content {
	return summary{
		Timestamp:   timestamp,
		TotalFiles:  totalFiles,
		TextFiles:   textFiles,
		BinaryFiles: binaryFiles,
		TotalLines:  totalLines,
	}
}

// ===== Primitive Content Types =====

// header represents a section header like "# File Index"
type header struct {
	Text string
}

func (h header) LineCount() int {
	return 1
}

func (h header) WriteTo(lt *writer.LineTracker) error {
	return lt.WriteLine("# " + h.Text)
}

func newHeader(text string) Content {
	return header{Text: text}
}

// divider represents the blank + separator + blank lines between sections
// emits:
//   (empty line)
//   # ----------------------------------------
//   (empty line)
type divider struct{}

func (d divider) LineCount() int {
	return 3
}

func (d divider) WriteTo(lt *writer.LineTracker) error {
	if err := lt.WriteLine(""); err != nil {
		return err
	}
	if err := lt.WriteLine("# ----------------------------------------"); err != nil {
		return err
	}
	return lt.WriteLine("")
}

func newDivider() Content {
	return divider{}
}

// emptyLine represents a blank line (used within sections, e.g. after summary)
type emptyLine struct{}

func (e emptyLine) LineCount() int {
	return 1
}

func (e emptyLine) WriteTo(lt *writer.LineTracker) error {
	return lt.WriteLine("")
}

func newEmptyLine() Content {
	return emptyLine{}
}

// ===== Content Types =====

// gitLog represents git log lines
type gitLog struct {
	Lines GitLogLines
}

func (g gitLog) LineCount() int {
	return len(g.Lines)
}

func (g gitLog) WriteTo(lt *writer.LineTracker) error {
	for _, line := range g.Lines {
		if err := lt.WriteLine(line); err != nil {
			return err
		}
	}
	return nil
}

func newGitLog(lines GitLogLines) Content {
	return gitLog{Lines: lines}
}

// index renders all file index entries
type index struct {
	Files []*file.File
}

func (idx index) LineCount() int {
	return len(idx.Files)
}

func (idx index) WriteTo(lt *writer.LineTracker) error {
	for _, f := range idx.Files {
		var line string
		endLine := f.StartLine + len(f.Lines) - 1
		lineCount := len(f.Lines)

		if f.IsBinary {
			sizeStr := formatSize(f.Size)
			line = fmt.Sprintf("%s [%d-%d] (binary, %s)",
				f.RelPath, f.StartLine, endLine, sizeStr)
		} else {
			sizeStr := formatSize(f.Size)
			line = fmt.Sprintf("%s [%d-%d] (%d lines, %s)",
				f.RelPath, f.StartLine, endLine, lineCount, sizeStr)
		}

		if err := lt.WriteLine(line); err != nil {
			return err
		}
	}
	return nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes == 0:
		return "0 bytes"
	case bytes == 1:
		return "1 byte"
	case bytes < KB:
		return fmt.Sprintf("%d bytes", bytes)
	case bytes < MB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	case bytes < GB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	}
}

func newIndex(files []*file.File) Content {
	return index{Files: files}
}

// fileContent renders a single file's content
type fileContent struct {
	File *file.File
}

func (f fileContent) LineCount() int {
	return len(f.File.Lines)
}

func (f fileContent) WriteTo(lt *writer.LineTracker) error {
	for _, line := range f.File.Lines {
		if err := lt.WriteLine(line); err != nil {
			return err
		}
	}
	return nil
}

func newFileContent(f *file.File) Content {
	return fileContent{File: f}
}
