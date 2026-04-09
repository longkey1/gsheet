package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// These variables are set at build time using ldflags
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if Version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if CommitSHA == "unknown" && s.Value != "" {
				CommitSHA = s.Value
			}
		case "vcs.time":
			if BuildTime == "unknown" && s.Value != "" {
				BuildTime = s.Value
			}
		}
	}
}

// Info returns version information as a string
func Info() string {
	return fmt.Sprintf("Version: %s\nCommit: %s\nBuild Time: %s\nGo Version: %s",
		Version, CommitSHA, BuildTime, GoVersion)
}

// Short returns a short version string
func Short() string {
	return Version
}
