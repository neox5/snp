package main

import (
	"context"
	"fmt"
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
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Print files that would be included without creating output",
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
			sourceDir := "."
			if c.NArg() > 0 {
				sourceDir = c.Args().First()
			}

			cfg := snapshot.Config{
				SourceDir:           sourceDir,
				OutputPath:          c.String("output"),
				IncludePatterns:     c.StringSlice("include"),
				ExcludePatterns:     c.StringSlice("exclude"),
				IncludeGitLog:       !c.Bool("exclude-git-log"),
				DryRun:              c.Bool("dry-run"),
				OutputExplicit:      c.IsSet("output"),
				ForceTextPatterns:   c.StringSlice("force-text"),
				ForceBinaryPatterns: c.StringSlice("force-binary"),
			}

			absSourceDir, absOutput, err := snapshot.ValidateAndResolve(cfg)
			if err != nil {
				return err
			}

			snap, err := snapshot.Build(ctx, cfg, absSourceDir, absOutput)
			if err != nil {
				return err
			}

			if cfg.DryRun {
				for _, f := range snap.Files {
					fmt.Println(f.RelPath)
				}
				return nil
			}

			outFile, err := os.Create(absOutput)
			if err != nil {
				return fmt.Errorf("cannot create output file %q: %w", absOutput, err)
			}
			defer outFile.Close()

			if err := snap.WriteTo(outFile); err != nil {
				return err
			}

			fmt.Printf("Files concatenated to %s\n", absOutput)
			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("snap: %v", err)
	}
}
