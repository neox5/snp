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

		// Check force overrides first
		isBinary, contentType, overridden := checkForceOverride(relUnix, cfg)
		if !overridden {
			// Detect binary status and content type
			var detectErr error
			isBinary, contentType, detectErr = isFileBinary(path, fileSize)
			if detectErr != nil {
				return nil
			}
		}

		files = append(files, FileInfo{
			RelPath:     relUnix,
			FullPath:    path,
			Size:        fileSize,
			IsBinary:    isBinary,
			ContentType: contentType,
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
	if cfg.IncludeGitLog && gitlog.HasRepo(absSourceDir) {
		if err := gitlog.Write(ctx, writer, absSourceDir); err != nil {
			return fmt.Errorf("failed to write git log: %w", err)
		}
	}

	// File list section
	if _, err := fmt.Fprintln(writer, "# Files"); err != nil {
		return err
	}
	for _, file := range files {
		if file.IsBinary {
			sizeStr := formatSize(file.Size)
			if _, err := fmt.Fprintf(writer, "%s [binary, %s]\n", file.RelPath, sizeStr); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintln(writer, file.RelPath); err != nil {
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
