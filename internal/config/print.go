package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Print writes a human-readable diagnostic representation of Config to stdout.
// Settings that have an entry in origins are followed by their source.
func (c Config) Print(origins Origins, indent ...string) {
	prefix := ""
	if len(indent) > 0 {
		prefix = indent[0]
	}
	line := func(key, label string, value any) {
		suffix := ""
		if src := origins[key]; src != "" {
			suffix = "  (" + src + ")"
		}
		fmt.Printf("%s%-15s%v%s\n", prefix, label, value, suffix)
	}

	generated := "-"
	if !c.Generated.IsZero() {
		generated = c.Generated.Local().Format("2006-01-02 15:04:05")
	}
	line("", "generated:", generated)
	line("", "source_dir:", c.SourceDir)
	line("", "dry_run:", c.DryRun)
	fmt.Println()
	line("depth", "depth:", c.Depth)
	line("filter_flags", "matcher_flags:", c.MatcherFlags)
	line("pick_paths", "pick_paths:", c.PickPaths)
	line("force_text_patterns", "force_text:", c.ForceTextPatterns)
	line("force_binary_patterns", "force_binary:", c.ForceBinaryPatterns)
	line("output_path", "output:", c.OutputPath)
	line("no_summary", "no_summary:", c.NoSummary)
	line("no_index", "no_index:", c.NoIndex)
	line("no_git_log", "no_git_log:", c.NoGitLog)
	line("no_content", "no_content:", c.NoContent)
	line("stdout", "stdout:", c.Stdout)
	line("silent", "silent:", c.Silent)
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
