package scaffold

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/lrndwy/gokil/version"
)

const frameworkModule = "github.com/lrndwy/gokil"

// GoModuleConfig describes how a new project should depend on gokil.
type GoModuleConfig struct {
	RequireVersion string
	ReplacePath    string
	UseReplace     bool
}

// ResolveGoModule decides whether to use a local replace directive or a published module version.
func ResolveGoModule(projectDir string) GoModuleConfig {
	requireVersion := version.RequireVersion()
	cfg := GoModuleConfig{RequireVersion: requireVersion}

	if shouldUsePublishedModule() {
		if requireVersion == "latest" {
			cfg.RequireVersion = "v0.0.0"
		}
		return cfg
	}

	root := FindFrameworkRoot()
	if root == "" || isModuleCache(root) {
		if requireVersion == "latest" {
			cfg.RequireVersion = "v0.0.0"
		}
		return cfg
	}

	absProject, err := filepath.Abs(projectDir)
	if err != nil {
		return cfg
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return cfg
	}
	rel, err := filepath.Rel(absProject, absRoot)
	if err != nil || rel == "." {
		return cfg
	}

	cfg.UseReplace = true
	cfg.ReplacePath = filepath.ToSlash(rel)
	cfg.RequireVersion = "v0.0.0"
	return cfg
}

func shouldUsePublishedModule() bool {
	tag := version.ModuleTag()
	if tag != "dev" && !isDevelBuild() {
		return true
	}
	return false
}

func isDevelBuild() bool {
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return false
	}
	v := build.Main.Version
	return v == "(devel)" || v == "" || strings.HasPrefix(v, "v0.0.0")
}

// FindFrameworkRoot locates a local checkout of github.com/lrndwy/gokil.
func FindFrameworkRoot() string {
	if env := strings.TrimSpace(os.Getenv("GOKIL_FRAMEWORK_ROOT")); env != "" {
		if isFrameworkModuleRoot(env) {
			return env
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if root := findModuleRoot(wd); root != "" {
			return root
		}
	}

	if exe, err := os.Executable(); err == nil {
		if root := findModuleRoot(filepath.Dir(exe)); root != "" {
			return root
		}
	}

	return ""
}

func findModuleRoot(start string) string {
	dir, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	for {
		if isFrameworkModuleRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isFrameworkModuleRoot(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "module "+frameworkModule {
			return true
		}
	}
	return false
}

func isModuleCache(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return strings.Contains(clean, "/pkg/mod/")
}
