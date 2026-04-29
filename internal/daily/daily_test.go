package daily

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestAddAppendsTimestampedEntry(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	now := time.Date(2026, 4, 29, 14, 5, 0, 0, time.UTC)

	path, err := Add(root, AddOptions{Text: "finished deploy", Now: now})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read daily: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "# Daily 2026-04-29") || !strings.Contains(text, "- 14:05 finished deploy") {
		t.Fatalf("daily content missing entry:\n%s", text)
	}
}

func TestShowReadsDate(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	date := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	path := filepath.Join(root, "daily", "2026-04-29.md")
	if err := os.WriteFile(path, []byte("# Daily\n"), 0o644); err != nil {
		t.Fatalf("write daily: %v", err)
	}

	content, gotPath, err := Show(root, ShowOptions{Date: date})
	if err != nil {
		t.Fatalf("Show() error = %v", err)
	}
	if gotPath != path || string(content) != "# Daily\n" {
		t.Fatalf("Show() = %q, %s", content, gotPath)
	}
}
