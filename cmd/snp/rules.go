package main

import (
	"strings"

	"github.com/neox5/snp/internal/filter"
)

// buildFilterRules reconstructs ordered filter rules from raw CLI args.
// If no baseline flag is present, seeds with implicit defaults:
// --include-all --exclude-defaults
// User --include/--exclude rules are always appended in order.
func buildFilterRules(args []string, includes, excludes []string) []filter.Rule {
	includeSet := make(map[string]int)
	excludeSet := make(map[string]int)
	for _, v := range includes {
		includeSet[v]++
	}
	for _, v := range excludes {
		excludeSet[v]++
	}

	// Check if any baseline flag is present
	hasBaseline := false
	for _, arg := range args {
		for _, bf := range baselineFlags {
			if arg == bf {
				hasBaseline = true
				break
			}
		}
		if hasBaseline {
			break
		}
	}

	// Seed with implicit defaults if no baseline present
	var rules []filter.Rule
	if !hasBaseline {
		rules = []filter.Rule{
			{Type: filter.RuleIncludeAll},
			{Type: filter.RuleExcludeDefaults},
		}
	}

	// Parse args in order — baseline flags append to rules,
	// --include/--exclude consume from their respective count maps
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
