package matcher

import (
	"fmt"
	"io"
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

// Print writes the rules in evaluation order, one per line. The last column
// states what a rule does when its pattern matches a directory, which is the
// only place ScopeSubtree and ScopeEntry behave differently.
func (r Rules) Print(w io.Writer, indent ...string) {
	prefix := ""
	if len(indent) > 0 {
		prefix = indent[0]
	}

	const (
		patternHeader  = "pattern"
		coverageHeader = "on a matching directory"
	)

	width := len(patternHeader)
	for _, rule := range r {
		if len(rule.Pattern) > width {
			width = len(rule.Pattern)
		}
	}

	fmt.Fprintf(w, "%slast match wins\n", prefix)
	fmt.Fprintf(w, "%s%-6s%-*s  %s\n", prefix, "rule", width, patternHeader, coverageHeader)

	for _, rule := range r {
		switch rule.Type {
		case RuleIncludeAll:
			fmt.Fprintf(w, "%s%-6sALL\n", prefix, "[+]")
		case RuleExcludeAll:
			fmt.Fprintf(w, "%s%-6sALL\n", prefix, "[-]")
		case RuleInclude, RuleExclude:
			sign := "[+]"
			if rule.Type == RuleExclude {
				sign = "[-]"
			}
			if rule.Scope == ScopeEntry {
				fmt.Fprintf(w, "%s%-6s%-*s  %-*s  (.gitignore)\n", prefix, sign, width, rule.Pattern, len(coverageHeader), "itself only")
				continue
			}
			fmt.Fprintf(w, "%s%-6s%-*s  with contents\n", prefix, sign, width, rule.Pattern)
		}
	}
}
