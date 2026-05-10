package filter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neox5/snp/internal/filter"
)

func TestShouldInclude_ImplicitDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.log\nsecrets/\n"), 0o644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	// Implicit default: --include-all --exclude-defaults
	rules := []filter.Rule{
		{Type: filter.RuleIncludeAll},
		{Type: filter.RuleExcludeDefaults},
	}

	tests := []struct {
		name   string
		path   string
		want   bool
		reason string
	}{
		{
			name:   "default excludes node_modules",
			path:   "node_modules/package.json",
			want:   false,
			reason: "exclude-defaults covers node_modules/",
		},
		{
			name:   "gitignore excludes log files",
			path:   "app.log",
			want:   false,
			reason: ".gitignore excludes *.log",
		},
		{
			name:   "normal file included",
			path:   "src/main.go",
			want:   true,
			reason: "include-all wins, not excluded by defaults",
		},
		{
			name:   "gitignore directory exclusion",
			path:   "secrets/password.txt",
			want:   false,
			reason: ".gitignore excludes secrets/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := filter.New(tmpDir, rules)
			if err != nil {
				t.Fatalf("New failed: %v", err)
			}
			got := m.ShouldInclude(tt.path)
			if got != tt.want {
				t.Errorf("ShouldInclude(%q) = %v, want %v\nReason: %s", tt.path, got, tt.want, tt.reason)
			}
		})
	}
}

func TestShouldInclude_OrderedRules(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name   string
		rules  []filter.Rule
		path   string
		want   bool
		reason string
	}{
		{
			name: "exclude-all then include Go — only Go files",
			rules: []filter.Rule{
				{Type: filter.RuleExcludeAll},
				{Type: filter.RuleInclude, Pattern: "**/*.go"},
			},
			path:   "src/main.go",
			want:   true,
			reason: "include **/*.go wins after exclude-all",
		},
		{
			name: "exclude-all then include Go — non-Go excluded",
			rules: []filter.Rule{
				{Type: filter.RuleExcludeAll},
				{Type: filter.RuleInclude, Pattern: "**/*.go"},
			},
			path:   "README.md",
			want:   false,
			reason: "exclude-all wins, no include rule matches",
		},
		{
			name: "last rule wins — rescue from exclude",
			rules: []filter.Rule{
				{Type: filter.RuleIncludeAll},
				{Type: filter.RuleExcludeDefaults},
				{Type: filter.RuleExclude, Pattern: "**/*_test.go"},
				{Type: filter.RuleInclude, Pattern: "internal/auth/auth_test.go"},
			},
			path:   "internal/auth/auth_test.go",
			want:   true,
			reason: "last include rule rescues specific test file",
		},
		{
			name: "last rule wins — exclude after include",
			rules: []filter.Rule{
				{Type: filter.RuleIncludeAll},
				{Type: filter.RuleExclude, Pattern: "**/*_test.go"},
			},
			path:   "internal/auth/auth_test.go",
			want:   false,
			reason: "exclude wins as last matching rule",
		},
		{
			name: "include-all baseline — everything included",
			rules: []filter.Rule{
				{Type: filter.RuleIncludeAll},
			},
			path:   "node_modules/package.json",
			want:   true,
			reason: "include-all with no further rules includes everything",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := filter.New(tmpDir, tt.rules)
			if err != nil {
				t.Fatalf("New failed: %v", err)
			}
			got := m.ShouldInclude(tt.path)
			if got != tt.want {
				t.Errorf("ShouldInclude(%q) = %v, want %v\nReason: %s", tt.path, got, tt.want, tt.reason)
			}
		})
	}
}

func TestShouldInclude_NilMatcher(t *testing.T) {
	var m *filter.Matcher
	if !m.ShouldInclude("any/path.go") {
		t.Error("nil Matcher should include all paths")
	}
}

func TestNew_InvalidDirectory(t *testing.T) {
	_, err := filter.New("/nonexistent/directory/12345", nil)
	if err == nil {
		t.Error("New should fail for nonexistent directory")
	}
}
