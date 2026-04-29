package sync

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/knowblazer/knowblazer/internal/scan"
)

type Result struct {
	Output string
}

func Status(repoRoot string) (Result, error) {
	if result, err := scan.Path(repoRoot); err != nil {
		return Result{}, err
	} else if result.Level == scan.High {
		return Result{}, errors.New("sensitive content detected; sync status blocked until quarantine is reviewed")
	}
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
	if _, err := git(repoRoot, "add", "AI-SETUP.md", "daily", "profile", "projects", "experience", "system", "inbox", "recall", ".knowblazer"); err != nil {
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
