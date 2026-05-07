package consolidate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/capture"
	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestRunSynthesizesFreshAutoMemories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "memory.md")
	if err := os.WriteFile(source, []byte("# MCP Debugging\n\nCompare Claude Code and shell environments first.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	captured, err := capture.MarkdownAuto(root, source)
	if err != nil {
		t.Fatalf("MarkdownAuto() error = %v", err)
	}

	result, err := Run(root)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("Count = %d, want 1", result.Count)
	}
	if !strings.Contains(result.Path, filepath.Join("experience", "synthesized")) {
		t.Fatalf("Path = %s, want synthesized path", result.Path)
	}
	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read synthesized: %v", err)
	}
	text := string(content)
	for _, want := range []string{`status: "synthesized"`, "Compare Claude Code and shell environments first.", "experience/auto/"} {
		if !strings.Contains(text, want) {
			t.Fatalf("synthesized memory missing %q:\n%s", want, text)
		}
	}

	sourceContent, err := os.ReadFile(captured.Path)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	if !strings.Contains(string(sourceContent), `status: "consolidated"`) {
		t.Fatalf("source not marked consolidated:\n%s", sourceContent)
	}
}

func TestRunPreservesMultilineMemoryBody(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "memory.md")
	content := "# Release Lessons\n\n- Run smoke tests after deploy.\n- Check queue depth before rollback.\n- Keep the deploy window documented.\n"
	if err := os.WriteFile(source, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	captured, err := capture.MarkdownAuto(root, source)
	if err != nil {
		t.Fatalf("MarkdownAuto() error = %v", err)
	}

	result, err := Run(root)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	textBytes, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read synthesized: %v", err)
	}
	text := string(textBytes)
	for _, want := range []string{"## Source Signals", "Release Lessons", "Run smoke tests after deploy.", "Check queue depth before rollback.", "Keep the deploy window documented.", "experience/auto/"} {
		if !strings.Contains(text, want) {
			t.Fatalf("synthesized memory missing %q:\n%s", want, text)
		}
	}
	sourceContent, err := os.ReadFile(captured.Path)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	if !strings.Contains(string(sourceContent), `status: "consolidated"`) {
		t.Fatalf("source not marked consolidated:\n%s", sourceContent)
	}
}

func TestRunAcceptsLegacyAutoPromotedStatus(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	path := filepath.Join(root, "experience", "auto", "legacy.md")
	content := `---
title: "Legacy"
type: "experience"
status: "auto_promoted"
---

# Legacy

Old automatic memory.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	result, err := Run(root)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("Count = %d, want 1", result.Count)
	}
}

func TestRunNoFreshMemoriesIsNoop(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}

	result, err := Run(root)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Count != 0 || result.Path != "" {
		t.Fatalf("result = %#v, want noop", result)
	}
}
