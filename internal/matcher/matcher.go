// Package matcher implements gitignore-style pattern matching and ordered filter rule evaluation.
package matcher

import (
	"fmt"
	"path"
	"strings"
)

// Matcher holds an ordered list of compiled rules and pre-computed traversal sets.
//
// Rules answer the question "should this file appear in output?" and are
// evaluated at call time by ShouldInclude.
//
// Traversal sets answer the question "should this directory be entered?" and
// are pre-computed once from the include patterns at construction time by
// ShouldTraverse.
type Matcher struct {
	rules             Rules
	parentPrefixes    map[string]bool // exact dirs required as passthrough
	recursivePrefixes map[string]bool // dirs whose entire subtrees must be entered
	traverseAll       bool            // any include pattern can match at any depth
	hasEntry          bool            // at least one rule has ScopeEntry
}

// New builds a Matcher from an ordered list of rules.
// Traversal sets are pre-computed from the include patterns.
func New(rs Rules) *Matcher {
	m := &Matcher{rules: rs}
	m.parentPrefixes, m.recursivePrefixes, m.traverseAll = buildTraversalSets(rs)
	for _, r := range rs {
		if r.Scope == ScopeEntry {
			m.hasEntry = true
			break
		}
	}
	return m
}

// ShouldInclude returns true if the file at relPath should appear in output.
// Rules are evaluated in order; last match wins.
// If no rule matches, the file is included by default.
// Only call for files — directories are handled by ShouldTraverse.
func (m *Matcher) ShouldInclude(relPath string) bool {
	if m == nil {
		return true
	}
	return m.evalRules(relPath, false)
}

// ShouldTraverse returns true if the directory at relPath should be entered.
//
// Two-stage evaluation:
//  1. Traversal sets — if any include pattern requires descending through this
//     directory, return true immediately regardless of rule outcomes.
//  2. Rule evaluation — normal include/exclude logic with isDir=true.
//
// Only call for directories — files are handled by ShouldInclude.
func (m *Matcher) ShouldTraverse(relPath string) bool {
	if m == nil {
		return true
	}

	// Stage 1: traversal sets express intent that cannot be blocked by rules.
	if m.traverseAll {
		return true
	}
	if m.parentPrefixes[relPath] {
		return true
	}
	for prefix := range m.recursivePrefixes {
		if relPath == prefix || strings.HasPrefix(relPath, prefix+"/") {
			return true
		}
	}

	// Stage 2: fall back to rule evaluation.
	return m.evalRules(relPath, true)
}

