package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/matcher"
)

var DefaultExcludePatterns = []string{
	// VCS and dependencies
	".git/",
	"node_modules/",
	".venv/",
	"venv/",
	"__pycache__/",
	".pytest_cache/",
	"dist/",
	"build/",
	"target/",
	"vendor/",

	// Common artifacts
	"*.log",
	"*.tmp",

	// Snapshot files themselves
	"**/*.snp",
	"**/*.snp.txt",
}

// PrintDefaultExcludes prints the default exclude patterns + the note that we also
// add the .gitignore patterns if present.
func PrintDefaultExcludes() {
	fmt.Println("[Default Exclude Patterns]")
	for _, p := range DefaultExcludePatterns {
		fmt.Printf("  %s\n", p)
	}
	fmt.Println()
	fmt.Println("  + .gitignore (if present)")
}

func buildExcludeDefaultRules(srcDir string) matcher.Rules {
	srcDirAbs, err := filepath.Abs(srcDir)
	if err != nil {
		return nil
	}
	path := filepath.Join(srcDirAbs, ".gitignore")
	patterns := append(DefaultExcludePatterns, loadGitignorePatterns(path)...)

	r := matcher.NewRules()
	for _, n := range patterns {
		r = r.AddExclude(n)
	}
	return r
}

// loadGitignorePatterns reads a .gitignore file and returns non-empty, non-comment lines.
func loadGitignorePatterns(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if scanner.Err() != nil {
		return lines
	}
	return lines
}
