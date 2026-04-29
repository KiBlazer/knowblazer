package projectmap

import (
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
