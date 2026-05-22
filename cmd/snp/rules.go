package main

import (
	"strings"
	"time"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/filter"
)

// buildFullConfig constructs a config.FullConfig from CLI args and loaded config.
func buildFullConfig(
	args []string,
	loaded config.FullConfig,
	includes, excludes, pickPaths []string,
	forceText, forceBinary []string,
	outputPath string,
	noSummary, noIndex, noGitLog, noContent, silent, dryRun bool,
	depth int,
	sourceDir string,
) config.FullConfig {
	cfg := config.FullConfig{
		SourceDir: sourceDir,
		DryRun:    dryRun,
	}

	// depth: CLI wins if explicitly set (not -1), else use loaded, else -1
	if depth != -1 {
		cfg.Depth = depth
	} else if loaded.Depth != -1 {
		cfg.Depth = loaded.Depth
	} else {
		cfg.Depth = -1
	}

	// pick mode — CLI picks override loaded picks entirely
	if len(pickPaths) > 0 {
		cfg.PickPaths = pickPaths
	} else if len(loaded.PickPaths) > 0 {
		cfg.PickPaths = loaded.PickPaths
	}

	// filter rules (traversal mode)
	if len(cfg.PickPaths) == 0 {
		hasBaselineCLI := hasBaselineFlag(args)
		hasBaselineConfig := config.HasBaseline(loaded)

		var rules []filter.Rule

		if !hasBaselineCLI && !hasBaselineConfig {
			rules = append(rules,
				filter.Rule{Type: filter.RuleIncludeAll},
				filter.Rule{Type: filter.RuleExcludeDefaults},
			)
		}

		rules = append(rules, loaded.FilterRules...)
		rules = append(rules, buildCLIFilterRules(args, includes, excludes)...)
		cfg.FilterRules = rules
	}

	// force overrides: merge loaded + CLI
	cfg.ForceTextPatterns = mergeUnique(loaded.ForceTextPatterns, forceText)
	cfg.ForceBinaryPatterns = mergeUnique(loaded.ForceBinaryPatterns, forceBinary)

	// output path: CLI wins, fallback to loaded
	if outputPath != "" {
		cfg.OutputPath = outputPath
	} else if loaded.OutputPath != "" {
		cfg.OutputPath = loaded.OutputPath
	}

	// boolean output flags: CLI wins over loaded
	cfg.NoSummary = noSummary || loaded.NoSummary
	cfg.NoIndex = noIndex || loaded.NoIndex
	cfg.NoGitLog = noGitLog || loaded.NoGitLog
	cfg.NoContent = noContent || loaded.NoContent
	cfg.Silent = silent || loaded.Silent

	return cfg
}

// buildSaveConfig constructs a FullConfig for --save-config from CLI args only.
func buildSaveConfig(
	args []string,
	includes, excludes, pickPaths []string,
	forceText, forceBinary []string,
	outputPath string,
	noSummary, noIndex, noGitLog, noContent, silent bool,
	depth int,
) config.FullConfig {
	cfg := config.FullConfig{
		Generated: time.Now(),
		Depth:     depth,
	}

	cfg.PickPaths = pickPaths
	cfg.FilterRules = buildCLIFilterRules(args, includes, excludes)
	cfg.ForceTextPatterns = forceText
	cfg.ForceBinaryPatterns = forceBinary
	cfg.OutputPath = outputPath
	cfg.NoSummary = noSummary
	cfg.NoIndex = noIndex
	cfg.NoGitLog = noGitLog
	cfg.NoContent = noContent
	cfg.Silent = silent

	return cfg
}

// buildCLIFilterRules extracts ordered traversal filter rules from raw CLI args.
func buildCLIFilterRules(args []string, includes, excludes []string) []filter.Rule {
	includeSet := make(map[string]int)
	excludeSet := make(map[string]int)
	for _, v := range includes {
		includeSet[v]++
	}
	for _, v := range excludes {
		excludeSet[v]++
	}

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
			if includeSet[value] > 0 {
				rules = append(rules, filter.Rule{Type: filter.RuleInclude, Pattern: value})
				includeSet[value]--
			}
		case "--exclude":
			if excludeSet[value] > 0 {
				rules = append(rules, filter.Rule{Type: filter.RuleExclude, Pattern: value})
				excludeSet[value]--
			}
		}
	}

	return rules
}

// hasBaselineFlag reports whether any baseline flag is present in args.
func hasBaselineFlag(args []string) bool {
	for _, arg := range args {
		for _, bf := range baselineFlags {
			if arg == bf {
				return true
			}
		}
	}
	return false
}

// mergeUnique merges two string slices, preserving order and removing duplicates.
func mergeUnique(base, override []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range append(base, override...) {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
