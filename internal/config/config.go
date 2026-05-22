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

const (
	ConfigFileName    = ".snpconfig"
	DefaultOutputPath = "snapshot.snp"
)

// FullConfig is the single configuration type for a snapshot run.
// It is used across CLI parsing, config file I/O, and snapshot execution.
type FullConfig struct {
	Generated time.Time
	// depth — -1 means full traversal
	Depth int
	// file selection — traversal mode (ordered, last-match-wins)
	FilterRules []filter.Rule
	// file selection — pick mode (mutually exclusive with FilterRules)
	PickPaths []string
	// binary overrides
	ForceTextPatterns   []string
	ForceBinaryPatterns []string
	// output
	SourceDir  string
	OutputPath string
	NoSummary  bool
	NoIndex    bool
	NoGitLog   bool
	NoContent  bool
	DryRun     bool
	Silent     bool
}

// IsPick reports whether the config is in pick mode.
func (c FullConfig) IsPick() bool {
	return len(c.PickPaths) > 0
}

// ApplyDefaults fills in default values for any unset fields.
func ApplyDefaults(cfg *FullConfig) {
	if cfg.OutputPath == "" {
		cfg.OutputPath = DefaultOutputPath
	}
}

// Load reads .snpconfig from dir and returns a FullConfig.
// Returns empty FullConfig (not an error) if the file does not exist.
// SourceDir and DryRun are not persisted — caller sets them after load.
func Load(dir string) (FullConfig, error) {
	path := filepath.Join(dir, ConfigFileName)

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return FullConfig{}, nil
	}
	if err != nil {
		return FullConfig{}, fmt.Errorf("opening %s: %w", ConfigFileName, err)
	}
	defer f.Close()

	var cfg FullConfig
	cfg.Depth = -1
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if err := parseLine(line, lineNum, &cfg); err != nil {
			return FullConfig{}, err
		}
	}

	if err := scanner.Err(); err != nil {
		return FullConfig{}, fmt.Errorf("reading %s: %w", ConfigFileName, err)
	}

	return cfg, nil
}

// Save writes cfg to .snpconfig in dir, overwriting any existing file.
// Only non-default values are written.
func Save(dir string, cfg FullConfig) error {
	path := filepath.Join(dir, ConfigFileName)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", ConfigFileName, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	// generated timestamp
	ts := cfg.Generated
	if ts.IsZero() {
		ts = time.Now()
	}
	if err := writeLine(w, "generated "+ts.Format(time.RFC3339)); err != nil {
		return err
	}

	// depth — only if not full traversal
	if cfg.Depth >= 0 {
		if err := writeLine(w, fmt.Sprintf("depth %d", cfg.Depth)); err != nil {
			return err
		}
	}

	// filter rules (traversal, ordered)
	for _, r := range cfg.FilterRules {
		line, err := serializeRule(r)
		if err != nil {
			return err
		}
		if err := writeLine(w, line); err != nil {
			return err
		}
	}

	// pick paths
	for _, p := range cfg.PickPaths {
		if err := writeLine(w, "pick "+p); err != nil {
			return err
		}
	}

	// output path — only if non-default
	if cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath {
		if err := writeLine(w, "output "+cfg.OutputPath); err != nil {
			return err
		}
	}

	// boolean output flags — only if true
	if cfg.NoSummary {
		if err := writeLine(w, "no-summary"); err != nil {
			return err
		}
	}
	if cfg.NoIndex {
		if err := writeLine(w, "no-index"); err != nil {
			return err
		}
	}
	if cfg.NoGitLog {
		if err := writeLine(w, "no-git-log"); err != nil {
			return err
		}
	}
	if cfg.NoContent {
		if err := writeLine(w, "no-content"); err != nil {
			return err
		}
	}
	if cfg.Silent {
		if err := writeLine(w, "silent"); err != nil {
			return err
		}
	}

	// force overrides
	for _, p := range cfg.ForceTextPatterns {
		if err := writeLine(w, "force-text "+p); err != nil {
			return err
		}
	}
	for _, p := range cfg.ForceBinaryPatterns {
		if err := writeLine(w, "force-binary "+p); err != nil {
			return err
		}
	}

	return w.Flush()
}

