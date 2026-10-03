package config

// Params captures the runtime-only input of a snapshot run: what to snapshot
// and how to run, not what the snapshot contains. Persistable settings travel
// as a Layer.
type Params struct {
	SourceDir    string
	NoConfig     bool
	SaveConfig   bool
	ShowConfig   bool
	DryRun       bool
	VerboseLevel int
}
