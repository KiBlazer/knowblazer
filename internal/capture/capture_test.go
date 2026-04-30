package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestMarkdownCleanFileGoesToInbox(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "deploy lesson.md")
	original := "# Deploy Lesson\n\nUse the existing release checklist.\n"
	if err := os.WriteFile(source, []byte(original), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := Markdown(root, source)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(result.Path, filepath.Join("inbox")) {
		t.Fatalf("Path = %s, want inbox path", result.Path)
	}
	if _, err := os.Stat(result.Path); err != nil {
		t.Fatalf("captured file missing: %v", err)
	}
	captured, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read captured: %v", err)
	}
	if !strings.Contains(string(captured), `status: "candidate"`) {
		t.Fatalf("captured file missing candidate status:\n%s", captured)
	}
	if !strings.Contains(string(captured), "# Deploy Lesson") {
		t.Fatalf("captured file missing original body:\n%s", captured)
	}
	gotSource, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	if string(gotSource) != original {
		t.Fatalf("source file changed")
	}
}

func TestMarkdownAutoCleanFileGoesToExperienceAuto(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "memory.md")
	if err := os.WriteFile(source, []byte("# Memory\n\nDeploys need smoke tests.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := MarkdownAuto(root, source)
	if err != nil {
		t.Fatalf("MarkdownAuto() error = %v", err)
	}

	if !strings.Contains(result.Path, filepath.Join("experience", "auto")) {
		t.Fatalf("Path = %s, want experience/auto path", result.Path)
	}
	captured, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read captured: %v", err)
	}
	for _, want := range []string{`type: "experience"`, `status: "fresh"`, `scan_level: "clean"`} {
		if !strings.Contains(string(captured), want) {
			t.Fatalf("captured file missing %q:\n%s", want, captured)
		}
	}
}

func TestMarkdownAutoHighRiskFileGoesToQuarantine(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(source, []byte("# Secret\n\npassword=super-secret-password\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := MarkdownAuto(root, source)
	if err != nil {
		t.Fatalf("MarkdownAuto() error = %v", err)
	}

	if !strings.Contains(result.Path, filepath.Join("quarantine")) {
		t.Fatalf("Path = %s, want quarantine path", result.Path)
	}
	if result.ScanLevel.String() != "high" {
		t.Fatalf("ScanLevel = %s, want high", result.ScanLevel.String())
	}
}

func TestMarkdownHighRiskFileGoesToQuarantine(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "secret.md")
	if err := os.WriteFile(source, []byte("# Secret\n\npassword=super-secret-password\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := Markdown(root, source)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(result.Path, filepath.Join("quarantine")) {
		t.Fatalf("Path = %s, want quarantine path", result.Path)
	}
	if result.ScanLevel.String() != "high" {
		t.Fatalf("ScanLevel = %s, want high", result.ScanLevel.String())
	}
}

func TestMarkdownRejectsNonMarkdownFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "lesson.txt")
	if err := os.WriteFile(source, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if _, err := Markdown(root, source); err == nil {
		t.Fatal("Markdown() error = nil, want error for non-Markdown file")
	}
}
