// Package ignore aggregates default, .gitignore, and CLI patterns into
// matchers that decide whether a given path should be included in a snapshot.
package ignore

import (
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// DefaultPatterns mirrors the shell script's default excludes.
var DefaultPatterns = []string{
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
	"**/snap.txt",
	"**/*.snap.txt",
}

// Matchers holds compiled ignore and include patterns.
type Matchers struct {
	ignore      *gitignore.GitIgnore
	include     *gitignore.GitIgnore
	hasIncludes bool
}

// NewMatchers builds ignore/include matchers from defaults, .gitignore,
// CLI excludes, and CLI includes.
func NewMatchers(sourceDir string, excludePatterns, includePatterns []string) (*Matchers, error) {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, err
	}

	var ignoreLines []string
	ignoreLines = append(ignoreLines, DefaultPatterns...)

	// CLI excludes
	ignoreLines = append(ignoreLines, excludePatterns...)

	// .gitignore at sourceDir (if present)
	gitignorePath := filepath.Join(absSourceDir, ".gitignore")
	if b, err := os.ReadFile(gitignorePath); err == nil {
		lines := strings.Split(string(b), "\n")
		ignoreLines = append(ignoreLines, lines...)
	}

	var ignoreMatcher *gitignore.GitIgnore
	if len(ignoreLines) > 0 {
		ignoreMatcher = gitignore.CompileIgnoreLines(ignoreLines...)
	}

	var includeMatcher *gitignore.GitIgnore
	if len(includePatterns) > 0 {
		includeMatcher = gitignore.CompileIgnoreLines(includePatterns...)
	}

	return &Matchers{
		ignore:      ignoreMatcher,
		include:     includeMatcher,
		hasIncludes: len(includePatterns) > 0,
	}, nil
}

// ShouldInclude decides if a relative path should be included in the snapshot.
//
// relPath must be a path relative to sourceDir, with forward slashes ("/").
func (m *Matchers) ShouldInclude(relPath string) bool {
	if m == nil {
		return true
	}

	ignored := false
	if m.ignore != nil && m.ignore.MatchesPath(relPath) {
		ignored = true
	}

	included := false
	if m.include != nil && m.include.MatchesPath(relPath) {
		included = true
	}

	if m.hasIncludes {
		// Include if rescued OR not ignored
		return included || !ignored
	}

	if ignored {
		return false
	}
	return true
}
