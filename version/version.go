package version

import (
	"runtime/debug"
	"strings"
)

const ModulePath = "github.com/lrndwy/gokil"

// Info holds version metadata resolved at runtime.
type Info struct {
	Version string
	Commit  string
	Date    string
	Dirty   bool
}

// Get returns version metadata from the running binary.
func Get() Info {
	info := Info{Version: "dev"}

	build, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	info.Version = resolveModuleVersion(build)
	readVCS(&info, build)

	return info
}

// Module returns the module version tag (e.g. v0.1.2).
func Module() string {
	return Get().Version
}

// ModuleTag returns a clean semver tag safe for go.mod (strips +dirty metadata).
func ModuleTag() string {
	v := Module()
	if v == "dev" {
		return v
	}
	if i := strings.Index(v, "+"); i >= 0 {
		return v[:i]
	}
	return v
}

// RequireVersion returns the version string for go.mod require directives.
func RequireVersion() string {
	tag := ModuleTag()
	if tag == "dev" {
		return "latest"
	}
	return tag
}

func resolveModuleVersion(build *debug.BuildInfo) string {
	if build.Main.Path == ModulePath && isReleaseVersion(build.Main.Version) {
		return build.Main.Version
	}
	for _, dep := range build.Deps {
		if dep.Path == ModulePath && isReleaseVersion(dep.Version) {
			return dep.Version
		}
	}
	if build.Main.Path == ModulePath && build.Main.Version != "" && build.Main.Version != "(devel)" {
		return build.Main.Version
	}
	return "dev"
}

func isReleaseVersion(v string) bool {
	return v != "" && v != "(devel)"
}

func readVCS(info *Info, build *debug.BuildInfo) {
	for _, s := range build.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) > 7 {
				info.Commit = s.Value[:7]
			} else {
				info.Commit = s.Value
			}
		case "vcs.time":
			info.Date = s.Value
		case "vcs.modified":
			info.Dirty = s.Value == "true"
		}
	}
}

// String formats version info for CLI output.
func (i Info) String() string {
	v := i.Version
	if i.Dirty && !strings.Contains(v, "+") {
		v += "+dirty"
	}
	out := "gokil " + v
	if i.Commit != "" {
		out += " (" + i.Commit
		if i.Dirty {
			out += ", dirty"
		}
		out += ")"
	}
	if i.Date != "" {
		out += " built " + i.Date
	}
	return out
}
