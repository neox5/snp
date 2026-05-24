package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Print writes a human-readable diagnostic representation of Config to stdout.
func (c Config) Print() {
	fmt.Println("[snp config]")
	fmt.Printf("  generated:     %s\n", func() string {
		if c.Generated.IsZero() {
			return "-"
		}
		return c.Generated.Local().Format("2006-01-02 15:04:05")
	}())
	fmt.Printf("  source_dir:    %s\n", c.SourceDir)
	fmt.Printf("  dry_run:       %v\n", c.DryRun)
	fmt.Println()
	fmt.Printf("  depth:         %d\n", c.Depth)
	fmt.Printf("  matcher_flags: %v\n", c.MatcherFlags)
	fmt.Printf("  pick_paths:    %v\n", c.PickPaths)
	fmt.Printf("  force_text:    %v\n", c.ForceTextPatterns)
	fmt.Printf("  force_binary:  %v\n", c.ForceBinaryPatterns)
	fmt.Printf("  output:        %s\n", c.OutputPath)
	fmt.Printf("  no_summary:    %v\n", c.NoSummary)
	fmt.Printf("  no_index:      %v\n", c.NoIndex)
	fmt.Printf("  no_git_log:    %v\n", c.NoGitLog)
	fmt.Printf("  no_content:    %v\n", c.NoContent)
	fmt.Printf("  silent:        %v\n", c.Silent)
}

// BuildCommand returns the equivalent CLI command string for cfg.
func (c Config) BuildCommand() string {
	var parts []string
	parts = append(parts, "snp")

	for _, f := range c.MatcherFlags {
		switch f.Type {
		case FlagTypeIncludeAll:
			parts = append(parts, "--include-all")
		case FlagTypeExcludeAll:
			parts = append(parts, "--exclude-all")
		case FlagTypeInclude:
			parts = append(parts, "--include", f.Value)
		case FlagTypeExclude:
			parts = append(parts, "--exclude", f.Value)
		}
	}

	for _, p := range c.PickPaths {
		parts = append(parts, "--pick", p)
	}

	if c.Depth >= 0 {
		parts = append(parts, "--depth", strconv.Itoa(c.Depth))
	}

	if c.NoSummary {
		parts = append(parts, "--no-summary")
	}
	if c.NoIndex {
		parts = append(parts, "--no-index")
	}
	if c.NoGitLog {
		parts = append(parts, "--no-git-log")
	}
	if c.NoContent {
		parts = append(parts, "--no-content")
	}
	if c.Silent {
		parts = append(parts, "--silent")
	}
	if c.OutputPath != "" && c.OutputPath != DefaultOutputPath {
		parts = append(parts, "--output", c.OutputPath)
	}

	for _, p := range c.ForceTextPatterns {
		parts = append(parts, "--force-text", p)
	}
	for _, p := range c.ForceBinaryPatterns {
		parts = append(parts, "--force-binary", p)
	}

	return strings.Join(parts, " ")
}
