package sync

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestStatusReturnsGitShortStatus(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "projects", "kiblazer.md"), []byte("# Kiblazer\n"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}

	result, err := Status(root)
	if err != nil {
		t.Fatalf("Status() error = %v, output = %s", err, result.Output)
	}
	if !strings.Contains(result.Output, "projects/") {
		t.Fatalf("status missing projects directory: %s", result.Output)
	}
}

func TestCommitBlocksHighRiskContent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "projects", "secret.md"), []byte("password=super-secret-password\n"), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	if _, err := Commit(root, "memory update"); err == nil {
		t.Fatal("Commit() error = nil, want high risk block")
	}
}

func TestCommitDoesNotStageQuarantineByDefault(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(root, "quarantine", "README.md"), []byte("# Quarantine\n"), 0o644); err != nil {
		t.Fatalf("write quarantine: %v", err)
	}

	if _, err := Commit(root, "memory update"); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	result := gitTrackedFiles(t, root)
	for _, want := range []string{"AI-SETUP.md", "experience/auto/README.md"} {
		if !strings.Contains(result, want) {
			t.Fatalf("commit missing tracked path %q:\n%s", want, result)
		}
	}
	if strings.Contains(result, "quarantine/README.md") {
		t.Fatalf("commit should not stage quarantine by default:\n%s", result)
	}
}

func gitTrackedFiles(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-files failed: %v\n%s", err, output)
	}
	return string(output)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
