package matcher_test

import (
	"strings"
	"testing"

	"github.com/neox5/snp/internal/matcher"
)

func TestShouldInclude_NilMatcher(t *testing.T) {
	var m *matcher.Matcher
	if !m.ShouldInclude("any/path.go") {
		t.Error("nil Matcher should include all paths")
	}
}

func TestShouldInclude_OrderedRules(t *testing.T) {
	tests := []struct {
		name   string
		rules  matcher.Rules
		path   string
		want   bool
		reason string
	}{
		{
			name: "exclude-all then include Go — only Go files",
			rules: []matcher.Rule{
				{Type: matcher.RuleExcludeAll},
				{Type: matcher.RuleInclude, Pattern: "**/*.go"},
			},
			path:   "src/main.go",
			want:   true,
			reason: "include **/*.go wins after exclude-all",
		},
		{
			name: "exclude-all then include Go — only Go files duplicate",
			rules: []matcher.Rule{
				{Type: matcher.RuleExcludeAll},
				{Type: matcher.RuleInclude, Pattern: "**/*.go"},
			},
			path:   "src/main.go",
			want:   true,
			reason: "include **/*.go wins after exclude-all",
		},
		{
			name: "exclude-all then include Go — non-Go excluded",
			rules: []matcher.Rule{
				{Type: matcher.RuleExcludeAll},
				{Type: matcher.RuleInclude, Pattern: "**/*.go"},
			},
			path:   "README.md",
			want:   false,
			reason: "exclude-all wins, no include rule matches",
		},
		{
			name: "last rule wins — rescue from exclude",
			rules: []matcher.Rule{
				{Type: matcher.RuleIncludeAll},
				{Type: matcher.RuleExclude, Pattern: "**/*_test.go"},
				{Type: matcher.RuleInclude, Pattern: "internal/auth/auth_test.go"},
			},
			path:   "internal/auth/auth_test.go",
			want:   true,
			reason: "last include rule rescues specific test file",
		},
		{
			name: "last rule wins — exclude after include",
			rules: []matcher.Rule{
				{Type: matcher.RuleIncludeAll},
				{Type: matcher.RuleExclude, Pattern: "**/*_test.go"},
			},
			path:   "internal/auth/auth_test.go",
			want:   false,
			reason: "exclude wins as last matching rule",
		},
		{
			name: "include-all baseline — everything included",
			rules: []matcher.Rule{
				{Type: matcher.RuleIncludeAll},
			},
			path:   "node_modules/package.json",
			want:   true,
			reason: "include-all with no further rules includes everything",
		},

		// ── anchored subpath regression ───────────────────────────────────────
		{
			name: "exclude-all with subpath include — file inside path is included",
			rules: []matcher.Rule{
				{Type: matcher.RuleExcludeAll},
				{Type: matcher.RuleInclude, Pattern: "internal/config"},
			},
			path:   "internal/config/config.go",
			want:   true,
			reason: "internal/config pattern matches files inside the directory",
		},
		{
			name: "exclude-all with subpath include — sibling path stays excluded",
			rules: []matcher.Rule{
				{Type: matcher.RuleExcludeAll},
				{Type: matcher.RuleInclude, Pattern: "internal/config"},
			},
			path:   "internal/snapshot/collect.go",
			want:   false,
			reason: "internal/config pattern does not match sibling directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := matcher.New(tt.rules)
			got := m.ShouldInclude(tt.path)
			if got != tt.want {
				t.Errorf("ShouldInclude(%q) = %v, want %v\nReason: %s", tt.path, got, tt.want, tt.reason)
			}
		})
	}
}

// gitignoreRules translates .gitignore lines into ScopeEntry rules the same way
// the config package does: "!pattern" re-includes, every other line excludes.
func gitignoreRules(lines ...string) matcher.Rules {
	r := matcher.NewRules()
	for _, line := range lines {
		if p, ok := strings.CutPrefix(line, "!"); ok {
			r = r.AddIncludeEntry(p)
			continue
		}
		r = r.AddExcludeEntry(line)
	}
	return r
}

