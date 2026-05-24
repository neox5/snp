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
		want    bool
	}{
		// ── Rule 1: trailing slash ────────────────────────────────────────────
		{
			name:    "trailing slash matches file under dir",
			pattern: "node_modules/",
			path:    "node_modules/package.json",
			want:    true,
		},
		{
			name:    "trailing slash matches file deep under dir",
			pattern: "node_modules/",
			path:    "node_modules/lodash/index.js",
			want:    true,
		},
		{
			name:    "trailing slash does not match dir entry itself",
			pattern: "node_modules/",
			path:    "node_modules",
			want:    false,
		},
		{
			name:    "trailing slash does not match sibling",
			pattern: "node_modules/",
			path:    "node_modules_extra/foo.js",
			want:    false,
		},

		// ── Rule 2: no slash — filename only ─────────────────────────────────
		{
			name:    "no-slash matches filename at root",
			pattern: "*.test.js",
			path:    "kernel.test.js",
			want:    true,
		},
		{
			name:    "no-slash matches filename deep",
			pattern: "*.test.js",
			path:    "packages/riker2/src/tests/kernel.test.js",
			want:    true,
		},
		{
			name:    "no-slash no match different extension",
			pattern: "*.test.js",
			path:    "packages/riker2/src/kernel.js",
			want:    false,
		},
		{
			name:    "no-slash exact filename deep",
			pattern: "README.md",
			path:    "docs/README.md",
			want:    true,
		},
		{
			name:    "no-slash question mark single char",
			pattern: "?.go",
			path:    "src/a.go",
			want:    true,
		},
		{
			name:    "no-slash question mark no match multiple chars",
			pattern: "?.go",
			path:    "src/ab.go",
			want:    false,
		},
		{
			name:    "no-slash character class match",
			pattern: "[abc].go",
			path:    "src/b.go",
			want:    true,
		},
		{
			name:    "no-slash character class no match",
			pattern: "[abc].go",
			path:    "src/d.go",
			want:    false,
		},

		// ── Rule 3: leading slash — anchored to root ─────────────────────────
		{
			name:    "leading slash matches at root",
			pattern: "/README.md",
			path:    "README.md",
			want:    true,
		},
		{
			name:    "leading slash does not match deep",
			pattern: "/README.md",
			path:    "docs/README.md",
			want:    false,
		},

		// ── Rule 4: middle slash — anchored full path ─────────────────────────
		{
			name:    "middle slash matches exact depth",
			pattern: "src/*.go",
			path:    "src/main.go",
			want:    true,
		},
		{
			name:    "middle slash does not match deeper",
			pattern: "src/*.go",
			path:    "src/internal/main.go",
			want:    false,
		},
		{
			name:    "middle slash does not match at different root",
			pattern: "src/*.go",
			path:    "pkg/src/main.go",
			want:    false,
		},

		// ── ** patterns ───────────────────────────────────────────────────────
		{
			name:    "leading doublestar matches deep",
			pattern: "**/*.test.js",
			path:    "packages/riker2/src/tests/kernel.test.js",
			want:    true,
		},
		{
			name:    "leading doublestar matches at root",
			pattern: "**/*.test.js",
			path:    "kernel.test.js",
			want:    true,
		},
		{
			name:    "leading doublestar no match wrong extension",
			pattern: "**/*.test.js",
			path:    "packages/kernel.js",
			want:    false,
		},
		{
			name:    "middle doublestar zero dirs",
			pattern: "a/**/b",
			path:    "a/b",
			want:    true,
		},
		{
			name:    "middle doublestar one dir",
			pattern: "a/**/b",
			path:    "a/x/b",
			want:    true,
		},
		{
			name:    "middle doublestar multiple dirs",
			pattern: "a/**/b",
			path:    "a/x/y/b",
			want:    true,
		},
		{
			name:    "middle doublestar no match wrong file",
			pattern: "a/**/b",
			path:    "a/x/c",
			want:    false,
		},
		{
			name:    "trailing doublestar matches file under dir",
			pattern: "internal/**",
			path:    "internal/filter/filter.go",
			want:    true,
		},
		{
			name:    "trailing doublestar no match outside dir",
			pattern: "internal/**",
			path:    "cmd/main.go",
			want:    false,
		},

		// ── snp default patterns ──────────────────────────────────────────────
		{
			name:    "snp pattern **.snp matches deep",
			pattern: "**/*.snp",
			path:    "some/path/snapshot.snp",
			want:    true,
		},
		{
			name:    "dist/ matches file under dist",
			pattern: "dist/",
			path:    "dist/snp-linux-amd64",
			want:    true,
		},
		{
			name:    "*.log matches log file deep",
			pattern: "*.log",
			path:    "logs/app.log",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matcher.Match(tt.pattern, tt.path)
			if err != nil {
				t.Fatalf("Match(%q, %q) error: %v", tt.pattern, tt.path, err)
			}
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
