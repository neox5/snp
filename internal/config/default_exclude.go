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
	// VCS, dependencies and caches, at any depth.
	// No trailing slash: a pattern with a trailing slash matches at the root only.
	".git",
	"node_modules",
	".venv",
	"venv",
	"__pycache__",
	".pytest_cache",

	// Build output, root only: these names are also used by source packages
	// (for example internal/build), so they must not match at depth.
	"dist/",
	"build/",
	"target/",
	"vendor/",

	// Common artifacts
	"*.log",
	"*.tmp",

	// snp config
	".snpconfig.json",

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

	r := matcher.NewRules()
	for _, p := range DefaultExcludePatterns {
		r = r.AddExclude(p)
	}
	r = r.WithSource(SourceDefault)

	// .gitignore lines use ScopeEntry (git semantics); "!pattern" re-includes.
	ignored := matcher.NewRules()
	for _, line := range loadGitignorePatterns(path) {
		if p, negate := strings.CutPrefix(line, "!"); negate {
			if p != "" {
				ignored = ignored.AddIncludeEntry(p)
			}
			continue
		}
		ignored = ignored.AddExcludeEntry(line)
	}
	return r.AddRules(ignored.WithSource(SourceGitignore))
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
