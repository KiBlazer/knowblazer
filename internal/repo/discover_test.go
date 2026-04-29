package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFindsCurrentOrParentRepo(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := Init(root); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	nested := filepath.Join(root, "projects", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	got, err := Discover(Options{StartDir: nested})
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if got != root {
		t.Fatalf("Discover() = %s, want %s", got, root)
	}
}

func TestDiscoverUsesEnvironment(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := Init(root); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	got, err := Discover(Options{StartDir: t.TempDir(), EnvRepo: root})
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if got != root {
		t.Fatalf("Discover() = %s, want %s", got, root)
	}
}

func TestDiscoverErrorsWhenNoRepoFound(t *testing.T) {
	if _, err := Discover(Options{StartDir: t.TempDir()}); err == nil {
		t.Fatal("Discover() error = nil, want error")
	}
}
