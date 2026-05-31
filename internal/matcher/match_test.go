package matcher_test

import (
	"testing"

	"github.com/neox5/snp/internal/matcher"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		isDir   bool
		want    bool
	}{
		// ── Rule 1: trailing slash ────────────────────────────────────────────
		{
			name:    "trailing slash matches file under dir",
			pattern: "node_modules/",
			path:    "node_modules/package.json",
			isDir:   false,
			want:    true,
		},
		{
			name:    "trailing slash matches dir itself",
			pattern: "node_modules/",
			path:    "node_modules",
			isDir:   true,
			want:    true,
		},
		{
			name:    "trailing slash does not match dir as file",
			pattern: "node_modules/",
			path:    "node_modules",
			isDir:   false,
			want:    false,
		},
		{
			name:    "trailing slash does not match sibling",
			pattern: "node_modules/",
			path:    "node_modules_extra/pkg.json",
			isDir:   false,
			want:    false,
		},

		// ── Rule 2: no slash ──────────────────────────────────────────────────
		{
			name:    "no-slash matches filename at root",
			pattern: "README.md",
			path:    "README.md",
			isDir:   false,
			want:    true,
		},
		{
			name:    "no-slash matches filename deep",
			pattern: "README.md",
			path:    "docs/README.md",
			isDir:   false,
			want:    true,
		},
		{
			name:    "no-slash matches dir segment deep",
			pattern: "internal",
			path:    "cmd/internal/main.go",
			isDir:   false,
			want:    true,
		},
		{
			name:    "no-slash wildcard matches deep",
			pattern: "*.test.js",
			path:    "packages/riker2/src/tests/kernel.test.js",
			isDir:   false,
			want:    true,
		},
		{
			name:    "no-slash no match",
			pattern: "README.md",
			path:    "docs/CONTRIBUTING.md",
			isDir:   false,
			want:    false,
		},
		{
			name:    "no-slash character class match",
			pattern: "[abc].go",
			path:    "src/a.go",
			isDir:   false,
			want:    true,
		},
		{
			name:    "no-slash character class no match",
			pattern: "[abc].go",
			path:    "src/d.go",
			isDir:   false,
			want:    false,
		},

		// ── Rule 3: leading slash — anchored to root ─────────────────────────
		{
			name:    "leading slash matches at root",
			pattern: "/README.md",
			path:    "README.md",
			isDir:   false,
			want:    true,
		},
		{
			name:    "leading slash does not match deep",
			pattern: "/README.md",
			path:    "docs/README.md",
			isDir:   false,
			want:    false,
		},

		// ── Rule 4: middle slash — anchored full path ─────────────────────────
		{
			name:    "middle slash matches exact depth",
			pattern: "src/*.go",
			path:    "src/main.go",
			isDir:   false,
			want:    true,
		},
		{
			name:    "middle slash does not match deeper",
			pattern: "src/*.go",
			path:    "src/internal/main.go",
			isDir:   false,
			want:    false,
		},
		{
			name:    "middle slash does not match at different root",
			pattern: "src/*.go",
			path:    "pkg/src/main.go",
			isDir:   false,
			want:    false,
		},

		// ── anchored pattern — prefix matching ────────────────────────────────
		// A pattern matching a path also matches everything beneath it.
		// "internal/config" includes the directory and all files inside it.
		{
			name:    "anchored pattern matches exact path",
			pattern: "internal/config",
			path:    "internal/config",
			isDir:   false,
			want:    true,
		},
		{
			name:    "anchored pattern matches file directly inside",
			pattern: "internal/config",
			path:    "internal/config/config.go",
			isDir:   false,
			want:    true,
		},
		{
			name:    "anchored pattern matches file deep inside",
			pattern: "internal/config",
			path:    "internal/config/sub/deep.go",
			isDir:   false,
			want:    true,
		},
		{
			name:    "anchored pattern does not match sibling dir",
			pattern: "internal/config",
			path:    "internal/config_other/file.go",
			isDir:   false,
			want:    false,
		},
		{
			name:    "anchored pattern does not match parent",
			pattern: "internal/config",
			path:    "internal/other.go",
			isDir:   false,
			want:    false,
		},
		{
			name:    "anchored pattern subdir matches its file",
			pattern: "internal/config",
			path:    "internal/config/sub",
			isDir:   true,
			want:    true,
		},

		// ── ** patterns ───────────────────────────────────────────────────────
		{
			name:    "leading doublestar matches deep",
			pattern: "**/*.test.js",
			path:    "packages/riker2/src/tests/kernel.test.js",
			isDir:   false,
			want:    true,
		},
		{
			name:    "leading doublestar matches at root",
			pattern: "**/*.test.js",
			path:    "kernel.test.js",
			isDir:   false,
			want:    true,
		},
		{
			name:    "leading doublestar no match wrong extension",
			pattern: "**/*.test.js",
			path:    "packages/kernel.js",
			isDir:   false,
			want:    false,
		},
		{
			name:    "middle doublestar zero dirs",
			pattern: "a/**/b",
			path:    "a/b",
			isDir:   false,
			want:    true,
		},
		{
			name:    "middle doublestar one dir",
			pattern: "a/**/b",
			path:    "a/x/b",
			isDir:   false,
			want:    true,
		},
		{
			name:    "middle doublestar multiple dirs",
			pattern: "a/**/b",
			path:    "a/x/y/b",
			isDir:   false,
			want:    true,
		},
		{
			name:    "middle doublestar no match wrong file",
			pattern: "a/**/b",
			path:    "a/x/c",
			isDir:   false,
			want:    false,
		},
		{
			name:    "trailing doublestar matches file under dir",
			pattern: "internal/**",
			path:    "internal/filter/filter.go",
			isDir:   false,
			want:    true,
		},
		{
			name:    "trailing doublestar no match outside dir",
			pattern: "internal/**",
			path:    "cmd/main.go",
			isDir:   false,
			want:    false,
		},

		// ── snp default patterns ──────────────────────────────────────────────
		{
			name:    "snp pattern **.snp matches deep",
			pattern: "**/*.snp",
			path:    "some/path/snapshot.snp",
			isDir:   false,
			want:    true,
		},
		{
			name:    "dist/ matches file under dist",
			pattern: "dist/",
			path:    "dist/snp-linux-amd64",
			isDir:   false,
			want:    true,
		},
		{
			name:    "*.log matches log file deep",
			pattern: "*.log",
			path:    "logs/app.log",
			isDir:   false,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matcher.Match(tt.pattern, tt.path, tt.isDir)
			if err != nil {
				t.Fatalf("Match(%q, %q, %v) error: %v", tt.pattern, tt.path, tt.isDir, err)
			}
			if got != tt.want {
				t.Errorf("Match(%q, %q, %v) = %v, want %v", tt.pattern, tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}
