package dream

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Run(repoRoot string) (string, error) {
	now := time.Now()
	var snippets []string
	for _, dir := range []string{"daily", "inbox"} {
		root := filepath.Join(repoRoot, dir)
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" || strings.Contains(path, string(filepath.Separator)+"rejected"+string(filepath.Separator)) {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			snippets = append(snippets, fmt.Sprintf("- `%s`: %s", rel(repoRoot, path), firstLine(string(content))))
			return nil
		})
	}
	if len(snippets) == 0 {
		return "", fmt.Errorf("no daily or inbox notes to summarize")
	}
	targetDir := filepath.Join(repoRoot, "inbox", now.Format("2006-01-02"))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(targetDir, now.Format("20060102-150405")+"-dream-suggestions.md")
	content := fmt.Sprintf(`---
title: "Dream Suggestions %s"
type: "inbox"
status: "suggested"
source: "knowblazer dream"
captured_at: "%s"
---

# Dynamic Memory Suggestions %s

Use these deterministic suggestions as inputs for the dynamic memory loop.

%s
`, now.Format("2006-01-02"), now.Format(time.RFC3339), now.Format("2006-01-02"), strings.Join(snippets, "\n"))
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return "", err
	}
	return target, nil
}

func firstLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" && !strings.HasPrefix(line, "---") {
			if len(line) > 160 {
				return line[:160] + "..."
			}
			return line
		}
	}
	return "empty note"
}

func rel(root string, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(value)
}
