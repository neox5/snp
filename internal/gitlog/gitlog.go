// Package gitlog detects Git repositories and writes a formatted Git log
// section into the snapshot output when requested.
package gitlog

import (
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

// Write writes a Git log section to w, similar to the shell script behaviour.
func Write(ctx context.Context, w io.Writer, root string) error {
	if _, err := fmt.Fprintln(w, "# Git Log (git adog)"); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", "-C", root, "log", "--all", "--decorate", "--oneline", "--graph")
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "\n# ----------------------------------------\n\n"); err != nil {
		return err
	}

	return nil
}
