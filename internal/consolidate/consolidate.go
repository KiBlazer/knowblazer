package consolidate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Result struct {
	Path  string
	Count int
}

type memory struct {
	path    string
	relPath string
	title   string
	body    string
}

func Run(repoRoot string) (Result, error) {
	fresh, err := freshMemories(repoRoot)
	if err != nil {
		return Result{}, err
	}
	if len(fresh) == 0 {
		return Result{}, nil
	}

	now := time.Now()
	targetDir := filepath.Join(repoRoot, "experience", "synthesized")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Result{}, err
	}
	target := filepath.Join(targetDir, now.Format("2006-01-02-150405")+"-synthesized-memory.md")
	content := synthesizedContent(fresh, now)
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return Result{}, err
	}
	for _, item := range fresh {
		if err := markConsolidated(item.path); err != nil {
			return Result{}, err
		}
	}
	return Result{Path: target, Count: len(fresh)}, nil
}

func CountFresh(repoRoot string) (int, error) {
	fresh, err := freshMemories(repoRoot)
	if err != nil {
		return 0, err
	}
	return len(fresh), nil
}

func CountSynthesized(repoRoot string) (int, error) {
	count := 0
	root := filepath.Join(repoRoot, "experience", "synthesized")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" || strings.EqualFold(entry.Name(), "README.md") {
			return nil
		}
		count++
		return nil
	})
	return count, err
}

func freshMemories(repoRoot string) ([]memory, error) {
	root := filepath.Join(repoRoot, "experience", "auto")
	var items []memory
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" || strings.EqualFold(entry.Name(), "README.md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		if !isFresh(text) {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		items = append(items, memory{
			path:    path,
			relPath: filepath.ToSlash(rel),
			title:   title(text),
			body:    firstMeaningfulLine(stripFrontMatter(text)),
		})
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].relPath < items[j].relPath })
	return items, nil
}

func isFresh(content string) bool {
	return strings.Contains(content, `status: "fresh"`) || strings.Contains(content, `status: "auto_promoted"`)
}

func synthesizedContent(items []memory, capturedAt time.Time) []byte {
	var out bytes.Buffer
	fmt.Fprintf(&out, `---
title: "Synthesized Memory %s"
type: "experience"
status: "synthesized"
source: "knowblazer consolidate"
captured_at: "%s"
---

`, capturedAt.Format("2006-01-02"), capturedAt.Format(time.RFC3339))
	fmt.Fprintf(&out, "# Synthesized Memory %s\n\n", capturedAt.Format("2006-01-02"))
	fmt.Fprintln(&out, "Automatically consolidated from fresh memory signals.")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "## Signals")
	fmt.Fprintln(&out)
	for _, item := range items {
		line := item.body
		if line == "" {
			line = item.title
		}
		fmt.Fprintf(&out, "- `%s`: %s\n", item.relPath, line)
	}
	return out.Bytes()
}

func markConsolidated(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(content)
	replacements := []string{`status: "fresh"`, `status: "auto_promoted"`}
	for _, old := range replacements {
		if strings.Contains(text, old) {
			text = strings.Replace(text, old, `status: "consolidated"`, 1)
			return os.WriteFile(path, []byte(text), 0o644)
		}
	}
	return nil
}

func title(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "title:")), `"`)
		}
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return "Memory"
}

func stripFrontMatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return content
	}
	return content[end+8:]
}

func firstMeaningfulLine(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}
		if len(line) > 180 {
			return line[:180] + "..."
		}
		return line
	}
	return ""
}
