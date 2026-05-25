package main

import (
	"context"
	"fmt"
	"os"
	"time"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/snapshot"
)

func cliToParams(c *cli.Command) config.Params {
	sourceDir := "."
	if c.NArg() > 0 {
		sourceDir = c.Args().First()
	}

	return config.Params{
		SourceDir:           sourceDir,
		Args:                os.Args[1:], // included for determine order of include/exclude flags
		NoConfig:            c.Bool("no-config"),
		SaveConfig:          c.Bool("save-config"),
		PrintConfig:         c.Bool("print-config"),
		DryRun:              c.Bool("dry-run"),
		Depth:               c.Int("depth"),
		PickPaths:           c.StringSlice("pick"),
		ForceTextPatterns:   c.StringSlice("force-text"),
		ForceBinaryPatterns: c.StringSlice("force-binary"),
		OutputPath:          c.String("output"),
		NoSummary:           c.Bool("no-summary"),
		NoIndex:             c.Bool("no-index"),
		NoGitLog:            c.Bool("no-git-log"),
		NoContent:           c.Bool("no-content"),
		Silent:              c.Bool("silent"),
		VerboseLevel:        verboseCount,
	}
}

func runAction(ctx context.Context, c *cli.Command) error {
	if c.Bool("show-defaults") {
		config.PrintDefaultExcludes()
		return nil
	}

	// ### process cli parameters and convert to config
	p := cliToParams(c)
	cfg, err := config.LoadConfig(p)
	if err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return err
	}

	// ### save configuration feature
	if p.SaveConfig {
		if err = cfg.Save(p.SourceDir); err != nil {
			return err
		}
		if !p.Silent {
			fmt.Printf("Saved config to %s\n", config.ConfigFileName)
		}
		return nil
	}

	// printHeader is used to give a uniform way of printing a header block.
	// it also adds a separation line if first != true
	printHeader := func(first bool, h string) {
		if !first {
			fmt.Println()
		}
		fmt.Printf("[%s]\n", h)
	}

	// ### print configuration feature
	if p.PrintConfig {
		printHeader(true, "snp config")
		cfg.Print()
		return nil
	}

	if !cfg.Silent {
		printHeader(true, "command")
		fmt.Println(cfg.BuildCommand())
	}

	if p.VerboseLevel >= 2 {
		printHeader(false, "Matcher Rules")
		cfg.BuildMatcherRules().Print("  ")
	}

	snap := snapshot.New(cfg)

	if err = snap.Collect(ctx); err != nil {
		return err
	}

	// ### dry run feature
	if p.DryRun {
		printHeader(false, "Dry Run")
		snap.PrintRawEntries()
		return nil
	}

	start := time.Now()

	// ### build Snapshot
	if err = snap.Build(); err != nil {
		return err
	}

	if err = snap.Write(); err != nil {
		return err
	}

	if !cfg.Silent {
		fmt.Printf("Snapshot written: %s (%s)\n", cfg.OutputPath, formatDuration(time.Since(start)))
	}

	return nil
}

func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
