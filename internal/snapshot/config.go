package snapshot

// Config holds the runtime configuration for a snapshot run.
type Config struct {
	SourceDir           string
	OutputPath          string
	IncludePatterns     []string
	ExcludePatterns     []string
	IncludeGitLog       bool
	DryRun              bool
	ForceTextPatterns   []string
	ForceBinaryPatterns []string
}
