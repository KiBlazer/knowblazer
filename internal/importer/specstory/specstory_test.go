package specstory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
	"github.com/knowblazer/knowblazer/internal/scan"
)

func TestImportMarkdownToInbox(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "history.md")
	if err := os.WriteFile(source, []byte("# Deploy Chat\n\nRemember smoke tests.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	results, err := Import(root, source)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if len(results) != 1 || results[0].Level != scan.Clean || !strings.Contains(results[0].Path, string(filepath.Separator)+"inbox"+string(filepath.Separator)) {
		t.Fatalf("results = %#v", results)
	}
	content, err := os.ReadFile(results[0].Path)
	if err != nil {
		t.Fatalf("read import: %v", err)
	}
	if !strings.Contains(string(content), `imported_from: "specstory"`) || !strings.Contains(string(content), "Remember smoke tests.") {
		t.Fatalf("imported content invalid:\n%s", content)
	}
}

func TestImportHighRiskToQuarantine(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(source, []byte(`{"text":"password=super-secret-password"}`), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	results, err := Import(root, source)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if len(results) != 1 || results[0].Level != scan.High || !strings.Contains(results[0].Path, string(filepath.Separator)+"quarantine"+string(filepath.Separator)) {
		t.Fatalf("results = %#v", results)
	}
}
