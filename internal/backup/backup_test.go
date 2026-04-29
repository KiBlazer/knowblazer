package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestCreateAndRestoreBackup(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "projects", "kiblazer.md"), []byte("# Kiblazer\n"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}
	backup := filepath.Join(t.TempDir(), "backup.tgz")
	if err := Create(root, backup, "secret"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	target := filepath.Join(t.TempDir(), "restore")
	if err := Restore(backup, target, "secret"); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "projects", "kiblazer.md")); err != nil {
		t.Fatalf("restored project missing: %v", err)
	}
}

func TestRestoreRejectsNonEmptyTarget(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write target file: %v", err)
	}
	if err := Restore(filepath.Join(t.TempDir(), "missing.tgz"), target, ""); err == nil {
		t.Fatal("Restore() error = nil, want non-empty target error")
	}
}