// TestShouldInclude_EntryRules: the expected values are the results of real git
// (git ls-files --others --exclude-standard) for the same .gitignore and files.
func TestShouldInclude_EntryRules(t *testing.T) {
	tests := []struct {
		name   string
		ignore []string
		want   map[string]bool // path -> included
	}{
		{
			name:   "starter repo allowlist",
			ignore: []string{"*", "!automation/", "!personal/", "!emergency/", "!automation/*.pub", "!personal/*.pub", "!emergency/*.pub", "!.gitkeep", "!README.md", "!.gitignore"},
			want: map[string]bool{
				"README.md":                         true,
				".gitignore":                        true,
				"automation/.gitkeep":               true,
				"automation/id_ed25519_ansible.pub": true,
				"personal/id_x.pub":                 true,
				"emergency/.gitkeep":                true,
				"automation/id_ed25519_ansible":     false,
				"personal/id_x":                     false,
				"sub/.gitkeep":                      false,
				"sub/x":                             false,
			},
		},
		{
			name:   "partial exclude, re-included directory keeps its contents",
			ignore: []string{"out/*", "!out/keep"},
			want: map[string]bool{
				"top.txt":             true,
				"out/keep/a.txt":      true,
				"out/keep/deep/b.txt": true,
				"out/other.txt":       false,
			},
		},
		{
			name:   "excluded parent blocks re-include of a file",
			ignore: []string{"secrets/", "!secrets/notes.md"},
			want: map[string]bool{
				"a.txt":               true,
				"secrets/notes.md":    false,
				"secrets/private.key": false,
			},
		},
		{
			name:   "excluded parent blocks re-include of a nested directory",
			ignore: []string{"a/", "!a/b/", "!a/b/c.txt"},
			want: map[string]bool{
				"z.txt":     true,
				"a/d.txt":   false,
				"a/b/c.txt": false,
			},
		},
		{
			name:   "allowlist over three levels",
			ignore: []string{"*", "!a/", "!a/b/", "!a/b/c.txt"},
			want: map[string]bool{
				"a/b/c.txt": true,
				"a/b/d.txt": false,
				"a/e.txt":   false,
			},
		},
		{
			name:   "directory excluded by a glob is not re-included by its parent's entry",
			ignore: []string{"*", "!d/", "!d/x/f.txt", "!d/y.txt"},
			want: map[string]bool{
				"d/y.txt":   true,
				"d/x/f.txt": false,
				"d/x/g.txt": false,
			},
		},
		{
			name:   "trailing slash with a name matches directories at any depth",
			ignore: []string{"__pycache__/"},
			want: map[string]bool{
				"top.txt":               true,
				"src/x.py":              true,
				"src/__pycache__/x.pyc": false,
				"__pycache__":           true, // a file, the pattern is for directories
			},
		},
		{
			name:   "trailing slash with a glob matches directories",
			ignore: []string{"*.egg-info/"},
			want: map[string]bool{
				"src/x.py":           true,
				"pkg.egg-info/PKG":   false,
				"a/pkg.egg-info/PKG": false,
			},
		},
		{
			name:   "trailing /** excludes the contents and allows re-including inside",
			ignore: []string{"docs/**", "!docs/keep.md"},
			want: map[string]bool{
				"top.txt":          true,
				"docs/keep.md":     true,
				"docs/other.md":    false,
				"docs/sub/x.md":    false,
				"docs/sub/keep.md": false,
			},
		},
		{
			name:   "name without slash matches a directory at any depth",
			ignore: []string{"tmp"},
			want: map[string]bool{
				"tmp.txt": true,
				"tmp/g":   false,
				"x/tmp/f": false,
			},
		},
		{
			name:   "leading slash anchors to the root",
			ignore: []string{"/top.txt"},
			want: map[string]bool{
				"top.txt":     false,
				"sub/top.txt": true,
			},
		},
		{
			name:   "glob exclude with re-include of one name",
			ignore: []string{"*.tmpx", "!keep.tmpx"},
			want: map[string]bool{
				"a.tmpx":      false,
				"d/b.tmpx":    false,
				"d/keep.tmpx": true,
				"d/c.txt":     true,
			},
		},
	}

	for _, tt := range tests {
		m := matcher.New(gitignoreRules(tt.ignore...))
		for path, want := range tt.want {
			t.Run(tt.name+"/"+path, func(t *testing.T) {
				if got := m.ShouldInclude(path); got != want {
					t.Errorf("ShouldInclude(%q) = %v, want %v (.gitignore: %v)", path, got, want, tt.ignore)
				}
			})
		}
	}
}

