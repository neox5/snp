// Package matcher aggregates ordered rules into a matcher that decides
// whether a given path should be included in a snapshot.
package matcher

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

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

// Match reports whether relPath matches the given gitignore-style pattern.
//
// Pattern rules:
//
//  1. Trailing slash (e.g. "node_modules/"):
//     matches any file whose path starts with the directory prefix.
//     "node_modules/" matches "node_modules/pkg/foo.js" but NOT "node_modules".
//
//  2. No slash (e.g. "*.test.js", "README.md"):
//     matched against the filename component only (last path segment),
//     at any depth.
//
//  3. Leading slash (e.g. "/README.md"):
//     anchored to root — matched against full relPath only.
//
//  4. Slash in middle (e.g. "src/*.go", "a/**/b"):
//     matched against full relPath, anchored to root.
//     ** as a path segment matches zero or more directory segments.
//
// Glob characters per segment: * ? [abc] [a-z]
// Returns an error only if the pattern is malformed.
func Match(pattern, relPath string) (bool, error) {
	if pattern == "" || relPath == "" {
		return false, nil
	}

	// Rule 1: trailing slash — directory prefix match (files only)
	if prefix, ok := strings.CutSuffix(pattern, "/"); ok {
		return strings.HasPrefix(relPath, prefix+"/"), nil
	}

	// Rule 2: no slash — match filename component only
	if !strings.Contains(pattern, "/") {
		return matchGlob(pattern, path.Base(relPath))
	}

	// Rules 3 & 4: slash present — anchored full path match
	// strip optional leading slash (anchoring is always implicit)
	pattern = strings.TrimPrefix(pattern, "/")
	return matchSegments(
		strings.Split(pattern, "/"),
		strings.Split(relPath, "/"),
	)
}

// matchSegments matches ordered pattern segments against path segments.
// ** as a segment matches zero or more path segments.
func matchSegments(pat, parts []string) (bool, error) {
	for len(pat) > 0 {
		seg := pat[0]
		pat = pat[1:]

		if seg == "**" {
			// ** at end of pattern matches everything remaining
			if len(pat) == 0 {
				return true, nil
			}
			// try matching remaining pattern at every position in parts
			for i := 0; i <= len(parts); i++ {
				ok, err := matchSegments(pat, parts[i:])
				if err != nil {
					return false, err
				}
				if ok {
					return true, nil
				}
			}
			return false, nil
		}

		// regular segment: consume one path part
		if len(parts) == 0 {
			return false, nil
		}
		ok, err := matchGlob(seg, parts[0])
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		parts = parts[1:]
	}

	// pattern exhausted — must have consumed all path parts
	return len(parts) == 0, nil
}

// matchGlob matches a single non-separator pattern against a single name segment.
func matchGlob(pattern, name string) (bool, error) {
	return path.Match(pattern, name)
}
