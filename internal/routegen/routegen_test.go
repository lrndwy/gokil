package routegen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lrndwy/gokil/internal/routegen"
)

func TestFolderToURLViaScan(t *testing.T) {
	dir := t.TempDir()
	writeRoute(t, dir, "users/route.go", "package users\nfunc GET() {}\nfunc POST() {}\n")
	writeRoute(t, dir, "users/_id/route.go", "package id\nfunc GET() {}\nfunc PUT() {}\nfunc DELETE() {}\n")
	writeRoute(t, dir, "posts/route.go", "package posts\nfunc GET() {}\n")

	routes, err := routegen.Scan(dir, "example")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, r := range routes {
		got[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{
		"GET /users",
		"POST /users",
		"GET /users/:id",
		"PUT /users/:id",
		"DELETE /users/:id",
		"GET /posts",
	} {
		if !got[want] {
			t.Fatalf("missing route %q in %#v", want, got)
		}
	}
}

func TestGenerate(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module demo\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	appDir := filepath.Join(root, "app")
	writeRoute(t, appDir, "users/route.go", "package users\nfunc GET() {}\nfunc POST() {}\n")
	writeRoute(t, appDir, "users/_id/route.go", "package id\nfunc GET() {}\nfunc DELETE() {}\n")

	out, n, err := routegen.Generate(root)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("route count = %d, want 4", n)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	src := string(data)
	for _, want := range []string{
		"package app",
		`framework.RegisterRoute("GET", "/users",`,
		`framework.RegisterRoute("POST", "/users",`,
		`framework.RegisterRoute("GET", "/users/:id",`,
		`framework.RegisterRoute("DELETE", "/users/:id",`,
		"demo/app/users",
		"demo/app/users/_id",
		"DO NOT EDIT",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("generated register.go missing %q:\n%s", want, src)
		}
	}
}

func writeRoute(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
