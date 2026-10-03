package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/neox5/snp/internal/matcher"
)

func writeGitignore(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestBuildExcludeDefaultRules_NoGitignore(t *testing.T) {
	rules := buildExcludeDefaultRules(t.TempDir())

	if len(rules) != len(DefaultExcludePatterns) {
		t.Fatalf("got %d rules, want %d default rules", len(rules), len(DefaultExcludePatterns))
	}
	for i, p := range DefaultExcludePatterns {
		want := matcher.Rule{Type: matcher.RuleExclude, Pattern: p, Scope: matcher.ScopeSubtree, Source: SourceDefault}
		if rules[i] != want {
			t.Errorf("rule %d = %+v, want %+v", i, rules[i], want)
		}
	}
}

func TestBuildExcludeDefaultRules_Gitignore(t *testing.T) {
	// comment, blank line, CRLF, negation, lone "!" and a pattern with a glob
	dir := writeGitignore(t, "# comment\n\n*\r\n!automation/\n!\n!.gitkeep\nbuild/*\n")

	rules := buildExcludeDefaultRules(dir)
	n := len(DefaultExcludePatterns)

	want := matcher.Rules{
		{Type: matcher.RuleExclude, Pattern: "*", Scope: matcher.ScopeEntry, Source: SourceGitignore},
		{Type: matcher.RuleInclude, Pattern: "automation/", Scope: matcher.ScopeEntry, Source: SourceGitignore},
		{Type: matcher.RuleInclude, Pattern: ".gitkeep", Scope: matcher.ScopeEntry, Source: SourceGitignore},
		{Type: matcher.RuleExclude, Pattern: "build/*", Scope: matcher.ScopeEntry, Source: SourceGitignore},
	}
	if len(rules) != n+len(want) {
		t.Fatalf("got %d rules, want %d", len(rules), n+len(want))
	}
	if !reflect.DeepEqual(rules[n:], want) {
		t.Errorf(".gitignore rules = %+v, want %+v", rules[n:], want)
	}
	for i, p := range DefaultExcludePatterns {
		if rules[i].Pattern != p || rules[i].Scope != matcher.ScopeSubtree {
			t.Errorf("default rule %d = %+v, want subtree exclude %q first", i, rules[i], p)
		}
	}
}

func TestDefaultExcludes_Depth(t *testing.T) {
	m := matcher.New(buildExcludeDefaultRules(t.TempDir()))

	excluded := []string{
		"node_modules/x.js",
		"packages/a/node_modules/x.js",
		"src/__pycache__/x.pyc",
		"a/b/.venv/bin/python",
		"sub/.git",
		"dist/out.bin",
		"build/out.bin",
		"app.log",
	}
	for _, p := range excluded {
		if m.ShouldInclude(p) {
			t.Errorf("ShouldInclude(%q) = true, want false", p)
		}
	}

	included := []string{
		"internal/build/build.go",
		"cmd/dist/main.go",
		"pkg/vendor/v.go",
		"src/main.go",
	}
	for _, p := range included {
		if !m.ShouldInclude(p) {
			t.Errorf("ShouldInclude(%q) = false, want true", p)
		}
	}

	for dir, want := range map[string]bool{
		"packages/a/node_modules": false,
		"packages/a":              true,
		"internal/build":          true,
		"build":                   false,
	} {
		if got := m.ShouldTraverse(dir); got != want {
			t.Errorf("ShouldTraverse(%q) = %v, want %v", dir, got, want)
		}
	}
}

func TestBuildExcludeDefaultRules_DoesNotChangeDefaults(t *testing.T) {
	before := append([]string(nil), DefaultExcludePatterns...)

	buildExcludeDefaultRules(writeGitignore(t, "a\nb\n"))
	buildExcludeDefaultRules(writeGitignore(t, "c\n"))

	if !reflect.DeepEqual(DefaultExcludePatterns, before) {
		t.Errorf("DefaultExcludePatterns changed: %v, want %v", DefaultExcludePatterns, before)
	}
}
