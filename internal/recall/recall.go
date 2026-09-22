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

	"github.com/knowblazer/knowblazer/internal/bm25"
	"github.com/knowblazer/knowblazer/internal/textutil"
)

const (
	maxPackBytes    = 20 * 1024
	DefaultMinScore = 0.15
)

type Options struct {
	Task     string
	Project  string
	MinScore float64
	Explain  bool
}

type ExplainCandidate struct {
	Path       string
	Breadcrumb string
	Score      float64
	Priority   int
	Admitted   bool
}

type Diagnostics struct {
	Terms       []string
	MinScore    float64
	TotalScored int
	Candidates  []ExplainCandidate
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

	fmt.Fprintln(&out, "## Tier 1: Mandatory Constraints & Governance")
	fmt.Fprintln(&out, "> Directives in this tier represent mandatory project constraints and developer principles that must be strictly preserved.")
	fmt.Fprintln(&out)

	writeSection(&out, "Developer Preferences", readExisting(
		filepath.Join(repoRoot, "profile", "preferences.md"),
		filepath.Join(repoRoot, "profile", "decision-principles.md"),
	))

	if opts.Project != "" {
		writeSection(&out, "Project Context", readExisting(filepath.Join(repoRoot, "projects", opts.Project+".md")))
	}

	fmt.Fprintln(&out, "## Tier 2: Empirical Reference & Past Cases")
	fmt.Fprintln(&out, "> Heuristics in this tier provide historical troubleshooting patterns and empirical reference lessons to consult, not rigid commands.")
	fmt.Fprintln(&out)

	minScore := opts.MinScore
	if minScore <= 0 {
		minScore = DefaultMinScore
	}
	experience, diag := relevantExperience(repoRoot, task, 5, minScore)
	writeSection(&out, "Relevant Experience", experience)

	daily := recentDaily(repoRoot)
	writeSection(&out, "Recent Daily Notes", daily)

	fmt.Fprintln(&out, "## Cautions")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "- This pack excludes inbox and quarantine by default.")
	fmt.Fprintln(&out, "- Tier 1 contains mandatory engineering constraints that must be preserved.")
	fmt.Fprintln(&out, "- Tier 2 provides historical lessons and reference heuristics to consult, not rigid commands.")
	fmt.Fprintln(&out, "- It prioritizes synthesized memory and uses fresh automatic memory as lower-confidence context.")
	fmt.Fprintln(&out, "- Verify commands and secrets before running anything.")

	if opts.Explain && diag != nil {
		fmt.Fprintln(&out)
		fmt.Fprintln(&out, "## Recall Diagnostics")
		fmt.Fprintln(&out)
		fmt.Fprintf(&out, "- Extracted Terms: %v\n", diag.Terms)
		fmt.Fprintf(&out, "- Minimum Score Threshold: %.2f\n", diag.MinScore)
		fmt.Fprintf(&out, "- Total Sections Evaluated: %d\n", diag.TotalScored)
		fmt.Fprintln(&out, "- Candidate Scoring:")
		if len(diag.Candidates) == 0 {
			fmt.Fprintln(&out, "  _No candidates found._")
		} else {
			for _, c := range diag.Candidates {
				status := "rejected"
				if c.Admitted {
					status = "admitted"
				}
				fmt.Fprintf(&out, "  - `%s` (BM25: %.2f, priority: %d, %s)\n", c.Breadcrumb, c.Score, c.Priority, status)
			}
		}
	}

	return limitBytes(out.Bytes(), maxPackBytes), nil
}

