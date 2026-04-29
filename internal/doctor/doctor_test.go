package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestRunHealthyRepoHasNoFailures(t *testing.T) {
	root := initRepo(t)
	writeGitConfig(t, root, `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = git@example.com:owner/repo.git
`)

	result := Run(root)
	if result.HasFailures() {
		t.Fatalf("expected no failures: %#v", result.Checks)
	}
	assertCheck(t, result, "git-repo", OK)
	assertCheck(t, result, "git-remote", OK)
}

func TestRunFailsWhenRequiredFileIsMissing(t *testing.T) {
	root := initRepo(t)
	if err := os.Remove(filepath.Join(root, "system", "privacy-policy.md")); err != nil {
		t.Fatalf("remove policy: %v", err)
	}

	result := Run(root)
	if !result.HasFailures() {
		t.Fatalf("expected failure: %#v", result.Checks)
	}
	assertCheck(t, result, "system/privacy-policy.md", Fail)
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

func TestRunWarnsWhenGitMetadataIsMissing(t *testing.T) {
	root := initRepo(t)

	result := Run(root)
	if result.HasFailures() {
		t.Fatalf("expected warnings without failure: %#v", result.Checks)
	}
	assertCheck(t, result, "git-repo", Warn)
	assertCheck(t, result, "git-remote", Warn)
}

func TestRunDetectsGitRemoteFromConfig(t *testing.T) {
	root := initRepo(t)
	writeGitConfig(t, root, `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = git@example.com:owner/repo.git
`)

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

func writeGitConfig(t *testing.T, root string, content string) {
	t.Helper()
	gitDir := filepath.Join(root, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(content), 0o644); err != nil {
		t.Fatalf("write git config: %v", err)
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
