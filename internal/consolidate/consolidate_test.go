package consolidate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

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

func TestRunPreservesMultiParagraphMemoryWithEmptyLines(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "memory.md")
	content := "# Troubleshooting\n\nFirst paragraph explaining the bug.\n\nSecond paragraph explaining the solution.\n\n```bash\nkubectl rollout restart\n```\n"
	if err := os.WriteFile(source, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if _, err := capture.MarkdownAuto(root, source); err != nil {
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
	for _, want := range []string{
		"First paragraph explaining the bug.",
		"Second paragraph explaining the solution.",
		"kubectl rollout restart",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("synthesized memory missing %q:\n%s", want, text)
		}
	}
}

func TestRunTruncatesWhenLineLimitExceeded(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "long.md")
	var b strings.Builder
	b.WriteString("# Long note\n\n")
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "Line item number %d\n", i)
	}
	if err := os.WriteFile(source, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if _, err := capture.MarkdownAuto(root, source); err != nil {
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
	if !strings.Contains(text, "[truncated]") {
		t.Fatalf("synthesized memory missing truncation marker for 40 lines:\n%s", text)
	}
}

func TestRunTruncatesUtf8WithoutCorruptingRunes(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	source := filepath.Join(t.TempDir(), "cjk_long.md")
	// Repeat 3-byte CJK runes so length in bytes is ~4800, but runes is 1600 (> 1500 limit).
	cjkText := strings.Repeat("中", 1600)
	content := "# 性能优化经验\n\n" + cjkText + "\n"
	if err := os.WriteFile(source, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if _, err := capture.MarkdownAuto(root, source); err != nil {
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
	if !utf8.Valid(textBytes) {
		t.Fatalf("synthesized file is not valid UTF-8, truncated mid-rune")
	}
	text := string(textBytes)
	if !strings.Contains(text, "[truncated]") {
		t.Fatalf("synthesized memory missing truncation marker:\n%s", text)
	}
}

