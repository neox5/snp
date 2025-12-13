// Package snapshot orchestrates directory traversal, ignore rules, optional
// Git log inclusion, and writing the final snapshot file.
package snapshot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/neox5/snap/internal/gitlog"
	"github.com/neox5/snap/internal/ignore"
)

// calculateStartLines computes the StartLine field for each file in the snapshot.
// This must be called after we know:
// - gitLogLines (how many lines the git log section takes)
// - len(files) (how many files we have)
//
// Note: This function calculates line numbers for where files will appear
// in the final output, which includes the Files list section that hasn't
// been written yet when this function is called.
// StartLine points to the first content line (after the "# filename" header).
func calculateStartLines(files []FileInfo, gitLogLines int) {
	currentLine := 1

	// Git log section (if present)
	if gitLogLines > 0 {
		currentLine += gitLogLines
	}

	// Files list section (that will be written after this calculation)
	currentLine++             // "# Files" header
	currentLine += len(files) // one line per file in the list
	currentLine += 3          // blank line + separator + blank line

	// File content sections
	for i := range files {
		currentLine++                    // The "# filename" header line
		files[i].StartLine = currentLine // Points to first content line (after header)
		currentLine += files[i].LineCount
		currentLine += 2 // two blank lines after content
	}
}

// Run performs the snapshot according to the given configuration.
func Run(ctx context.Context, cfg Config) error {
	srcInfo, err := os.Stat(cfg.SourceDir)
	if err != nil {
		return fmt.Errorf("cannot stat directory %q: %w", cfg.SourceDir, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("path %q is not a directory", cfg.SourceDir)
	}

	absSourceDir, err := filepath.Abs(cfg.SourceDir)
	if err != nil {
		return fmt.Errorf("cannot resolve source directory: %w", err)
	}

	absOutput, defaultAbs, err := resolveOutputPath(cfg.OutputPath)
	if err != nil {
		return err
	}

	// Overwrite semantics (skip in dry-run mode):
	// - If output path is the default ./snap.txt in PWD -> always overwrite.
	// - Else, if file exists and output flag was NOT explicitly set -> refuse.
	if !cfg.DryRun && absOutput != defaultAbs && !cfg.OutputExplicit {
		if _, err := os.Stat(absOutput); err == nil {
			return fmt.Errorf("refusing to overwrite existing file %q; use --output to override", absOutput)
		}
	}

	matchers, err := ignore.NewMatchers(absSourceDir, cfg.ExcludePatterns, cfg.IncludePatterns)
	if err != nil {
		return err
	}

	// Collect all file metadata in single walk
	var files []FileInfo
	err = filepath.WalkDir(absSourceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Skip permission errors silently, propagate others
			if errors.Is(walkErr, fs.ErrPermission) {
				return nil
			}
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil
		}
		if samePath(absPath, absOutput) {
			return nil
		}

		relPath, err := filepath.Rel(absSourceDir, path)
		if err != nil {
			return nil
		}
		relUnix := filepath.ToSlash(relPath)

		if !matchers.ShouldInclude(relUnix) {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fileSize := info.Size()

		var isBinary bool
		var contentType string
		var lineCount int

		// Check force overrides first
		overridden := false
		isBinary, contentType, overridden = checkForceOverride(relUnix, cfg)

		if !overridden {
			// Detect binary status, content type, and count lines
			var detectErr error
			isBinary, contentType, lineCount, detectErr = isFileBinary(path, fileSize)
			if detectErr != nil {
				return nil
			}
		} else {
			// If overridden to binary, set LineCount = 1
			if isBinary {
				lineCount = 1
			}
			// If overridden to text, LineCount remains 0 (we don't count lines for forced-text files)
		}

		files = append(files, FileInfo{
			RelPath:     relUnix,
			FullPath:    path,
			Size:        fileSize,
			IsBinary:    isBinary,
			ContentType: contentType,
			LineCount:   lineCount,
			StartLine:   0, // Will be set by calculateStartLines
		})

		return nil
	})
	if err != nil {
		return err
	}

	// Handle dry-run (all info already available)
	if cfg.DryRun {
		for _, file := range files {
			fmt.Println(file.RelPath)
		}
		return nil
	}

	// Normal mode: create output file
	outFile, err := os.Create(absOutput)
	if err != nil {
		return fmt.Errorf("cannot create output file %q: %w", absOutput, err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	// Git log section
	gitLogLines := 0
	if cfg.IncludeGitLog && gitlog.HasRepo(absSourceDir) {
		gitLogLines, err = gitlog.Write(ctx, writer, absSourceDir)
		if err != nil {
			return fmt.Errorf("failed to write git log: %w", err)
		}
	}

	// Calculate StartLine for all files now that we know gitLogLines
	calculateStartLines(files, gitLogLines)

	// File list section with line ranges
	if _, err := fmt.Fprintln(writer, "# Files"); err != nil {
		return err
	}
	for _, file := range files {
		// StartLine now points to first content line directly
		contentStart := file.StartLine
		var contentEnd int
		if file.IsBinary {
			contentEnd = contentStart // Binary files have single content line
		} else {
			contentEnd = contentStart + file.LineCount - 1
		}

		// Format: filename [start-end] (optional metadata)
		if file.IsBinary {
			sizeStr := formatSize(file.Size)
			if _, err := fmt.Fprintf(writer, "%s [%d-%d] (binary, %s)\n",
				file.RelPath, contentStart, contentEnd, sizeStr); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(writer, "%s [%d-%d]\n",
				file.RelPath, contentStart, contentEnd); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprint(writer, "\n# ----------------------------------------\n\n"); err != nil {
		return err
	}

	// Write file contents (metadata already known)
	for _, file := range files {
		if err := appendFileWithMetadata(writer, file); err != nil {
			return err
		}
	}

	fmt.Printf("Files concatenated to %s\n", absOutput)
	return nil
}

func samePath(a, b string) bool {
	ra := filepath.Clean(a)
	rb := filepath.Clean(b)
	return ra == rb
}
