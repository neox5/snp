package snp

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func (s *Snapshot) processDir(e *Entry) error {
	// ### itemCount processing
	items, err := os.ReadDir(e.Path)
	if err != nil {
		return err
	}
	e.ItemCount = len(items)

	// ### size calculation - travers complete folder and sum up all file sizes
	var total int64
	err = filepath.WalkDir(e.Path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			total += info.Size()
		}
		return nil
	})
	e.Size = total

	// ### content processing
	e.Lines = []string{fmt.Sprintf("[%d items, %s]", e.ItemCount, formatSize(e.Size))}

	return nil
}

func (s *Snapshot) processFile(e *Entry) error {
	f, err := os.Open(e.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// ### isBinary processing
	var isBinary bool
	isBinaryOverride, overridden := checkForceOverride(e.Path, s.Config.ForceBinaryPatterns, s.Config.ForceTextPatterns)
	if overridden {
		isBinary = isBinaryOverride
	} else {
		isBinary, err = detectBinary(f)
		if err != nil {
			return err
		}
	}

	// ### content processing
	if isBinary {
		e.Lines = []string{fmt.Sprintf("[Binary file - %s - content omitted]", formatSize(e.Size))}
		return nil
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	e.Lines = lines

	return nil
}
