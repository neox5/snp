package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snp/internal/filter"
	"github.com/neox5/snp/internal/snapshot"
	"github.com/neox5/snp/internal/version"
)

// baselineFlags are the flags that set the starting state for traversal.
var baselineFlags = []string{"--include-all", "--exclude-all", "--exclude-defaults"}

func main() {
	app := &cli.Command{
		Name:    "snp",
		Usage:   "Concatenate readable source/text files into one snapshot file.",
		Version: version.String(),
		UsageText: `snp [OPTIONS] [DIRECTORY]

Concatenates readable source/text files into one snapshot file.
If DIRECTORY is omitted, '.' is used.

Two modes (mutually exclusive):
  Traversal (default): walk directory tree with ordered filter rules
  Pick:                directly address files by path or glob, no traversal`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output",
				Usage: "Set output file path",
				Value: snapshot.DefaultOutputName,
			},
			// Mode 1 — Traversal baseline
			&cli.BoolFlag{
				Name:  "include-all",
				Usage: "Baseline: include all files (positional, ordered)",
			},
			&cli.BoolFlag{
				Name:  "exclude-all",
				Usage: "Baseline: exclude all files (positional, ordered)",
			},
			&cli.BoolFlag{
				Name:  "exclude-defaults",
				Usage: "Baseline: exclude default patterns and .gitignore (positional, ordered)",
			},
			// Mode 1 — Traversal filters
			&cli.StringSliceFlag{
				Name:  "include",
				Usage: "Include files matching glob pattern (repeatable, ordered)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "Exclude files matching glob pattern (repeatable, ordered)",
			},
			// Mode 1 — Utility
			&cli.BoolFlag{
				Name:  "show-defaults",
				Usage: "Print default exclude patterns and exit",
			},
			// Mode 2 — Pick
			&cli.StringSliceFlag{
				Name:  "pick",
				Usage: "Directly include file by path or glob, no traversal (repeatable)",
			},
			// Shared
			&cli.BoolFlag{
				Name:  "exclude-git-log",
				Usage: "Omit the Git log section (included by default)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Print files that would be included without creating output",
			},
			&cli.BoolFlag{
				Name:  "silent",
				Usage: "Suppress all output (exit codes only)",
			},
			&cli.StringSliceFlag{
				Name:  "force-text",
				Usage: "Force files matching glob pattern to be treated as text (repeatable)",
			},
			&cli.StringSliceFlag{
				Name:  "force-binary",
				Usage: "Force files matching glob pattern to be treated as binary (repeatable)",
			},
		},
		ArgsUsage: "[DIRECTORY]",
		Action: func(ctx context.Context, c *cli.Command) error {
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
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("snp: %v", err)
	}
}

// buildFilterRules reconstructs ordered filter rules from raw CLI args.
// If no baseline flag is present, injects implicit defaults:
// --include-all --exclude-defaults
func buildFilterRules(args []string, includes, excludes []string) []filter.Rule {
	includeSet := make(map[string]bool)
	excludeSet := make(map[string]bool)
	for _, v := range includes {
		includeSet[v] = true
	}
	for _, v := range excludes {
		excludeSet[v] = true
	}

	// Check if any baseline flag is present
	hasBaseline := false
	for _, arg := range args {
		for _, bf := range baselineFlags {
			if arg == bf {
				hasBaseline = true
				break
			}
		}
		if hasBaseline {
			break
		}
	}

	// Inject implicit defaults if no baseline present
	if !hasBaseline {
		return []filter.Rule{
			{Type: filter.RuleIncludeAll},
			{Type: filter.RuleExcludeDefaults},
		}
	}

	// Parse args in order
	var rules []filter.Rule
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--include-all":
			rules = append(rules, filter.Rule{Type: filter.RuleIncludeAll})
			continue
		case "--exclude-all":
			rules = append(rules, filter.Rule{Type: filter.RuleExcludeAll})
			continue
		case "--exclude-defaults":
			rules = append(rules, filter.Rule{Type: filter.RuleExcludeDefaults})
			continue
		}

		var flagName, value string

		if strings.HasPrefix(arg, "--include=") {
			flagName = "--include"
			value = strings.TrimPrefix(arg, "--include=")
		} else if strings.HasPrefix(arg, "--exclude=") {
			flagName = "--exclude"
			value = strings.TrimPrefix(arg, "--exclude=")
		} else if arg == "--include" && i+1 < len(args) {
			flagName = "--include"
			i++
			value = args[i]
		} else if arg == "--exclude" && i+1 < len(args) {
			flagName = "--exclude"
			i++
			value = args[i]
		} else {
			continue
		}

		switch flagName {
		case "--include":
			if includeSet[value] {
				rules = append(rules, filter.Rule{Type: filter.RuleInclude, Pattern: value})
			}
		case "--exclude":
			if excludeSet[value] {
				rules = append(rules, filter.Rule{Type: filter.RuleExclude, Pattern: value})
			}
		}
	}

	return rules
}

// formatDuration formats duration as milliseconds or seconds
func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
