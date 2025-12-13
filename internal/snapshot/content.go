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

// Header creates a new header content item
func Header(text string) Content {
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

// Separator creates a new separator content item
func Separator() Content {
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

// EmptyLine creates a new empty line content item
func EmptyLine() Content {
	return emptyLine{}
}

// ===== Content Types =====

// gitLogContent represents git log lines
type gitLogContent struct {
	Lines GitLogLines
}

func (g gitLogContent) LineCount() int {
	return len(g.Lines)
}

func (g gitLogContent) WriteTo(lt *writer.LineTracker) error {
	for _, line := range g.Lines {
		if err := lt.WriteLine(line); err != nil {
			return err
		}
	}
	return nil
}

// GitLogContent creates a new git log content item
func GitLogContent(lines GitLogLines) Content {
	return gitLogContent{Lines: lines}
}

// fileListContent renders all file list entries
type fileListContent struct {
	Files []*file.File
}

func (fl fileListContent) LineCount() int {
	return len(fl.Files)
}

func (fl fileListContent) WriteTo(lt *writer.LineTracker) error {
	for _, f := range fl.Files {
		var line string
		endLine := f.StartLine + len(f.Lines) - 1

		if f.IsBinary {
			sizeStr := file.FormatSize(f.Size)
			line = fmt.Sprintf("%s [%d-%d] (binary, %s)",
				f.RelPath, f.StartLine, endLine, sizeStr)
		} else {
			line = fmt.Sprintf("%s [%d-%d]",
				f.RelPath, f.StartLine, endLine)
		}

		if err := lt.WriteLine(line); err != nil {
			return err
		}
	}
	return nil
}

// FileListContent creates a new file list content item
func FileListContent(files []*file.File) Content {
	return fileListContent{Files: files}
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

// FileContent creates a new file content item
func FileContent(f *file.File) Content {
	return fileContent{File: f}
}
