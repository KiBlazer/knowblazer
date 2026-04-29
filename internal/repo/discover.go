package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Options struct {
	StartDir    string
	EnvRepo     string
	DefaultRepo string
}

func Discover(opts Options) (string, error) {
	start := opts.StartDir
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	startAbs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for dir := startAbs; ; dir = filepath.Dir(dir) {
		if isKnowblazerRepo(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	for _, candidate := range []string{opts.EnvRepo, opts.DefaultRepo} {
		if candidate == "" {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return "", err
		}
		if isKnowblazerRepo(abs) {
			return abs, nil
		}
	}

	return "", errors.New("Knowblazer repo not found")
}

func MustBeRepo(path string) error {
	if path == "" {
		return errors.New("repo path is required")
	}
	if !isKnowblazerRepo(path) {
		return fmt.Errorf("%s is not a Knowblazer repo", path)
	}
	return nil
}
