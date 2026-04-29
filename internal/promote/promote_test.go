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

func TestPromoteRejectsSymlinkEscape(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	link := filepath.Join(root, "experience", "outside-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("symlink outside: %v", err)
	}
	source := filepath.Join(root, "inbox", "note.md")
	if err := os.WriteFile(source, []byte("# Note\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if _, err := File(root, source, "experience/outside-link"); err == nil {
		t.Fatal("File() error = nil, want symlink escape error")
	}
}

func TestPromoteUpdatesExistingStatusFieldOnce(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(root, "inbox", "note.md")
	content := `---
title: "Note"
status: "candidate"
---

# Note

No secrets.
`
	if err := os.WriteFile(source, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := File(root, source, "profile")
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	got, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read promoted: %v", err)
	}
	text := string(got)
	if strings.Count(text, "status:") != 1 {
		t.Fatalf("status field count = %d, want 1:\n%s", strings.Count(text, "status:"), text)
	}
	if !strings.Contains(text, `status: "promoted"`) {
		t.Fatalf("status was not promoted:\n%s", text)
	}
}

func TestPromoteWithoutFrontMatterAddsPromotedHeader(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(root, "inbox", "note.md")
	if err := os.WriteFile(source, []byte("# Note\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result, err := File(root, source, "projects")
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	got, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read promoted: %v", err)
	}
	if !strings.HasPrefix(string(got), "---\nstatus: \"promoted\"") {
		t.Fatalf("missing promoted header:\n%s", got)
	}
}
