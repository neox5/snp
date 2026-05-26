package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Print writes a human-readable diagnostic representation of Config to stdout.
func (c Config) Print(indent ...string) {
	prefix := ""
	if len(indent) > 0 {
		prefix = indent[0]
	}
	fmt.Printf("%sgenerated:     %s\n", prefix, func() string {
		if c.Generated.IsZero() {
			return "-"
		}
		return c.Generated.Local().Format("2006-01-02 15:04:05")
	}())
	fmt.Printf("%ssource_dir:    %s\n", prefix, c.SourceDir)
	fmt.Printf("%sdry_run:       %v\n", prefix, c.DryRun)
	fmt.Println()
	fmt.Printf("%sdepth:         %d\n", prefix, c.Depth)
	fmt.Printf("%smatcher_flags: %s\n", prefix, c.MatcherFlags)
	fmt.Printf("%spick_paths:    %v\n", prefix, c.PickPaths)
	fmt.Printf("%sforce_text:    %v\n", prefix, c.ForceTextPatterns)
	fmt.Printf("%sforce_binary:  %v\n", prefix, c.ForceBinaryPatterns)
	fmt.Printf("%soutput:        %s\n", prefix, c.OutputPath)
	fmt.Printf("%sno_summary:    %v\n", prefix, c.NoSummary)
	fmt.Printf("%sno_index:      %v\n", prefix, c.NoIndex)
	fmt.Printf("%sno_git_log:    %v\n", prefix, c.NoGitLog)
	fmt.Printf("%sno_content:    %v\n", prefix, c.NoContent)
	fmt.Printf("%sstdout:        %v\n", prefix, c.Stdout)
	fmt.Printf("%ssilent:        %v\n", prefix, c.Silent)
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
	if c.Stdout {
		parts = append(parts, "--stdout")
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
