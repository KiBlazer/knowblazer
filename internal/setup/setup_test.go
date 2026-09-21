package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestCodexGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}

	result, err := Codex(CodexOptions{
		RepoRoot: root,
		Global:   true,
		SkipMCP:  true,
	})
	if err != nil {
		t.Fatalf("Codex() error = %v", err)
	}

	expectedPath := filepath.Join(home, ".agents", "AGENTS.md")
	if result.AgentsMD != expectedPath {
		t.Errorf("AgentsMD = %q, want %q", result.AgentsMD, expectedPath)
	}

	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "Use Knowblazer as the local engineering memory source across all workspaces.") {
		t.Fatalf("AGENTS.md missing global prompt:\n%s", text)
	}

	// Test idempotence
	_, err = Codex(CodexOptions{
		RepoRoot: root,
		Global:   true,
		SkipMCP:  true,
	})
	if err != nil {
		t.Fatalf("second Codex() error = %v", err)
	}
	secondContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read second AGENTS.md: %v", err)
	}
	if strings.Count(string(secondContent), "KNOWBLAZER-CODEX-SETUP:START") != 1 {
		t.Fatalf("duplicate setup block found:\n%s", string(secondContent))
	}
}

func TestClaudeGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}

	result, err := Claude(ClaudeOptions{
		RepoRoot: root,
		Global:   true,
		SkipMCP:  true,
	})
	if err != nil {
		t.Fatalf("Claude() error = %v", err)
	}

	expectedPath := filepath.Join(home, ".claude", "CLAUDE.md")
	if result.ClaudeMD != expectedPath {
		t.Errorf("ClaudeMD = %q, want %q", result.ClaudeMD, expectedPath)
	}

	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "Use Knowblazer as the local engineering memory source across all workspaces.") {
		t.Fatalf("CLAUDE.md missing global prompt:\n%s", text)
	}
}

func TestCodexGlobalPreservesExistingUserRules(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	agentsDir := filepath.Join(home, ".agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	initialContent := `# My Custom Company Rules

## 1. Decision Discipline
* Inspect code before deciding.

## 2. Security Rule
* Never leak API keys.
`
	agentsPath := filepath.Join(agentsDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(initialContent), 0o644); err != nil {
		t.Fatalf("write initial AGENTS.md: %v", err)
	}

	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}

	_, err := Codex(CodexOptions{
		RepoRoot: root,
		Global:   true,
		SkipMCP:  true,
	})
	if err != nil {
		t.Fatalf("Codex() error = %v", err)
	}

	updated, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	updatedStr := string(updated)

	// Verify original rules are 100% preserved at the top
	if !strings.HasPrefix(updatedStr, initialContent) {
		t.Fatalf("original rules were modified or clobbered:\n%s", updatedStr)
	}
	// Verify knowblazer block was appended
	if !strings.Contains(updatedStr, "KNOWBLAZER-CODEX-SETUP:START") {
		t.Fatalf("knowblazer setup block missing:\n%s", updatedStr)
	}

	// Update with new repo path and verify original rules still intact
	newRoot := filepath.Join(t.TempDir(), "memory-v2")
	if err := repo.Init(newRoot); err != nil {
		t.Fatalf("repo.Init() v2 error = %v", err)
	}

	_, err = Codex(CodexOptions{
		RepoRoot: newRoot,
		Global:   true,
		SkipMCP:  true,
	})
	if err != nil {
		t.Fatalf("second Codex() error = %v", err)
	}

	secondUpdated, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read second AGENTS.md: %v", err)
	}
	secondStr := string(secondUpdated)
	if !strings.HasPrefix(secondStr, initialContent) {
		t.Fatalf("original rules were modified on second update:\n%s", secondStr)
	}
	if !strings.Contains(secondStr, newRoot) {
		t.Fatalf("updated repo path missing:\n%s", secondStr)
	}
	if strings.Count(secondStr, "KNOWBLAZER-CODEX-SETUP:START") != 1 {
		t.Fatalf("multiple setup blocks found:\n%s", secondStr)
	}
}
