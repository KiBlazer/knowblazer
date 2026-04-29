package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanFileDetectsHighRiskSecrets(t *testing.T) {
	file := filepath.Join(t.TempDir(), "lesson.md")
	content := []byte(`# Deploy

password=super-secret-password
DATABASE_URL=postgres://user:pass@example.com/app
-----BEGIN PRIVATE KEY-----
abc123
-----END PRIVATE KEY-----
`)
	if err := os.WriteFile(file, content, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Path(file)
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}

	if result.Level != High {
		t.Fatalf("Level = %v, want High", result.Level)
	}
	if len(result.Findings) < 3 {
		t.Fatalf("Findings len = %d, want at least 3", len(result.Findings))
	}
}

func TestScanFileRedactsSecretValues(t *testing.T) {
	file := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(file, []byte("password=super-secret-password\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Path(file)
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("Findings len = %d, want 1", len(result.Findings))
	}
	if strings.Contains(result.Findings[0].Snippet, "super-secret-password") {
		t.Fatalf("snippet leaked secret: %s", result.Findings[0].Snippet)
	}
	if !strings.Contains(result.Findings[0].Snippet, "****") {
		t.Fatalf("snippet was not redacted: %s", result.Findings[0].Snippet)
	}
}

func TestScanCleanFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(file, []byte("# Lesson\n\nNo secrets here.\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Path(file)
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}
	if result.Level != Clean {
		t.Fatalf("Level = %v, want Clean", result.Level)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("Findings len = %d, want 0", len(result.Findings))
	}
}
