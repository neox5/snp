package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultOutputName is the default snapshot output filename
const DefaultOutputName = "snapshot.snp"

func resolveOutputPath(outputPath string) (absOutput string, err error) {
	if outputPath == "" {
		outputPath = DefaultOutputName
	}

	absOutput, err = filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve output path %q: %w", outputPath, err)
	}

	return absOutput, nil
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
	absOutput, err = resolveOutputPath(cfg.OutputPath)
	if err != nil {
		return "", "", err
	}

	return absSourceDir, absOutput, nil
}
