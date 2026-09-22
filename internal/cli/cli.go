package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/knowblazer/knowblazer/internal/backup"
	"github.com/knowblazer/knowblazer/internal/capture"
	"github.com/knowblazer/knowblazer/internal/consolidate"
	"github.com/knowblazer/knowblazer/internal/daily"
	"github.com/knowblazer/knowblazer/internal/doctor"
	"github.com/knowblazer/knowblazer/internal/dream"
	"github.com/knowblazer/knowblazer/internal/importer/specstory"
	"github.com/knowblazer/knowblazer/internal/index"
	"github.com/knowblazer/knowblazer/internal/mcp"
	"github.com/knowblazer/knowblazer/internal/projectmap"
	"github.com/knowblazer/knowblazer/internal/recall"
	"github.com/knowblazer/knowblazer/internal/repo"
	"github.com/knowblazer/knowblazer/internal/review"
	"github.com/knowblazer/knowblazer/internal/scan"
	"github.com/knowblazer/knowblazer/internal/setup"
	ksync "github.com/knowblazer/knowblazer/internal/sync"
	"github.com/knowblazer/knowblazer/internal/update"
	"github.com/knowblazer/knowblazer/internal/version"
)

var setupLookPath = func(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	// --- Core Daily Workflow ---
	case "setup":
		return runSetup(args[1:], stdout, stderr)
	case "remember":
		return runRemember(args[1:], stdout, stderr)
	case "recall":
		return runRecall(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "sync":
		return runSync(args[1:], stdout, stderr)

	// --- Knowledge Management & Curation ---
	case "memory":
		return runMemory(args[1:], stdout, stderr)
	case "daily":
		return runDaily(args[1:], stdout, stderr)
	case "project":
		return runProject(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "backup":
		return runBackup(args[1:], stdout, stderr)

	// --- System & Integration ---
	case "mcp":
		return runMCP(args[1:], stdout, stderr)
	case "version", "-v", "--version":
		return runVersion(args[1:], stdout, stderr)
	case "update", "upgrade":
		return runUpdate(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0

	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		printUsage(stderr)
		return 2
	}
}

// ---------------------------------------------------------------------------
// Core Commands
// ---------------------------------------------------------------------------

func runSetup(args []string, stdout io.Writer, stderr io.Writer) int {
	isGlobal := hasFlag(args, "--global") || hasFlag(args, "-g")
	skipMCP := hasFlag(args, "--skip-mcp")
	tool := flagValue(args, "--tool")
	positionals := nonFlagArgs(args)

	// Check if first positional is a supported tool
	if tool == "" && len(positionals) > 0 {
		switch positionals[0] {
		case "claude", "codex", "antigravity", "agy":
			tool = positionals[0]
			positionals = positionals[1:]
		}
	}

	customPath := ""
	if len(positionals) > 0 {
		customPath = positionals[0]
	}

	repoRoot := flagValue(args, "--repo")
	if repoRoot == "" && customPath != "" {
		repoRoot = customPath
	}
	if repoRoot == "" {
		repoRoot = os.Getenv("KNOWBLAZER_REPO")
	}
	if repoRoot == "" {
		repoRoot = filepath.Join(os.Getenv("HOME"), "knowblazer-notes")
	}

	// Ensure repository is initialized
	if err := repo.Init(repoRoot); err != nil {
		fmt.Fprintf(stderr, "setup failed to init repo %s: %v\n", repoRoot, err)
		return 1
	}

	workspace := flagValue(args, "--path")
	if workspace == "" && customPath == "" && !isGlobal {
		workspace, _ = os.Getwd()
	}

	project := flagValue(args, "--project")
	if project == "" && workspace != "" {
		project = inferProjectName(workspace)
	}

	// 1. Global mode: inject into user-level instructions
	if isGlobal {
		codexRes, err := setup.Codex(setup.CodexOptions{
			RepoRoot: repoRoot,
			Global:   true,
			SkipMCP:  skipMCP,
		})
		if err != nil {
			fmt.Fprintf(stderr, "setup global codex failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Global Codex/Agent instructions updated: %s\n", codexRes.AgentsMD)

		claudeRes, err := setup.Claude(setup.ClaudeOptions{
			RepoRoot: repoRoot,
			Global:   true,
			Scope:    "user",
			SkipMCP:  skipMCP,
		})
		if err != nil {
			fmt.Fprintf(stderr, "setup global claude failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Global Claude instructions updated: %s\n", claudeRes.ClaudeMD)

		_, _ = setup.Antigravity(setup.AntigravityOptions{
			RepoRoot: repoRoot,
			SkipMCP:  skipMCP,
		})

		fmt.Fprintln(stdout, "Global Knowblazer memory instructions ready across all workspaces.")
		return 0
	}

	// 2. Specific tool mode
	if tool != "" {
		switch tool {
		case "codex":
			result, err := setup.Codex(setup.CodexOptions{
				RepoRoot:      repoRoot,
				Workspace:     workspace,
				Project:       project,
				SkipMCP:       skipMCP,
				EnsureProject: true,
			})
			if err != nil {
				fmt.Fprintf(stderr, "setup codex failed: %v\n", err)
				return 1
			}
			fmt.Fprintln(stdout, "Knowblazer is ready for Codex in this project.")
			fmt.Fprintf(stdout, "Memory repo: %s\n", repoRoot)
			fmt.Fprintf(stdout, "Project: %s\n", project)
			printCodexStartResult(stdout, result)
			fmt.Fprintln(stdout, "Next: run `codex` from this project.")
			return 0

		case "claude":
			result, err := setup.Claude(setup.ClaudeOptions{
				RepoRoot:      repoRoot,
				Workspace:     workspace,
				Project:       project,
				Scope:         flagValue(args, "--scope"),
				SkipMCP:       skipMCP,
				EnsureProject: true,
			})
			if err != nil {
				fmt.Fprintf(stderr, "setup claude failed: %v\n", err)
				return 1
			}
			fmt.Fprintln(stdout, "Knowblazer is ready for Claude Code in this project.")
			fmt.Fprintf(stdout, "Memory repo: %s\n", repoRoot)
			fmt.Fprintf(stdout, "Project: %s\n", project)
			printClaudeStartResult(stdout, result)
			fmt.Fprintln(stdout, "Next: run `claude` from this project.")
			return 0

		case "antigravity", "agy":
			result, err := setup.Antigravity(setup.AntigravityOptions{
				RepoRoot:  repoRoot,
				Workspace: workspace,
				Project:   project,
				SkipMCP:   skipMCP,
			})
			if err != nil {
				fmt.Fprintf(stderr, "setup antigravity failed: %v\n", err)
				return 1
			}
			if result.MCPConfigured {
				fmt.Fprintln(stdout, "Antigravity MCP configured: knowblazer")
			}
			return 0

		default:
			fmt.Fprintf(stderr, "unknown tool: %s (supported: claude, codex, antigravity)\n", tool)
			return 2
		}
	}

	// 3. Workspace auto-detection mode
	if workspace == "" {
		fmt.Fprintf(stdout, "Initialized Knowblazer memory repo: %s\n", repoRoot)
		fmt.Fprintln(stdout, "Next steps:")
		fmt.Fprintf(stdout, "  1. Connect your workspace: knowblazer setup\n")
		fmt.Fprintf(stdout, "  2. Remember lessons: knowblazer remember \"<lesson>\"\n")
		fmt.Fprintf(stdout, "  3. Recall context: knowblazer recall \"<task>\"\n")
		return 0
	}

	claudeInstalled := setupLookPath("claude")
	codexInstalled := setupLookPath("codex")
	antigravityInstalled := setupLookPath("antigravity") || setupLookPath("agy") || isAntigravityEnv()

	if skipMCP {
		claudeInstalled = true
		codexInstalled = true
	}

	if !claudeInstalled && !codexInstalled && !antigravityInstalled {
		if project != "" {
			_ = setup.EnsureProjectFile(repoRoot, project, workspace)
			_ = projectmap.Set(repoRoot, workspace, project)
		}
		fmt.Fprintf(stdout, "Initialized Knowblazer memory repo at: %s\n", repoRoot)
		if project != "" {
			fmt.Fprintf(stdout, "Project ready: %s (%s)\n", project, workspace)
		}
		return 0
	}

	var claudeRes setup.ClaudeResult
	var codexRes setup.CodexResult
	var agyRes setup.AntigravityResult
	var claudeErr, codexErr, agyErr error

	if claudeInstalled {
		claudeRes, claudeErr = setup.Claude(setup.ClaudeOptions{
			RepoRoot:      repoRoot,
			Workspace:     workspace,
			Project:       project,
			Scope:         flagValue(args, "--scope"),
			SkipMCP:       skipMCP,
			EnsureProject: true,
		})
	}
	if codexInstalled {
		codexRes, codexErr = setup.Codex(setup.CodexOptions{
			RepoRoot:      repoRoot,
			Workspace:     workspace,
			Project:       project,
			SkipMCP:       skipMCP,
			EnsureProject: true,
		})
	}
	if antigravityInstalled {
		agyRes, agyErr = setup.Antigravity(setup.AntigravityOptions{
			RepoRoot:  repoRoot,
			Workspace: workspace,
			Project:   project,
			SkipMCP:   skipMCP,
		})
	}

	fmt.Fprintln(stdout, startReadySummary(claudeInstalled && claudeErr == nil, codexInstalled && codexErr == nil, antigravityInstalled && agyErr == nil))
	fmt.Fprintf(stdout, "Memory repo: %s\n", repoRoot)
	if project != "" {
		fmt.Fprintf(stdout, "Project: %s\n", project)
	}
	if claudeInstalled {
		if claudeErr != nil {
			fmt.Fprintf(stdout, "Claude Code MCP failed: %v\n", claudeErr)
		} else {
			printClaudeStartResult(stdout, claudeRes)
		}
	}
	if codexInstalled {
		if codexErr != nil {
			fmt.Fprintf(stdout, "Codex MCP failed: %v\n", codexErr)
		} else {
			printCodexStartResult(stdout, codexRes)
		}
	}
	if antigravityInstalled {
		if agyErr != nil {
			fmt.Fprintf(stdout, "Antigravity MCP failed: %v\n", agyErr)
		} else {
			if agyRes.MCPConfigured {
				fmt.Fprintln(stdout, "Antigravity MCP configured: knowblazer")
			} else if agyRes.MCPSkipped {
				fmt.Fprintln(stdout, "Antigravity MCP skipped")
			} else {
				fmt.Fprintln(stdout, "Antigravity MCP already configured: knowblazer")
			}
		}
	}
	fmt.Fprintf(stdout, "Next: run %s from this project.\n", startNextCommand(claudeInstalled && claudeErr == nil, codexInstalled && codexErr == nil, antigravityInstalled && agyErr == nil))
	return 0
}

func startReadySummary(claude bool, codex bool, antigravity bool) string {
	var parts []string
	if claude {
		parts = append(parts, "Claude Code")
	}
	if codex {
		parts = append(parts, "Codex")
	}
	if antigravity {
		parts = append(parts, "Antigravity")
	}
	if len(parts) == 0 {
		return "Knowblazer is ready in this project."
	}
	return fmt.Sprintf("Knowblazer is ready for %s in this project.", strings.Join(parts, " and "))
}

func startNextCommand(claude bool, codex bool, antigravity bool) string {
	var parts []string
	if claude {
		parts = append(parts, "`claude`")
	}
	if codex {
		parts = append(parts, "`codex`")
	}
	if antigravity {
		parts = append(parts, "`agy`")
	}
	if len(parts) == 0 {
		return "`knowblazer recall`"
	}
	return strings.Join(parts, " or ")
}

func isAntigravityEnv() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".gemini", "antigravity"))
	return err == nil
}

func runRemember(args []string, stdout io.Writer, stderr io.Writer) int {
	isDaily := hasFlag(args, "--daily") || hasFlag(args, "-d")
	nonFlags := nonFlagArgs(args)
	if len(nonFlags) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer remember <text|file> [-d|--daily] [--repo <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	text := strings.TrimSpace(strings.Join(nonFlags, " "))
	if isDaily {
		path, err := daily.Add(repoRoot, daily.AddOptions{Text: text})
		if err != nil {
			fmt.Fprintf(stderr, "remember daily failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Remembered to daily notes: %s\n", path)
		return 0
	}

	// Check if single positional argument is an existing markdown file
	if len(nonFlags) == 1 {
		candidate := nonFlags[0]
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			result, err := capture.MarkdownAuto(repoRoot, candidate)
			if err != nil {
				fmt.Fprintf(stderr, "remember failed: %v\n", err)
				return 1
			}
			if result.ScanLevel == scan.High {
				fmt.Fprintf(stderr, "sensitive content detected; saved to quarantine: %s\n", result.Path)
				return 1
			}
			fmt.Fprintf(stdout, "Remembered to fresh memory: %s\n", result.Path)
			if !autoConsolidate(repoRoot, stdout, stderr) {
				return 1
			}
			return 0
		}
	}

	path, err := rememberText(repoRoot, text)
	if err != nil {
		fmt.Fprintf(stderr, "remember failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Remembered to fresh memory: %s\n", path)
	if !autoConsolidate(repoRoot, stdout, stderr) {
		return 1
	}
	return 0
}

func rememberText(repoRoot string, text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("memory text is empty")
	}
	file, err := os.CreateTemp("", "knowblazer-remember-*.md")
	if err != nil {
		return "", err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err := fmt.Fprintf(file, "# Memory\n\n%s\n", text); err != nil {
		_ = file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	result, err := capture.MarkdownAuto(repoRoot, name)
	if err != nil {
		return "", err
	}
	if result.ScanLevel == scan.High {
		return "", fmt.Errorf("sensitive content detected; saved to quarantine: %s", result.Path)
	}
	return result.Path, nil
}

func autoConsolidate(repoRoot string, stdout io.Writer, stderr io.Writer) bool {
	result, err := consolidate.Run(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "automatic consolidation failed: %v\n", err)
		return false
	}
	if result.Count > 0 {
		fmt.Fprintf(stdout, "Auto-consolidated %d fresh memories into: %s\n", result.Count, result.Path)
	}
	return true
}

func runRecall(args []string, stdout io.Writer, stderr io.Writer) int {
	task := flagValue(args, "--task")
	if task == "" {
		task = strings.TrimSpace(strings.Join(nonFlagArgs(args), " "))
	}
	if task == "" {
		fmt.Fprintln(stderr, "usage: knowblazer recall <task> [--project <name>] [--output <file>] [--repo <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}

	project := flagValue(args, "--project")
	if project == "" {
		if cwd, err := os.Getwd(); err == nil {
			if mapped, ok, err := projectmap.Resolve(repoRoot, cwd); err == nil && ok {
				project = mapped
			} else {
				project = inferProjectName(cwd)
			}
		}
	}

	output := flagValue(args, "--output")
	pack, err := recall.Generate(repoRoot, recall.Options{Task: task, Project: project})
	if err != nil {
		fmt.Fprintf(stderr, "recall failed: %v\n", err)
		return 1
	}
	if output != "" {
		if err := os.WriteFile(output, pack, 0o644); err != nil {
			fmt.Fprintf(stderr, "write recall output failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Recall pack written to: %s\n", output)
		return 0
	}
	fmt.Fprint(stdout, string(pack))
	return 0
}

func runStatus(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	workspace := flagValue(args, "--path")
	if workspace == "" {
		workspace, _ = os.Getwd()
	}

	project := ""
	isAutoInferred := false
	if workspace != "" {
		if mapped, ok, err := projectmap.Resolve(repoRoot, workspace); err == nil && ok {
			project = mapped
		} else {
			project = inferProjectName(workspace)
			isAutoInferred = true
		}
	}

	fmt.Fprintf(stdout, "Memory repo: %s\n", repoRoot)
	if project != "" {
		if isAutoInferred {
			fmt.Fprintf(stdout, "Project: %s (auto-detected)\n", project)
		} else {
			fmt.Fprintf(stdout, "Project: %s\n", project)
		}
		fmt.Fprintf(stdout, "Project memory: %s\n", projectMemoryState(repoRoot, project))
	} else {
		fmt.Fprintln(stdout, "Project: not mapped")
		fmt.Fprintln(stdout, "Project memory: unmapped")
	}

	// Check Claude instructions
	claudeMD := filepath.Join(workspace, "CLAUDE.md")
	if checkFileContains(claudeMD, "KNOWBLAZER-CLAUDE-SETUP:START") {
		fmt.Fprintf(stdout, "Claude instructions: %s\n", claudeMD)
	} else if globalClaude := filepath.Join(os.Getenv("HOME"), ".claude", "CLAUDE.md"); checkFileContains(globalClaude, "KNOWBLAZER-") {
		fmt.Fprintf(stdout, "Claude instructions: %s (global)\n", globalClaude)
	} else {
		fmt.Fprintln(stdout, "Claude instructions: not configured")
	}

	// Check Codex instructions
	agentsMD := filepath.Join(workspace, "AGENTS.md")
	if checkFileContains(agentsMD, "KNOWBLAZER-") {
		fmt.Fprintf(stdout, "Codex instructions: %s\n", agentsMD)
	} else if globalAgents := filepath.Join(os.Getenv("HOME"), ".agents", "AGENTS.md"); checkFileContains(globalAgents, "KNOWBLAZER-") {
		fmt.Fprintf(stdout, "Codex instructions: %s (global)\n", globalAgents)
	} else {
		fmt.Fprintln(stdout, "Codex instructions: not configured")
	}

	// Check Gemini instructions
	geminiMD := filepath.Join(workspace, "GEMINI.md")
	if checkFileContains(geminiMD, "KNOWBLAZER-") {
		fmt.Fprintf(stdout, "Gemini instructions: %s\n", geminiMD)
	} else if globalGemini := filepath.Join(os.Getenv("HOME"), ".gemini", "GEMINI.md"); checkFileContains(globalGemini, "KNOWBLAZER-") {
		fmt.Fprintf(stdout, "Gemini instructions: %s (global)\n", globalGemini)
	} else {
		fmt.Fprintln(stdout, "Gemini instructions: not configured")
	}

	freshCount, _ := consolidate.CountFresh(repoRoot)
	synthesizedCount, _ := consolidate.CountSynthesized(repoRoot)
	candidates, _ := review.List(repoRoot)

	fmt.Fprintf(stdout, "Fresh auto memories: %d\n", freshCount)
	fmt.Fprintf(stdout, "Synthesized memories: %d\n", synthesizedCount)
	fmt.Fprintf(stdout, "Inbox candidates: %d\n", len(candidates))
	return 0
}

func runSync(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	action := "status"
	positionals := nonFlagArgs(args)
	if len(positionals) > 0 {
		action = positionals[0]
	}

	switch action {
	case "status":
		res, err := ksync.Status(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "sync status failed: %v\n", err)
			return 1
		}
		if strings.TrimSpace(res.Output) == "" {
			fmt.Fprintln(stdout, "Working tree clean.")
			return 0
		}
		fmt.Fprint(stdout, res.Output)
		return 0

	case "commit":
		message := flagValue(args, "--message")
		if message == "" {
			message = flagValue(args, "-m")
		}
		if message == "" {
			message = "sync: update knowblazer memory"
		}
		res, err := ksync.Commit(repoRoot, message)
		if err != nil {
			fmt.Fprintf(stderr, "sync commit failed: %v\n", err)
			return 1
		}
		if strings.TrimSpace(res.Output) == "" {
			fmt.Fprintln(stdout, "Committed memory changes.")
		} else {
			fmt.Fprint(stdout, res.Output)
		}
		return 0

	case "push":
		res, err := ksync.Push(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "sync push failed: %v\n%s\n", err, res.Output)
			return 1
		}
		fmt.Fprint(stdout, res.Output)
		return 0

	case "pull":
		res, err := ksync.Pull(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "sync pull failed: %v\n%s\n", err, res.Output)
			return 1
		}
		fmt.Fprint(stdout, res.Output)
		return 0

	default:
		fmt.Fprintf(stderr, "unknown sync action: %s (supported: status, commit, push, pull)\n", action)
		return 2
	}
}

// ---------------------------------------------------------------------------
// Knowledge Management & Curation Commands
// ---------------------------------------------------------------------------

func runMemory(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer memory <list|review|promote|reject|consolidate|search|capture|import|scan> [args]")
		return 2
	}
	if args[0] == "scan" {
		values := nonFlagArgs(args[1:])
		if len(values) < 1 {
			fmt.Fprintln(stderr, "usage: knowblazer memory scan <path>")
			return 2
		}
		result, err := scan.Path(values[0])
		if err != nil {
			fmt.Fprintf(stderr, "scan failed: %v\n", err)
			return 1
		}
		highCount := 0
		for _, finding := range result.Findings {
			if finding.Level == scan.High {
				highCount++
			}
			fmt.Fprintf(stdout, "%s  %s  %s\n", levelString(finding.Level), finding.Rule, finding.Snippet)
		}
		fmt.Fprintf(stdout, "Summary: %s, %d findings, %d high-risk\n", result.Level.String(), len(result.Findings), highCount)
		if result.Level == scan.High {
			return 1
		}
		return 0
	}

	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}

	switch args[0] {
	case "list":
		candidates, err := review.List(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "memory list failed: %v\n", err)
			return 1
		}
		if len(candidates) == 0 {
			fmt.Fprintln(stdout, "No inbox candidates.")
			return 0
		}
		for _, candidate := range candidates {
			fmt.Fprintf(stdout, "%s  %s\n", levelString(candidate.Level), candidate.Path)
		}
		return 0

	case "review":
		if len(args) > 1 && args[1] == "suggest" {
			path, err := dream.Run(repoRoot)
			if err != nil {
				fmt.Fprintf(stderr, "memory review suggest failed: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "Memory suggestions generated: %s\n", path)
			return 0
		}
		candidates, err := review.List(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "memory review failed: %v\n", err)
			return 1
		}
		for _, candidate := range candidates {
			fmt.Fprintf(stdout, "%s  %s\n", levelString(candidate.Level), candidate.Path)
		}
		return 0

	case "promote":
		values := nonFlagArgs(args[1:])
		target := flagValue(args[1:], "--to")
		if len(values) != 1 || target == "" {
			fmt.Fprintln(stderr, "usage: knowblazer memory promote <file> --to <target>")
			return 2
		}
		result, err := review.Promote(repoRoot, values[0], target)
		if err != nil {
			fmt.Fprintf(stderr, "memory promote failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Promoted to: %s\n", result.Path)
		return 0

	case "reject":
		values := nonFlagArgs(args[1:])
		if len(values) != 1 {
			fmt.Fprintln(stderr, "usage: knowblazer memory reject <file>")
			return 2
		}
		path, err := review.Reject(repoRoot, values[0])
		if err != nil {
			fmt.Fprintf(stderr, "memory reject failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Rejected to: %s\n", path)
		return 0

	case "consolidate":
		result, err := consolidate.Run(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "consolidate failed: %v\n", err)
			return 1
		}
		if result.Count == 0 {
			fmt.Fprintln(stdout, "No fresh auto memories to consolidate.")
			return 0
		}
		fmt.Fprintf(stdout, "Consolidated %d fresh memories into: %s\n", result.Count, result.Path)
		return 0

	case "search":
		query := strings.TrimSpace(strings.Join(nonFlagArgs(args[1:]), " "))
		if query == "" {
			fmt.Fprintln(stderr, "usage: knowblazer memory search <query>")
			return 2
		}
		hits, err := index.Search(repoRoot, query)
		if err != nil {
			fmt.Fprintf(stderr, "memory search failed: %v\n", err)
			return 1
		}
		if len(hits) == 0 {
			fmt.Fprintln(stdout, "No matching memories found.")
			return 0
		}
		for _, hit := range hits {
			fmt.Fprintf(stdout, "%d  %s  %s\n", hit.Score, hit.Path, strings.ReplaceAll(hit.Snippet, "\n", " "))
		}
		return 0

	case "index":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "usage: knowblazer memory index <build|search> [args]")
			return 2
		}
		if args[1] == "build" {
			idx, err := index.Build(repoRoot)
			if err != nil {
				fmt.Fprintf(stderr, "index build failed: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "Indexed %d files.\n", len(idx.Documents))
			return 0
		}
		if args[1] == "search" {
			query := strings.TrimSpace(strings.Join(nonFlagArgs(args[2:]), " "))
			if query == "" {
				fmt.Fprintln(stderr, "usage: knowblazer memory index search <query>")
				return 2
			}
			hits, err := index.Search(repoRoot, query)
			if err != nil {
				fmt.Fprintf(stderr, "index search failed: %v\n", err)
				return 1
			}
			if len(hits) == 0 {
				fmt.Fprintln(stdout, "No matching memories found.")
				return 0
			}
			for _, hit := range hits {
				fmt.Fprintf(stdout, "%d  %s  %s\n", hit.Score, hit.Path, strings.ReplaceAll(hit.Snippet, "\n", " "))
			}
			return 0
		}
		fmt.Fprintf(stderr, "unknown index action: %s\n", args[1])
		return 2

	case "dream":
		path, err := dream.Run(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "memory dream failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Dynamic memory suggestions written to: %s\n", path)
		return 0

	case "capture":
		values := nonFlagArgs(args[1:])
		if len(values) < 1 {
			fmt.Fprintln(stderr, "usage: knowblazer memory capture <file>")
			return 2
		}
		result, err := capture.Markdown(repoRoot, values[0])
		if err != nil {
			fmt.Fprintf(stderr, "capture failed: %v\n", err)
			return 1
		}
		if result.ScanLevel == scan.High {
			fmt.Fprintf(stdout, "Moved to quarantine: %s\n", result.Path)
			return 1
		}
		fmt.Fprintf(stdout, "Captured to inbox: %s\n", result.Path)
		return 0

	case "import":
		if len(args) < 3 || args[1] != "specstory" {
			fmt.Fprintln(stderr, "usage: knowblazer memory import specstory <path>")
			return 2
		}
		results, err := specstory.Import(repoRoot, args[2])
		if err != nil {
			fmt.Fprintf(stderr, "import failed: %v\n", err)
			return 1
		}
		for _, r := range results {
			fmt.Fprintf(stdout, "%s  %s -> %s\n", levelString(r.Level), r.Source, r.Path)
		}
		return 0

	default:
		fmt.Fprintf(stderr, "unknown memory command: %s\n", args[0])
		return 2
	}
}

func runDaily(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	action := "show"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		action = args[0]
	}

	switch action {
	case "show":
		dateStr := flagValue(args, "--date")
		date, err := daily.ParseDate(dateStr)
		if err != nil {
			fmt.Fprintf(stderr, "invalid date: %v\n", err)
			return 2
		}
		content, _, err := daily.Show(repoRoot, daily.ShowOptions{Date: date})
		if err != nil {
			fmt.Fprintf(stderr, "daily show failed: %v\n", err)
			return 1
		}
		fmt.Fprint(stdout, string(content))
		return 0

	case "add":
		text := strings.TrimSpace(strings.Join(nonFlagArgs(args[1:]), " "))
		if text == "" {
			fmt.Fprintln(stderr, "usage: knowblazer daily add <text>")
			return 2
		}
		path, err := daily.Add(repoRoot, daily.AddOptions{Text: text})
		if err != nil {
			fmt.Fprintf(stderr, "daily add failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Daily note updated: %s\n", path)
		return 0

	default:
		// If text passed directly: `knowblazer daily "did something"`
		text := strings.TrimSpace(strings.Join(nonFlagArgs(args), " "))
		if text != "" {
			path, err := daily.Add(repoRoot, daily.AddOptions{Text: text})
			if err != nil {
				fmt.Fprintf(stderr, "daily add failed: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "Daily note updated: %s\n", path)
			return 0
		}
		fmt.Fprintln(stderr, "usage: knowblazer daily [show|add] [text] [--date YYYY-MM-DD]")
		return 2
	}
}

func runProject(args []string, stdout io.Writer, stderr io.Writer) int {
	action := "list"
	projectArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		action = args[0]
		projectArgs = args[1:]
	}
	repoRoot, err := repoForArgs(projectArgs)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}

	workspace := flagValue(projectArgs, "--path")
	if workspace == "" {
		workspace, _ = os.Getwd()
	}

	switch action {
	case "list", "show":
		mappings, err := projectmap.List(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "project list failed: %v\n", err)
			return 1
		}
		for _, m := range mappings {
			fmt.Fprintf(stdout, "%s -> %s\n", m.Path, m.Project)
		}
		return 0

	case "set":
		nonFlags := nonFlagArgs(projectArgs)
		if len(nonFlags) != 1 {
			fmt.Fprintln(stderr, "usage: knowblazer project set <name> [--path <dir>]")
			return 2
		}
		project := nonFlags[0]
		if err := setup.EnsureProjectFile(repoRoot, project, workspace); err != nil {
			fmt.Fprintf(stderr, "project set failed: %v\n", err)
			return 1
		}
		if err := projectmap.Set(repoRoot, workspace, project); err != nil {
			fmt.Fprintf(stderr, "project set failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Project mapping set: %s -> %s\n", workspace, project)
		return 0

	case "clear":
		if err := projectmap.Clear(repoRoot, workspace); err != nil {
			fmt.Fprintf(stderr, "project clear failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Cleared project mapping for: %s\n", workspace)
		return 0

	default:
		fmt.Fprintln(stderr, "usage: knowblazer project <list|set|clear> [args]")
		return 2
	}
}

func runDoctor(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, ok := valueForFlag(args, "--repo")
	if !ok || repoRoot == "" {
		var err error
		repoRoot, err = discoverRepo()
		if err != nil {
			fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
			return 2
		}
	}
	report := doctor.Run(repoRoot)
	for _, check := range report.Checks {
		fmt.Fprintf(stdout, "%s  %s  %s\n", doctorStatusString(check.Status), check.Name, check.Message)
	}
	if report.HasFailures() {
		return 1
	}
	return 0
}

func runBackup(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer backup <create|restore> [--output <file>] [--input <file>] [--target <path>] [--passphrase <text>]")
		return 2
	}
	passphrase := flagValue(args[1:], "--passphrase")
	switch args[0] {
	case "create":
		repoRoot, err := repoForArgs(args[1:])
		if err != nil {
			fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
			return 2
		}
		output := flagValue(args[1:], "--output")
		if output == "" {
			output = flagValue(args[1:], "-o")
		}
		if output == "" {
			fmt.Fprintln(stderr, "backup create requires --output <file>")
			return 2
		}
		if err := backup.Create(repoRoot, output, passphrase); err != nil {
			fmt.Fprintf(stderr, "backup create failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Backup written to: %s\n", output)
		return 0

	case "restore":
		input := flagValue(args[1:], "--input")
		if input == "" {
			input = flagValue(args[1:], "-i")
		}
		target := flagValue(args[1:], "--target")
		if target == "" {
			target = flagValue(args[1:], "-t")
		}
		if input == "" || target == "" {
			fmt.Fprintln(stderr, "backup restore requires --input <file> --target <path>")
			return 2
		}
		if err := backup.Restore(input, target, passphrase); err != nil {
			fmt.Fprintf(stderr, "backup restore failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Backup restored to: %s\n", target)
		return 0

	default:
		fmt.Fprintln(stderr, "usage: knowblazer backup <create|restore> [--output <file>] [--input <file>] [--target <path>]")
		return 2
	}
}

// ---------------------------------------------------------------------------
// System & Infrastructure Commands
// ---------------------------------------------------------------------------

func runMCP(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer setup`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	if err := mcp.Serve(repoRoot, os.Stdin, stdout); err != nil {
		fmt.Fprintf(stderr, "mcp serve failed: %v\n", err)
		return 1
	}
	return 0
}

func runVersion(args []string, stdout io.Writer, stderr io.Writer) int {
	fmt.Fprintln(stdout, version.FullString())
	return 0
}

func runUpdate(args []string, stdout io.Writer, stderr io.Writer) int {
	force := hasFlag(args, "--force") || hasFlag(args, "-f")
	cfg := update.Config{
		Force: force,
	}
	if err := update.Run(context.Background(), cfg, stdout); err != nil {
		fmt.Fprintf(stderr, "update error: %v\n", err)
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// Helper Functions
// ---------------------------------------------------------------------------

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: knowblazer <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Core Commands:")
	fmt.Fprintln(w, "  setup       Initialize memory repo and connect AI tools (--global supported)")
	fmt.Fprintln(w, "  remember    Save lessons, solutions, markdown files, or daily notes")
	fmt.Fprintln(w, "  recall      Retrieve tailored context pack for an engineering task")
	fmt.Fprintln(w, "  status      Check current workspace mapping and memory connection")
	fmt.Fprintln(w, "  sync        Sync memory repo with remote Git (status|commit|push|pull)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Management:")
	fmt.Fprintln(w, "  memory      Manage memories (list|review|promote|reject|consolidate|search|capture)")
	fmt.Fprintln(w, "  daily       View and manage daily work logs (show|add)")
	fmt.Fprintln(w, "  project     Manage workspace-to-project mappings (list|set|clear)")
	fmt.Fprintln(w, "  doctor      Diagnose memory repo integrity and secrets")
	fmt.Fprintln(w, "  backup      Encrypted archive creation and restoration (create|restore)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "System:")
	fmt.Fprintln(w, "  mcp         Run Model Context Protocol server for AI tools")
	fmt.Fprintln(w, "  update      Self-update knowblazer to latest release")
	fmt.Fprintln(w, "  version     Print version and build metadata")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -h, --help     Show help")
	fmt.Fprintln(w, "  -v, --version  Show version")
}

func printClaudeStartResult(stdout io.Writer, result setup.ClaudeResult) {
	fmt.Fprintf(stdout, "Claude instructions: %s\n", result.ClaudeMD)
	if result.MCPConfigured {
		fmt.Fprintln(stdout, "Claude Code MCP configured: knowblazer")
	} else if result.MCPSkipped {
		fmt.Fprintln(stdout, "Claude Code MCP skipped")
	} else {
		fmt.Fprintln(stdout, "Claude Code MCP already configured: knowblazer")
	}
}

func printCodexStartResult(stdout io.Writer, result setup.CodexResult) {
	fmt.Fprintf(stdout, "Codex instructions: %s\n", result.AgentsMD)
	if result.MCPConfigured {
		fmt.Fprintln(stdout, "Codex MCP configured: knowblazer")
	} else if result.MCPSkipped {
		fmt.Fprintln(stdout, "Codex MCP skipped")
	} else {
		fmt.Fprintln(stdout, "Codex MCP already configured: knowblazer")
	}
}

func checkFileContains(path string, substr string) bool {
	if content, err := os.ReadFile(path); err == nil {
		return strings.Contains(string(content), substr)
	}
	return false
}

func levelString(level scan.Level) string {
	switch level {
	case scan.High:
		return "HIGH"
	case scan.Warning:
		return "WARNING"
	default:
		return "CLEAN"
	}
}

func doctorStatusString(status doctor.Status) string {
	switch status {
	case doctor.Warn:
		return "WARN"
	case doctor.Fail:
		return "FAIL"
	default:
		return "OK"
	}
}

func valueForFlag(args []string, name string) (string, bool) {
	for i := 0; i < len(args); i++ {
		if args[i] == name {
			if i+1 >= len(args) {
				return "", false
			}
			return args[i+1], true
		}
		if strings.HasPrefix(args[i], name+"=") {
			return strings.TrimPrefix(args[i], name+"="), true
		}
	}
	return "", false
}

func repoForArgs(args []string) (string, error) {
	repoRoot, ok := valueForFlag(args, "--repo")
	if !ok || repoRoot == "" {
		return discoverRepo()
	}
	if err := repo.MustBeRepo(repoRoot); err != nil {
		return "", err
	}
	return repoRoot, nil
}

func flagValue(args []string, name string) string {
	value, _ := valueForFlag(args, name)
	return value
}

func nonFlagArgs(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				continue
			}
			if flagTakesValue(arg) && i+1 < len(args) {
				i++
			}
			continue
		}
		out = append(out, arg)
	}
	return out
}

func hasFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
		if strings.HasPrefix(arg, name+"=") {
			return true
		}
	}
	return false
}

func inferProjectName(workspace string) string {
	if root, ok := gitRoot(workspace); ok {
		return filepath.Base(root)
	}
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return filepath.Base(filepath.Clean(workspace))
	}
	return filepath.Base(abs)
}

func gitRoot(workspace string) (string, bool) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = filepath.Clean(workspace)
	output, err := cmd.Output()
	if err != nil {
		return "", false
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", false
	}
	return root, true
}

func flagTakesValue(name string) bool {
	switch name {
	case "--daily", "-d", "--skip-mcp", "--force", "-f", "--global", "-g", "--help", "-h", "--version", "-v":
		return false
	default:
		return true
	}
}

func discoverRepo() (string, error) {
	return repo.Discover(repo.Options{
		EnvRepo:     os.Getenv("KNOWBLAZER_REPO"),
		DefaultRepo: filepath.Join(os.Getenv("HOME"), "knowblazer-notes"),
	})
}

func projectMemoryState(repoRoot string, project string) string {
	if project == "" {
		return "unmapped"
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, "projects", project+".md"))
	if os.IsNotExist(err) {
		return "missing"
	}
	if err != nil {
		return "unknown"
	}
	text := string(content)
	if strings.Contains(text, "Add what this project is for.") || strings.Contains(text, "Add durable project context here.") {
		return "scaffold"
	}
	return "active"
}
