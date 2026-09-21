package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Get returns the cleaned version string without 'v' prefix (e.g. "0.2.0" or "dev").
func Get() string {
	if Version != "" && Version != "dev" {
		return strings.TrimPrefix(Version, "v")
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}

// FullString returns formatted version information.
func FullString() string {
	v := Get()
	res := fmt.Sprintf("knowblazer v%s (%s/%s)", v, runtime.GOOS, runtime.GOARCH)
	if Commit != "" && Commit != "none" {
		res += fmt.Sprintf(" commit=%s", Commit)
	}
	if Date != "" && Date != "unknown" {
		res += fmt.Sprintf(" built=%s", Date)
	}
	return res
}
