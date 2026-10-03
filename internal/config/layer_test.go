package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/neox5/snp/internal/clirec"
)

func occ(flag, value string) clirec.Occurrence {
	return clirec.Occurrence{Flag: flag, Typed: flag, Value: value}
}

func cliLayer(t *testing.T, occs ...clirec.Occurrence) Layer {
	t.Helper()
	l, err := LayerFromOccurrences(occs)
	if err != nil {
		t.Fatalf("LayerFromOccurrences: %v", err)
	}
	return l
}

func writeFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func load(t *testing.T, dir string, cli Layer) *Resolved {
	t.Helper()
	res, err := Load(Params{SourceDir: dir}, cli)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return res
}

func TestDefaultsOnly(t *testing.T) {
	res := load(t, t.TempDir(), Layer{})
	c := res.Config
	if c.Depth != -1 || c.OutputPath != DefaultOutputPath || c.NoSummary || c.Stdout {
		t.Errorf("defaults wrong: %+v", c)
	}
	if res.Origins["depth"] != SourceDefault || res.Origins["output_path"] != SourceDefault {
		t.Errorf("origins = %v", res.Origins)
	}
}

// The file's output_path must win over the default when the CLI is silent.
func TestFileOutputWinsOverDefault(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"output_path":"out/x.snp"}`)
	res := load(t, dir, cliLayer(t))
	if res.Config.OutputPath != "out/x.snp" || res.Origins["output_path"] != SourceFile {
		t.Errorf("output = %q origin = %q", res.Config.OutputPath, res.Origins["output_path"])
	}
}

func TestCLIOutputWinsOverFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"output_path":"out/x.snp"}`)
	res := load(t, dir, cliLayer(t, occ("output", "y.snp")))
	if res.Config.OutputPath != "y.snp" || res.Origins["output_path"] != SourceCLI {
		t.Errorf("output = %q origin = %q", res.Config.OutputPath, res.Origins["output_path"])
	}
}

// File booleans survive a CLI that does not mention them; an explicit false
// on the CLI overrides a true from the file.
func TestFileBoolsSurviveAndCanBeOverridden(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"no_summary":true,"stdout":true}`)

	res := load(t, dir, cliLayer(t))
	if !res.Config.NoSummary || !res.Config.Stdout {
		t.Errorf("file bools lost: %+v", res.Config)
	}

	res = load(t, dir, cliLayer(t, occ("no-summary", "false")))
	if res.Config.NoSummary || !res.Config.Stdout {
		t.Errorf("--no-summary=false: NoSummary=%v Stdout=%v", res.Config.NoSummary, res.Config.Stdout)
	}
	if res.Origins["no_summary"] != SourceCLI || res.Origins["stdout"] != SourceFile {
		t.Errorf("origins = %v", res.Origins)
	}
}

func TestLastScalarOccurrenceWins(t *testing.T) {
	l := cliLayer(t, occ("depth", "2"), occ("depth", "5"), occ("output", "a"), occ("output", "b"),
		occ("no-index", "true"), occ("no-index", "false"))
	if *l.Depth != 5 || *l.OutputPath != "b" || *l.NoIndex {
		t.Errorf("layer = %+v", l)
	}
}

func TestMatcherFlagsAppendInOrder(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"filter_flags":[{"type":"exclude-all","value":""},{"type":"include","value":"*.go"}]}`)
	res := load(t, dir, cliLayer(t,
		occ("exclude", "*_test.go"), occ("include-all", "true"), occ("include-all", "false"), occ("include", "a,b")))
	want := []Flag{
		{FlagTypeExcludeAll, ""},
		{FlagTypeInclude, "*.go"},
		{FlagTypeExclude, "*_test.go"},
		{FlagTypeIncludeAll, ""},
		{FlagTypeInclude, "a,b"}, // include is not split on commas
	}
	if !reflect.DeepEqual(res.Config.MatcherFlags, want) {
		t.Errorf("flags:\n got %v\nwant %v", res.Config.MatcherFlags, want)
	}
	if res.Origins["filter_flags"] != "file+cli" {
		t.Errorf("origin = %q", res.Origins["filter_flags"])
	}
}

