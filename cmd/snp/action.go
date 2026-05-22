package main

import (
	"context"
	"fmt"
	"os"
	"time"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snp/internal/filter"
	"github.com/neox5/snp/internal/snapshot"
)

func runAction(ctx context.Context, c *cli.Command) error {
	if c.Bool("show-defaults") {
		fmt.Println("Default exclude patterns:")
		for _, p := range filter.DefaultPatterns {
			fmt.Println(" ", p)
		}
		fmt.Println("\nPlus patterns from .gitignore (if present).")
		return nil
	}

	sourceDir := "."
	if c.NArg() > 0 {
		sourceDir = c.Args().First()
	}

	silent := c.Bool("silent")
	pickPaths := c.StringSlice("pick")
	includes := c.StringSlice("include")
	excludes := c.StringSlice("exclude")

	hasPick := len(pickPaths) > 0
	hasTraversal := len(includes) > 0 || len(excludes) > 0 ||
		c.Bool("include-all") || c.Bool("exclude-all") || c.Bool("exclude-defaults")

	if hasPick && hasTraversal {
		return fmt.Errorf("--pick cannot be combined with --include, --exclude, --include-all, --exclude-all, or --exclude-defaults")
	}

	mode := snapshot.ModeTraversal
	if hasPick {
		mode = snapshot.ModePick
	}

	var filterRules []filter.Rule
	if mode == snapshot.ModeTraversal {
		filterRules = buildFilterRules(os.Args[1:], includes, excludes)
	}

	cfg := snapshot.Config{
		Mode:                mode,
		SourceDir:           sourceDir,
		OutputPath:          c.String("output"),
		IncludeGitLog:       !c.Bool("exclude-git-log"),
		DryRun:              c.Bool("dry-run"),
		FilterRules:         filterRules,
		PickPaths:           pickPaths,
		ForceTextPatterns:   c.StringSlice("force-text"),
		ForceBinaryPatterns: c.StringSlice("force-binary"),
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
		if !silent {
			for _, f := range snap.Files {
				fmt.Println(f.RelPath)
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

	elapsed := time.Since(start)

	if !silent {
		fmt.Printf("Snapshot created: %s (%s)\n", absOutput, formatDuration(elapsed))
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
