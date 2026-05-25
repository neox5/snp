package snp

// Entry represents an entry in the snapshot.
// depending on IsDir it represents a directory or a file.
type Entry struct {
	IsDir bool

	// common
	// RelPath  string
	// FullPath string
	Path string
	Size int64 // number of bytes

	// file specific
	IsBinary  bool // decides if it has a line or metadata representation
	Lines     []string
	StartLine int // startline in the snapshot

	// dir specific
	ItemCount int // the number of direct children in the directory
}

func NewDir(path string, size int64, itemCount int) *Entry {
	return &Entry{
		IsDir: true,
		Path:  path,
		// RelPath:   relPath,
		// FullPath:  fullPath,
		Size:      size,
		ItemCount: itemCount,
	}
}

func NewFile(path string, size int64, isBinary bool) *Entry {
	return &Entry{
		IsDir: false,
		Path:  path,
		// RelPath:  relPath,
		// FullPath: fullPath,
		Size:     size,
		IsBinary: isBinary,
	}
}