func TestListsAreUnions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"pick_paths":["a","b"],"force_text_patterns":["*.x"],"force_binary_patterns":["*.y"]}`)
	res := load(t, dir, cliLayer(t, occ("pick", "b,c"), occ("force-text", "*.x"), occ("force-text", "*.z")))
	c := res.Config
	if !reflect.DeepEqual(c.PickPaths, []string{"a", "b", "c"}) {
		t.Errorf("pick = %v", c.PickPaths)
	}
	if !reflect.DeepEqual(c.ForceTextPatterns, []string{"*.x", "*.z"}) {
		t.Errorf("force-text = %v", c.ForceTextPatterns)
	}
	if !reflect.DeepEqual(c.ForceBinaryPatterns, []string{"*.y"}) {
		t.Errorf("force-binary = %v", c.ForceBinaryPatterns)
	}
}

func TestOnlyExpansion(t *testing.T) {
	l := cliLayer(t, occ("only-content", "true"))
	if *l.NoSummary != true || *l.NoIndex != true || *l.NoGitLog != true || *l.NoContent != false {
		t.Errorf("only-content: %v %v %v %v", *l.NoSummary, *l.NoIndex, *l.NoGitLog, *l.NoContent)
	}

	l = cliLayer(t, occ("only-content", "false"))
	if l.NoSummary != nil || l.NoContent != nil {
		t.Errorf("only-content=false stated something: %+v", l)
	}

	// only-content overrides a file's no_content, and an explicit
	// --no-content on the CLI is kept.
	dir := t.TempDir()
	writeFile(t, dir, `{"no_content":true}`)
	if res := load(t, dir, cliLayer(t, occ("only-content", "true"))); res.Config.NoContent {
		t.Error("only-content did not override file no_content")
	}
	l = cliLayer(t, occ("only-content", "true"), occ("no-content", "true"))
	if !*l.NoContent {
		t.Error("explicit --no-content dropped")
	}
}

func TestIncludeValueKeepsEqualsSign(t *testing.T) {
	l := cliLayer(t, occ("include", "=weird"), occ("include", "*.md"))
	if l.MatcherFlags[0].Value != "=weird" || l.MatcherFlags[1].Value != "*.md" {
		t.Errorf("flags = %v", l.MatcherFlags)
	}
}

func TestInvalidValuesFail(t *testing.T) {
	for _, o := range []clirec.Occurrence{occ("depth", "x"), occ("no-index", "maybe"), occ("only-index", "z")} {
		if _, err := LayerFromOccurrences([]clirec.Occurrence{o}); err == nil {
			t.Errorf("%s=%s: want error", o.Flag, o.Value)
		}
	}
}

func TestUnknownFlagsIgnored(t *testing.T) {
	l := cliLayer(t, occ("dry-run", "true"), occ("verbose", "true"), occ("no-config", "true"))
	if !reflect.DeepEqual(l, Layer{}) {
		t.Errorf("layer = %+v", l)
	}
}

func TestNoConfigIgnoresFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"output_path":"x","no_index":true}`)
	res, err := Load(Params{SourceDir: dir, NoConfig: true}, cliLayer(t))
	if err != nil {
		t.Fatal(err)
	}
	if res.Config.OutputPath != DefaultOutputPath || res.Config.NoIndex {
		t.Errorf("file not ignored: %+v", res.Config)
	}
}

func TestSaveOmitsDefaultsAndRuntimeFields(t *testing.T) {
	dir := t.TempDir()
	res, err := Load(Params{SourceDir: dir, DryRun: true}, cliLayer(t, occ("exclude", "x"), occ("silent", "true")))
	if err != nil {
		t.Fatal(err)
	}
	if err = Save(dir, res.Persist); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ConfigFileName))
	s := string(data)
	for _, bad := range []string{"depth", "output_path", "generated", "source_dir", "dry_run", "no_summary"} {
		if strings.Contains(s, bad) {
			t.Errorf("saved file contains %q:\n%s", bad, s)
		}
	}
	for _, good := range []string{"filter_flags", `"silent": true`} {
		if !strings.Contains(s, good) {
			t.Errorf("saved file lacks %q:\n%s", good, s)
		}
	}
}

// Saving and loading again reproduces the same effective config.
func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	first := load(t, dir, cliLayer(t, occ("depth", "2"), occ("include", "*.go"), occ("no-index", "true"), occ("pick", "a")))
	if err := Save(dir, first.Persist); err != nil {
		t.Fatal(err)
	}
	second := load(t, dir, cliLayer(t))
	a, b := *first.Config, *second.Config
	a.Generated, b.Generated = time.Time{}, time.Time{}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("round trip differs:\n%+v\n%+v", a, b)
	}
}

// Files written by earlier versions carry every key, including runtime ones.
func TestLegacyFileLoads(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{"generated":"2026-01-01T00:00:00Z","source_dir":".","depth":-1,"filter_flags":null,
"pick_paths":null,"force_text_patterns":null,"force_binary_patterns":null,"output_path":"snapshot.snp",
"no_summary":false,"no_index":false,"no_git_log":false,"no_content":false,"stdout":false,"dry_run":false,"silent":false}`)
	res := load(t, dir, cliLayer(t, occ("no-index", "true")))
	if !res.Config.NoIndex || res.Config.Depth != -1 || res.Config.DryRun {
		t.Errorf("config = %+v", res.Config)
	}
}

func TestInvalidFileFails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, `{not json`)
	if _, err := Load(Params{SourceDir: dir}, Layer{}); err == nil {
		t.Error("want error for invalid JSON")
	}
}
