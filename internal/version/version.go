// Package version holds the snap CLI version. It is overridden at build time
// via -ldflags in the Makefile.
package version

// Version is the application version. Default is "dev" for local builds
// without explicit version injection.
var Version = "dev"
