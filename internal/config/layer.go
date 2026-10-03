package config

// Layer is a partial statement of settings: one source (defaults, the config
// file, the command line) says what it sets and leaves the rest unsaid.
// Scalars are pointers (nil = not stated); lists are nil when not stated.
type Layer struct {
	Depth               *int     `json:"depth,omitempty"`
	MatcherFlags        []Flag   `json:"filter_flags,omitempty"`
	PickPaths           []string `json:"pick_paths,omitempty"`
	ForceTextPatterns   []string `json:"force_text_patterns,omitempty"`
	ForceBinaryPatterns []string `json:"force_binary_patterns,omitempty"`
	OutputPath          *string  `json:"output_path,omitempty"`
	NoSummary           *bool    `json:"no_summary,omitempty"`
	NoIndex             *bool    `json:"no_index,omitempty"`
	NoGitLog            *bool    `json:"no_git_log,omitempty"`
	NoContent           *bool    `json:"no_content,omitempty"`
	Stdout              *bool    `json:"stdout,omitempty"`
	Silent              *bool    `json:"silent,omitempty"`
}

// Source names for the layers snp resolves.
const (
	SourceDefault = "default"
	SourceFile    = "file"
	SourceCLI     = "cli"
)

// Source is a named layer. Sources are resolved lowest precedence first.
type Source struct {
	Name  string
	Layer Layer
}

// DefaultLayer returns the built-in defaults.
func DefaultLayer() Layer {
	depth := -1
	out := DefaultOutputPath
	return Layer{Depth: &depth, OutputPath: &out}
}
