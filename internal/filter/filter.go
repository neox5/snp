// Package filter aggregates ordered rules into a matcher that decides
// whether a given path should be included in a snapshot.
package filter

import (
	"fmt"
	"os"
	"path/filepath"
)

// RuleType defines the kind of rule.
type RuleType int

const (
	RuleInclude    RuleType = iota // --include <pattern>
	RuleExclude                    // --exclude <pattern>
	RuleIncludeAll                 // --include-all
	RuleExcludeAll                 // --exclude-all
)

// Rule represents a single ordered filter rule.
type Rule struct {
	Type    RuleType
	Pattern string // only used for RuleInclude and RuleExclude
}

type Rules []Rule

func NewRules() Rules {
	return Rules{}
}

func (r Rules) AddRules(rs Rules) Rules {
	return append(r, rs...)
}

func (r Rules) AddExcludeAll() Rules {
	return append(r, Rule{Type: RuleExcludeAll})
}

func (r Rules) AddExclude(p string) Rules {
	return append(r, Rule{Type: RuleExclude, Pattern: p})
}

func (r Rules) AddIncludeAll() Rules {
	return append(r, Rule{Type: RuleIncludeAll})
}

func (r Rules) AddInclude(p string) Rules {
	return append(r, Rule{Type: RuleInclude, Pattern: p})
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
