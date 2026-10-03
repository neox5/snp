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

func TestRulesPrint_Sources(t *testing.T) {
	rules := matcher.NewRules().AddExclude(".git/").WithSource("default").
		AddRules(matcher.NewRules().AddExcludeEntry("*").WithSource(".gitignore")).
		AddRules(matcher.NewRules().AddExcludeAll().WithSource("file")).
		AddRules(matcher.NewRules().AddInclude("**/*.go").WithSource("file")).
		AddRules(matcher.NewRules().AddInclude("README.md").WithSource("cli"))

	want := strings.Join([]string{
		"  last match wins",
		"  rule  pattern    source      on a matching directory",
		"  [-]   .git/      default     with contents",
		"  [-]   *          .gitignore  itself only",
		"  [-]   ALL        file",
		"  [+]   **/*.go    file        with contents",
		"  [+]   README.md  cli         with contents",
		"",
	}, "\n")

	var sb strings.Builder
	rules.Print(&sb, "  ")
	if got := sb.String(); got != want {
		t.Errorf("Print output mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRulesWithSource_KeepsExistingLabels(t *testing.T) {
	rules := matcher.NewRules().AddExclude("a").WithSource("one").AddExclude("b").WithSource("two")
	if rules[0].Source != "one" || rules[1].Source != "two" {
		t.Errorf("sources = %q, %q", rules[0].Source, rules[1].Source)
	}
	orig := matcher.NewRules().AddExclude("a")
	_ = orig.WithSource("x")
	if orig[0].Source != "" {
		t.Error("WithSource modified the receiver")
	}
}
