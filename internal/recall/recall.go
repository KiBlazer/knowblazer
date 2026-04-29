package recall

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const maxPackBytes = 20 * 1024

type Options struct {
	Task    string
	Project string
}

func Generate(repoRoot string, opts Options) ([]byte, error) {
	task := strings.TrimSpace(opts.Task)
	if task == "" {
		return nil, errors.New("task is required")
	}

	var out bytes.Buffer
	fmt.Fprintln(&out, "# Knowblazer Recall Pack")
	fmt.Fprintln(&out)
	fmt.Fprintf(&out, "Task: %s\n", task)
	fmt.Fprintf(&out, "Generated at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintln(&out)

	writeSection(&out, "Developer Preferences", readExisting(
		filepath.Join(repoRoot, "profile", "preferences.md"),
		filepath.Join(repoRoot, "profile", "decision-principles.md"),
	))

	if opts.Project != "" {
		writeSection(&out, "Project Context", readExisting(filepath.Join(repoRoot, "projects", opts.Project+".md")))
	}

	experience := relevantExperience(repoRoot, task, 5)
	writeSection(&out, "Relevant Experience", experience)

	daily := recentDaily(repoRoot)
	writeSection(&out, "Recent Daily Notes", daily)

	fmt.Fprintln(&out, "## Cautions")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "- This pack excludes inbox and quarantine by default.")
	fmt.Fprintln(&out, "- Verify commands and secrets before running anything.")
	return limitBytes(out.Bytes(), maxPackBytes), nil
}

func writeSection(out *bytes.Buffer, title string, parts []string) {
	fmt.Fprintf(out, "## %s\n\n", title)
	if len(parts) == 0 {
		fmt.Fprintln(out, "_No matching memory found._")
		fmt.Fprintln(out)
		return
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fmt.Fprintln(out, part)
		fmt.Fprintln(out)
	}
}

func readExisting(paths ...string) []string {
	var parts []string
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			parts = append(parts, limitString(string(content), 8*1024))
		}
	}
	return parts
}

func relevantExperience(repoRoot string, task string, maxFiles int) []string {
	type candidate struct {
		path  string
		score int
	}

	keywords := keywords(task)
	var candidates []candidate
	root := filepath.Join(repoRoot, "experience")
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" || strings.EqualFold(entry.Name(), "README.md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		score := scoreText(path+"\n"+string(content), keywords)
		if score > 0 {
			candidates = append(candidates, candidate{path: path, score: score})
		}
		return nil
	})
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].path < candidates[j].path
		}
		return candidates[i].score > candidates[j].score
	})

	var parts []string
	for i, candidate := range candidates {
		if i >= maxFiles {
			break
		}
		content, err := os.ReadFile(candidate.path)
		if err == nil {
			parts = append(parts, limitString(string(content), 8*1024))
		}
	}
	return parts
}

func recentDaily(repoRoot string) []string {
	now := time.Now()
	paths := []string{
		filepath.Join(repoRoot, "daily", now.Format("2006-01-02")+".md"),
		filepath.Join(repoRoot, "daily", now.AddDate(0, 0, -1).Format("2006-01-02")+".md"),
	}
	return readExisting(paths...)
}

func keywords(task string) []string {
	fields := strings.FieldsFunc(strings.ToLower(task), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	var out []string
	seen := map[string]bool{}
	for _, field := range fields {
		if len(field) < 2 || seen[field] {
			continue
		}
		seen[field] = true
		out = append(out, field)
	}
	return out
}

func scoreText(text string, keywords []string) int {
	text = strings.ToLower(text)
	score := 0
	for _, keyword := range keywords {
		score += strings.Count(text, keyword)
	}
	return score
}

func limitString(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "\n\n[truncated]"
}

func limitBytes(value []byte, max int) []byte {
	if len(value) <= max {
		return value
	}
	truncated := []byte("\n\n[truncated]")
	if max <= len(truncated) {
		return value[:max]
	}
	out := make([]byte, max)
	copy(out, value[:max-len(truncated)])
	copy(out[max-len(truncated):], truncated)
	return out
}
