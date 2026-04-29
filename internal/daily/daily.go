package daily

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AddOptions struct {
	Text string
	Now  time.Time
}

type ShowOptions struct {
	Date time.Time
}

func Add(repoRoot string, opts AddOptions) (string, error) {
	text := strings.TrimSpace(opts.Text)
	if text == "" {
		return "", fmt.Errorf("daily text is required")
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	path := filepath.Join(repoRoot, "daily", now.Format("2006-01-02")+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte("# Daily "+now.Format("2006-01-02")+"\n\n"), 0o644); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	entry := fmt.Sprintf("- %s %s\n", now.Format("15:04"), text)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(entry); err != nil {
		return "", err
	}
	return path, nil
}

func Show(repoRoot string, opts ShowOptions) ([]byte, string, error) {
	date := opts.Date
	if date.IsZero() {
		date = time.Now()
	}
	path := filepath.Join(repoRoot, "daily", date.Format("2006-01-02")+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, path, err
	}
	return content, path, nil
}

func ParseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}
