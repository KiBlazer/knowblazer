package repo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesDefaultMemoryRepo(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")

	if err := Init(root); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	wantPaths := []string{
		".knowblazer/config.json",
		"AI-SETUP.md",
		"inbox",
		"daily",
		"projects",
		"experience",
		"experience/auto",
		"quarantine",
	}

	for _, rel := range wantPaths {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected %s to exist: %v", rel, err)
		}
	}
}

func TestInitUsesDefaultTemplateContent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")

	if err := Init(root); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, "AI-SETUP.md"))
	if err != nil {
		t.Fatalf("read AI-SETUP.md: %v", err)
	}
	if !strings.Contains(string(content), "treat synthesized project notes as more reliable than raw automatic notes") {
		t.Fatalf("AI-SETUP.md does not look like template content:\n%s", content)
	}
}

func TestInitDoesNotOverwriteExistingFiles(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := Init(root); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	setup := filepath.Join(root, "AI-SETUP.md")
	custom := []byte("# Custom AI Setup\n\nDo not overwrite this.\n")
	if err := os.WriteFile(setup, custom, 0o644); err != nil {
		t.Fatalf("write custom AI setup: %v", err)
	}

	if err := Init(root); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	got, err := os.ReadFile(setup)
	if err != nil {
		t.Fatalf("read AI setup: %v", err)
	}
	if string(got) != string(custom) {
		t.Fatalf("AI setup was overwritten\ngot:\n%s\nwant:\n%s", got, custom)
	}
}

func TestInitRejectsNonEmptyNonKnowblazerDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not-memory")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Existing\n"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	if err := Init(root); err == nil {
		t.Fatal("Init() error = nil, want error for non-empty non-Knowblazer directory")
	}
}
