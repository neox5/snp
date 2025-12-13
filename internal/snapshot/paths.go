package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
)

func resolveOutputPath(outputPath string) (absOutput string, defaultAbs string, err error) {
	const defaultName = "snap.txt"

	defaultAbs, err = filepath.Abs(defaultName)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve default output path: %w", err)
	}

	if outputPath == "" {
		outputPath = defaultName
	}

	absOutput, err = filepath.Abs(outputPath)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve output path %q: %w", outputPath, err)
	}

	return absOutput, defaultAbs, nil
}

// ValidateAndResolve validates the config and resolves paths
func ValidateAndResolve(cfg Config) (absSourceDir, absOutput string, err error) {
	// Validate source directory
	srcInfo, err := os.Stat(cfg.SourceDir)
	if err != nil {
		return "", "", fmt.Errorf("cannot stat directory %q: %w", cfg.SourceDir, err)
	}
	if !srcInfo.IsDir() {
		return "", "", fmt.Errorf("path %q is not a directory", cfg.SourceDir)
	}

	absSourceDir, err = filepath.Abs(cfg.SourceDir)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve source directory: %w", err)
	}

	// Resolve output path
	absOutput, defaultAbs, err := resolveOutputPath(cfg.OutputPath)
	if err != nil {
		return "", "", err
	}

	// Check overwrite semantics (skip in dry-run mode)
	if !cfg.DryRun && absOutput != defaultAbs && !cfg.OutputExplicit {
		if _, err := os.Stat(absOutput); err == nil {
			return "", "", fmt.Errorf("refusing to overwrite existing file %q; use --output to override", absOutput)
		}
	}

	return absSourceDir, absOutput, nil
}
