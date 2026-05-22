package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neox5/snp/internal/filter"
)

const ConfigFileName = ".snpconfig"

// Load reads .snpconfig from dir and returns ordered filter rules.
// Returns nil rules (not an error) if the file does not exist.
func Load(dir string) ([]filter.Rule, error) {
	path := filepath.Join(dir, ConfigFileName)

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", ConfigFileName, err)
	}
	defer f.Close()

	var rules []filter.Rule
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule, err := parseLine(line, lineNum)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", ConfigFileName, err)
	}

	return rules, nil
}

// Save writes rules to .snpconfig in dir, overwriting any existing file.
func Save(dir string, rules []filter.Rule) error {
	path := filepath.Join(dir, ConfigFileName)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", ConfigFileName, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	for _, r := range rules {
		line, err := serializeRule(r)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return fmt.Errorf("writing %s: %w", ConfigFileName, err)
		}
	}

	return w.Flush()
}

// HasBaseline reports whether any rule in rules is a baseline rule.
func HasBaseline(rules []filter.Rule) bool {
	for _, r := range rules {
		switch r.Type {
		case filter.RuleIncludeAll, filter.RuleExcludeAll, filter.RuleExcludeDefaults:
			return true
		}
	}
	return false
}

func parseLine(line string, lineNum int) (filter.Rule, error) {
	switch line {
	case "include-all":
		return filter.Rule{Type: filter.RuleIncludeAll}, nil
	case "exclude-all":
		return filter.Rule{Type: filter.RuleExcludeAll}, nil
	case "exclude-defaults":
		return filter.Rule{Type: filter.RuleExcludeDefaults}, nil
	}

	if pattern, ok := strings.CutPrefix(line, "include "); ok {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			return filter.Rule{}, fmt.Errorf("%s line %d: include requires a pattern", ConfigFileName, lineNum)
		}
		return filter.Rule{Type: filter.RuleInclude, Pattern: pattern}, nil
	}

	if pattern, ok := strings.CutPrefix(line, "exclude "); ok {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			return filter.Rule{}, fmt.Errorf("%s line %d: exclude requires a pattern", ConfigFileName, lineNum)
		}
		return filter.Rule{Type: filter.RuleExclude, Pattern: pattern}, nil
	}

	return filter.Rule{}, fmt.Errorf("%s line %d: unknown directive %q", ConfigFileName, lineNum, line)
}

func serializeRule(r filter.Rule) (string, error) {
	switch r.Type {
	case filter.RuleIncludeAll:
		return "include-all", nil
	case filter.RuleExcludeAll:
		return "exclude-all", nil
	case filter.RuleExcludeDefaults:
		return "exclude-defaults", nil
	case filter.RuleInclude:
		return "include " + r.Pattern, nil
	case filter.RuleExclude:
		return "exclude " + r.Pattern, nil
	default:
		return "", fmt.Errorf("unknown rule type: %d", r.Type)
	}
}
