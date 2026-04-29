package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configJSON = `{
  "version": 1,
  "created_by": "knowblazer",
  "memory_repo": true
}
`

var defaultDirs = []string{
	".knowblazer",
	"inbox",
	"daily",
	"profile",
	"projects",
	"experience",
	"experience/deployment",
	"experience/frontend",
	"experience/backend",
	"experience/ai-tools",
	"experience/operations",
	"system",
	"quarantine",
	"recall",
}

var defaultFiles = map[string]string{
	".knowblazer/config.json":         configJSON,
	"AI-SETUP.md":                     "# AI Setup\n\nThis is a private Knowblazer engineering memory repo.\n",
	"profile/preferences.md":          "# Preferences\n\nStore stable engineering preferences here.\n",
	"profile/decision-principles.md":  "# Decision Principles\n\nStore durable engineering decision principles here.\n",
	"system/memory-policy.md":         "# Memory Policy\n\nMarkdown and Git are the source of truth for this memory repo.\n",
	"system/privacy-policy.md":        "# Privacy Policy\n\nDo not upload memory to any official Knowblazer service.\n",
	"inbox/README.md":                 "# Inbox\n\nRaw candidate memory goes here.\n",
	"daily/README.md":                 "# Daily Notes\n\nShort-term working notes go here.\n",
	"projects/README.md":              "# Projects\n\nDurable project context goes here.\n",
	"experience/README.md":            "# Experience\n\nReusable engineering lessons go here.\n",
	"experience/deployment/README.md": "# Deployment Experience\n\nReusable deployment lessons go here.\n",
	"experience/frontend/README.md":   "# Frontend Experience\n\nReusable frontend lessons go here.\n",
	"experience/backend/README.md":    "# Backend Experience\n\nReusable backend lessons go here.\n",
	"experience/ai-tools/README.md":   "# AI Tools Experience\n\nReusable AI coding tool lessons go here.\n",
	"experience/operations/README.md": "# Operations Experience\n\nReusable operations lessons go here.\n",
	"quarantine/README.md":            "# Quarantine\n\nSensitive or risky content goes here.\n",
	"recall/README.md":                "# Recall\n\nGenerated task context packs can be written here.\n",
}

func Init(root string) error {
	if root == "" {
		return errors.New("repo path is required")
	}

	root = filepath.Clean(root)
	if err := ensureCanInitialize(root); err != nil {
		return err
	}

	for _, dir := range defaultDirs {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	for rel, content := range defaultFiles {
		if err := writeFileIfMissing(filepath.Join(root, rel), []byte(content)); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
	}

	return nil
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