// TestShouldInclude_FlagsOverGitignore: flags come after .gitignore rules and
// are not restricted by excluded parent directories.
func TestShouldInclude_FlagsOverGitignore(t *testing.T) {
	ignore := []string{"secrets/", "!secrets/notes.md"}

	tests := []struct {
		name  string
		rules matcher.Rules
		want  map[string]bool
	}{
		{
			name:  "no flags",
			rules: gitignoreRules(ignore...),
			want:  map[string]bool{"a.md": true, "secrets/notes.md": false, "secrets/private.key": false},
		},
		{
			name:  "include *.md reaches notes.md only",
			rules: gitignoreRules(ignore...).AddInclude("*.md"),
			want:  map[string]bool{"a.md": true, "secrets/notes.md": true, "secrets/private.key": false},
		},
		{
			name:  "include secrets takes the whole directory",
			rules: gitignoreRules(ignore...).AddInclude("secrets"),
			want:  map[string]bool{"a.md": true, "secrets/notes.md": true, "secrets/private.key": true},
		},
		{
			name:  "include one file inside the excluded directory",
			rules: gitignoreRules(ignore...).AddInclude("secrets/notes.md"),
			want:  map[string]bool{"secrets/notes.md": true, "secrets/private.key": false},
		},
	}

	for _, tt := range tests {
		m := matcher.New(tt.rules)
		for path, want := range tt.want {
			t.Run(tt.name+"/"+path, func(t *testing.T) {
				if got := m.ShouldInclude(path); got != want {
					t.Errorf("ShouldInclude(%q) = %v, want %v", path, got, want)
				}
			})
		}
	}
}

// TestShouldTraverse_EntryRules: entry includes never widen traversal, subtree
// includes do.
func TestShouldTraverse_EntryRules(t *testing.T) {
	tests := []struct {
		name  string
		rules matcher.Rules
		dir   string
		want  bool
	}{
		{"entry include of a name does not enter other directories", gitignoreRules("*", "!.gitkeep"), "sub", false},
		{"re-included directory is entered", gitignoreRules("*", "!automation/", "!automation/*.pub"), "automation", true},
		{"directory that is not re-included is pruned", gitignoreRules("*", "!automation/", "!automation/*.pub"), "personal", false},
		{"subtree include of a name enters every directory", matcher.NewRules().AddExcludeAll().AddInclude(".gitkeep"), "sub", true},
		{"entry include after an excluded directory keeps it pruned", matcher.NewRules().AddExclude("sub/").AddIncludeEntry(".gitkeep"), "sub", false},
		{"trailing /** keeps the directory itself traversable", gitignoreRules("docs/**", "!docs/keep.md"), "docs", true},
		{"trailing /** prunes subdirectories", gitignoreRules("docs/**", "!docs/keep.md"), "docs/sub", false},
		{"subtree include after an excluded directory enters it", matcher.NewRules().AddExclude("sub/").AddInclude(".gitkeep"), "sub", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matcher.New(tt.rules).ShouldTraverse(tt.dir); got != tt.want {
				t.Errorf("ShouldTraverse(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}
