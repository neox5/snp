package snapshot

import (
	"fmt"

	"github.com/neox5/snap/internal/file"
	"github.com/neox5/snap/internal/writer"
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
	TotalLines  *int // Pointer to allow updating after layout construction
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

// newSummary creates a new summary content item with mutable totalLines
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

// header represents a section header like "# Git Log (git adog)"
type header struct {
	Text string
}

func (h header) LineCount() int {
	return 1
}

func (h header) WriteTo(lt *writer.LineTracker) error {
	return lt.WriteLine("# " + h.Text)
}

// newHeader creates a new header content item
func newHeader(text string) Content {
	return header{Text: text}
}

// separator represents the "# ----------------------------------------" line
type separator struct{}

func (s separator) LineCount() int {
	return 1
}

func (s separator) WriteTo(lt *writer.LineTracker) error {
	return lt.WriteLine("# ----------------------------------------")
}

// newSeparator creates a new separator content item
func newSeparator() Content {
	return separator{}
}

// emptyLine represents a blank line
type emptyLine struct{}

func (e emptyLine) LineCount() int {
	return 1
}

func (e emptyLine) WriteTo(lt *writer.LineTracker) error {
	return lt.WriteLine("")
}

// newEmptyLine creates a new empty line content item
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

// newGitLog creates a new git log content item
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

// formatSize formats byte size in human-readable format
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

// newIndex creates a new file index content item
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

// newFileContent creates a new file content item
func newFileContent(f *file.File) Content {
	return fileContent{File: f}
}
