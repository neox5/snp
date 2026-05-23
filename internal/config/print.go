package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neox5/snp/internal/filter"
)

// Print writes a human-readable diagnostic representation of Config to stdout.
// label identifies the stage, e.g. "file config" or "merged config".
func (cfg Config) Print(label string) {
	fmt.Println("[" + label + "]")
	fmt.Printf("  generated:     %s\n", func() string {
		if cfg.Generated.IsZero() {
			return "-"
		}
		return cfg.Generated.Local().Format("2006-01-02 15:04:05")
	}())
	fmt.Printf("  source_dir:    %s\n", cfg.SourceDir)
	fmt.Printf("  dry_run:       %v\n", cfg.DryRun)
	fmt.Println()
	fmt.Printf("  depth:         %d\n", cfg.Depth)
	fmt.Printf("  filter_rules:  [%s]\n", func() string {
		var parts []string
		for _, r := range cfg.FilterRules {
			if s, err := serializeRule(r); err == nil {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	}())
	fmt.Printf("  pick_paths:    %v\n", cfg.PickPaths)
	fmt.Printf("  force_text:    %v\n", cfg.ForceTextPatterns)
	fmt.Printf("  force_binary:  %v\n", cfg.ForceBinaryPatterns)
	fmt.Printf("  output:        %s\n", cfg.OutputPath)
	fmt.Printf("  no_summary:    %v\n", cfg.NoSummary)
	fmt.Printf("  no_index:      %v\n", cfg.NoIndex)
	fmt.Printf("  no_git_log:    %v\n", cfg.NoGitLog)
	fmt.Printf("  no_content:    %v\n", cfg.NoContent)
	fmt.Printf("  silent:        %v\n", cfg.Silent)
}

// Serialize writes cfg in .snpconfig file format to stdout.
func (cfg Config) Serialize() {
	if !cfg.Generated.IsZero() {
		fmt.Println("# generated: " + cfg.Generated.Local().Format("2006-01-02 15:04:05"))
	}

	if cfg.Depth >= 0 {
		fmt.Println()
		fmt.Println("# depth")
		fmt.Printf("depth %d\n", cfg.Depth)
	}

	if len(cfg.FilterRules) > 0 {
		fmt.Println()
		fmt.Println("# filters")
		for _, r := range cfg.FilterRules {
			if line, err := serializeRule(r); err == nil {
				fmt.Println(line)
			}
		}
	}

	if len(cfg.PickPaths) > 0 {
		fmt.Println()
		fmt.Println("# pick")
		for _, p := range cfg.PickPaths {
			fmt.Println("pick " + p)
		}
	}

	hasOutput := (cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath) ||
		cfg.Silent || cfg.NoSummary || cfg.NoIndex || cfg.NoGitLog || cfg.NoContent
	if hasOutput {
		fmt.Println()
		fmt.Println("# output")
		if cfg.NoSummary {
			fmt.Println("no-summary")
		}
		if cfg.NoIndex {
			fmt.Println("no-index")
		}
		if cfg.NoGitLog {
			fmt.Println("no-git-log")
		}
		if cfg.NoContent {
			fmt.Println("no-content")
		}
		if cfg.Silent {
			fmt.Println("silent")
		}
		if cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath {
			fmt.Println("output " + cfg.OutputPath)
		}
	}

	if len(cfg.ForceTextPatterns) > 0 || len(cfg.ForceBinaryPatterns) > 0 {
		fmt.Println()
		fmt.Println("# overrides")
		for _, p := range cfg.ForceTextPatterns {
			fmt.Println("force-text " + p)
		}
		for _, p := range cfg.ForceBinaryPatterns {
			fmt.Println("force-binary " + p)
		}
	}

	fmt.Println()
	fmt.Println("# equivalent command")
	fmt.Println(cfg.BuildCommand())
}

// BuildCommand returns the equivalent CLI command string for cfg.
func (cfg Config) BuildCommand() string {
	var parts []string
	parts = append(parts, "snp")

	for _, r := range cfg.FilterRules {
		switch r.Type {
		case filter.RuleIncludeAll:
			parts = append(parts, "--include-all")
		case filter.RuleExcludeAll:
			parts = append(parts, "--exclude-all")
		case filter.RuleExcludeDefaults:
			parts = append(parts, "--exclude-defaults")
		case filter.RuleInclude:
			parts = append(parts, "--include", r.Pattern)
		case filter.RuleExclude:
			parts = append(parts, "--exclude", r.Pattern)
		}
	}

	for _, p := range cfg.PickPaths {
		parts = append(parts, "--pick", p)
	}

	if cfg.Depth >= 0 {
		parts = append(parts, "--depth", strconv.Itoa(cfg.Depth))
	}

	if cfg.NoSummary {
		parts = append(parts, "--no-summary")
	}
	if cfg.NoIndex {
		parts = append(parts, "--no-index")
	}
	if cfg.NoGitLog {
		parts = append(parts, "--no-git-log")
	}
	if cfg.NoContent {
		parts = append(parts, "--no-content")
	}
	if cfg.Silent {
		parts = append(parts, "--silent")
	}
	if cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath {
		parts = append(parts, "--output", cfg.OutputPath)
	}

	for _, p := range cfg.ForceTextPatterns {
		parts = append(parts, "--force-text", p)
	}
	for _, p := range cfg.ForceBinaryPatterns {
		parts = append(parts, "--force-binary", p)
	}

	return strings.Join(parts, " ")
}
