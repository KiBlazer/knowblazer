package sync

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/knowblazer/knowblazer/internal/scan"
)

type Result struct {
	Output string
}

func Status(repoRoot string) (Result, error) {
	return git(repoRoot, "status", "--short")
}

func Commit(repoRoot string, message string) (Result, error) {
	if strings.TrimSpace(message) == "" {
		return Result{}, errors.New("commit message is required")
	}
	if result, err := scan.Path(repoRoot); err != nil {
		return Result{}, err
	} else if result.Level == scan.High {
		return Result{}, errors.New("sensitive content detected; commit blocked")
	}
	paths, err := existingStagePaths(repoRoot)
	if err != nil {
		return Result{}, err
	}
	if len(paths) == 0 {
		return Result{}, errors.New("no Knowblazer paths found to stage")
	}
	args := append([]string{"add"}, paths...)
	if _, err := git(repoRoot, args...); err != nil {
		return Result{}, err
	}
	return git(repoRoot, "commit", "-m", message)
}

func Push(repoRoot string) (Result, error) {
	if result, err := scan.Path(repoRoot); err != nil {
		return Result{}, err
	} else if result.Level == scan.High {
		return Result{}, errors.New("sensitive content detected; push blocked")
	}
	return git(repoRoot, "push")
}

func Pull(repoRoot string) (Result, error) {
	return git(repoRoot, "pull")
}

func existingStagePaths(repoRoot string) ([]string, error) {
	known := []string{
		"AI-SETUP.md",
		"daily",
		"projects",
		"experience",
		"inbox",
		".knowblazer",
		"profile",
		"system",
		"recall",
	}
	var paths []string
	for _, path := range known {
		_, err := os.Stat(filepath.Join(repoRoot, path))
		if err == nil {
			paths = append(paths, path)
			continue
		}
		if os.IsNotExist(err) {
			continue
		}
		return nil, err
	}
	return paths, nil
}

func git(repoRoot string, args ...string) (Result, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = filepath.Clean(repoRoot)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return Result{Output: out.String()}, fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return Result{Output: out.String()}, nil
}
