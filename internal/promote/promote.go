package promote

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knowblazer/knowblazer/internal/scan"
)

type Result struct {
	Path string
}

func File(repoRoot string, sourcePath string, target string) (Result, error) {
	scanResult, err := scan.Path(sourcePath)
	if err != nil {
		return Result{}, err
	}
	if scanResult.Level == scan.High {
		return Result{}, errors.New("sensitive content detected; file was not promoted")
	}

	targetDir, err := safeTargetDir(repoRoot, target)
	if err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Result{}, err
	}

	targetPath := filepath.Join(targetDir, filepath.Base(sourcePath))
	if _, err := os.Stat(targetPath); err == nil {
		return Result{}, fmt.Errorf("target already exists: %s", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Result{}, err
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return Result{}, err
	}
	content = markPromoted(content, target, time.Now())
	if err := os.WriteFile(targetPath, content, 0o644); err != nil {
		return Result{}, err
	}
	if err := os.Remove(sourcePath); err != nil {
		return Result{}, err
	}

	return Result{Path: targetPath}, nil
}

func safeTargetDir(repoRoot string, target string) (string, error) {
	cleanTarget := filepath.Clean(target)
	if filepath.IsAbs(cleanTarget) || strings.HasPrefix(cleanTarget, "..") {
		return "", fmt.Errorf("target path is outside the Knowblazer repo and is not allowed")
	}

	first := strings.Split(cleanTarget, string(filepath.Separator))[0]
	switch first {
	case "experience", "projects", "profile":
	default:
		return "", fmt.Errorf("target must be under experience, projects, or profile")
	}

	rootAbs, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(filepath.Join(repoRoot, cleanTarget))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("target path is outside the Knowblazer repo and is not allowed")
	}
	return targetAbs, nil
}

func markPromoted(content []byte, target string, promotedAt time.Time) []byte {
	body := string(content)
	if !strings.HasPrefix(body, "---\n") {
		header := fmt.Sprintf(`---
status: "promoted"
promoted_at: "%s"
promoted_to: "%s"
---

`, promotedAt.Format(time.RFC3339), target)
		return []byte(header + body)
	}

	end := strings.Index(body[4:], "\n---")
	if end < 0 {
		return content
	}
	end += 4
	front := body[:end]
	rest := body[end:]
	front = replaceOrAppendField(front, "status", "promoted")
	front = replaceOrAppendField(front, "promoted_at", promotedAt.Format(time.RFC3339))
	front = replaceOrAppendField(front, "promoted_to", target)
	return []byte(front + rest)
}

func replaceOrAppendField(front string, key string, value string) string {
	lines := strings.Split(front, "\n")
	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, key+":") {
			lines[i] = fmt.Sprintf(`%s: "%s"`, key, value)
			replaced = true
		}
	}
	if !replaced {
		lines = append(lines, fmt.Sprintf(`%s: "%s"`, key, value))
	}
	return strings.Join(lines, "\n")
}