// evalRules evaluates all rules against relPath in order; last match wins.
// isDir controls Rule 1 (trailing slash) matching in Match.
//
// ScopeEntry rules follow git: they cannot re-include a path whose parent
// directory is excluded by ScopeEntry rules. Rules of other scopes (--include,
// --exclude) are not restricted by this and are evaluated normally.
func (m *Matcher) evalRules(relPath string, isDir bool) bool {
	result := true
	blocked := m.hasEntry && m.parentExcluded(relPath)

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

		if r.Scope == ScopeEntry && blocked {
			if r.Type == RuleExclude {
				result = false
			}
			continue
		}

		matched, err := match(r.Pattern, relPath, isDir, r.Scope)
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

// parentExcluded reports whether any parent directory of relPath is excluded
// by the ScopeEntry rules.
func (m *Matcher) parentExcluded(relPath string) bool {
	parts := strings.Split(relPath, "/")
	for i := 1; i < len(parts); i++ {
		if m.entryExcluded(strings.Join(parts[:i], "/")) {
			return true
		}
	}
	return false
}

// entryExcluded evaluates the ScopeEntry rules for the directory dir.
// The last matching rule wins; no match means not excluded.
func (m *Matcher) entryExcluded(dir string) bool {
	excluded := false
	for _, r := range m.rules {
		if r.Scope != ScopeEntry {
			continue
		}
		if ok, err := match(r.Pattern, dir, true, ScopeEntry); err == nil && ok {
			excluded = r.Type == RuleExclude
		}
	}
	return excluded
}

// buildTraversalSets pre-computes directory traversal intent from include rules.
//
// ScopeEntry rules are skipped: they must not widen traversal (a rule such as
// "!.gitkeep" would otherwise set traverseAll and disable directory pruning).
//
// Each RuleInclude pattern is classified by scanning its directory segments
// (all segments except the last) for the first glob character or **:
//
//   - No slash or leading ** → traverseAll (match possible at any depth)
//   - Exact prefix before first glob → parentPrefixes (known ancestor chain)
//   - Glob at or beyond pivot → recursivePrefixes (subtree must be entered)
//   - Trailing ** as last segment → recursivePrefixes on the prefix
func buildTraversalSets(rules Rules) (parentPrefixes map[string]bool, recursivePrefixes map[string]bool, traverseAll bool) {
	parentPrefixes = make(map[string]bool)
	recursivePrefixes = make(map[string]bool)

	for _, r := range rules {
		if r.Type != RuleInclude || r.Scope == ScopeEntry || r.Pattern == "" {
			continue
		}

		pattern := r.Pattern

		// No slash: Match rule 2 handles these at any depth.
		if !strings.Contains(pattern, "/") {
			traverseAll = true
			continue
		}

		// Normalize: strip leading and trailing slashes.
		pattern = strings.TrimPrefix(pattern, "/")
		pattern = strings.TrimSuffix(pattern, "/")

		parts := strings.Split(pattern, "/")

		// Single segment after normalization: root-level target, no parents needed.
		if len(parts) <= 1 {
			continue
		}

		// Directory segments are everything except the last segment.
		// For file patterns: last segment is the file matcher.
		// For trailing-slash patterns (already trimmed): last segment was the target dir.
		// In both cases the last segment is matched directly by Match — not our concern here.
		dirParts := parts[:len(parts)-1]

		// Find pivot: first dir segment that is ** or contains a glob character.
		pivot := len(dirParts)
		for i, seg := range dirParts {
			if seg == "**" || containsGlob(seg) {
				pivot = i
				break
			}
		}

		if pivot == 0 {
			// First dir segment is already a glob — any directory could match.
			traverseAll = true
			continue
		}

		// Exact prefix: all dir segments before the pivot.
		exactPrefix := strings.Join(dirParts[:pivot], "/")
		addParentChain(parentPrefixes, exactPrefix)

		// If glob segments exist at or beyond the pivot, matching may extend
		// to arbitrary depth — the subtree under exactPrefix must be entered.
		if pivot < len(dirParts) {
			recursivePrefixes[exactPrefix] = true
		}

		// Trailing **: last segment of the full pattern (not just dirParts).
		// e.g. "internal/**" — the subtree under exactPrefix must be entered.
		if parts[len(parts)-1] == "**" {
			recursivePrefixes[exactPrefix] = true
		}
	}

	return
}

// addParentChain adds dirPath and every ancestor path to parents.
// "a/b/c" adds "a", "a/b", and "a/b/c".
func addParentChain(parents map[string]bool, dirPath string) {
	parts := strings.Split(dirPath, "/")
	for i := 1; i <= len(parts); i++ {
		parents[strings.Join(parts[:i], "/")] = true
	}
}

// containsGlob reports whether s contains any glob metacharacter.
func containsGlob(s string) bool {
	return strings.ContainsAny(s, "*?[")
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
//  2. No slash (e.g. "*.test.js", "README.md", "internal"):
//     Matches any path segment at any depth.
//     "*.test.js" matches "src/tests/kernel.test.js" and "kernel.test.js".
//     "internal" matches "internal", "internal/config", "internal/config/config.go",
//     and "cmd/internal/main.go".
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
//
// Match uses ScopeSubtree: a matched path also matches everything beneath it.
func Match(pattern, relPath string, isDir bool) (bool, error) {
	return match(pattern, relPath, isDir, ScopeSubtree)
}

// match implements Match for the given scope. With ScopeEntry the pattern
// matches only the entry itself:
//
//	Rule 1: the directory itself (isDir=true), not the paths beneath it; the
//	        pattern without its slash is matched as by Rules 2-4.
//	Rule 2: the last path segment only, not every segment.
//	Rules 3 & 4: the pattern must consume the whole path.
func match(pattern, relPath string, isDir bool, scope Scope) (bool, error) {
	if pattern == "" || relPath == "" {
		return false, nil
	}

	// Rule 1: trailing slash — match directory and its contents
	if prefix, ok := strings.CutSuffix(pattern, "/"); ok {
		if scope == ScopeEntry {
			// git: a trailing slash restricts the match to directories; the
			// rest is matched like the pattern without the slash.
			if !isDir {
				return false, nil
			}
			return match(prefix, relPath, true, ScopeEntry)
		}
		return isDir && relPath == prefix ||
			strings.HasPrefix(relPath, prefix+"/"), nil
	}

	// Rule 2: no slash — match against every path segment
	// "internal" matches any path containing "internal" as a segment:
	// "internal", "internal/config", "cmd/internal/main.go"
	if !strings.Contains(pattern, "/") {
		segments := strings.Split(relPath, "/")
		if scope == ScopeEntry {
			segments = segments[len(segments)-1:]
		}
		for _, segment := range segments {
			ok, err := path.Match(pattern, segment)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}

	// Rules 3 & 4: slash present — anchored full path match
	pattern = strings.TrimPrefix(pattern, "/")
	return matchSegments(
		strings.Split(pattern, "/"),
		strings.Split(relPath, "/"),
		scope,
	)
}

// matchSegments matches ordered pattern segments against path segments.
// ** as a segment matches zero or more path segments.
// When the pattern is exhausted, remaining path parts are allowed for
// ScopeSubtree — a matched prefix also matches everything beneath it.
// With ScopeEntry the pattern must consume the whole path.
func matchSegments(pat, parts []string, scope Scope) (bool, error) {
	for len(pat) > 0 {
		seg := pat[0]
		pat = pat[1:]

		if seg == "**" {
			// ** at end of pattern matches everything remaining.
			// With ScopeEntry something must remain: "dir/**" matches what is
			// inside dir, not dir itself (git).
			if len(pat) == 0 {
				return scope == ScopeSubtree || len(parts) > 0, nil
			}
			// try matching remaining pattern at every position in parts
			for i := 0; i <= len(parts); i++ {
				ok, err := matchSegments(pat, parts[i:], scope)
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

	// Pattern exhausted — the path starts with the matched prefix.
	// ScopeSubtree: remaining parts mean the path is inside the matched directory.
	// ScopeEntry: remaining parts mean the path is a different, deeper entry.
	return scope == ScopeSubtree || len(parts) == 0, nil
}
