package matcher

import (
	"fmt"
	"io"
	"strings"
)

// RuleType defines the kind of rule.
type RuleType int

// String converts RuleType to string name
func (r RuleType) String() string {
	switch r {
	case RuleInclude:
		return "include"
	case RuleExclude:
		return "exclude"
	case RuleIncludeAll:
		return "include-all"
	case RuleExcludeAll:
		return "exclude-all"
	default:
		return "unknown"
	}
}

const (
	RuleInclude    RuleType = iota // --include <pattern>
	RuleExclude                    // --exclude <pattern>
	RuleIncludeAll                 // --include-all
	RuleExcludeAll                 // --exclude-all
)

// Scope defines what a rule pattern matches once it has matched a path.
type Scope int

const (
	// ScopeSubtree: a matched path also matches everything beneath it.
	// Used by --include, --exclude and the default patterns.
	ScopeSubtree Scope = iota
	// ScopeEntry: a pattern matches only the entry it names, not the entries
	// beneath it. Used by .gitignore lines, as git evaluates them.
	ScopeEntry
)

// Rule represents a single ordered filter rule.
type Rule struct {
	Type    RuleType
	Pattern string // only used for RuleInclude and RuleExclude
	Scope   Scope  // zero value is ScopeSubtree
	Source  string // display label of where the rule came from; never evaluated
}

type Rules []Rule

func NewRules() Rules {
	return Rules{}
}

func (r Rules) AddRules(rs Rules) Rules {
	return append(r, rs...)
}

func (r Rules) AddExcludeAll() Rules {
	return append(r, Rule{Type: RuleExcludeAll})
}

func (r Rules) AddExclude(p string) Rules {
	return append(r, Rule{Type: RuleExclude, Pattern: p})
}

func (r Rules) AddIncludeAll() Rules {
	return append(r, Rule{Type: RuleIncludeAll})
}

func (r Rules) AddInclude(p string) Rules {
	return append(r, Rule{Type: RuleInclude, Pattern: p})
}

// AddExcludeEntry adds an exclude rule with ScopeEntry.
func (r Rules) AddExcludeEntry(p string) Rules {
	return append(r, Rule{Type: RuleExclude, Pattern: p, Scope: ScopeEntry})
}

// AddIncludeEntry adds an include rule with ScopeEntry.
func (r Rules) AddIncludeEntry(p string) Rules {
	return append(r, Rule{Type: RuleInclude, Pattern: p, Scope: ScopeEntry})
}

// WithSource returns a copy of r in which every rule without a source label
// gets the given one.
func (r Rules) WithSource(source string) Rules {
	out := make(Rules, len(r))
	for i, rule := range r {
		if rule.Source == "" {
			rule.Source = source
		}
		out[i] = rule
	}
	return out
}

// Print writes the rules in evaluation order, one per line. The last column
// states what a rule does when its pattern matches a directory, which is the
// only place ScopeSubtree and ScopeEntry behave differently. When any rule
// carries a source label, a source column follows the pattern; otherwise
// .gitignore rules are marked at the end of the line.
func (r Rules) Print(w io.Writer, indent ...string) {
	prefix := ""
	if len(indent) > 0 {
		prefix = indent[0]
	}

	const (
		patternHeader  = "pattern"
		sourceHeader   = "source"
		coverageHeader = "on a matching directory"
	)

	width := len(patternHeader)
	sourceWidth := len(sourceHeader)
	labelled := false
	for _, rule := range r {
		width = max(width, len(rule.Pattern))
		sourceWidth = max(sourceWidth, len(rule.Source))
		labelled = labelled || rule.Source != ""
	}

	fmt.Fprintf(w, "%slast match wins\n", prefix)
	if labelled {
		fmt.Fprintf(w, "%s%-6s%-*s  %-*s  %s\n", prefix, "rule", width, patternHeader, sourceWidth, sourceHeader, coverageHeader)
	} else {
		fmt.Fprintf(w, "%s%-6s%-*s  %s\n", prefix, "rule", width, patternHeader, coverageHeader)
	}

	for _, rule := range r {
		sign := "[+]"
		if rule.Type == RuleExclude || rule.Type == RuleExcludeAll {
			sign = "[-]"
		}

		pattern, coverage := "ALL", ""
		if rule.Type == RuleInclude || rule.Type == RuleExclude {
			pattern, coverage = rule.Pattern, "with contents"
			if rule.Scope == ScopeEntry {
				coverage = "itself only"
			}
		}

		var line string
		switch {
		case labelled:
			line = fmt.Sprintf("%-6s%-*s  %-*s  %s", sign, width, pattern, sourceWidth, rule.Source, coverage)
		case rule.Scope == ScopeEntry && coverage != "":
			line = fmt.Sprintf("%-6s%-*s  %-*s  (.gitignore)", sign, width, pattern, len(coverageHeader), coverage)
		default:
			line = fmt.Sprintf("%-6s%-*s  %s", sign, width, pattern, coverage)
		}
		fmt.Fprintf(w, "%s%s\n", prefix, strings.TrimRight(line, " "))
	}
}