func writeSection(out *bytes.Buffer, title string, parts []string) {
	fmt.Fprintf(out, "### %s\n\n", title)
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

type sectionItem struct {
	path       string
	breadcrumb string
	priority   int
	content    string
	bmDoc      bm25.Document
}

func parseSections(relPath string, content string, priority int) []sectionItem {
	lines := strings.Split(content, "\n")
	docTitle := ""

	inFrontmatter := false
	var bodyLines []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && trimmed == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if trimmed == "---" {
				inFrontmatter = false
				continue
			}
			if strings.HasPrefix(trimmed, "title:") {
				docTitle = strings.Trim(strings.TrimPrefix(trimmed, "title:"), ` "'`)
			}
			continue
		}
		if docTitle == "" && strings.HasPrefix(trimmed, "# ") {
			docTitle = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
		bodyLines = append(bodyLines, line)
	}

	if docTitle == "" {
		docTitle = strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath))
	}

	type rawSection struct {
		heading string
		lines   []string
	}
	var rawSections []rawSection
	var currentHeading string
	var currentLines []string

	for _, line := range bodyLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			if currentHeading != "" && len(currentLines) > 0 {
				rawSections = append(rawSections, rawSection{heading: currentHeading, lines: currentLines})
			}
			if strings.HasPrefix(trimmed, "## ") {
				currentHeading = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			} else {
				currentHeading = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			}
			currentLines = nil
			continue
		}
		if strings.HasPrefix(trimmed, "# ") && strings.TrimSpace(strings.TrimPrefix(trimmed, "# ")) == docTitle {
			continue
		}
		currentLines = append(currentLines, line)
	}
	if currentHeading != "" && len(currentLines) > 0 {
		rawSections = append(rawSections, rawSection{heading: currentHeading, lines: currentLines})
	}

	// If no subheadings found, treat entire document as a single section
	if len(rawSections) == 0 {
		freqs, length := textutil.TermFrequencies(relPath + "\n" + content)
		return []sectionItem{
			{
				path:       relPath,
				breadcrumb: docTitle,
				priority:   priority,
				content:    content,
				bmDoc: bm25.Document{
					Length: length,
					Freqs:  freqs,
				},
			},
		}
	}

	var out []sectionItem
	for _, sec := range rawSections {
		body := strings.TrimSpace(strings.Join(sec.lines, "\n"))
		if body == "" {
			continue
		}
		breadcrumb := docTitle + " > " + sec.heading
		rendered := fmt.Sprintf("#### %s\n\n%s", breadcrumb, body)
		freqs, length := textutil.TermFrequencies(relPath + "\n" + breadcrumb + "\n" + body)
		out = append(out, sectionItem{
			path:       relPath,
			breadcrumb: breadcrumb,
			priority:   priority,
			content:    rendered,
			bmDoc: bm25.Document{
				Length: length,
				Freqs:  freqs,
			},
		})
	}
	return out
}

func relevantExperience(repoRoot string, task string, maxFiles int, minScore float64) ([]string, *Diagnostics) {
	keywords := keywords(task)
	if len(keywords) == 0 {
		return nil, nil
	}

	diag := &Diagnostics{
		Terms:    keywords,
		MinScore: minScore,
	}

	var allSections []sectionItem
	var bmDocs []bm25.Document

	root := filepath.Join(repoRoot, "experience")
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" || strings.EqualFold(entry.Name(), "README.md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)
		priority := experiencePriority(repoRoot, path)
		if priority == 2 && !isFreshAutoMemory(text) {
			return nil
		}

		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			rel = path
		}
		secs := parseSections(filepath.ToSlash(rel), text, priority)
		for _, sec := range secs {
			allSections = append(allSections, sec)
			bmDocs = append(bmDocs, sec.bmDoc)
		}
		return nil
	})

	if len(bmDocs) == 0 {
		return nil, diag
	}

	diag.TotalScored = len(allSections)
	corpus := bm25.NewCorpus(bmDocs)
	type candidate struct {
		content  string
		path     string
		score    float64
		priority int
	}
	var candidates []candidate
	for _, sec := range allSections {
		score := corpus.Score(sec.bmDoc, keywords)
		admitted := score >= minScore
		diag.Candidates = append(diag.Candidates, ExplainCandidate{
			Path:       sec.path,
			Breadcrumb: sec.breadcrumb,
			Score:      score,
			Priority:   sec.priority,
			Admitted:   admitted,
		})
		if admitted {
			candidates = append(candidates, candidate{
				content:  sec.content,
				path:     sec.path,
				score:    score,
				priority: sec.priority,
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority < candidates[j].priority
		}
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
		parts = append(parts, limitString(candidate.content, 8*1024))
	}
	return LongContextReorder(parts), diag
}

// LongContextReorder reorders items according to the U-shaped attention distribution:
// the most relevant items are placed at the beginning and end, while less relevant
// items are placed in the middle.
func LongContextReorder(items []string) []string {
	if len(items) <= 2 {
		out := make([]string, len(items))
		copy(out, items)
		return out
	}

	var left []string
	var right []string

	for i, item := range items {
		if i%2 == 0 {
			left = append(left, item)
		} else {
			right = append(right, item)
		}
	}

	for i, j := 0, len(right)-1; i < j; i, j = i+1, j-1 {
		right[i], right[j] = right[j], right[i]
	}

	return append(left, right...)
}

func experiencePriority(repoRoot string, path string) int {
	rel, err := filepath.Rel(filepath.Join(repoRoot, "experience"), path)
	if err != nil {
		return 1
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "synthesized/") {
		return 0
	}
	if strings.HasPrefix(rel, "auto/") {
		return 2
	}
	return 1
}

func isFreshAutoMemory(content string) bool {
	return strings.Contains(content, `status: "fresh"`) || strings.Contains(content, `status: "auto_promoted"`)
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
	return textutil.Tokenize(task)
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
