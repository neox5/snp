package ignore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neox5/snp/internal/ignore"
)

func TestShouldInclude_Precedence(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a .gitignore file
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	gitignoreContent := "*.log\nsecrets/\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	tests := []struct {
		name            string
		excludePatterns []string
		includePatterns []string
		path            string
		want            bool
		reason          string
	}{
		{
			name:            "default pattern excludes node_modules",
			excludePatterns: nil,
			includePatterns: nil,
			path:            "node_modules/package.json",
			want:            false,
			reason:          "default patterns should exclude node_modules/",
		},
		{
			name:            "gitignore excludes log files",
			excludePatterns: nil,
			includePatterns: nil,
			path:            "app.log",
			want:            false,
			reason:          ".gitignore should exclude *.log",
		},
		{
			name:            "include overrides gitignore",
			excludePatterns: nil,
			includePatterns: []string{"*.log"},
			path:            "app.log",
			want:            true,
			reason:          "--include should override .gitignore exclusion",
		},
		{
			name:            "include overrides default patterns",
			excludePatterns: nil,
			includePatterns: []string{"node_modules/**"},
			path:            "node_modules/package.json",
			want:            true,
			reason:          "--include should override default pattern exclusion",
		},
		{
			name:            "exclude overrides include",
			excludePatterns: []string{"secret.log"},
			includePatterns: []string{"*.log"},
			path:            "secret.log",
			want:            false,
			reason:          "--exclude should override --include",
		},
		{
			name:            "exclude is final even with broad include",
			excludePatterns: []string{"**/secret/**"},
			includePatterns: []string{"**/*.go"},
			path:            "src/secret/main.go",
			want:            false,
			reason:          "--exclude should be final decision",
		},
		{
			name:            "normal file without any patterns",
			excludePatterns: nil,
			includePatterns: nil,
			path:            "src/main.go",
			want:            true,
			reason:          "normal files should be included by default",
		},
		{
			name:            "gitignore directory exclusion",
			excludePatterns: nil,
			includePatterns: nil,
			path:            "secrets/password.txt",
			want:            false,
			reason:          ".gitignore should exclude secrets/ directory",
		},
		{
			name:            "include rescues from gitignore directory",
			excludePatterns: nil,
			includePatterns: []string{"secrets/config.txt"},
			path:            "secrets/config.txt",
			want:            true,
			reason:          "--include should rescue specific file from gitignored directory",
		},
		{
			name:            "exclude blocks rescue attempt",
			excludePatterns: []string{"secrets/config.txt"},
			includePatterns: []string{"secrets/config.txt"},
			path:            "secrets/config.txt",
			want:            false,
			reason:          "--exclude should block even explicit --include",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchers, err := ignore.NewMatchers(tmpDir, tt.excludePatterns, tt.includePatterns)
			if err != nil {
				t.Fatalf("NewMatchers failed: %v", err)
			}

			got := matchers.ShouldInclude(tt.path)
			if got != tt.want {
				t.Errorf("ShouldInclude(%q) = %v, want %v\nReason: %s", tt.path, got, tt.want, tt.reason)
			}
		})
	}
}

func TestShouldInclude_NoGitignore(t *testing.T) {
	// Test with directory that has no .gitignore
	tmpDir := t.TempDir()

	tests := []struct {
		name            string
		excludePatterns []string
		includePatterns []string
		path            string
		want            bool
	}{
		{
			name:            "default excludes still work",
			excludePatterns: nil,
			includePatterns: nil,
			path:            "node_modules/test.js",
			want:            false,
		},
		{
			name:            "normal file included",
			excludePatterns: nil,
			includePatterns: nil,
			path:            "src/main.go",
			want:            true,
		},
		{
			name:            "CLI exclude works",
			excludePatterns: []string{"*.tmp"},
			includePatterns: nil,
			path:            "temp.tmp",
			want:            false,
		},
		{
			name:            "CLI include works",
			excludePatterns: nil,
			includePatterns: []string{"special/"},
			path:            "special/file.txt",
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchers, err := ignore.NewMatchers(tmpDir, tt.excludePatterns, tt.includePatterns)
			if err != nil {
				t.Fatalf("NewMatchers failed: %v", err)
			}

			got := matchers.ShouldInclude(tt.path)
			if got != tt.want {
				t.Errorf("ShouldInclude(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestShouldInclude_NilMatchers(t *testing.T) {
	var matchers *ignore.Matchers
	if !matchers.ShouldInclude("any/path.go") {
		t.Error("nil Matchers should include all paths")
	}
}

func TestNewMatchers_InvalidDirectory(t *testing.T) {
	_, err := ignore.NewMatchers("/nonexistent/directory/12345", nil, nil)
	if err == nil {
		t.Error("NewMatchers should fail for nonexistent directory")
	}
}
