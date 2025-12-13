// Package gitlog detects Git repositories and writes a formatted Git log
// section into the snapshot output when requested.
package gitlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// HasRepo reports whether a .git directory exists under root.
func HasRepo(root string) bool {
	info, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil && info.IsDir()
}

// Write writes a Git log section to w and returns the number of lines written.
func Write(ctx context.Context, w io.Writer, root string) (int, error) {
	var buf bytes.Buffer

	cmd := exec.CommandContext(ctx, "git", "-C", root, "log", "--all", "--decorate", "--oneline", "--graph")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Count lines in git output
	gitLogLines := bytes.Count(buf.Bytes(), []byte{'\n'})

	// Write header
	if _, err := fmt.Fprintln(w, "# Git Log (git adog)"); err != nil {
		return 0, err
	}
	lineCount := 1

	// Write git log content
	if _, err := w.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	lineCount += gitLogLines

	// Write separator
	if _, err := fmt.Fprint(w, "\n# ----------------------------------------\n\n"); err != nil {
		return 0, err
	}
	lineCount += 3 // blank line + separator + blank line

	return lineCount, nil
}
