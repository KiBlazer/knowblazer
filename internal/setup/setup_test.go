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
