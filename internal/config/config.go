package config

import (
	"github.com/neox5/snp/internal/filter"
	"time"
)

const (
	ConfigFileName    = ".snpconfig"
	DefaultOutputPath = "snapshot.snp"
)

// Config is the single configuration type for a snapshot run.
type Config struct {
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
func (c Config) IsPick() bool {
	return len(c.PickPaths) > 0
}
