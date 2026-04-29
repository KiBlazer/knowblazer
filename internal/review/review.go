package review

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knowblazer/knowblazer/internal/promote"
	"github.com/knowblazer/knowblazer/internal/scan"
)

type Candidate struct {
	Path  string
	Level scan.Level
}

func List(repoRoot string) ([]Candidate, error) {
	var candidates []Candidate
	root := filepath.Join(repoRoot, "inbox")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" || isIgnoredInboxFile(path) {
			return nil
		}
		result, err := scan.Path(path)
		if err != nil {
			return err
		}
		candidates = append(candidates, Candidate{Path: path, Level: result.Level})
		return nil
	})
	return candidates, err
}

func isIgnoredInboxFile(path string) bool {
	return filepath.Base(path) == "README.md" || strings.Contains(path, string(filepath.Separator)+"rejected"+string(filepath.Separator))
}

func Promote(repoRoot string, source string, target string) (promote.Result, error) {
	return promote.File(repoRoot, source, target)
}

func Reject(repoRoot string, source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("cannot reject directory")
	}
	rejectedDir := filepath.Join(repoRoot, "inbox", "rejected", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(rejectedDir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(rejectedDir, filepath.Base(source))
	if _, err := os.Stat(target); err == nil {
		target = filepath.Join(rejectedDir, fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(source)))
	}
	if err := os.Rename(source, target); err != nil {
		return "", err
	}
	return target, nil
}
