package filter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neox5/snp/internal/filter"
)

func TestShouldInclude_Defaults(t *testing.T) {
	tmpDir := t.TempDir()

	gitignoreContent := "*.log\nsecrets/\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
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
			reason: "default patterns exclude node_modules/",
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
			reason: "no rule matches, default is include",
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
			m, err := filter.New(tmpDir, false, nil)
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
			name: "last rule wins — include after exclude",
			rules: []filter.Rule{
				{Pattern: "**/*.go", Exclude: false},
				{Pattern: "**/*_test.go", Exclude: true},
				{Pattern: "internal/auth/auth_test.go", Exclude: false},
			},
			path:   "internal/auth/auth_test.go",
			want:   true,
			reason: "last matching rule is include",
		},
		{
			name: "last rule wins — exclude after include",
			rules: []filter.Rule{
				{Pattern: "**/*.go", Exclude: false},
				{Pattern: "**/*_test.go", Exclude: true},
			},
			path:   "internal/auth/auth_test.go",
			want:   false,
			reason: "last matching rule is exclude",
		},
		{
			name: "no match defaults to include",
			rules: []filter.Rule{
				{Pattern: "**/*.go", Exclude: false},
			},
			path:   "README.md",
			want:   true,
			reason: "no rule matches, default is include",
		},
		{
			name: "single exclude rule",
			rules: []filter.Rule{
				{Pattern: "**/*.log", Exclude: true},
			},
			path:   "app.log",
			want:   false,
			reason: "matched by exclude rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := filter.New(tmpDir, true, tt.rules)
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

func TestShouldInclude_NoDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.log"), 0o644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	m, err := filter.New(tmpDir, true, nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tests := []struct {
		path string
		want bool
	}{
		{"node_modules/package.json", true},
		{"app.log", true},
		{"src/main.go", true},
	}

	for _, tt := range tests {
		got := m.ShouldInclude(tt.path)
		if got != tt.want {
			t.Errorf("ShouldInclude(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestShouldInclude_NilMatcher(t *testing.T) {
	var m *filter.Matcher
	if !m.ShouldInclude("any/path.go") {
		t.Error("nil Matcher should include all paths")
	}
}

func TestNew_InvalidDirectory(t *testing.T) {
	_, err := filter.New("/nonexistent/directory/12345", false, nil)
	if err == nil {
		t.Error("New should fail for nonexistent directory")
	}
}
