// Package ignore aggregates default, .gitignore, and CLI patterns into
// matchers that decide whether a given path should be included in a snapshot.
package ignore

import (
	"fmt"
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

// Matchers holds compiled ignore and include patterns with proper precedence.
//
// Precedence order:
//  1. Base ignore (defaults + .gitignore)
//  2. Include patterns override base ignore
//  3. Exclude patterns are final (cannot be overridden)
type Matchers struct {
	baseIgnore  *gitignore.GitIgnore // defaults + .gitignore
	include     *gitignore.GitIgnore // CLI --include
	exclude     *gitignore.GitIgnore // CLI --exclude (final)
	hasIncludes bool
	hasExcludes bool
}

// NewMatchers builds ignore/include matchers from defaults, .gitignore,
// CLI excludes, and CLI includes.
func NewMatchers(sourceDir string, excludePatterns, includePatterns []string) (*Matchers, error) {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, err
	}

	// Validate directory exists
	info, err := os.Stat(absSourceDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", absSourceDir)
	}

	// Base ignore: defaults + .gitignore
	var baseIgnoreLines []string
	baseIgnoreLines = append(baseIgnoreLines, DefaultPatterns...)

	gitignorePath := filepath.Join(absSourceDir, ".gitignore")
	if b, err := os.ReadFile(gitignorePath); err == nil {
		lines := strings.Split(string(b), "\n")
		baseIgnoreLines = append(baseIgnoreLines, lines...)
	}

	var baseIgnoreMatcher *gitignore.GitIgnore
	if len(baseIgnoreLines) > 0 {
		baseIgnoreMatcher = gitignore.CompileIgnoreLines(baseIgnoreLines...)
	}

	// Include patterns
	var includeMatcher *gitignore.GitIgnore
	if len(includePatterns) > 0 {
		includeMatcher = gitignore.CompileIgnoreLines(includePatterns...)
	}

	// Exclude patterns (final override)
	var excludeMatcher *gitignore.GitIgnore
	if len(excludePatterns) > 0 {
		excludeMatcher = gitignore.CompileIgnoreLines(excludePatterns...)
	}

	return &Matchers{
		baseIgnore:  baseIgnoreMatcher,
		include:     includeMatcher,
		exclude:     excludeMatcher,
		hasIncludes: len(includePatterns) > 0,
		hasExcludes: len(excludePatterns) > 0,
	}, nil
}

// ShouldInclude decides if a relative path should be included in the snapshot.
//
// relPath must be a path relative to sourceDir, with forward slashes ("/").
//
// Logic:
//  1. If matched by --exclude: always exclude (final decision)
//  2. If matched by --include: include (overrides base ignore)
//  3. If matched by base ignore (defaults + .gitignore): exclude
//  4. Otherwise: include
func (m *Matchers) ShouldInclude(relPath string) bool {
	if m == nil {
		return true
	}

	// Step 1: Check final excludes (highest priority - cannot be overridden)
	if m.hasExcludes && m.exclude != nil && m.exclude.MatchesPath(relPath) {
		return false
	}

	// Step 2: Check includes (overrides base ignore)
	if m.hasIncludes && m.include != nil && m.include.MatchesPath(relPath) {
		return true
	}

	// Step 3: Check base ignore (defaults + .gitignore)
	if m.baseIgnore != nil && m.baseIgnore.MatchesPath(relPath) {
		return false
	}

	// Step 4: Default to include
	return true
}
