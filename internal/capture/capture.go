package capture

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/knowblazer/knowblazer/internal/scan"
)

type Result struct {
	Path      string
	ScanLevel scan.Level
}

func Markdown(repoRoot string, sourcePath string) (Result, error) {
	if !isMarkdown(sourcePath) {
		return Result{}, errors.New("only Markdown files are supported in MVP")
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return Result{}, err
	}

	scanResult, err := scan.Path(sourcePath)
	if err != nil {
		return Result{}, err
	}

	now := time.Now()
	layer := "inbox"
	status := "candidate"
	if scanResult.Level == scan.High {
		layer = "quarantine"
		status = "quarantined"
	}

	title := titleFromMarkdown(content, sourcePath)
	slug := slugify(title)
	if slug == "" {
		slug = "note"
	}

	targetDir := filepath.Join(repoRoot, layer, now.Format("2006-01-02"))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Result{}, err
	}

	targetPath := filepath.Join(targetDir, fmt.Sprintf("%s-%s.md", now.Format("20060102-150405"), slug))
	wrapped := addFrontMatter(content, title, layer, status, sourcePath, now)
	if err := os.WriteFile(targetPath, wrapped, 0o644); err != nil {
		return Result{}, err
	}

	return Result{Path: targetPath, ScanLevel: scanResult.Level}, nil
}

func isMarkdown(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

func titleFromMarkdown(content []byte, sourcePath string) string {
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	base := filepath.Base(sourcePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func addFrontMatter(content []byte, title string, noteType string, status string, sourcePath string, capturedAt time.Time) []byte {
	body := string(content)
	if strings.HasPrefix(body, "---\n") {
		return content
	}
	frontMatter := fmt.Sprintf(`---
title: "%s"
type: "%s"
status: "%s"
source: "%s"
captured_at: "%s"
---

`, escapeYAML(title), noteType, status, escapeYAML(sourcePath), capturedAt.Format(time.RFC3339))
	return []byte(frontMatter + body)
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	value = re.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func escapeYAML(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}
