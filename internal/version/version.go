// Package version holds the snp CLI version. It is overridden at build time
// via -ldflags in the Makefile, and falls back to module build metadata.
package version

import "runtime/debug"

// Version is the application version. It may be overridden at build time via
// -ldflags. When left as "dev", String() will attempt to read the module
// version from Go build info.
var Version = "dev"

// String returns the best available version string.
//
// Priority:
//  1. Explicit Version set via -ldflags (not "dev" and not empty)
//  2. Go module build info (debug.ReadBuildInfo) main module version
//  3. Fallback "dev"
func String() string {
	// 1) Explicitly injected version
	if Version != "" && Version != "dev" {
		return Version
	}

	// 2) Module build info (go install ...@vX.Y.Z)
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}

	// 3) Fallback
	return "dev"
}
