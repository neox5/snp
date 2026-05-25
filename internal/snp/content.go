package snp

import (
	"fmt"
	"sort"

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
	TotalSize   int64
	TotalLines  *int
}

func (s summary) LineCount() int {
	return 4
}

func (s summary) WriteTo(lt *writer.LineTracker) error {
	if err := lt.WriteLine("Generated: " + s.Timestamp); err != nil {
		return err
	}
	if err := lt.WriteLine(fmt.Sprintf("Total files: %d (%d text, %d binary)",
		s.TotalFiles, s.TextFiles, s.BinaryFiles)); err != nil {
		return err
	}
	if err := lt.WriteLine("Total size: " + formatSize(s.TotalSize)); err != nil {
		return err
	}
	return lt.WriteLine(fmt.Sprintf("Total lines: %d", *s.TotalLines))
}

func newSummary(timestamp string, totalFiles, textFiles, binaryFiles int, totalSize int64, totalLines *int) Content {
	return summary{
		Timestamp:   timestamp,
		TotalFiles:  totalFiles,
		TextFiles:   textFiles,
		BinaryFiles: binaryFiles,
		TotalSize:   totalSize,
		TotalLines:  totalLines,
	}
}

// ===== Primitive Content Types =====

// header represents a section header
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

// ===== Content Types =====

// // gitLog represents git log lines
// type gitLog struct {
// 	Lines GitLogLines
// }

// func (g gitLog) LineCount() int {
// 	return len(g.Lines)
// }

// func (g gitLog) WriteTo(lt *writer.LineTracker) error {
// 	for _, line := range g.Lines {
// 		if err := lt.WriteLine(line); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func newGitLog(lines GitLogLines) Content {
// 	return gitLog{Lines: lines}
// }

// index renders all file index entries and collapsed dir entries
type index struct {
	Files         []*file.File
	CollapsedDirs []file.DirEntry
}

func (idx index) LineCount() int {
	return len(idx.Files) + len(idx.CollapsedDirs)
}

func (idx index) WriteTo(lt *writer.LineTracker) error {
	type entry struct {
		path string
		line string
	}

	entries := make([]entry, 0, len(idx.Files)+len(idx.CollapsedDirs))

	for _, f := range idx.Files {
		endLine := f.StartLine + len(f.Lines) - 1
		var line string
		if f.IsBinary {
			line = fmt.Sprintf("%s [%d-%d] (binary, %s)",
				f.RelPath, f.StartLine, endLine, formatSize(f.Size))
		} else {
			line = fmt.Sprintf("%s [%d-%d] (%d lines, %s)",
				f.RelPath, f.StartLine, endLine, len(f.Lines), formatSize(f.Size))
		}
		entries = append(entries, entry{path: f.RelPath, line: line})
	}

	for _, d := range idx.CollapsedDirs {
		line := fmt.Sprintf("%s (%d items, %s)", d.RelPath, d.ItemCount, formatSize(d.Size))
		entries = append(entries, entry{path: d.RelPath, line: line})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].path < entries[j].path
	})

	for _, e := range entries {
		if err := lt.WriteLine(e.line); err != nil {
			return err
		}
	}
	return nil
}

func newIndex(files []*file.File, collapsedDirs []file.DirEntry) Content {
	return index{Files: files, CollapsedDirs: collapsedDirs}
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
