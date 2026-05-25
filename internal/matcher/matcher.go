// Package matcher aggregates ordered rules into a matcher that decides
// whether a given path should be included in a snapshot.
package matcher

import (
	"fmt"
	"path"
	"path/filepath"
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
func (m *Matcher) ShouldInclude(relPath string, isDir bool) bool {
	if m == nil {
		return true
	}

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

		matched, err := Match(r.Pattern, relPath, isDir)
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
// The matching considers whether the path is a directory (isDir) for patterns
// ending with a trailing slash.
//
// Pattern rules:
//
//  1. Trailing slash (e.g. "node_modules/"):
//     Matches the directory itself (only if isDir=true) and all paths under it.
//     "node_modules/" matches "node_modules" (if isDir) and "node_modules/pkg/foo.js".
//
//  2. No slash (e.g. "*.test.js", "README.md"):
//     Matches the last path segment (filename) at any depth.
//     "*.test.js" matches "src/tests/kernel.test.js" and "kernel.test.js".
//
//  3. Leading slash (e.g. "/README.md"):
//     Anchored to the start of relPath. Matches only the exact path.
//     "/README.md" matches "README.md" but not "docs/README.md".
//
//  4. Slash in middle (e.g. "src/*.go", "a/**/b"):
//     Anchored to the start of relPath. Matches paths starting with the pattern.
//     "src/*.go" matches "src/main.go" but not "pkg/src/main.go".
//
// Glob characters per segment: * ? [abc] [a-z]
// Returns an error only if the pattern is malformed.
func Match(pattern, relPath string, isDir bool) (bool, error) {
	if pattern == "" || relPath == "" {
		return false, nil
	}

	// Rule 1: trailing slash — match directory and its contents
	if prefix, ok := strings.CutSuffix(pattern, "/"); ok {
		return isDir && relPath == prefix ||
			strings.HasPrefix(relPath, prefix+"/"), nil
	}

	// Rule 2: no slash — match filename component only
	if !strings.Contains(pattern, "/") {
		return path.Match(pattern, filepath.Base(relPath))
	}

	// Rules 3 & 4: slash present — anchored full path match
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
