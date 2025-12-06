package main

import (
	"context"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"

	"github.com/neox5/snap/internal/snapshot"
	"github.com/neox5/snap/internal/version"
)

func main() {
	app := &cli.Command{
		Name:    "snap",
		Usage:   "Concatenate readable source/text files into one snapshot file.",
		Version: version.String(),
		UsageText: `snap [OPTIONS] [DIRECTORY]

Concatenates readable source/text files into one snapshot file.
If DIRECTORY is omitted, '.' is used.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output",
				Usage: "Set output file path (default: ./snap.txt)",
				Value: "snap.txt",
			},
			&cli.StringSliceFlag{
				Name:  "include",
				Usage: "Include files matching this glob pattern (repeatable)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "Exclude files matching this glob pattern (repeatable)",
			},
			&cli.BoolFlag{
				Name:  "exclude-git-log",
				Usage: "Omit the Git log section (included by default)",
			},
		},
		ArgsUsage: "[DIRECTORY]",
		Action: func(ctx context.Context, c *cli.Command) error {
			sourceDir := "."
			if c.NArg() > 0 {
				sourceDir = c.Args().First()
			}

			cfg := snapshot.Config{
				SourceDir:       sourceDir,
				OutputPath:      c.String("output"),
				IncludePatterns: c.StringSlice("include"),
				ExcludePatterns: c.StringSlice("exclude"),
				IncludeGitLog:   !c.Bool("exclude-git-log"),
				// Was the flag explicitly set?
				OutputExplicit: c.IsSet("output"),
			}

			return snapshot.Run(ctx, cfg)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("snap: %v", err)
	}
}
