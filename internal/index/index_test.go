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
	deployDir := filepath.Join(root, "experience", "deployment")
	if err := os.MkdirAll(deployDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deployDir, "deploy.md"), []byte("# Deploy\n\nRun smoke tests after deploy.\n"), 0o644); err != nil {
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

func TestBuildAndSearchChineseQuery(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "deployment")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}
	if err := os.WriteFile(filepath.Join(expDir, "plane.md"), []byte("# Plane 部署经验\n\nPlane 平台使用 Docker Compose 部署。\n"), 0o644); err != nil {
		t.Fatalf("write plane.md: %v", err)
	}

	hits, err := Search(root, "Plane部署")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) == 0 {
		t.Fatalf("Search('Plane部署') returned 0 hits")
	}
	if hits[0].Path != "experience/deployment/plane.md" {
		t.Fatalf("hits[0].Path = %q, want experience/deployment/plane.md", hits[0].Path)
	}
}

func TestSearchPrioritizesShortFocusedNoteOverLongDilutedNote(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "networking")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}

	focused := "# gRPC Timeout\n\nFix grpc deadline exceeded by increasing client timeout to 10s.\n"
	if err := os.WriteFile(filepath.Join(expDir, "z_focused.md"), []byte(focused), 0o644); err != nil {
		t.Fatalf("write z_focused.md: %v", err)
	}

	diluted := "# Network Overview Log\n\n"
	for i := 0; i < 500; i++ {
		diluted += "unrelated gateway routing traffic information packet header\n"
	}
	for i := 0; i < 6; i++ {
		diluted += "casually mentioned fix grpc deadline exceeded and timeout in background worker\n"
		for j := 0; j < 50; j++ {
			diluted += "irrelevant log lines metrics counters\n"
		}
	}
	if err := os.WriteFile(filepath.Join(expDir, "a_diluted.md"), []byte(diluted), 0o644); err != nil {
		t.Fatalf("write a_diluted.md: %v", err)
	}

	hits, err := Search(root, "fix grpc deadline exceeded timeout")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) < 2 {
		t.Fatalf("expected at least 2 hits, got %d", len(hits))
	}
	if hits[0].Path != "experience/networking/z_focused.md" {
		t.Fatalf("expected top hit to be z_focused.md, got %s (score %v vs %v)", hits[0].Path, hits[0].Score, hits[1].Score)
	}

}


