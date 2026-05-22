package main

import (
	"strings"

	"github.com/neox5/snp/internal/config"
	"github.com/neox5/snp/internal/filter"
)

// buildFilterRules reconstructs ordered filter rules from raw CLI args,
// injecting configRules between the implicit baseline and CLI flags.
//
// Order:
//
//	[implicit baseline]   — suppressed if baseline present in args OR configRules
//	[configRules]
//	[CLI args in order]
func buildFilterRules(args []string, includes, excludes []string, configRules []filter.Rule) []filter.Rule {
	hasBaselineCLI := hasBaselineFlag(args)
	hasBaselineConfig := config.HasBaseline(configRules)

	var rules []filter.Rule

	// seed implicit baseline only if no baseline anywhere
	if !hasBaselineCLI && !hasBaselineConfig {
		rules = []filter.Rule{
			{Type: filter.RuleIncludeAll},
			{Type: filter.RuleExcludeDefaults},
		}
	}

	// inject config rules
	rules = append(rules, configRules...)

	// append CLI rules in order
	rules = append(rules, buildCLIRules(args, includes, excludes)...)

	return rules
}

// buildCLIRules extracts only the traversal rules from raw CLI args in order.
// Used both by buildFilterRules and --save-config.
func buildCLIRules(args []string, includes, excludes []string) []filter.Rule {
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
