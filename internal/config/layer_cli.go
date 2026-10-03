package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neox5/snp/internal/clirec"
)

// LayerFromOccurrences builds the CLI layer from the flags typed on the command
// line, in order. Only typed flags produce statements: scalars take the last
// occurrence, matcher flags keep their order, pick and force lists collect
// every occurrence. --only-<section> expands to --no-<other sections>.
func LayerFromOccurrences(occs []clirec.Occurrence) (Layer, error) {
	var l Layer
	only := map[string]bool{}

	for _, o := range occs {
		switch o.Flag {
		case "depth":
			n, err := strconv.Atoi(o.Value)
			if err != nil {
				return Layer{}, fmt.Errorf("--depth: invalid integer %q", o.Value)
			}
			l.Depth = &n
		case "output":
			v := o.Value
			l.OutputPath = &v
		case "include":
			l.MatcherFlags = append(l.MatcherFlags, Flag{Type: FlagTypeInclude, Value: o.Value})
		case "exclude":
			l.MatcherFlags = append(l.MatcherFlags, Flag{Type: FlagTypeExclude, Value: o.Value})
		case "include-all", "exclude-all":
			on, err := boolValue(o)
			if err != nil {
				return Layer{}, err
			}
			if on {
				t := FlagTypeIncludeAll
				if o.Flag == "exclude-all" {
					t = FlagTypeExcludeAll
				}
				l.MatcherFlags = append(l.MatcherFlags, Flag{Type: t})
			}
		case "pick":
			l.PickPaths = append(l.PickPaths, splitList(o.Value)...)
		case "force-text":
			l.ForceTextPatterns = append(l.ForceTextPatterns, splitList(o.Value)...)
		case "force-binary":
			l.ForceBinaryPatterns = append(l.ForceBinaryPatterns, splitList(o.Value)...)
		case "no-summary", "no-index", "no-git-log", "no-content", "stdout", "silent":
			on, err := boolValue(o)
			if err != nil {
				return Layer{}, err
			}
			*boolMember(&l, o.Flag) = &on
		case "only-summary", "only-index", "only-git-log", "only-content":
			on, err := boolValue(o)
			if err != nil {
				return Layer{}, err
			}
			only[strings.TrimPrefix(o.Flag, "only-")] = on
		}
	}

	expandOnly(&l, only)
	return l, nil
}

// expandOnly turns --only-<section> into statements about the sections: every
// section that is not selected is suppressed; a selected section is stated as
// shown unless --no-<section> was typed explicitly.
func expandOnly(l *Layer, only map[string]bool) {
	has := false
	for _, on := range only {
		has = has || on
	}
	if !has {
		return
	}
	for _, section := range []string{"summary", "index", "git-log", "content"} {
		m := boolMember(l, "no-"+section)
		if only[section] {
			if *m == nil {
				shown := false
				*m = &shown
			}
			continue
		}
		suppressed := true
		*m = &suppressed
	}
}

// boolMember returns the Layer member behind a boolean flag name.
func boolMember(l *Layer, flag string) **bool {
	switch flag {
	case "no-summary":
		return &l.NoSummary
	case "no-index":
		return &l.NoIndex
	case "no-git-log":
		return &l.NoGitLog
	case "no-content":
		return &l.NoContent
	case "stdout":
		return &l.Stdout
	default:
		return &l.Silent
	}
}

func boolValue(o clirec.Occurrence) (bool, error) {
	b, err := strconv.ParseBool(o.Value)
	if err != nil {
		return false, fmt.Errorf("--%s: invalid boolean %q", o.Flag, o.Value)
	}
	return b, nil
}

// splitList splits a comma separated flag value, as urfave does for slice
// flags, and drops empty parts.
func splitList(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
