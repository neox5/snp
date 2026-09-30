package matcher_test

import (
	"strings"
	"testing"

	"github.com/neox5/snp/internal/matcher"
)

func TestRulesPrint(t *testing.T) {
	rules := matcher.NewRules().
		AddExclude(".git/").
		AddExcludeEntry("*").
		AddIncludeEntry("automation/").
		AddExcludeAll().
		AddInclude("**/*.go").
		AddInclude("README.md")

	want := strings.Join([]string{
		"  last match wins",
		"  rule  pattern      on a matching directory",
		"  [-]   .git/        with contents",
		"  [-]   *            itself only              (.gitignore)",
		"  [+]   automation/  itself only              (.gitignore)",
		"  [-]   ALL",
		"  [+]   **/*.go      with contents",
		"  [+]   README.md    with contents",
		"",
	}, "\n")

	var sb strings.Builder
	rules.Print(&sb, "  ")
	if got := sb.String(); got != want {
		t.Errorf("Print output mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRulesPrint_Empty(t *testing.T) {
	var sb strings.Builder
	matcher.NewRules().Print(&sb)

	want := "last match wins\nrule  pattern  on a matching directory\n"
	if got := sb.String(); got != want {
		t.Errorf("Print output mismatch\n got: %q\nwant: %q", got, want)
	}
}
