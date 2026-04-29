package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestListAndRejectCandidates(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(root, "inbox", "note.md")
	if err := os.WriteFile(source, []byte("# Note\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	candidates, err := List(root)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].Path != source {
		t.Fatalf("candidates = %#v", candidates)
	}
	rejected, err := Reject(root, source)
	if err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if !strings.Contains(rejected, filepath.Join("inbox", "rejected")) {
		t.Fatalf("rejected path = %s", rejected)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists or stat error mismatch: %v", err)
	}
}
