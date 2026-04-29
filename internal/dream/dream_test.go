package dream

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestRunCreatesSuggestedInboxCandidate(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "daily", "2026-04-29.md"), []byte("# Daily\n\nDeployed frontend.\n"), 0o644); err != nil {
		t.Fatalf("write daily: %v", err)
	}

	path, err := Run(root)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read dream: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `status: "suggested"`) || !strings.Contains(text, "daily/2026-04-29.md") {
		t.Fatalf("dream content invalid:\n%s", text)
	}
}
