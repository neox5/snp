package clirec_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snp/internal/clirec"
)

var verbose int

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "output path"},
		&cli.IntFlag{Name: "depth", Value: -1, Sources: cli.EnvVars("SNP_DEPTH")},
		&cli.BoolFlag{Name: "no-index", Usage: "omit index"},
		&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Config: cli.BoolConfig{Count: &verbose}},
		&cli.BoolFlag{Name: "dry-run", Aliases: []string{"n"}},
		&cli.StringSliceFlag{Name: "include", Aliases: []string{"i"}},
		&cli.StringSliceFlag{Name: "exclude"},
	}
}

func run(t *testing.T, wrap bool, args ...string) (*clirec.Recorder, *cli.Command, []string, string) {
	t.Helper()
	rec := &clirec.Recorder{}
	fl := flags()
	if wrap {
		fl = clirec.WrapUrfave(rec, fl)
	}
	var help bytes.Buffer
	var pos []string
	var got *cli.Command
	app := &cli.Command{
		Name:   "snp",
		Flags:  fl,
		Writer: &help,
		Action: func(_ context.Context, c *cli.Command) error {
			pos = c.Args().Slice()
			got = c
			return nil
		},
	}
	if err := app.Run(context.Background(), append([]string{"snp"}, args...)); err != nil {
		t.Fatalf("run: %v", err)
	}
	return rec, got, pos, help.String()
}

func TestOccurrences(t *testing.T) {
	rec, c, pos, _ := run(t, true,
		"-o", "a.snp", "--depth=3", "--include", "*.go", "--no-index",
		"--exclude", "vendor", "-i=*.md", "--output", "b.snp", "dir")

	want := []clirec.Occurrence{
		{Flag: "output", Typed: "o", Value: "a.snp"},
		{Flag: "depth", Typed: "depth", Value: "3"},
		{Flag: "include", Typed: "include", Value: "*.go"},
		{Flag: "no-index", Typed: "no-index", Value: "true"},
		{Flag: "exclude", Typed: "exclude", Value: "vendor"},
		{Flag: "include", Typed: "i", Value: "*.md"},
		{Flag: "output", Typed: "output", Value: "b.snp"},
	}
	if got := rec.Occurrences(); !reflect.DeepEqual(got, want) {
		t.Errorf("occurrences:\n got %+v\nwant %+v", got, want)
	}
	if !reflect.DeepEqual(pos, []string{"dir"}) {
		t.Errorf("positional = %v, want [dir]", pos)
	}
	// urfave itself still works as before.
	if c.String("output") != "b.snp" || c.Int("depth") != 3 || !c.Bool("no-index") {
		t.Errorf("urfave values changed: %q %d %v", c.String("output"), c.Int("depth"), c.Bool("no-index"))
	}
	if !reflect.DeepEqual(c.StringSlice("include"), []string{"*.go", "*.md"}) {
		t.Errorf("include = %v", c.StringSlice("include"))
	}
}

// A bare bool must not swallow the next argument (IsBoolFlag forwarded).
func TestBareBoolDoesNotTakeNextArg(t *testing.T) {
	_, c, pos, _ := run(t, true, "--no-index", "dir")
	if !c.Bool("no-index") || !reflect.DeepEqual(pos, []string{"dir"}) {
		t.Errorf("no-index=%v positional=%v", c.Bool("no-index"), pos)
	}
}

func TestExplicitFalse(t *testing.T) {
	rec, c, _, _ := run(t, true, "--no-index=false")
	occ := rec.Occurrences()
	if len(occ) != 1 || occ[0].Value != "false" || c.Bool("no-index") {
		t.Errorf("occ=%+v value=%v", occ, c.Bool("no-index"))
	}
}

// Values from Sources are not command-line occurrences.
func TestEnvSourceNotRecorded(t *testing.T) {
	t.Setenv("SNP_DEPTH", "9")
	rec, c, _, _ := run(t, true)
	if n := len(rec.Occurrences()); n != 0 {
		t.Errorf("recorded %d occurrences from env, want 0", n)
	}
	if c.Int("depth") != 9 || !c.IsSet("depth") {
		t.Errorf("urfave: depth=%d IsSet=%v (documents that IsSet is true for env)", c.Int("depth"), c.IsSet("depth"))
	}
}

func TestNothingTyped(t *testing.T) {
	rec, _, _, _ := run(t, true)
	if n := len(rec.Occurrences()); n != 0 {
		t.Errorf("got %d occurrences, want 0", n)
	}
}

// Help output must be identical with and without the wrapper.
func TestHelpUnchanged(t *testing.T) {
	_, _, _, plain := run(t, false, "--help")
	_, _, _, wrapped := run(t, true, "--help")
	if plain == "" || plain != wrapped {
		t.Errorf("help differs:\n--- plain ---\n%s\n--- wrapped ---\n%s", plain, wrapped)
	}
}

func TestBadValueStillFails(t *testing.T) {
	rec := &clirec.Recorder{}
	app := &cli.Command{Name: "snp", Flags: clirec.WrapUrfave(rec, flags()),
		Action: func(context.Context, *cli.Command) error { return nil }}
	if err := app.Run(context.Background(), []string{"snp", "--depth", "x"}); err == nil {
		t.Error("want parse error for --depth x")
	}
	if n := len(rec.Occurrences()); n != 0 {
		t.Errorf("failed parse recorded %d occurrences", n)
	}
}

// Grouped short flags are recorded under the canonical name and keep counting.
func TestGroupedShortFlags(t *testing.T) {
	verbose = 0
	rec := &clirec.Recorder{}
	app := &cli.Command{Name: "snp", UseShortOptionHandling: true,
		Flags:  clirec.WrapUrfave(rec, flags()),
		Action: func(context.Context, *cli.Command) error { return nil }}
	if err := app.Run(context.Background(), []string{"snp", "-vvn"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	var names []string
	for _, o := range rec.Occurrences() {
		names = append(names, o.Flag)
	}
	if !reflect.DeepEqual(names, []string{"verbose", "verbose", "dry-run"}) {
		t.Errorf("recorded %v", names)
	}
	if verbose != 2 {
		t.Errorf("verbose count = %d, want 2", verbose)
	}
}
