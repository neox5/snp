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

	if p.PrintConfig {
		cfg.Print()
		return nil
	}

	if !cfg.Silent {
		fmt.Println("[command]")
		fmt.Println(cfg.BuildCommand())
	}

	if p.VerboseLevel >= 2 {
		fmt.Println()
		fmt.Println("[Matcher Rules]")
		cfg.BuildMatcherRules().Print("  ")
	}

	entires, err := snp.Collect(cfg)
	if err != nil {
		return err
	}

	for _, e := range entires {
		if e.IsDir {
			fmt.Printf("%s (%d items, %s)\n", e.Path, e.ItemCount, formatSize(e.Size))
			continue
		}
		fmt.Printf("%s\n", e.Path)
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

// formatSize formats byte size in human-readable format
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes == 0:
		return "0 bytes"
	case bytes == 1:
		return "1 byte"
	case bytes < KB:
		return fmt.Sprintf("%d bytes", bytes)
	case bytes < MB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	case bytes < GB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	}
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
