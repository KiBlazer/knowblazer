package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/knowblazer/knowblazer/internal/bm25"
	"github.com/knowblazer/knowblazer/internal/textutil"
)

type Document struct {
	Path    string         `json:"path"`
	Terms   []string       `json:"terms,omitempty"`
	Freqs   map[string]int `json:"freqs,omitempty"`
	Length  int            `json:"length,omitempty"`
	Snippet string         `json:"snippet"`
}

type Index struct {
	Documents []Document `json:"documents"`
}

type Hit struct {
	Path    string  `json:"path"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet"`
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
			text := string(content)
			freqs, length := textutil.TermFrequencies(rel + "\n" + text)
			idx.Documents = append(idx.Documents, Document{
				Path:    filepath.ToSlash(rel),
				Terms:   uniqueTerms(text),
				Freqs:   freqs,
				Length:  length,
				Snippet: snippet(text),
			})
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

	var bmDocs []bm25.Document
	for _, doc := range idx.Documents {
		freqs := doc.Freqs
		length := doc.Length
		if freqs == nil || length == 0 {
			freqs, length = textutil.TermFrequencies(doc.Path + "\n" + strings.Join(doc.Terms, " "))
		}
		bmDocs = append(bmDocs, bm25.Document{
			Length: length,
			Freqs:  freqs,
		})
	}

	corpus := bm25.NewCorpus(bmDocs)
	var hits []Hit
	for i, doc := range idx.Documents {
		score := corpus.Score(bmDocs[i], terms)
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
	out := textutil.Tokenize(value)
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
