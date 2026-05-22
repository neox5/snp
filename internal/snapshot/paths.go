package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neox5/snp/internal/config"
)

// ValidateAndResolve validates the config and resolves paths.
func ValidateAndResolve(cfg config.FullConfig) (absSourceDir, absOutput string, err error) {
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

	outputPath := cfg.OutputPath
	if outputPath == "" {
		outputPath = config.DefaultOutputPath
	}
	absOutput, err = filepath.Abs(outputPath)
	if err != nil {
		return "", "", fmt.Errorf("cannot resolve output path %q: %w", outputPath, err)
	}

	return absSourceDir, absOutput, nil
}
