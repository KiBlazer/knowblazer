package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestBuildAndSearchReviewedMemory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "experience", "deployment", "deploy.md"), []byte("# Deploy\n\nRun smoke tests after deploy.\n"), 0o644); err != nil {
		t.Fatalf("write experience: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "quarantine", "secret.md"), []byte("# Secret\n\ndeploy password\n"), 0o644); err != nil {
		t.Fatalf("write quarantine: %v", err)
	}

	idx, err := Build(root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(idx.Documents) == 0 {
		t.Fatal("Build() produced no documents")
	}
	hits, err := Search(root, "deploy smoke")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) == 0 || hits[0].Path != "experience/deployment/deploy.md" {
		t.Fatalf("hits = %#v", hits)
	}
	for _, hit := range hits {
		if hit.Path == "quarantine/secret.md" {
			t.Fatalf("search included quarantine: %#v", hits)
		}
	}
}
