// Package filter aggregates default and CLI rules into an ordered matcher
// that decides whether a given path should be included in a snapshot.
package filter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// DefaultPatterns are the default exclude patterns applied unless --no-defaults is set.
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
	"**/*.snp",
}

// Rule represents a single ordered include or exclude rule.
type Rule struct {
	Pattern string
	Exclude bool // true = exclude, false = include
}

// rule is a compiled internal rule.
type rule struct {
	matcher *gitignore.GitIgnore
	exclude bool
}

// Matcher holds an ordered list of compiled rules. Last match wins.
type Matcher struct {
	rules []rule
}

// New builds a Matcher from defaults, .gitignore, and CLI rules.
// If noDefaults is true, defaults and .gitignore are skipped.
func New(sourceDir string, noDefaults bool, rules []Rule) (*Matcher, error) {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absSourceDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", absSourceDir)
	}

	var compiled []rule

	if !noDefaults {
		var defaultLines []string
		defaultLines = append(defaultLines, DefaultPatterns...)

		gitignorePath := filepath.Join(absSourceDir, ".gitignore")
		if b, err := os.ReadFile(gitignorePath); err == nil {
			lines := strings.Split(string(b), "\n")
			defaultLines = append(defaultLines, lines...)
		}

		if len(defaultLines) > 0 {
			m := gitignore.CompileIgnoreLines(defaultLines...)
			compiled = append(compiled, rule{matcher: m, exclude: true})
		}
	}

	for _, r := range rules {
		m := gitignore.CompileIgnoreLines(r.Pattern)
		compiled = append(compiled, rule{matcher: m, exclude: r.Exclude})
	}

	return &Matcher{rules: compiled}, nil
}

// ShouldInclude returns true if relPath should be included.
// Rules are evaluated in order; last match wins.
// If no rule matches, the file is included by default.
func (m *Matcher) ShouldInclude(relPath string) bool {
	if m == nil {
		return true
	}

	result := true

	for _, r := range m.rules {
		if r.matcher.MatchesPath(relPath) {
			result = !r.exclude
		}
	}

	return result
}
