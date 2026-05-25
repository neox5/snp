package matcher_test

import (
	"testing"

	"github.com/neox5/snp/internal/matcher"
)

func TestShouldInclude_NilMatcher(t *testing.T) {
	var m *matcher.Matcher
	if !m.ShouldInclude("any/path.go", false) {
		t.Error("nil Matcher should include all paths")
	}
}

func TestShouldInclude_OrderedRules(t *testing.T) {
	tests := []struct {
		name   string
		rules  matcher.Rules
		path   string
		isDir  bool
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
			isDir:  false,
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
			isDir:  false,
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
			isDir:  false,
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
			isDir:  false,
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
			isDir:  false,
			want:   false,
			reason: "exclude wins as last matching rule",
		},
		{
			name: "include-all baseline — everything included",
			rules: []matcher.Rule{
				{Type: matcher.RuleIncludeAll},
			},
			path:   "node_modules/package.json",
			isDir:  false,
			want:   true,
			reason: "include-all with no further rules includes everything",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := matcher.New(tt.rules)
			got := m.ShouldInclude(tt.path, tt.isDir)
			if got != tt.want {
				t.Errorf("ShouldInclude(%q, %v) = %v, want %v\nReason: %s", tt.path, tt.isDir, got, tt.want, tt.reason)
			}
		})
	}
}
