package main

import (
	"context"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snp/internal/version"
)

var verboseCount int

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}
}

func main() {
	app := &cli.Command{
		Name:                   "snp",
		Usage:                  "Concatenate readable source/text files into one snapshot file.",
		Version:                version.String(),
		UseShortOptionHandling: true,
		UsageText: `snp [OPTIONS] [DIRECTORY]

Concatenates readable source/text files into one snapshot file.
If DIRECTORY is omitted, '.' is used.

Examples:
  snp                        # snapshot current directory
  snp ./myproject            # snapshot specific directory
  snp --dry-run              # list files without writing
  snp --pick go.mod go.sum   # include only specific files`,
		Flags: []cli.Flag{
			// Verbosity
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose output (-v) or debug output (-vv)",
				Config:  cli.BoolConfig{Count: &verboseCount},
			},
			// Depth
			&cli.IntFlag{
				Name:        "depth",
				Usage:       "Limit traversal depth (0 = root only, -1 = unlimited)",
				DefaultText: "unlimited",
				Value:       -1,
			},
			// File selection — traversal
			&cli.BoolFlag{
				Name:  "include-all",
				Usage: "Baseline: include all files",
			},
			&cli.BoolFlag{
				Name:  "exclude-all",
				Usage: "Baseline: exclude all files",
			},
			&cli.StringSliceFlag{
				Name:  "include",
				Usage: "Include files matching glob pattern (repeatable, ordered)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "Exclude files matching glob pattern (repeatable, ordered)",
			},
			// File selection — pick
			&cli.StringSliceFlag{
				Name:  "pick",
				Usage: "Include only these exact paths (repeatable, mutually exclusive with traversal flags)",
			},
			// Output sections
			&cli.BoolFlag{
				Name:  "no-summary",
				Usage: "Omit summary section",
			},
			&cli.BoolFlag{
				Name:  "no-index",
				Usage: "Omit file index section",
			},
			&cli.BoolFlag{
				Name:  "no-git-log",
				Usage: "Omit git log section",
			},
			&cli.BoolFlag{
				Name:  "no-content",
				Usage: "Omit file content sections",
			},
			// Output control
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path",
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "List files that would be included without writing output",
			},
			&cli.BoolFlag{
				Name:  "silent",
				Usage: "Suppress all output messages",
			},
			// Diagnostics
			&cli.BoolFlag{
				Name:  "show-defaults",
				Usage: "Print default exclude patterns and exit",
			},
			// Binary overrides
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
