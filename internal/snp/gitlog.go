package snp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GitData represents collected git log information
type GitData struct {
	LogLines []string
}

func (d GitData) Print() {
	for _, l := range d.LogLines {
		fmt.Println(l)
	}
}

// isGitRepo reports whether a .git directory exists under root
func isGitRepo(root string) bool {
	info, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil && info.IsDir()
}

// collectGitlog retrieves git log output as lines
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
