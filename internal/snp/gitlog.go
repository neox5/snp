package snp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// GitData holds git log output and implements Content.
type GitData struct {
	LogLines []string
}

// LineCount implements Content.
func (d *GitData) LineCount() int { return len(d.LogLines) }

// WriteTo implements Content.
func (d *GitData) WriteTo(w io.Writer) (int, error) {
	for _, line := range d.LogLines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func (d *GitData) Print() {
	for _, l := range d.LogLines {
		fmt.Println(l)
	}
}

// isGitRepo reports whether a .git directory exists under root.
func isGitRepo(root string) bool {
	info, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil && info.IsDir()
}

// collectGitData retrieves git log output as lines.
func collectGitData(ctx context.Context, root string) (*GitData, error) {
	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "-C", root, "log", "--all", "--decorate", "--oneline", "--graph")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &GitData{LogLines: lines}, nil
}
