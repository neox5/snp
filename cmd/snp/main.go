package main

import (
	"context"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"

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
				Usage: "Set output file path (default: snapshot.snp)",
			},
			// Depth
			&cli.IntFlag{
				Name:  "depth",
				Usage: "Limit traversal depth (0 = root files only, -1 = full traversal)",
				Value: -1,
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
			// Output sections
			&cli.BoolFlag{
				Name:  "no-summary",
				Usage: "Omit the summary header",
			},
			&cli.BoolFlag{
				Name:  "no-index",
				Usage: "Omit the file index",
			},
			&cli.BoolFlag{
				Name:  "no-git-log",
				Usage: "Omit the Git log section",
			},
			&cli.BoolFlag{
				Name:  "no-content",
				Usage: "Omit file contents (index and metadata only)",
			},
			// Shared
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
			// Config
			&cli.BoolFlag{
				Name:  "no-config",
				Usage: "Skip .snpconfig file even if present",
			},
			&cli.BoolFlag{
				Name:  "save-config",
				Usage: "Save current flags to .snpconfig in source directory and exit",
			},
			&cli.BoolFlag{
				Name:  "print-config",
				Usage: "Print current config and equivalent snp command, then exit",
			},
		},
		ArgsUsage: "[DIRECTORY]",
		Action:    runAction,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("snp: %v", err)
	}
}
