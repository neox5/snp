package matcher

import "testing"

// Internal test: the scope-aware match function is unexported.

// TestMatchScopes runs the same pattern and path through both scopes.
// ScopeSubtree: a matched path also matches everything beneath it.
// ScopeEntry:   only the named entry matches (git semantics).
func TestMatchScopes(t *testing.T) {
	tests := []struct {
		pattern     string
		path        string
		isDir       bool
		wantSubtree bool
		wantEntry   bool
	}{
		// trailing slash
		{"automation/", "automation", true, true, true},
		{"automation/", "automation", false, false, false},
		{"automation/", "automation/key", false, true, false},

		// no slash: name at any depth
		{"README.md", "README.md", false, true, true},
		{"README.md", "docs/README.md", false, true, true},
		{"*.tmpx", "d/b.tmpx", false, true, true},
		{"internal", "cmd/internal/main.go", false, true, false},

		// slash in the middle: anchored
		{"build/*", "build/other.txt", false, true, true},
		{"build/*", "build/keep/a.txt", false, true, false},
		{"build/keep", "build/keep", true, true, true},
		{"build/keep", "build/keep/a.txt", false, true, false},
		{"automation/*.pub", "automation/x.pub", false, true, true},
		{"automation/*.pub", "automation/key", false, false, false},

		// leading slash: anchored to the root
		{"/top.txt", "top.txt", false, true, true},
		{"/top.txt", "sub/top.txt", false, false, false},

		// **
		{"**/foo", "a/b/foo", false, true, true},
		{"**/foo", "a/foo/x", false, true, false},
		{"a/**/b", "a/x/y/b", false, true, true},
		{"a/**/b", "a/b", false, true, true},

		// trailing ** : subtree also matches the directory itself
		{"docs/**", "docs", true, true, false},
		{"docs/**", "docs/a", false, true, true},
		{"__pycache__/", "src/__pycache__", true, false, true}, // subtree: root only (unchanged)
	}

	for _, tt := range tests {
		name := tt.pattern + " vs " + tt.path
		t.Run(name, func(t *testing.T) {
			got, err := match(tt.pattern, tt.path, tt.isDir, ScopeSubtree)
			if err != nil || got != tt.wantSubtree {
				t.Errorf("subtree: match(%q, %q, %v) = %v, %v; want %v", tt.pattern, tt.path, tt.isDir, got, err, tt.wantSubtree)
			}
			got, err = match(tt.pattern, tt.path, tt.isDir, ScopeEntry)
			if err != nil || got != tt.wantEntry {
				t.Errorf("entry: match(%q, %q, %v) = %v, %v; want %v", tt.pattern, tt.path, tt.isDir, got, err, tt.wantEntry)
			}
		})
	}
}

// TestMatchEntry covers entry-scope cases without a subtree counterpart.
func TestMatchEntry(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		isDir   bool
		want    bool
	}{
		{"/dist/", "dist", true, true},
		{"/dist/", "dist", false, false},
		{"/dist", "dist", false, true},
		{"a/b/", "a/b", true, true},
		{"a/b/", "a/b/c", false, false},
		{"tmp", "x/tmp", true, true},
		{"tmp", "tmp/g", false, false},
		{"*", "a/b/c", false, true},
		{"", "a", false, false},

		// trailing slash: directories only, rest matched like the bare pattern
		{"__pycache__/", "__pycache__", true, true},
		{"__pycache__/", "src/__pycache__", true, true},
		{"__pycache__/", "src/__pycache__", false, false},
		{"__pycache__/", "src/__pycache__/x.pyc", false, false},
		{"*.egg-info/", "pkg.egg-info", true, true},
		{"*.egg-info/", "a/pkg.egg-info", true, true},
		{"*.egg-info/", "pkg.egg-info", false, false},
		{"/", "a", true, false},

		// trailing /**: what is inside, not the directory itself
		{"docs/**", "docs", true, false},
		{"docs/**", "docs/a", false, true},
		{"docs/**", "docs/a/b", false, true},
		{"docs/**", "other/a", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+" vs "+tt.path, func(t *testing.T) {
			got, err := match(tt.pattern, tt.path, tt.isDir, ScopeEntry)
			if err != nil || got != tt.want {
				t.Errorf("match(%q, %q, %v) = %v, %v; want %v", tt.pattern, tt.path, tt.isDir, got, err, tt.want)
			}
		})
	}
}
