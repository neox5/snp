package snapshot

// Config holds the runtime configuration for a snapshot run.
type Config struct {
	SourceDir       string
	OutputPath      string
	IncludePatterns []string
	ExcludePatterns []string
	IncludeGitLog   bool
	// OutputExplicit indicates if --output was explicitly set.
	OutputExplicit bool
}
