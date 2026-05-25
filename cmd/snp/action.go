package main

import (
	"context"
	"fmt"
	"os"
	"time"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/snp"
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

	p := cliToParams(c)
	cfg, err := config.LoadConfig(p)
	if err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return err
	}

	if p.SaveConfig {
		if err = cfg.Save(p.SourceDir); err != nil {
			return err
		}
		if !p.Silent {
			fmt.Printf("Saved config to %s\n", config.ConfigFileName)
		}
		return nil
	}
	printHeader := func(first bool, h string) {
		if !first {
			fmt.Println()
		}
		fmt.Printf("[%s]\n", h)
	}

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

	snap := snp.New(cfg)

	if err = snap.Collect(ctx); err != nil {
		return err
	}

	if p.DryRun {
		printHeader(false, "Dry Run")
		snap.PrintRawEntries()
		return nil
	}

	printHeader(false, "Git Log")
	snap.GitData.Print()

	// Process

	return nil
}

func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// 	absSourceDir, absOutput, err := snapshot.ValidateAndResolve(cfg)
// 	if err != nil {
// 		return err
// 	}

// 	start := time.Now()

// 	snap, err := snapshot.Build(ctx, cfg, absSourceDir, absOutput)
// 	if err != nil {
// 		return err
// 	}

// 	if cfg.DryRun {
// 		if !cfg.Silent {
// 			fmt.Println()
// 			fmt.Println("[Dry-Run]")
// 			for _, f := range snap.Files {
// 				if f.IsBinary {
// 					fmt.Printf("%s (binary, %s)\n", f.RelPath, file.FormatSize(f.Size))
// 				} else {
// 					fmt.Printf("%s (%d lines, %s)\n", f.RelPath, len(f.Lines), file.FormatSize(f.Size))
// 				}
// 			}
// 			for _, d := range snap.CollapsedDirs {
// 				fmt.Printf("%s (%d items, %s)\n", d.RelPath, d.ItemCount, file.FormatSize(d.Size))
// 			}
// 		}
// 		return nil
// 	}

// 	outFile, err := os.Create(absOutput)
// 	if err != nil {
// 		return fmt.Errorf("cannot create output file %q: %w", absOutput, err)
// 	}
// 	defer outFile.Close()

// 	if _, err := snap.WriteTo(outFile); err != nil {
// 		return err
// 	}

// 	if !cfg.Silent {
// 		fmt.Printf("Snapshot created: %s (%s)\n", absOutput, formatDuration(time.Since(start)))
// 	}

// 	return nil
// }
