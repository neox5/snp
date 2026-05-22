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

	pickPaths := c.StringSlice("pick")
	includes := c.StringSlice("include")
	excludes := c.StringSlice("exclude")
	forceText := c.StringSlice("force-text")
	forceBinary := c.StringSlice("force-binary")
	outputPath := c.String("output")
	noSummary := c.Bool("no-summary")
	noIndex := c.Bool("no-index")
	noGitLog := c.Bool("no-git-log")
	noContent := c.Bool("no-content")
	silent := c.Bool("silent")
	dryRun := c.Bool("dry-run")
	depth := c.Int("depth")

	hasPick := len(pickPaths) > 0
	hasTraversal := len(includes) > 0 || len(excludes) > 0 ||
		c.Bool("include-all") || c.Bool("exclude-all") || c.Bool("exclude-defaults")

	if hasPick && hasTraversal {
		return fmt.Errorf("--pick cannot be combined with --include, --exclude, --include-all, --exclude-all, or --exclude-defaults")
	}
	if hasPick && depth != -1 {
		return fmt.Errorf("--depth cannot be combined with --pick")
	}

	// --save-config
	if c.Bool("save-config") {
		saveCfg := buildSaveConfig(
			os.Args[1:],
			includes, excludes, pickPaths,
			forceText, forceBinary,
			outputPath,
			noSummary, noIndex, noGitLog, noContent, silent,
			depth,
		)
		saveCfg.Generated = time.Now()
		if err := config.Save(sourceDir, saveCfg); err != nil {
			return err
		}
		if !silent {
			fmt.Printf("Saved config to %s\n", config.ConfigFileName)
		}
		return nil
	}

	// --print-config
	if c.Bool("print-config") {
		var loaded config.FullConfig
		if !c.Bool("no-config") {
			var err error
			loaded, err = config.Load(sourceDir)
			if err != nil {
				return err
			}
		}
		config.Print(loaded)
		return nil
	}

	// load .snpconfig unless --no-config
	var loaded config.FullConfig
	if !c.Bool("no-config") {
		var err error
		loaded, err = config.Load(sourceDir)
		if err != nil {
			return err
		}
	}

	cfg := buildFullConfig(
		os.Args[1:],
		loaded,
		includes, excludes, pickPaths,
		forceText, forceBinary,
		outputPath,
		noSummary, noIndex, noGitLog, noContent, silent, dryRun,
		depth,
		sourceDir,
	)

	config.ApplyDefaults(&cfg)

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

	elapsed := time.Since(start)

	if !cfg.Silent {
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
