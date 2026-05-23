package main

import (
	"context"
	"fmt"
	"os"
	"time"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/file"
	"github.com/neox5/snp/internal/filter"
	"github.com/neox5/snp/internal/snapshot"
)

func buildParamsFromCLI(c *cli.Command) config.Params {
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
		Includes:            c.StringSlice("include"),
		Excludes:            c.StringSlice("exclude"),
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
		fmt.Println("Default exclude patterns:")
		for _, p := range filter.DefaultPatterns {
			fmt.Println(" ", p)
		}
		fmt.Println("\nPlus patterns from .gitignore (if present).")
		return nil
	}

	p := buildParamsFromCLI(c)

	if p.VerboseLevel >= 1 {
		p.Print()
		fmt.Println()
	}

	hasPick := len(p.PickPaths) > 0
	hasTraversal := len(p.Includes) > 0 || len(p.Excludes) > 0 ||
		c.Bool("include-all") || c.Bool("exclude-all") || c.Bool("exclude-defaults")

	if hasPick && hasTraversal {
		return fmt.Errorf("--pick cannot be combined with --include, --exclude, --include-all, --exclude-all, or --exclude-defaults")
	}
	if hasPick && p.Depth != -1 {
		return fmt.Errorf("--depth cannot be combined with --pick")
	}

	if p.SaveConfig {
		cfg := config.FromParamsOnly(p, false)
		cfg.Generated = time.Now()
		if err := cfg.Save(p.SourceDir); err != nil {
			return err
		}
		if !p.Silent {
			fmt.Printf("Saved config to %s\n", config.ConfigFileName)
		}
		return nil
	}

	if p.PrintConfig {
		cfg, err := config.Load(p.SourceDir)
		if err != nil {
			return err
		}
		cfg.Serialize()
		return nil
	}

	cfg, err := config.FromParams(p, p.VerboseLevel >= 1)
	if err != nil {
		return err
	}

	if !cfg.Silent {
		fmt.Println("[command]")
		fmt.Println(cfg.BuildCommand())
	}

	absSourceDir, absOutput, err := snapshot.ValidateAndResolve(cfg)
	if err != nil {
		return err
	}

	start := time.Now()

	snap, err := snapshot.Build(ctx, cfg, absSourceDir, absOutput)
	if err != nil {
		return err
	}

	if cfg.DryRun {
		if !cfg.Silent {
			fmt.Println()
			fmt.Println("[Dry-Run]")
			for _, f := range snap.Files {
				if f.IsBinary {
					fmt.Printf("%s (binary, %s)\n", f.RelPath, file.FormatSize(f.Size))
				} else {
					fmt.Printf("%s (%d lines, %s)\n", f.RelPath, len(f.Lines), file.FormatSize(f.Size))
				}
			}
			for _, d := range snap.CollapsedDirs {
				fmt.Printf("%s (%d items, %s)\n", d.RelPath, d.ItemCount, file.FormatSize(d.Size))
			}
		}
		return nil
	}

	outFile, err := os.Create(absOutput)
	if err != nil {
		return fmt.Errorf("cannot create output file %q: %w", absOutput, err)
	}
	defer outFile.Close()

	if _, err := snap.WriteTo(outFile); err != nil {
		return err
	}

	if !cfg.Silent {
		fmt.Printf("Snapshot created: %s (%s)\n", absOutput, formatDuration(time.Since(start)))
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
