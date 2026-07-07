package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lrndwy/gokil/internal/scaffold"
	"github.com/lrndwy/gokil/version"
)

func TestResolveGoModulePublished(t *testing.T) {
	if version.ModuleTag() == "dev" {
		t.Skip("running from framework source")
	}

	dir := t.TempDir()
	cfg := scaffold.ResolveGoModule(dir)
	if cfg.UseReplace {
		t.Fatalf("expected no local replace, got replace => %q", cfg.ReplacePath)
	}
	if cfg.RequireVersion == "" {
		t.Fatal("expected require version")
	}
}

func TestFindFrameworkRootFromCWD(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := scaffold.FindFrameworkRoot()
	if root == "" {
		t.Fatal("expected framework root when running tests inside repo")
	}
	if !strings.Contains(root, "gokil") {
		t.Fatalf("unexpected framework root: %s (cwd=%s)", root, wd)
	}
}

func TestResolveGoModuleLocalDev(t *testing.T) {
	if version.ModuleTag() != "dev" {
		t.Skip("release build")
	}

	root := scaffold.FindFrameworkRoot()
	if root == "" {
		t.Fatal("framework root not found")
	}

	projectDir := filepath.Join(t.TempDir(), "sampleapp")
	cfg := scaffold.ResolveGoModule(projectDir)
	if !cfg.UseReplace {
		t.Fatal("expected local replace when developing framework")
	}
	if cfg.ReplacePath == "" {
		t.Fatal("expected replace path")
	}
	if strings.Contains(cfg.ReplacePath, "pkg/mod") {
		t.Fatalf("replace path must not point to module cache: %s", cfg.ReplacePath)
	}
}
