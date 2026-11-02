package projectmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestSetListResolveAndClear(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	workspace := filepath.Join(t.TempDir(), "work", "project")
	child := filepath.Join(workspace, "src")

	if err := Set(root, workspace, "kiblazer"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	mappings, err := List(root)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(mappings) != 1 || mappings[0].Project != "kiblazer" {
		t.Fatalf("Mappings = %#v", mappings)
	}
	project, ok, err := Resolve(root, child)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !ok || project != "kiblazer" {
		t.Fatalf("Resolve() = %q, %v", project, ok)
	}
	if err := Clear(root, workspace); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	mappings, err = List(root)
	if err != nil {
		t.Fatalf("List() after clear error = %v", err)
	}
	if len(mappings) != 0 {
		t.Fatalf("Mappings after clear = %#v", mappings)
	}
}

func TestLocalMappingsMigrateAndAreIsolated(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(root, ".knowblazer", "projects.json")
	original := []byte(`{"mappings":[{"path":"/old/work","project":"old"}]}`)
	if err := os.WriteFile(legacy, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "/new/work", "new"); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(legacy); err != nil || string(got) != string(original) {
		t.Fatalf("legacy changed: %s, %v", got, err)
	}
	if err := Clear(root, "/old/work"); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := Resolve(root, "/old/work"); err != nil || ok {
		t.Fatalf("legacy mapping resurrected: %v %v", ok, err)
	}
	if got, ok, err := Resolve(root, "/new/work/src"); err != nil || !ok || got != "new" {
		t.Fatalf("resolve = %q %v %v", got, ok, err)
	}
	other := filepath.Join(t.TempDir(), "other")
	if _, ok, err := Resolve(other, "/new/work"); err != nil || ok {
		t.Fatalf("mapping leaked: %v %v", ok, err)
	}
}

func TestResolveUsesDirectoryBoundariesAndMostSpecificMapping(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	if err := Set(root, "/work/app", "app"); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "/work/app/sub", "sub"); err != nil {
		t.Fatal(err)
	}
	if got, ok, err := Resolve(root, "/work/app/sub/file"); err != nil || !ok || got != "sub" {
		t.Fatalf("resolve = %q %v %v", got, ok, err)
	}
	if _, ok, err := Resolve(root, "/work/app2"); err != nil || ok {
		t.Fatalf("sibling matched: %v %v", ok, err)
	}
	if got, ok, err := Resolve(root, "/work/app/..hidden"); err != nil || !ok || got != "app" {
		t.Fatalf("dot directory = %q %v %v", got, ok, err)
	}
}
