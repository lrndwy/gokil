package version_test

import (
	"strings"
	"testing"

	"github.com/lrndwy/gokil/version"
)

func TestModulePath(t *testing.T) {
	if version.ModulePath != "github.com/lrndwy/gokil" {
		t.Fatalf("unexpected module path: %s", version.ModulePath)
	}
}

func TestModuleNotEmpty(t *testing.T) {
	if version.Module() == "" {
		t.Fatal("expected non-empty version")
	}
}

func TestModuleTagStripsMetadata(t *testing.T) {
	// When running from source, version may be dev or vX.Y.Z+dirty
	tag := version.ModuleTag()
	if strings.Contains(tag, "+") {
		t.Fatalf("ModuleTag should not contain build metadata: %s", tag)
	}
}

func TestStringFormat(t *testing.T) {
	s := version.Info{Version: "v0.1.2", Commit: "abc1234"}.String()
	if !strings.Contains(s, "gokil v0.1.2") {
		t.Fatalf("unexpected format: %s", s)
	}
}
