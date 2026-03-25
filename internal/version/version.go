package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using ldflags
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// Info returns version information as a string
func Info() string {
	return fmt.Sprintf("Version: %s\nCommit: %s\nBuild Time: %s\nGo Version: %s",
		Version, CommitSHA, BuildTime, GoVersion)
}

// Short returns a short version string
func Short() string {
	return Version
}
