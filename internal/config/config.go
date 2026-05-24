package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neox5/snp/internal/matcher"
)

const (
	DefaultOutputPath = "snapshot.snp"
)

type FlagType int

const (
	FlagTypeExcludeAll FlagType = iota
	FlagTypeIncludeAll
	FlagTypeInclude
	FlagTypeExclude
)

func (f FlagType) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *FlagType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*f = FlagTypeFromString(s)
	return nil
}

func (f FlagType) String() string {
	return [...]string{"exclude-all", "include-all", "include", "exclude"}[f]
}

func FlagTypeFromString(s string) FlagType {
	switch s {
	case "exclude-all":
		return FlagTypeExcludeAll
	case "include-all":
		return FlagTypeIncludeAll
	case "include":
		return FlagTypeInclude
	case "exclude":
		return FlagTypeExclude
	default:
		return FlagTypeExcludeAll
	}
}

type Flag struct {
	Type  FlagType `json:"type"`
	Value string   `json:"value"`
}

type Config struct {
	Generated           time.Time `json:"generated"`
	SourceDir           string    `json:"source_dir"`
	Depth               int       `json:"depth"`
	MatcherFlags        []Flag    `json:"filter_flags"`
	PickPaths           []string  `json:"pick_paths"`
	ForceTextPatterns   []string  `json:"force_text_patterns"`
	ForceBinaryPatterns []string  `json:"force_binary_patterns"`
	OutputPath          string    `json:"output_path"`
	NoSummary           bool      `json:"no_summary"`
	NoIndex             bool      `json:"no_index"`
	NoGitLog            bool      `json:"no_git_log"`
	NoContent           bool      `json:"no_content"`
	DryRun              bool      `json:"dry_run"`
	Silent              bool      `json:"silent"`
}

// Merge merges other over c and returns a new Config instance
func (c Config) Merge(other Config) *Config {
	depth := other.Depth
	if other.Depth == -1 {
		depth = c.Depth
	}

	out := other.OutputPath
	if other.OutputPath == "" {
		out = c.OutputPath
	}

	return &Config{
		Generated:           other.Generated,
		SourceDir:           other.SourceDir,
		Depth:               depth,
		MatcherFlags:        append(c.MatcherFlags, other.MatcherFlags...),
		PickPaths:           mergeUnique(c.PickPaths, other.PickPaths),
		ForceTextPatterns:   mergeUnique(c.ForceTextPatterns, other.ForceTextPatterns),
		ForceBinaryPatterns: mergeUnique(c.ForceBinaryPatterns, other.ForceBinaryPatterns),
		OutputPath:          out,
		NoSummary:           other.NoSummary,
		NoIndex:             other.NoIndex,
		NoGitLog:            other.NoGitLog,
		NoContent:           other.NoContent,
		DryRun:              other.DryRun,
		Silent:              other.Silent,
	}
}

func (c Config) Validate() error {
	if len(c.MatcherFlags) > 0 && len(c.PickPaths) > 0 {
		return fmt.Errorf("--pick cannot be combined with --include/exclude(-all)")
	}
	if len(c.PickPaths) > 0 && c.Depth > -1 {
		return fmt.Errorf("--depth cannot be combined with --pick")
	}
	return nil
}

func (c Config) buildFilterRules() matcher.Rules {
	r := buildExcludeDefaultRules(c.SourceDir) // apply default excludes first

	for _, f := range c.MatcherFlags {
		switch f.Type {
		case FlagTypeExcludeAll:
			r.AddExcludeAll()
			continue
		case FlagTypeIncludeAll:
			r.AddIncludeAll()
			continue
		case FlagTypeExclude:
			r.AddExclude(f.Value)
			continue
		case FlagTypeInclude:
			r.AddInclude(f.Value)
		}
	}

	return r
}

func mergeUnique(base, override []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range append(base, override...) {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
