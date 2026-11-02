package projectmap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Mapping struct {
	Path    string `json:"path"`
	Project string `json:"project"`
}

type Store struct {
	Mappings []Mapping `json:"mappings"`
}

func Set(repoRoot string, workspacePath string, project string) error {
	if strings.TrimSpace(project) == "" {
		return errors.New("project name is required")
	}
	abs, err := filepath.Abs(workspacePath)
	if err != nil {
		return err
	}
	store, err := load(repoRoot)
	if err != nil {
		return err
	}
	for i, mapping := range store.Mappings {
		if mapping.Path == abs {
			store.Mappings[i].Project = project
			return save(repoRoot, store)
		}
	}
	store.Mappings = append(store.Mappings, Mapping{Path: abs, Project: project})
	sort.Slice(store.Mappings, func(i, j int) bool { return store.Mappings[i].Path < store.Mappings[j].Path })
	return save(repoRoot, store)
}

func Clear(repoRoot string, workspacePath string) error {
	abs, err := filepath.Abs(workspacePath)
	if err != nil {
		return err
	}
	store, err := load(repoRoot)
	if err != nil {
		return err
	}
	var out []Mapping
	for _, mapping := range store.Mappings {
		if mapping.Path != abs {
			out = append(out, mapping)
		}
	}
	store.Mappings = out
	return save(repoRoot, store)
}

func List(repoRoot string) ([]Mapping, error) {
	store, err := load(repoRoot)
	if err != nil {
		return nil, err
	}
	return store.Mappings, nil
}

func Resolve(repoRoot string, workspacePath string) (string, bool, error) {
	abs, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", false, err
	}
	store, err := load(repoRoot)
	if err != nil {
		return "", false, err
	}
	bestLen := -1
	best := ""
	for _, mapping := range store.Mappings {
		rel, err := filepath.Rel(mapping.Path, abs)
		if err != nil {
			continue
		}
		if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)) {
			if len(mapping.Path) > bestLen {
				bestLen = len(mapping.Path)
				best = mapping.Project
			}
		}
	}
	if best == "" {
		return "", false, nil
	}
	return best, true, nil
}

func load(repoRoot string) (Store, error) {
	path, err := storePath(repoRoot)
	if err != nil {
		return Store{}, err
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		content, err = os.ReadFile(filepath.Join(repoRoot, ".knowblazer", "projects.json"))
		if errors.Is(err, os.ErrNotExist) {
			return Store{}, nil
		}
	}
	if err != nil {
		return Store{}, err
	}
	var store Store
	if err := json.Unmarshal(content, &store); err != nil {
		return Store{}, err
	}
	return store, nil
}

func save(repoRoot string, store Store) error {
	path, err := storePath(repoRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o600)
}

// storePath namespaces machine-local mappings by the canonical memory repository path.
func storePath(repoRoot string) (string, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(root))
	return filepath.Join(configDir, "knowblazer", "projects", hex.EncodeToString(sum[:])+".json"), nil
}
