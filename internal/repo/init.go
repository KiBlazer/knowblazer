package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/knowblazer/knowblazer/templates"
)

const templateRoot = "default-memory-repo"

func Init(root string) error {
	if root == "" {
		return errors.New("repo path is required")
	}

	root = filepath.Clean(root)
	if err := ensureCanInitialize(root); err != nil {
		return err
	}

	return writeTemplate(root)
}

func writeTemplate(root string) error {
	return fs.WalkDir(templates.FS, templateRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == templateRoot {
			return nil
		}

		rel, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(rel))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		content, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := writeFileIfMissing(target, content); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
		return nil
	})
}

func ensureCanInitialize(root string) error {
	info, err := os.Stat(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", root)
	}
	if isKnowblazerRepo(root) {
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s is non-empty and is not a Knowblazer repo", root)
	}
	return nil
}

func isKnowblazerRepo(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".knowblazer", "config.json"))
	return err == nil
}

func writeFileIfMissing(path string, content []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
