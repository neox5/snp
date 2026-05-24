// Package filter aggregates ordered rules into a matcher that decides
// whether a given path should be included in a snapshot.
package matcher

// RuleType defines the kind of rule.
type RuleType int

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
