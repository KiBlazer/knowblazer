package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type Document struct {
	Path    string   `json:"path"`
	Terms   []string `json:"terms"`
	Snippet string   `json:"snippet"`
}

type Index struct {
	Documents []Document `json:"documents"`
}

type Hit struct {
	Path    string
	Score   int
	Snippet string
}

func Build(repoRoot string) (Index, error) {
	var idx Index
	for _, layer := range []string{"profile", "projects", "experience", "daily"} {
		root := filepath.Join(repoRoot, layer)
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || filepath.Ext(path) != ".md" || strings.EqualFold(entry.Name(), "README.md") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return nil
			}
			idx.Documents = append(idx.Documents, Document{Path: filepath.ToSlash(rel), Terms: uniqueTerms(string(content)), Snippet: snippet(string(content))})
			return nil
		})
	}
	if err := Save(repoRoot, idx); err != nil {
		return Index{}, err
	}
	return idx, nil
}

func Save(repoRoot string, idx Index) error {
	path := filepath.Join(repoRoot, ".knowblazer", "index.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func Load(repoRoot string) (Index, error) {
	content, err := os.ReadFile(filepath.Join(repoRoot, ".knowblazer", "index.json"))
	if err != nil {
		return Index{}, err
	}
	var idx Index
	if err := json.Unmarshal(content, &idx); err != nil {
		return Index{}, err
	}
	return idx, nil
}

func Search(repoRoot string, query string) ([]Hit, error) {
	idx, err := Load(repoRoot)
	if err != nil {
		idx, err = Build(repoRoot)
		if err != nil {
			return nil, err
		}
	}
	terms := uniqueTerms(query)
	if len(terms) == 0 {
		return nil, fmt.Errorf("search query is required")
	}
	var hits []Hit
	for _, doc := range idx.Documents {
		score := 0
		termSet := map[string]bool{}
		for _, term := range doc.Terms {
			termSet[term] = true
		}
		for _, term := range terms {
			if termSet[term] || strings.Contains(strings.ToLower(doc.Path), term) {
				score++
			}
		}
		if score > 0 {
			hits = append(hits, Hit{Path: doc.Path, Score: score, Snippet: doc.Snippet})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Path < hits[j].Path
		}
		return hits[i].Score > hits[j].Score
	})
	return hits, nil
}

func uniqueTerms(value string) []string {
	seen := map[string]bool{}
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	var out []string
	for _, field := range fields {
		if len(field) < 2 || seen[field] {
			continue
		}
		seen[field] = true
		out = append(out, field)
	}
	sort.Strings(out)
	return out
}

func snippet(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 240 {
		return value
	}
	return value[:240] + "..."
}
