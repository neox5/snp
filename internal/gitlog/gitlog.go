package gitlog

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// HasRepo reports whether a .git directory exists under root
func HasRepo(root string) bool {
	info, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil && info.IsDir()
}

// GitLogData represents collected git log information
type GitLogData struct {
	Lines []string
}

// Collect retrieves git log output as lines
func Collect(ctx context.Context, root string) (*GitLogData, error) {
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

	return &GitLogData{Lines: lines}, nil
}
