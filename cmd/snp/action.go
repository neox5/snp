package main

import (
	"bufio"
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
		SourceDir:    sourceDir,
		NoConfig:     c.Bool("no-config"),
		SaveConfig:   c.Bool("save-config"),
		ShowConfig:   c.Bool("show-config"),
		DryRun:       c.Bool("dry-run"),
		VerboseLevel: verboseCount,
	}
}

func runAction(ctx context.Context, c *cli.Command) error {
	if c.Bool("show-defaults") {
		config.PrintDefaultExcludes()
		return nil
	}

	p := cliToParams(c)
	cliLayer, err := config.LayerFromOccurrences(recorder.Occurrences())
	if err != nil {
		return err
	}

	res, err := config.Load(p, cliLayer)
	if err != nil {
		return err
	}
	cfg := res.Config

	if err = cfg.Validate(); err != nil {
		return err
	}

	if p.SaveConfig {
		if err = config.Save(p.SourceDir, res.Persist); err != nil {
			return err
		}
		if !cfg.Silent {
			fmt.Printf("Saved config to %s\n", config.ConfigFileName)
		}
		return nil
	}

	// status writes to stderr when stdout carries snapshot content
	status := os.Stdout
	if cfg.Stdout {
		status = os.Stderr
	}

	printHeader := func(first bool, h string) {
		if !first {
			fmt.Fprintln(status)
		}
		fmt.Fprintf(status, "[%s]\n", h)
	}

	if p.ShowConfig {
		printHeader(true, "snp config")
		cfg.Print(res.Origins, "  ")
		printHeader(false, "command")
		fmt.Fprintln(status, cfg.BuildCommand())
		return nil
	}

	if !cfg.Silent {
		printHeader(true, "command")
		fmt.Fprintln(status, cfg.BuildCommand())
	}

	if p.VerboseLevel >= 2 {
		printHeader(false, "Matcher Rules")
		cfg.BuildMatcherRules().Print(status, "  ")
	}

	snap := snapshot.New(cfg)

	if err = snap.Collect(ctx); err != nil {
		return err
	}

	if p.DryRun {
		printHeader(false, "Dry Run")
		snap.PrintRawEntries(status)
		return nil
	}

	start := time.Now()

	if err = snap.Build(); err != nil {
		return err
	}

	if cfg.Stdout {
		bw := bufio.NewWriter(os.Stdout)
		if _, err = snap.WriteTo(bw); err != nil {
			return err
		}
		return bw.Flush()
	}

	if err = snap.Write(); err != nil {
		return err
	}

	if !cfg.Silent {
		fmt.Fprintf(status, "Snapshot written: %s (%s)\n", cfg.OutputPath, formatDuration(time.Since(start)))
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
