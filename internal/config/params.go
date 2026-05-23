package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neox5/snp/internal/filter"
)

// Params captures all CLI input for a snapshot run.
// It is the input to FromParams and FromParamsCLI and is never persisted directly.
type Params struct {
	// runtime-only
	SourceDir    string
	Args         []string
	NoConfig     bool
	SaveConfig   bool
	PrintConfig  bool
	DryRun       bool
	VerboseLevel int
	// persistable
	Depth               int
	Includes            []string
	Excludes            []string
	PickPaths           []string
	ForceTextPatterns   []string
	ForceBinaryPatterns []string
	OutputPath          string
	NoSummary           bool
	NoIndex             bool
	NoGitLog            bool
	NoContent           bool
	Silent              bool
}

// Print writes a human-readable representation of Params to stdout.
func (p Params) Print() {
	fmt.Println("[params]")
	fmt.Printf("  source_dir:    %s\n", p.SourceDir)
	fmt.Printf("  args:          [%s]\n", strings.Join(p.Args, " "))
	fmt.Printf("  no_config:     %v\n", p.NoConfig)
	fmt.Printf("  save_config:   %v\n", p.SaveConfig)
	fmt.Printf("  print_config:  %v\n", p.PrintConfig)
	fmt.Printf("  dry_run:       %v\n", p.DryRun)
	fmt.Printf("  verbose:       %d\n", p.VerboseLevel)
	fmt.Println()
	fmt.Printf("  depth:         %d\n", p.Depth)
	fmt.Printf("  includes:      %v\n", p.Includes)
	fmt.Printf("  excludes:      %v\n", p.Excludes)
	fmt.Printf("  pick_paths:    %v\n", p.PickPaths)
	fmt.Printf("  force_text:    %v\n", p.ForceTextPatterns)
	fmt.Printf("  force_binary:  %v\n", p.ForceBinaryPatterns)
	fmt.Printf("  output:        %s\n", p.OutputPath)
	fmt.Printf("  no_summary:    %v\n", p.NoSummary)
	fmt.Printf("  no_index:      %v\n", p.NoIndex)
	fmt.Printf("  no_git_log:    %v\n", p.NoGitLog)
	fmt.Printf("  no_content:    %v\n", p.NoContent)
	fmt.Printf("  silent:        %v\n", p.Silent)
}

// FromParams builds a Config from CLI params and the loaded .snpconfig file.
func FromParams(p Params, verbose bool) (Config, error) {
	return fromParams(p, !p.NoConfig, verbose)
}

// FromParamsOnly builds a Config from CLI params only, without loading .snpconfig.
func FromParamsOnly(p Params, verbose bool) Config {
	cfg, _ := fromParams(p, false, verbose)
	return cfg
}

// fromParams is the shared internal builder.
// If load is true, it loads and merges the .snpconfig file from p.SourceDir.
// If verbose is true, prints file config and merged config after each stage.
func fromParams(p Params, load bool, verbose bool) (Config, error) {
	loaded := newConfig()
	if load {
		var err error
		loaded, err = Load(p.SourceDir)
		if err != nil {
			return newConfig(), fmt.Errorf("loading config: %w", err)
		}
	}

	if verbose {
		loaded.Print("file config")
		fmt.Println()
	}

	cfg := newConfig()
	cfg.SourceDir = p.SourceDir
	cfg.DryRun = p.DryRun

	// depth: CLI wins if explicitly set (not -1), else use loaded, else -1
	if p.Depth != -1 {
		cfg.Depth = p.Depth
	} else if loaded.Depth != -1 {
		cfg.Depth = loaded.Depth
	} else {
		cfg.Depth = -1
	}

	// pick mode — CLI picks override loaded picks entirely
	if len(p.PickPaths) > 0 {
		cfg.PickPaths = p.PickPaths
	} else if len(loaded.PickPaths) > 0 {
		cfg.PickPaths = loaded.PickPaths
	}

	// filter rules (traversal mode)
	if len(cfg.PickPaths) == 0 {
		hasBaselineCLI := hasBaseline(p.Args)
		hasBaselineFile := hasBaselineRules(loaded.FilterRules)

		var rules []filter.Rule

		if !hasBaselineCLI && !hasBaselineFile {
			rules = append(
				rules,
				filter.Rule{Type: filter.RuleIncludeAll},
				filter.Rule{Type: filter.RuleExcludeDefaults},
			)
		}

		rules = append(rules, loaded.FilterRules...)
		rules = append(rules, buildFilterRules(p.Args, p.Includes, p.Excludes)...)
		cfg.FilterRules = rules
	}

	// force overrides: merge loaded + CLI
	cfg.ForceTextPatterns = mergeUnique(loaded.ForceTextPatterns, p.ForceTextPatterns)
	cfg.ForceBinaryPatterns = mergeUnique(loaded.ForceBinaryPatterns, p.ForceBinaryPatterns)

	// output path: CLI wins, fallback to loaded, fallback to default
	if p.OutputPath != "" {
		cfg.OutputPath = p.OutputPath
	} else if loaded.OutputPath != "" {
		cfg.OutputPath = loaded.OutputPath
	} else {
		cfg.OutputPath = DefaultOutputPath
	}

	// boolean output flags: CLI wins over loaded
	cfg.NoSummary = p.NoSummary || loaded.NoSummary
	cfg.NoIndex = p.NoIndex || loaded.NoIndex
	cfg.NoGitLog = p.NoGitLog || loaded.NoGitLog
	cfg.NoContent = p.NoContent || loaded.NoContent
	cfg.Silent = p.Silent || loaded.Silent

	if verbose {
		cfg.Print("merged config")
		fmt.Println()
	}

	return cfg, nil
}

// — unexported helpers ————————————————————————————————————————————————————

var baselineFlags = []string{"--include-all", "--exclude-all", "--exclude-defaults"}

func hasBaseline(args []string) bool {
	for _, arg := range args {
		if slices.Contains(baselineFlags, arg) {
			return true
		}
	}
	return false
}

func hasBaselineRules(rules []filter.Rule) bool {
	for _, r := range rules {
		switch r.Type {
		case filter.RuleIncludeAll, filter.RuleExcludeAll, filter.RuleExcludeDefaults:
			return true
		}
	}
	return false
}

func buildFilterRules(args []string, includes, excludes []string) []filter.Rule {
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
