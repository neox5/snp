package matcher

import "fmt"

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

// Rule represents a single ordered filter rule.
type Rule struct {
	Type    RuleType
	Pattern string // only used for RuleInclude and RuleExclude
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

func (r Rules) Print(indent ...string) {
	prefix := ""
	if len(indent) > 0 {
		prefix = indent[0]
	}
	for _, rule := range r {
		switch rule.Type {
		case RuleInclude:
			fmt.Printf("%s[+] %s\n", prefix, rule.Pattern)
		case RuleExclude:
			fmt.Printf("%s[-] %s\n", prefix, rule.Pattern)
		case RuleIncludeAll:
			fmt.Printf("%s[+] ALL\n", prefix)
		case RuleExcludeAll:
			fmt.Printf("%s[-] ALL\n", prefix)
		}
	}
}