// Print writes a human-readable representation of cfg to stdout,
// including the equivalent snp command.
func Print(cfg FullConfig) {
	// generated
	if !cfg.Generated.IsZero() {
		fmt.Println("# generated: " + cfg.Generated.Local().Format("2006-01-02 15:04:05"))
	}

	// depth
	if cfg.Depth >= 0 {
		fmt.Println()
		fmt.Println("# depth")
		fmt.Printf("depth %d\n", cfg.Depth)
	}

	// filters
	if len(cfg.FilterRules) > 0 {
		fmt.Println()
		fmt.Println("# filters")
		for _, r := range cfg.FilterRules {
			line, _ := serializeRule(r)
			fmt.Println(line)
		}
	}

	// pick paths
	if len(cfg.PickPaths) > 0 {
		fmt.Println()
		fmt.Println("# pick")
		for _, p := range cfg.PickPaths {
			fmt.Println("pick " + p)
		}
	}

	// output section
	hasOutput := (cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath) ||
		cfg.Silent || cfg.NoSummary || cfg.NoIndex || cfg.NoGitLog || cfg.NoContent
	if hasOutput {
		fmt.Println()
		fmt.Println("# output")
		if cfg.NoSummary {
			fmt.Println("no-summary")
		}
		if cfg.NoIndex {
			fmt.Println("no-index")
		}
		if cfg.NoGitLog {
			fmt.Println("no-git-log")
		}
		if cfg.NoContent {
			fmt.Println("no-content")
		}
		if cfg.Silent {
			fmt.Println("silent")
		}
		if cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath {
			fmt.Println("output " + cfg.OutputPath)
		}
	}

	// overrides
	if len(cfg.ForceTextPatterns) > 0 || len(cfg.ForceBinaryPatterns) > 0 {
		fmt.Println()
		fmt.Println("# overrides")
		for _, p := range cfg.ForceTextPatterns {
			fmt.Println("force-text " + p)
		}
		for _, p := range cfg.ForceBinaryPatterns {
			fmt.Println("force-binary " + p)
		}
	}

	// equivalent command
	fmt.Println()
	fmt.Println("# equivalent command")
	fmt.Println(buildCommand(cfg))
}

// HasBaseline reports whether any rule in cfg.FilterRules is a baseline rule.
func HasBaseline(cfg FullConfig) bool {
	for _, r := range cfg.FilterRules {
		switch r.Type {
		case filter.RuleIncludeAll, filter.RuleExcludeAll, filter.RuleExcludeDefaults:
			return true
		}
	}
	return false
}

func parseLine(line string, lineNum int, cfg *FullConfig) error {
	// generated
	if val, ok := strings.CutPrefix(line, "generated "); ok {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(val))
		if err != nil {
			return fmt.Errorf("%s line %d: invalid generated timestamp: %w", ConfigFileName, lineNum, err)
		}
		cfg.Generated = t
		return nil
	}

	// depth
	if val, ok := strings.CutPrefix(line, "depth "); ok {
		val = strings.TrimSpace(val)
		d, err := strconv.Atoi(val)
		if err != nil || d < 0 {
			return fmt.Errorf("%s line %d: depth must be a non-negative integer", ConfigFileName, lineNum)
		}
		cfg.Depth = d
		return nil
	}

	// output path
	if val, ok := strings.CutPrefix(line, "output "); ok {
		cfg.OutputPath = strings.TrimSpace(val)
		return nil
	}

	// pick path
	if val, ok := strings.CutPrefix(line, "pick "); ok {
		val = strings.TrimSpace(val)
		if val == "" {
			return fmt.Errorf("%s line %d: pick requires a path", ConfigFileName, lineNum)
		}
		cfg.PickPaths = append(cfg.PickPaths, val)
		return nil
	}

	// force overrides
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

	// filter rules and boolean flags
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
		return "include " + r.Pattern, nil
	case filter.RuleExclude:
		return "exclude " + r.Pattern, nil
	default:
		return "", fmt.Errorf("unknown rule type: %d", r.Type)
	}
}

func writeLine(w *bufio.Writer, line string) error {
	_, err := fmt.Fprintln(w, line)
	return err
}

func buildCommand(cfg FullConfig) string {
	var parts []string
	parts = append(parts, "snp")

	if cfg.Depth >= 0 {
		parts = append(parts, "--depth", fmt.Sprintf("%d", cfg.Depth))
	}

	for _, r := range cfg.FilterRules {
		switch r.Type {
		case filter.RuleIncludeAll:
			parts = append(parts, "--include-all")
		case filter.RuleExcludeAll:
			parts = append(parts, "--exclude-all")
		case filter.RuleExcludeDefaults:
			parts = append(parts, "--exclude-defaults")
		case filter.RuleInclude:
			parts = append(parts, "--include", quoteArg(r.Pattern))
		case filter.RuleExclude:
			parts = append(parts, "--exclude", quoteArg(r.Pattern))
		}
	}

	for _, p := range cfg.PickPaths {
		parts = append(parts, "--pick", quoteArg(p))
	}

	if cfg.NoSummary {
		parts = append(parts, "--no-summary")
	}
	if cfg.NoIndex {
		parts = append(parts, "--no-index")
	}
	if cfg.NoGitLog {
		parts = append(parts, "--no-git-log")
	}
	if cfg.NoContent {
		parts = append(parts, "--no-content")
	}
	if cfg.Silent {
		parts = append(parts, "--silent")
	}
	if cfg.OutputPath != "" && cfg.OutputPath != DefaultOutputPath {
		parts = append(parts, "--output", quoteArg(cfg.OutputPath))
	}
	for _, p := range cfg.ForceTextPatterns {
		parts = append(parts, "--force-text", quoteArg(p))
	}
	for _, p := range cfg.ForceBinaryPatterns {
		parts = append(parts, "--force-binary", quoteArg(p))
	}

	return strings.Join(parts, " ")
}

func quoteArg(s string) string {
	if strings.ContainsAny(s, " \t*?[]{}") {
		return `"` + s + `"`
	}
	return s
}
