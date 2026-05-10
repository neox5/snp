package snapshot

import "github.com/neox5/snp/internal/filter"

// Mode defines the operating mode of a snapshot run.
type Mode int

const (
	ModeTraversal Mode = iota
	ModePick
)

// Config holds the runtime configuration for a snapshot run.
type Config struct {
	Mode          Mode
	SourceDir     string
	OutputPath    string
	IncludeGitLog bool
	DryRun        bool

	// Mode 1 — Traversal
	FilterRules         []filter.Rule
	ForceTextPatterns   []string
	ForceBinaryPatterns []string

	// Mode 2 — Pick
	PickPaths []string
}
