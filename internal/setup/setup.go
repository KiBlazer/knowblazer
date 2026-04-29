package setup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/knowblazer/knowblazer/internal/projectmap"
	"github.com/knowblazer/knowblazer/internal/repo"
)

type ClaudeOptions struct {
	RepoRoot      string
	Workspace     string
	Project       string
	Scope         string
	SkipMCP       bool
	EnsureProject bool
}

type ClaudeResult struct {
	ClaudeMD      string
	MCPConfigured bool
	MCPSkipped    bool
	ProjectMapped bool
}

func Claude(opts ClaudeOptions) (ClaudeResult, error) {
	if err := repo.MustBeRepo(opts.RepoRoot); err != nil {
		return ClaudeResult{}, err
	}
	workspace := opts.Workspace
	if workspace == "" {
		var err error
		workspace, err = os.Getwd()
		if err != nil {
			return ClaudeResult{}, err
		}
	}
	workspaceAbs, err := filepath.Abs(workspace)
	if err != nil {
		return ClaudeResult{}, err
	}
	info, err := os.Stat(workspaceAbs)
	if err != nil {
		return ClaudeResult{}, err
	}
	if !info.IsDir() {
		return ClaudeResult{}, fmt.Errorf("workspace path is not a directory: %s", workspaceAbs)
	}
	scope := opts.Scope
	if scope == "" {
		scope = "local"
	}
	if scope != "local" && scope != "user" && scope != "project" {
		return ClaudeResult{}, fmt.Errorf("unsupported Claude MCP scope: %s", scope)
	}

	claudeMD, err := ensureClaudeInstructions(workspaceAbs, opts.RepoRoot)
	if err != nil {
		return ClaudeResult{}, err
	}

	result := ClaudeResult{ClaudeMD: claudeMD}
	if opts.Project != "" {
		if opts.EnsureProject {
			if err := ensureProjectFile(opts.RepoRoot, opts.Project); err != nil {
				return ClaudeResult{}, err
			}
		}
		if err := projectmap.Set(opts.RepoRoot, workspaceAbs, opts.Project); err != nil {
			return ClaudeResult{}, err
		}
		result.ProjectMapped = true
	}
	if opts.SkipMCP {
		result.MCPSkipped = true
		return result, nil
	}
	configured, skipped, err := configureClaudeMCP(opts.RepoRoot, scope)
	if err != nil {
		return ClaudeResult{}, err
	}
	result.MCPConfigured = configured
	result.MCPSkipped = skipped
	return result, nil
}

func ensureProjectFile(repoRoot string, project string) error {
	path := filepath.Join(repoRoot, "projects", project+".md")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	title := strings.TrimSpace(project)
	if title == "" {
		title = "Project"
	}
	content := fmt.Sprintf("# %s\n\nAdd durable project context here.\n", title)
	return os.WriteFile(path, []byte(content), 0o644)
}

func ensureClaudeInstructions(workspace string, repoRoot string) (string, error) {
	path := filepath.Join(workspace, "CLAUDE.md")
	block := claudeInstructionBlock(repoRoot)
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	start := []byte("<!-- KNOWBLAZER-CLAUDE-SETUP:START -->")
	end := []byte("<!-- KNOWBLAZER-CLAUDE-SETUP:END -->")
	if startAt := bytes.Index(content, start); startAt >= 0 {
		if endAt := bytes.Index(content[startAt:], end); endAt >= 0 {
			endAt += startAt + len(end)
			next := append([]byte{}, content[:startAt]...)
			next = append(next, []byte(block)...)
			if endAt < len(content) {
				next = append(next, content[endAt:]...)
			}
			return path, os.WriteFile(path, next, 0o644)
		}
		return path, nil
	}
	var next []byte
	if len(bytes.TrimSpace(content)) > 0 {
		next = append(next, content...)
		if !bytes.HasSuffix(next, []byte("\n")) {
			next = append(next, '\n')
		}
		next = append(next, '\n')
	}
	next = append(next, []byte(block)...)
	return path, os.WriteFile(path, next, 0o644)
}

func claudeInstructionBlock(repoRoot string) string {
	return fmt.Sprintf("<!-- KNOWBLAZER-CLAUDE-SETUP:START -->\n"+
		"## Knowblazer Memory\n\n"+
		"Use Knowblazer as the local engineering memory source for this project.\n\n"+
		"- At the start of non-trivial coding tasks, call the Knowblazer MCP `knowblazer_context` tool with the task.\n"+
		"- Automatically remember durable lessons, project constraints, and reusable debugging or deployment facts with `knowblazer_remember` when they emerge during work; do not wait for an explicit reminder.\n"+
		"- Do not remember secrets, credentials, customer data, or temporary task chatter. If sensitive content is detected, Knowblazer will quarantine it.\n"+
		"- Use `knowblazer_status` if you need to check whether project memory is connected.\n"+
		"- If MCP is unavailable, run `knowblazer recall \"<task>\" --repo %s` and use the returned Markdown as context.\n"+
		"- Never read from or include `quarantine/` in task context.\n"+
		"<!-- KNOWBLAZER-CLAUDE-SETUP:END -->\n", repoRoot)
}

func configureClaudeMCP(repoRoot string, scope string) (bool, bool, error) {
	if _, err := exec.LookPath("claude"); err != nil {
		return false, false, fmt.Errorf("claude command not found in PATH")
	}
	executable, err := os.Executable()
	if err != nil {
		return false, false, err
	}
	if get := exec.Command("claude", "mcp", "get", "knowblazer"); get.Run() == nil {
		return false, true, nil
	}
	cmd := exec.Command("claude", "mcp", "add", "--scope", scope, "knowblazer", "--", executable, "mcp", "serve", "--repo", repoRoot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, false, fmt.Errorf("claude mcp add failed: %v\n%s", err, strings.TrimSpace(string(output)))
	}
	return true, false, nil
}
