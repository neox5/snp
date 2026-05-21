// Package filter aggregates ordered rules into a matcher that decides
// whether a given path should be included in a snapshot.
package filter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultPatterns are the patterns applied by --exclude-defaults.
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

// RuleType defines the kind of rule.
type RuleType int

const (
	RuleInclude         RuleType = iota // --include <pattern>
	RuleExclude                         // --exclude <pattern>
	RuleIncludeAll                      // --include-all
	RuleExcludeAll                      // --exclude-all
	RuleExcludeDefaults                 // --exclude-defaults
)

// Rule represents a single ordered filter rule.
type Rule struct {
	Type    RuleType
	Pattern string // only used for RuleInclude and RuleExclude
}

// compiled is an internal evaluated rule.
type compiled struct {
	patterns []string // empty means match-all (include-all / exclude-all)
	exclude  bool
}

// Matcher holds an ordered list of compiled rules. Last match wins.
type Matcher struct {
	rules []compiled
}

// New builds a Matcher from an ordered list of rules.
// sourceDir is used to load .gitignore for RuleExcludeDefaults.
func New(sourceDir string, rules []Rule) (*Matcher, error) {
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

	var cr []compiled

	for _, r := range rules {
		switch r.Type {
		case RuleIncludeAll:
			cr = append(cr, compiled{patterns: nil, exclude: false})

		case RuleExcludeAll:
			cr = append(cr, compiled{patterns: nil, exclude: true})

		case RuleExcludeDefaults:
			patterns := append([]string{}, DefaultPatterns...)
			patterns = append(patterns, loadGitignore(filepath.Join(absSourceDir, ".gitignore"))...)
			cr = append(cr, compiled{patterns: patterns, exclude: true})

		case RuleInclude:
			cr = append(cr, compiled{patterns: []string{r.Pattern}, exclude: false})

		case RuleExclude:
			cr = append(cr, compiled{patterns: []string{r.Pattern}, exclude: true})
		}
	}

	return &Matcher{rules: cr}, nil
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
		if len(r.patterns) == 0 {
			// include-all or exclude-all — matches everything
			result = !r.exclude
			continue
		}

		for _, pattern := range r.patterns {
			matched, err := Match(pattern, relPath)
			if err != nil || !matched {
				continue
			}
			result = !r.exclude
			break
		}
	}

	return result
}

// loadGitignore reads a .gitignore file and returns non-empty, non-comment lines.
func loadGitignore(path string) []string {
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
	return lines
}
