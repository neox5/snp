package snapshot

import (
	"fmt"
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
