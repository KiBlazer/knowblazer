package specstory

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/knowblazer/knowblazer/internal/scan"
)

type Result struct {
	Source string
	Path   string
	Level  scan.Level
}

func Import(repoRoot string, source string) ([]Result, error) {
	info, err := os.Stat(source)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		result, err := importFile(repoRoot, source)
		if err != nil {
			return nil, err
		}
		return []Result{result}, nil
	}
	var results []Result
	err = filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		result, err := importFile(repoRoot, path)
		if err != nil {
			return err
		}
		results = append(results, result)
		return nil
	})
	return results, err
}

func importFile(repoRoot string, source string) (Result, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return Result{}, err
	}
	scanResult, err := scan.Path(source)
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
	title := titleFromContent(content, source)
	body := normalizeBody(content, source)
	wrapped := []byte(fmt.Sprintf(`---
title: "%s"
type: "%s"
status: "%s"
source: "%s"
imported_from: "specstory"
imported_at: "%s"
---

%s`, escape(title), layer, status, escape(source), now.Format(time.RFC3339), body))
	targetDir := filepath.Join(repoRoot, layer, now.Format("2006-01-02"))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Result{}, err
	}
	target := uniquePath(filepath.Join(targetDir, fmt.Sprintf("%s-%s.md", now.Format("20060102-150405"), slugify(title))))
	if err := os.WriteFile(target, wrapped, 0o644); err != nil {
		return Result{}, err
	}
	return Result{Source: source, Path: target, Level: scanResult.Level}, nil
}

func normalizeBody(content []byte, source string) string {
	text := string(content)
	ext := strings.ToLower(filepath.Ext(source))
	if ext == ".md" || ext == ".markdown" {
		return text
	}
	return "# " + titleFromContent(content, source) + "\n\n```text\n" + text + "\n```\n"
}

func titleFromContent(content []byte, source string) string {
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	base := filepath.Base(source)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	value = re.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "specstory"
	}
	return value
}

func escape(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
