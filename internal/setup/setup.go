package setup

import (
	"bytes"
	"encoding/json"
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

type CodexOptions struct {
	RepoRoot      string
	Workspace     string
	Project       string
	SkipMCP       bool
	EnsureProject bool
}

type CodexResult struct {
	AgentsMD      string
	MCPConfigured bool
	MCPSkipped    bool
	ProjectMapped bool
}

func Claude(opts ClaudeOptions) (ClaudeResult, error) {
	if err := repo.MustBeRepo(opts.RepoRoot); err != nil {
		return ClaudeResult{}, err
	}
	workspaceAbs, err := workspaceAbs(opts.Workspace)
	if err != nil {
		return ClaudeResult{}, err
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
			if err := ensureProjectFile(opts.RepoRoot, opts.Project, workspaceAbs); err != nil {
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

func Codex(opts CodexOptions) (CodexResult, error) {
	if err := repo.MustBeRepo(opts.RepoRoot); err != nil {
		return CodexResult{}, err
	}
	workspaceAbs, err := workspaceAbs(opts.Workspace)
	if err != nil {
		return CodexResult{}, err
	}

	agentsMD, err := ensureCodexInstructions(workspaceAbs, opts.RepoRoot)
	if err != nil {
		return CodexResult{}, err
	}

	result := CodexResult{AgentsMD: agentsMD}
	if opts.Project != "" {
		if opts.EnsureProject {
			if err := ensureProjectFile(opts.RepoRoot, opts.Project, workspaceAbs); err != nil {
				return CodexResult{}, err
			}
		}
		if err := projectmap.Set(opts.RepoRoot, workspaceAbs, opts.Project); err != nil {
			return CodexResult{}, err
		}
		result.ProjectMapped = true
	}
	if opts.SkipMCP {
		result.MCPSkipped = true
		return result, nil
	}
	configured, skipped, err := configureCodexMCP(opts.RepoRoot)
	if err != nil {
		return CodexResult{}, err
	}
	result.MCPConfigured = configured
	result.MCPSkipped = skipped
	return result, nil
}

func workspaceAbs(workspace string) (string, error) {
	if workspace == "" {
		var err error
		workspace, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace path is not a directory: %s", abs)
	}
	return abs, nil
}

func ensureProjectFile(repoRoot string, project string, workspace string) error {
	path := filepath.Join(repoRoot, "projects", project+".md")
	title := strings.TrimSpace(project)
	if title == "" {
		title = "Project"
	}
	content := projectScaffold(title, workspace)
	if existing, err := os.ReadFile(path); err == nil {
		if string(existing) == oldProjectStub(title) {
			return os.WriteFile(path, []byte(content), 0o644)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func oldProjectStub(title string) string {
	return fmt.Sprintf("# %s\n\nAdd durable project context here.\n", title)
}

func projectScaffold(title string, workspace string) string {
	return fmt.Sprintf(`# %s

## Workspace

- Path: %s

## Purpose

- Add what this project is for.

## Durable Constraints

- Add constraints future agents must preserve.

## Common Commands

- Add build, test, lint, deploy, and smoke-test commands.

## Operating Notes

- Add recurring debugging, deployment, or environment lessons.

## Open Questions

- Add project-memory gaps worth filling later.
`, title, workspace)
}

func ensureClaudeInstructions(workspace string, repoRoot string) (string, error) {
	path := filepath.Join(workspace, "CLAUDE.md")
	block := claudeInstructionBlock(repoRoot, workspace)
	return ensureInstructionBlock(path, block, "KNOWBLAZER-CLAUDE-SETUP")
}

func ensureCodexInstructions(workspace string, repoRoot string) (string, error) {
	path := filepath.Join(workspace, "AGENTS.md")
	block := codexInstructionBlock(repoRoot, workspace)
	return ensureInstructionBlock(path, block, "KNOWBLAZER-CODEX-SETUP")
}

func ensureInstructionBlock(path string, block string, marker string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	start := []byte("<!-- " + marker + ":START -->")
	end := []byte("<!-- " + marker + ":END -->")
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

func claudeInstructionBlock(repoRoot string, workspace string) string {
	return instructionBlock("KNOWBLAZER-CLAUDE-SETUP", repoRoot, workspace)
}

func codexInstructionBlock(repoRoot string, workspace string) string {
	return instructionBlock("KNOWBLAZER-CODEX-SETUP", repoRoot, workspace)
}

func instructionBlock(marker string, repoRoot string, workspace string) string {
	return fmt.Sprintf("<!-- %s:START -->\n"+
		"## Knowblazer Memory\n\n"+
		"Use Knowblazer as the local engineering memory source for this project.\n\n"+
		"- At the start of non-trivial coding tasks, call the Knowblazer MCP `knowblazer_context` tool with the task and `workspace: %q`.\n"+
		"- Before the final response on non-trivial tasks, do a silent end-of-task memory candidate review.\n"+
		"- Call `knowblazer_remember` when durable project constraints, debugging lessons, build/test/deploy caveats, environment setup facts, architectural decisions, or reusable tool behaviors emerge.\n"+
		"- Do not remember one-off task chatter, temporary implementation details, secrets, credentials, customer data, or facts already present in recalled context.\n"+
		"- Use `knowblazer_status` if you need to check whether project memory is connected.\n"+
		"- If MCP is unavailable, run `knowblazer recall \"<task>\" --repo %s` and use the returned Markdown as context.\n"+
		"- Never read from or include `quarantine/` in task context.\n"+
		"<!-- %s:END -->\n", marker, workspace, repoRoot, marker)
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

func configureCodexMCP(repoRoot string) (bool, bool, error) {
	if _, err := exec.LookPath("codex"); err != nil {
		return false, false, fmt.Errorf("codex command not found in PATH")
	}
	executable, err := os.Executable()
	if err != nil {
		return false, false, err
	}
	if get := exec.Command("codex", "mcp", "get", "knowblazer").Run(); get == nil {
		return false, true, nil
	}
	cmd := exec.Command("codex", "mcp", "add", "knowblazer", "--", executable, "mcp", "serve", "--repo", repoRoot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, false, fmt.Errorf("codex mcp add failed: %v\n%s", err, strings.TrimSpace(string(output)))
	}
	return true, false, nil
}

type AntigravityOptions struct {
	RepoRoot  string
	Workspace string
	Project   string
	SkipMCP   bool
}

type AntigravityResult struct {
	MCPConfigured bool
	MCPSkipped    bool
}

func Antigravity(opts AntigravityOptions) (AntigravityResult, error) {
	if err := repo.MustBeRepo(opts.RepoRoot); err != nil {
		return AntigravityResult{}, err
	}
	if opts.SkipMCP {
		return AntigravityResult{MCPSkipped: true}, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return AntigravityResult{}, err
	}

	paths := []string{
		filepath.Join(home, ".gemini", "antigravity", "mcp_config.json"),
		filepath.Join(home, ".gemini", "config", "mcp_config.json"),
	}

	executable, err := os.Executable()
	if err != nil {
		executable = "knowblazer"
	}

	serverConfig := map[string]any{
		"command": executable,
		"args":    []string{"mcp", "serve", "--repo", opts.RepoRoot},
	}

	configured := false
	for _, p := range paths {
		data := make(map[string]any)
		if content, err := os.ReadFile(p); err == nil {
			_ = json.Unmarshal(content, &data)
		}

		mcpServers, ok := data["mcpServers"].(map[string]any)
		if !ok {
			mcpServers = make(map[string]any)
			data["mcpServers"] = mcpServers
		}

		mcpServers["knowblazer"] = serverConfig

		dir := filepath.Dir(p)
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}

		content, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			continue
		}
		content = append(content, '\n')
		if err := os.WriteFile(p, content, 0644); err == nil {
			configured = true
		}
	}

	return AntigravityResult{MCPConfigured: configured}, nil
}
