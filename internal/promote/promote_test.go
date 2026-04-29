package promote

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/capture"
	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestCleanFilePromotesToExperience(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "deploy.md")
	if err := os.WriteFile(source, []byte("# Deploy Lesson\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	captured, err := capture.Markdown(root, source)
	if err != nil {
		t.Fatalf("capture.Markdown() error = %v", err)
	}

	result, err := File(root, captured.Path, "experience/deployment")
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}

	if !strings.Contains(result.Path, filepath.Join("experience", "deployment")) {
		t.Fatalf("Path = %s, want experience/deployment", result.Path)
	}
	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read promoted: %v", err)
	}
	if !strings.Contains(string(content), `status: "promoted"`) {
		t.Fatalf("promoted file missing status:\n%s", content)
	}
}

func TestHighRiskFileCannotPromote(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(source, []byte("# Secret\n\npassword=super-secret-password\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	captured, err := capture.Markdown(root, source)
	if err != nil {
		t.Fatalf("capture.Markdown() error = %v", err)
	}

	if _, err := File(root, captured.Path, "experience/deployment"); err == nil {
		t.Fatal("File() error = nil, want high risk error")
	}
}

func TestPromoteRejectsPathEscape(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "deploy.md")
	if err := os.WriteFile(source, []byte("# Deploy Lesson\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	captured, err := capture.Markdown(root, source)
	if err != nil {
		t.Fatalf("capture.Markdown() error = %v", err)
	}

	if _, err := File(root, captured.Path, "../outside"); err == nil {
		t.Fatal("File() error = nil, want path escape error")
	}
}
