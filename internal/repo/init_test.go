package repo

import (
	"os"
	"path/filepath"
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
		"profile/preferences.md",
		"profile/decision-principles.md",
		"projects",
		"experience/deployment",
		"experience/frontend",
		"experience/backend",
		"experience/ai-tools",
		"experience/operations",
		"system/memory-policy.md",
		"system/privacy-policy.md",
		"quarantine",
		"recall",
	}

	for _, rel := range wantPaths {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("expected %s to exist: %v", rel, err)
		}
	}
}

func TestInitDoesNotOverwriteExistingFiles(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := Init(root); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	preferences := filepath.Join(root, "profile", "preferences.md")
	custom := []byte("# My Preferences\n\nDo not overwrite this.\n")
	if err := os.WriteFile(preferences, custom, 0o644); err != nil {
		t.Fatalf("write custom preferences: %v", err)
	}

	if err := Init(root); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	got, err := os.ReadFile(preferences)
	if err != nil {
		t.Fatalf("read preferences: %v", err)
	}
	if string(got) != string(custom) {
		t.Fatalf("preferences was overwritten\ngot:\n%s\nwant:\n%s", got, custom)
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
