package snapshot

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/config"
)

func Collect(c config.Config) ([]Entry, error) {
	entries := []Entry{}
	root := strings.Trim(c.SourceDir, "/\\")

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, wErr error) error {
		if wErr != nil {
			if errors.Is(wErr, fs.ErrPermission) {
				return nil
			}
			return wErr
		}

		fmt.Println(path)

		return nil
	})

	return entries, err
}

func pathDepth(path string) int {
	return strings.Count(path, "/") + 1
}
