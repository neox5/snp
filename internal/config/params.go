package config

import (
	"fmt"
	"strings"
	"time"
)

// Params captures all CLI input parameters for a snapshot run.
type Params struct {
	// runtime-only
	SourceDir    string
	Args         []string
	NoConfig     bool
	SaveConfig   bool
	ShowConfig   bool
	DryRun       bool
	VerboseLevel int
	// persistable
	Depth               int
	PickPaths           []string
	ForceTextPatterns   []string
	ForceBinaryPatterns []string
	OutputPath          string
	NoSummary           bool
	NoIndex             bool
	NoGitLog            bool
	NoContent           bool
	OnlySummary         bool
	OnlyIndex           bool
	OnlyGitLog          bool
	OnlyContent         bool
	Stdout              bool
	Silent              bool
}

func (p Params) ToConfig() *Config {
	hasOnly := p.OnlySummary || p.OnlyIndex || p.OnlyGitLog || p.OnlyContent

	return &Config{
		Generated:           time.Now(),
		SourceDir:           p.SourceDir,
		Depth:               p.Depth,
		MatcherFlags:        p.buildMatcherFlags(),
		PickPaths:           p.PickPaths,
		ForceTextPatterns:   p.ForceTextPatterns,
		ForceBinaryPatterns: p.ForceBinaryPatterns,
		OutputPath:          p.OutputPath,
		NoSummary:           (hasOnly && !p.OnlySummary) || p.NoSummary,
		NoIndex:             (hasOnly && !p.OnlyIndex) || p.NoIndex,
		NoGitLog:            (hasOnly && !p.OnlyGitLog) || p.NoGitLog,
		NoContent:           (hasOnly && !p.OnlyContent) || p.NoContent,
		Stdout:              p.Stdout,
		DryRun:              p.DryRun,
		Silent:              p.Silent,
	}
}

// Print writes a human-readable representation of Params to stdout.
func (p Params) Print() {
	fmt.Println("[params]")
	fmt.Printf("  source_dir:    %s\n", p.SourceDir)
	fmt.Printf("  args:          [%s]\n", strings.Join(p.Args, " "))
	fmt.Printf("  no_config:     %v\n", p.NoConfig)
	fmt.Printf("  save_config:   %v\n", p.SaveConfig)
	fmt.Printf("  show_config:  %v\n", p.ShowConfig)
	fmt.Printf("  dry_run:       %v\n", p.DryRun)
	fmt.Printf("  verbose:       %d\n", p.VerboseLevel)
	fmt.Println()
	fmt.Printf("  depth:         %d\n", p.Depth)
	fmt.Printf("  pick_paths:    %v\n", p.PickPaths)
	fmt.Printf("  force_text:    %v\n", p.ForceTextPatterns)
	fmt.Printf("  force_binary:  %v\n", p.ForceBinaryPatterns)
	fmt.Printf("  output:        %s\n", p.OutputPath)
	fmt.Printf("  no_summary:    %v\n", p.NoSummary)
	fmt.Printf("  no_index:      %v\n", p.NoIndex)
	fmt.Printf("  no_git_log:    %v\n", p.NoGitLog)
	fmt.Printf("  no_content:    %v\n", p.NoContent)
	fmt.Printf("  only_summary:  %v\n", p.OnlySummary)
	fmt.Printf("  only_index:    %v\n", p.OnlyIndex)
	fmt.Printf("  only_git_log:  %v\n", p.OnlyGitLog)
	fmt.Printf("  only_content:  %v\n", p.OnlyContent)
	fmt.Printf("  stdout:        %v\n", p.Stdout)
	fmt.Printf("  silent:        %v\n", p.Silent)
}

func (p Params) buildMatcherFlags() []Flag {
	f := []Flag{}

	for i := 0; i < len(p.Args); i++ {
		arg := p.Args[i]

		if arg == "--exclude-all" {
			f = append(f, Flag{Type: FlagTypeExcludeAll})
			continue
		}
		if arg == "--include-all" {
			f = append(f, Flag{Type: FlagTypeIncludeAll})
			continue
		}

		if arg == "--exclude" {
			i++
			f = append(f, Flag{Type: FlagTypeExclude, Value: p.Args[i]})
			continue
		}

		if arg == "--include" {
			i++
			f = append(f, Flag{Type: FlagTypeInclude, Value: p.Args[i]})
			continue
		}

		if v, match := strings.CutPrefix(arg, "--exclude"); match {
			f = append(f, Flag{Type: FlagTypeExclude, Value: v})
			continue
		}

		if v, match := strings.CutPrefix(arg, "--include"); match {
			f = append(f, Flag{Type: FlagTypeInclude, Value: v})
			continue
		}
	}

	return f
}
