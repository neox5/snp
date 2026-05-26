package snapshot

import (
	"fmt"
	"io"
	"sort"
)

// index renders all entries as a sorted single-line-per-entry file index.
type index struct {
	Entries []*Entry
}

func (idx *index) LineCount() int { return len(idx.Entries) }

func (idx *index) WriteTo(w io.Writer) (int, error) {
	type item struct {
		path string
		line string
	}

	items := make([]item, 0, len(idx.Entries))
	for _, e := range idx.Entries {
		var line string
		switch {
		case e.IsDir:
			line = fmt.Sprintf("%s (%d items, %s)", e.RelPath, e.ItemCount, formatSize(e.Size))
		case e.IsBinary:
			line = fmt.Sprintf("%s [%d-%d] (binary, %s)", e.RelPath, e.StartLine, e.EndLine, formatSize(e.Size))
		case e.StartLine == 0 && e.EndLine == 0:
			line = fmt.Sprintf("%s (%d lines, %s)", e.RelPath, len(e.Lines), formatSize(e.Size))
		default:
			line = fmt.Sprintf("%s [%d-%d] (%d lines, %s)", e.RelPath, e.StartLine, e.EndLine, len(e.Lines), formatSize(e.Size))
		}
		items = append(items, item{path: e.RelPath, line: line})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].path < items[j].path
	})

	for _, it := range items {
		if _, err := fmt.Fprintln(w, it.line); err != nil {
			return 0, err
		}
	}
	return 0, nil
}
