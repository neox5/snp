package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/neox5/snp/internal/filter"
)

// Load reads .snpconfig from dir and returns a Config.
// Returns empty Config (not an error) if the file does not exist.
func Load(dir string) (Config, error) {
	path := filepath.Join(dir, ConfigFileName)

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return newConfig(), nil
	}
	if err != nil {
		return newConfig(), fmt.Errorf("opening %s: %w", ConfigFileName, err)
	}
	defer f.Close()

	cfg := newConfig()
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := parseLine(line, lineNum, &cfg); err != nil {
			return newConfig(), err
		}
	}

	if err := scanner.Err(); err != nil {
		return newConfig(), fmt.Errorf("reading %s: %w", ConfigFileName, err)
	}

	return cfg, nil
}

// Save writes cfg to .snpconfig in dir, overwriting any existing file.
func (cfg Config) Save(dir string) error {
	path := filepath.Join(dir, ConfigFileName)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", ConfigFileName, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	writeLine := func(s string) {
		fmt.Fprintln(w, s)
	}

	writeLine("generated " + cfg.Generated.UTC().Format(time.RFC3339))

	if cfg.Depth >= 0 {
		writeLine(fmt.Sprintf("depth %d", cfg.Depth))
	}

	for _, r := range cfg.FilterRules {
		if line, err := serializeRule(r); err == nil {
			writeLine(line)
		}
	}

	for _, p := range cfg.PickPaths {
		writeLine("pick " + p)
	}

	if cfg.NoSummary {
		writeLine("no-summary")
	}
	if cfg.NoIndex {
		writeLine("no-index")
	}
	if cfg.NoGitLog {
		writeLine("no-git-log")
	}
	if cfg.NoContent {
		writeLine("no-content")
	}
	if cfg.Silent {
		writeLine("silent")
	}
	if cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath {
		writeLine("output " + cfg.OutputPath)
	}

	for _, p := range cfg.ForceTextPatterns {
		writeLine("force-text " + p)
	}
	for _, p := range cfg.ForceBinaryPatterns {
		writeLine("force-binary " + p)
	}

	return w.Flush()
}

// — unexported helpers ————————————————————————————————————————————————————

func parseLine(line string, lineNum int, cfg *Config) error {
	if val, ok := strings.CutPrefix(line, "generated "); ok {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(val))
		if err != nil {
			return fmt.Errorf("%s line %d: invalid generated timestamp: %w", ConfigFileName, lineNum, err)
		}
		cfg.Generated = t
		return nil
	}

	if val, ok := strings.CutPrefix(line, "depth "); ok {
		d, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil || d < 0 {
			return fmt.Errorf("%s line %d: depth must be a non-negative integer", ConfigFileName, lineNum)
		}
		cfg.Depth = d
		return nil
	}

	if val, ok := strings.CutPrefix(line, "output "); ok {
		cfg.OutputPath = strings.TrimSpace(val)
		return nil
	}

	if val, ok := strings.CutPrefix(line, "pick "); ok {
		val = strings.TrimSpace(val)
		if val == "" {
			return fmt.Errorf("%s line %d: pick requires a path", ConfigFileName, lineNum)
		}
		cfg.PickPaths = append(cfg.PickPaths, val)
		return nil
	}

	if val, ok := strings.CutPrefix(line, "force-text "); ok {
		val = strings.TrimSpace(val)
		if val == "" {
			return fmt.Errorf("%s line %d: force-text requires a pattern", ConfigFileName, lineNum)
		}
		cfg.ForceTextPatterns = append(cfg.ForceTextPatterns, val)
		return nil
	}

	if val, ok := strings.CutPrefix(line, "force-binary "); ok {
		val = strings.TrimSpace(val)
		if val == "" {
			return fmt.Errorf("%s line %d: force-binary requires a pattern", ConfigFileName, lineNum)
		}
		cfg.ForceBinaryPatterns = append(cfg.ForceBinaryPatterns, val)
		return nil
	}

	switch line {
	case "include-all":
		cfg.FilterRules = append(cfg.FilterRules, filter.Rule{Type: filter.RuleIncludeAll})
	case "exclude-all":
		cfg.FilterRules = append(cfg.FilterRules, filter.Rule{Type: filter.RuleExcludeAll})
	case "exclude-defaults":
		cfg.FilterRules = append(cfg.FilterRules, filter.Rule{Type: filter.RuleExcludeDefaults})
	case "no-summary":
		cfg.NoSummary = true
	case "no-index":
		cfg.NoIndex = true
	case "no-git-log":
		cfg.NoGitLog = true
	case "no-content":
		cfg.NoContent = true
	case "silent":
		cfg.Silent = true
	default:
		if val, ok := strings.CutPrefix(line, "include "); ok {
			val = strings.TrimSpace(val)
			if val == "" {
				return fmt.Errorf("%s line %d: include requires a pattern", ConfigFileName, lineNum)
			}
			cfg.FilterRules = append(cfg.FilterRules, filter.Rule{Type: filter.RuleInclude, Pattern: val})
			return nil
		}
		if val, ok := strings.CutPrefix(line, "exclude "); ok {
			val = strings.TrimSpace(val)
			if val == "" {
				return fmt.Errorf("%s line %d: exclude requires a pattern", ConfigFileName, lineNum)
			}
			cfg.FilterRules = append(cfg.FilterRules, filter.Rule{Type: filter.RuleExclude, Pattern: val})
			return nil
		}
		return fmt.Errorf("%s line %d: unknown directive %q", ConfigFileName, lineNum, line)
	}

	return nil
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
		if r.Pattern == "" {
			return "", fmt.Errorf("include rule missing pattern")
		}
		return "include " + r.Pattern, nil
	case filter.RuleExclude:
		if r.Pattern == "" {
			return "", fmt.Errorf("exclude rule missing pattern")
		}
		return "exclude " + r.Pattern, nil
	default:
		return "", fmt.Errorf("unknown rule type: %v", r.Type)
	}
}
