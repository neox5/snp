// Package matcher aggregates ordered rules into a matcher that decides
// whether a given path should be included in a snapshot.
package matcher

import (
	"fmt"
	"path"
	"strings"
)

// Matcher holds an ordered list of compiled rules. Last match wins.
type Matcher struct {
	rules Rules
}

// New builds a Matcher from an ordered list of rules.
func New(rs Rules) *Matcher {
	return &Matcher{rules: rs}
}

// ShouldInclude returns true if relPath should be included.
// Rules are evaluated in order; last match wins.
// If no rule matches, the file is included by default.
func (m *Matcher) ShouldInclude(relPath string) bool {
	// default
	result := true

	for _, r := range m.rules {
		if len(r.Pattern) == 0 {
			if r.Type == RuleExcludeAll {
				result = false
				continue
			}
			if r.Type == RuleIncludeAll {
				result = true
				continue
			}
			panic(fmt.Sprintf("empty pattern for type %s", r.Type))
		}

		matched, err := Match(r.Pattern, relPath)
		if err != nil {
			fmt.Printf("warn: error in match with pattern '%s' on path '%s': %v", r.Pattern, relPath, err)
			continue
		}
		if !matched {
			continue
		}
		if r.Type == RuleExclude {
			result = false
			continue
		}
		if r.Type == RuleInclude {
			result = true
			continue
		}
		panic(fmt.Sprintf("pattern '%s' set on type '%s'", r.Pattern, r.Type))
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
		return path.Match(pattern, path.Base(relPath))
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
		ok, err := path.Match(seg, parts[0])
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
