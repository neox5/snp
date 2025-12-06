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

	// Overwrite semantics:
	// - If output path is the default ./snap.txt in PWD -> always overwrite.
	// - Else, if file exists and output flag was NOT explicitly set -> refuse.
	if absOutput != defaultAbs && !cfg.OutputExplicit {
		if _, err := os.Stat(absOutput); err == nil {
			return fmt.Errorf("refusing to overwrite existing file %q; use --output to override", absOutput)
		}
	}

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

	matchers, err := ignore.NewMatchers(absSourceDir, cfg.ExcludePatterns, cfg.IncludePatterns)
	if err != nil {
		return err
	}

	// Walk directory tree and concatenate files
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

		return appendFile(writer, path, relUnix)
	})
	if err != nil {
		return err
	}

	fmt.Printf("Files concatenated to %s\n", absOutput)
	return nil
}

func samePath(a, b string) bool {
	ra := filepath.Clean(a)
	rb := filepath.Clean(b)
	return ra == rb
}
