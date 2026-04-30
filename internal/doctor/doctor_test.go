package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestRunHealthyRepoHasNoFailures(t *testing.T) {
	root := initRepo(t)
	addGitRemote(t, root)

	result := Run(root)
	if result.HasFailures() {
		t.Fatalf("expected no failures: %#v", result.Checks)
	}
	assertCheck(t, result, "git-repo", OK)
	assertCheck(t, result, "git-remote", OK)
}

func TestRunFailsWhenRequiredFileIsMissing(t *testing.T) {
	root := initRepo(t)
	if err := os.Remove(filepath.Join(root, "AI-SETUP.md")); err != nil {
		t.Fatalf("remove AI setup: %v", err)
	}

	result := Run(root)
	if !result.HasFailures() {
		t.Fatalf("expected failure: %#v", result.Checks)
	}
	assertCheck(t, result, "AI-SETUP.md", Fail)
}

func TestRunFailsWhenRequiredDirectoryIsMissing(t *testing.T) {
	root := initRepo(t)
	if err := os.RemoveAll(filepath.Join(root, "projects")); err != nil {
		t.Fatalf("remove projects: %v", err)
	}

	result := Run(root)
	if !result.HasFailures() {
		t.Fatalf("expected failure: %#v", result.Checks)
	}
	assertCheck(t, result, "projects", Fail)
}

func TestRunWarnsForQuarantineFiles(t *testing.T) {
	root := initRepo(t)
	quarantineDir := filepath.Join(root, "quarantine", "2026-04-29")
	if err := os.MkdirAll(quarantineDir, 0o755); err != nil {
		t.Fatalf("mkdir quarantine date dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(quarantineDir, "secret.md"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("write quarantine file: %v", err)
	}

	result := Run(root)
	if result.HasFailures() {
		t.Fatalf("expected warning without failure: %#v", result.Checks)
	}
	assertCheck(t, result, "quarantine-files", Warn)
}

func TestRunWarnsWhenGitRemoteIsMissing(t *testing.T) {
	root := initRepo(t)

	result := Run(root)
	if result.HasFailures() {
		t.Fatalf("expected warnings without failure: %#v", result.Checks)
	}
	assertCheck(t, result, "git-repo", OK)
	assertCheck(t, result, "git-remote", Warn)
}

func TestRunDetectsGitRemoteFromConfig(t *testing.T) {
	root := initRepo(t)
	addGitRemote(t, root)

	result := Run(root)
	assertCheck(t, result, "git-repo", OK)
	assertCheck(t, result, "git-remote", OK)
}

func initRepo(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("init repo: %v", err)
	}
	return root
}

func addGitRemote(t *testing.T, root string) {
	t.Helper()
	cmd := exec.Command("git", "remote", "add", "origin", "git@example.com:owner/repo.git")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git remote add failed: %v\n%s", err, output)
	}
}

func assertCheck(t *testing.T, result Result, name string, status Status) {
	t.Helper()
	for _, check := range result.Checks {
		if check.Name == name {
			if check.Status != status {
				t.Fatalf("check %s status = %s, want %s", name, check.Status.String(), status.String())
			}
			return
		}
	}
	t.Fatalf("check %s not found in %#v", name, result.Checks)
}
