package snp

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Entry represents an entry in the snapshot.
// depending on IsDir it represents a directory or a file.
type Entry struct {
	IsDir bool

	// common
	Path string
	Size int64 // number of bytes

	// file specific
	IsBinary  bool // decides if it has a line or metadata representation
	Lines     []string
	StartLine int // startline in the snapshot

	// dir specific
	ItemCount int // the number of direct children in the directory
}

func NewDir(path string) *Entry {
	return &Entry{
		IsDir: true,
		Path:  path,
	}
}

func NewFile(path string) *Entry {
	return &Entry{
		IsDir: false,
		Path:  path,
	}
}

type DirInfo struct {
	ItemCount int
	TotalSize int64
}

func (e *Entry) Process() error {
	return nil
}

func dirInfo(path string) (DirInfo, error) {
	var info DirInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return info, err
	}
	info.ItemCount = len(entries)

	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}
			info.TotalSize += fileInfo.Size()
		}
		return nil
	})

	return info, err
}
