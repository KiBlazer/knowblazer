package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/knowblazer/knowblazer/internal/adapter"
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
	"github.com/knowblazer/knowblazer/internal/promote"
	"github.com/knowblazer/knowblazer/internal/recall"
	"github.com/knowblazer/knowblazer/internal/repo"
	"github.com/knowblazer/knowblazer/internal/review"
	"github.com/knowblazer/knowblazer/internal/scan"
	"github.com/knowblazer/knowblazer/internal/setup"
	ksync "github.com/knowblazer/knowblazer/internal/sync"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "start":
		return runStart(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "adapter":
		return runAdapter(args[1:], stdout, stderr)
	case "backup":
		return runBackup(args[1:], stdout, stderr)
	case "dream":
		return runDream(args[1:], stdout, stderr)
	case "import":
		return runImport(args[1:], stdout, stderr)
	case "index":
		return runIndex(args[1:], stdout, stderr)
	case "mcp":
		return runMCP(args[1:], stdout, stderr)
	case "remember":
		return runRemember(args[1:], stdout, stderr)
	case "review":
		return runReview(args[1:], stdout, stderr)
	case "scan":
		return runScan(args[1:], stdout, stderr)
	case "setup":
		return runSetup(args[1:], stdout, stderr)
	case "capture":
		return runCapture(args[1:], stdout, stderr)
	case "daily":
		return runDaily(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "project":
		return runProject(args[1:], stdout, stderr)
	case "promote":
		return runPromote(args[1:], stdout, stderr)
	case "consolidate":
		return runConsolidate(args[1:], stdout, stderr)
	case "sync":
		return runSync(args[1:], stdout, stderr)
	case "recall":
		return runRecall(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runStart(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForStart(args)
	if err != nil {
		fmt.Fprintf(stderr, "start failed: %v\n", err)
		return 1
	}
	workspace := flagValue(args, "--path")
	if workspace == "" {
		workspace, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "start failed: %v\n", err)
			return 1
		}
	}
	project := flagValue(args, "--project")
	if project == "" {
		project = inferProjectName(workspace)
	}
	result, err := setup.Claude(setup.ClaudeOptions{
		RepoRoot:      repoRoot,
		Workspace:     workspace,
		Project:       project,
		Scope:         flagValue(args, "--scope"),
		SkipMCP:       hasFlag(args, "--skip-mcp"),
		EnsureProject: true,
	})
	if err != nil {
		fmt.Fprintf(stderr, "start failed: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Knowblazer is ready for Claude Code in this project.")
	fmt.Fprintf(stdout, "Memory repo: %s\n", repoRoot)
	fmt.Fprintf(stdout, "Project: %s\n", project)
	fmt.Fprintf(stdout, "Claude instructions: %s\n", result.ClaudeMD)
	if result.MCPConfigured {
		fmt.Fprintln(stdout, "Claude Code MCP configured: knowblazer")
	} else if result.MCPSkipped {
		fmt.Fprintln(stdout, "Claude Code MCP skipped")
	} else {
		fmt.Fprintln(stdout, "Claude Code MCP already configured: knowblazer")
	}
	fmt.Fprintln(stdout, "Next: run `claude` from this project.")
	return 0
}

func runStatus(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer start`, `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	workspace := flagValue(args, "--path")
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	project := ""
	if workspace != "" {
		if mapped, ok, err := projectmap.Resolve(repoRoot, workspace); err == nil && ok {
			project = mapped
		}
	}
	fmt.Fprintf(stdout, "Memory repo: %s\n", repoRoot)
	if project != "" {
		fmt.Fprintf(stdout, "Project: %s\n", project)
	} else {
		fmt.Fprintln(stdout, "Project: not mapped")
	}
	claudeMD := ""
	if workspace != "" {
		claudeMD = filepath.Join(workspace, "CLAUDE.md")
	}
	if claudeMD != "" {
		if content, err := os.ReadFile(claudeMD); err == nil && strings.Contains(string(content), "KNOWBLAZER-CLAUDE-SETUP:START") {
			fmt.Fprintf(stdout, "Claude instructions: %s\n", claudeMD)
		} else {
			fmt.Fprintln(stdout, "Claude instructions: not configured")
		}
	}
	freshCount, err := consolidate.CountFresh(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "status failed: %v\n", err)
		return 1
	}
	synthesizedCount, err := consolidate.CountSynthesized(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "status failed: %v\n", err)
		return 1
	}
	candidates, err := review.List(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "status failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Fresh auto memories: %d\n", freshCount)
	fmt.Fprintf(stdout, "Synthesized memories: %d\n", synthesizedCount)
	fmt.Fprintf(stdout, "Inbox candidates: %d\n", len(candidates))
	return 0
}

func runSetup(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "claude" {
		fmt.Fprintln(stderr, "usage: knowblazer setup claude [--repo <path>] [--project <name>] [--path <dir>] [--scope <local|user|project>] [--skip-mcp]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	workspace := flagValue(args[1:], "--path")
	project := flagValue(args[1:], "--project")
	scope := flagValue(args[1:], "--scope")
	result, err := setup.Claude(setup.ClaudeOptions{
		RepoRoot:  repoRoot,
		Workspace: workspace,
		Project:   project,
		Scope:     scope,
		SkipMCP:   hasFlag(args[1:], "--skip-mcp"),
	})
	if err != nil {
		fmt.Fprintf(stderr, "setup claude failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Claude instructions updated: %s\n", result.ClaudeMD)
	if result.ProjectMapped {
		fmt.Fprintf(stdout, "Project mapping configured: %s\n", project)
	}
	if result.MCPConfigured {
		fmt.Fprintln(stdout, "Claude Code MCP configured: knowblazer")
	} else if result.MCPSkipped {
		fmt.Fprintln(stdout, "Claude Code MCP skipped")
	}
	fmt.Fprintln(stdout, "Claude Code integration ready.")
	return 0
}

func runBackup(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer backup <create|restore> [--repo <path>] [--output <file>] [--input <file>] [--target <path>] [--passphrase <text>]")
		return 2
	}
	passphrase := flagValue(args[1:], "--passphrase")
	switch args[0] {
	case "create":
		repoRoot, err := repoForArgs(args[1:])
		if err != nil {
			fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
			return 2
		}
		output := flagValue(args[1:], "--output")
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
		target := flagValue(args[1:], "--target")
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
		fmt.Fprintln(stderr, "usage: knowblazer backup <create|restore> [--repo <path>] [--output <file>] [--input <file>] [--target <path>] [--passphrase <text>]")
		return 2
	}
}

func runDream(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	path, err := dream.Run(repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "dream failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Dynamic memory suggestions written to: %s\n", path)
	return 0
}

func runMCP(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 || args[0] != "serve" {
		fmt.Fprintln(stderr, "usage: knowblazer mcp serve [--repo <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	if err := mcp.Serve(repoRoot, os.Stdin, stdout); err != nil {
		fmt.Fprintf(stderr, "mcp serve failed: %v\n", err)
		return 1
	}
	return 0
}

func runRemember(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(nonFlagArgs(args)) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer remember <file|text> [--repo <path>] [--daily]")
		return 2
	}
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	text := strings.TrimSpace(strings.Join(nonFlagArgs(args), " "))
	if hasFlag(args, "--daily") {
		path, err := daily.Add(repoRoot, daily.AddOptions{Text: text})
		if err != nil {
			fmt.Fprintf(stderr, "remember daily failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Remembered in daily notes: %s\n", path)
		return 0
	}
	if info, err := os.Stat(text); err == nil && !info.IsDir() {
		result, err := capture.MarkdownAuto(repoRoot, text)
		if err != nil {
			fmt.Fprintf(stderr, "remember failed: %v\n", err)
			return 1
		}
		if result.ScanLevel == scan.High {
			fmt.Fprintf(stderr, "sensitive content detected; saved to quarantine: %s\n", result.Path)
			return 1
		}
		fmt.Fprintf(stdout, "Remembered to fresh memory: %s\n", result.Path)
		return 0
	}
	path, err := rememberText(repoRoot, text)
	if err != nil {
		fmt.Fprintf(stderr, "remember failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Remembered to fresh memory: %s\n", path)
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
		file.Close()
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

func runConsolidate(args []string, stdout io.Writer, stderr io.Writer) int {
	repoRoot, err := repoForArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
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
}

func runReview(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer review <list|promote|reject> [file] [--to <target>] [--repo <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	switch args[0] {
	case "list":
		candidates, err := review.List(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "review list failed: %v\n", err)
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
			fmt.Fprintln(stderr, "review promote requires <file> --to <target>")
			return 2
		}
		result, err := review.Promote(repoRoot, values[0], target)
		if err != nil {
			fmt.Fprintf(stderr, "review promote failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Promoted to: %s\n", result.Path)
		return 0
	case "reject":
		values := nonFlagArgs(args[1:])
		if len(values) != 1 {
			fmt.Fprintln(stderr, "review reject requires <file>")
			return 2
		}
		path, err := review.Reject(repoRoot, values[0])
		if err != nil {
			fmt.Fprintf(stderr, "review reject failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Rejected to: %s\n", path)
		return 0
	default:
		fmt.Fprintln(stderr, "usage: knowblazer review <list|promote|reject> [file] [--to <target>] [--repo <path>]")
		return 2
	}
}

func runAdapter(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer adapter <claude|codex|gemini|cursor> [--repo <path>] [--output <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	content, err := adapter.Generate(args[0], repoRoot)
	if err != nil {
		fmt.Fprintf(stderr, "adapter failed: %v\n", err)
		return 1
	}
	output := flagValue(args[1:], "--output")
	if output != "" {
		if err := os.WriteFile(output, content, 0o644); err != nil {
			fmt.Fprintf(stderr, "adapter write failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Adapter written to: %s\n", output)
		return 0
	}
	fmt.Fprint(stdout, string(content))
	return 0
}

func runImport(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 2 || args[0] != "specstory" {
		fmt.Fprintln(stderr, "usage: knowblazer import specstory <path> [--repo <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args[2:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	results, err := specstory.Import(repoRoot, args[1])
	if err != nil {
		fmt.Fprintf(stderr, "import failed: %v\n", err)
		return 1
	}
	for _, result := range results {
		fmt.Fprintf(stdout, "%s  %s -> %s\n", levelString(result.Level), result.Source, result.Path)
	}
	return 0
}

func runIndex(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer index <build|search> [query] [--repo <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	switch args[0] {
	case "build":
		idx, err := index.Build(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "index build failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Indexed %d documents\n", len(idx.Documents))
		return 0
	case "search":
		query := strings.TrimSpace(strings.Join(nonFlagArgs(args[1:]), " "))
		hits, err := index.Search(repoRoot, query)
		if err != nil {
			fmt.Fprintf(stderr, "index search failed: %v\n", err)
			return 1
		}
		for _, hit := range hits {
			fmt.Fprintf(stdout, "%d  %s  %s\n", hit.Score, hit.Path, strings.ReplaceAll(hit.Snippet, "\n", " "))
		}
		return 0
	default:
		fmt.Fprintln(stderr, "usage: knowblazer index <build|search> [query] [--repo <path>]")
		return 2
	}
}

func runDaily(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer daily <add|show> [text] [--repo <path>] [--date YYYY-MM-DD]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	switch args[0] {
	case "add":
		text := strings.TrimSpace(strings.Join(nonFlagArgs(args[1:]), " "))
		path, err := daily.Add(repoRoot, daily.AddOptions{Text: text})
		if err != nil {
			fmt.Fprintf(stderr, "daily add failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Daily note updated: %s\n", path)
		return 0
	case "show":
		date, err := daily.ParseDate(flagValue(args[1:], "--date"))
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
	default:
		fmt.Fprintln(stderr, "usage: knowblazer daily <add|show> [text] [--repo <path>] [--date YYYY-MM-DD]")
		return 2
	}
}

func runProject(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer project <set|show|clear> [name] [--path <dir>] [--repo <path>]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	workspace := flagValue(args[1:], "--path")
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	switch args[0] {
	case "set":
		values := nonFlagArgs(args[1:])
		if len(values) != 1 {
			fmt.Fprintln(stderr, "project set requires a project name")
			return 2
		}
		if err := projectmap.Set(repoRoot, workspace, values[0]); err != nil {
			fmt.Fprintf(stderr, "project set failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Mapped %s to project %s\n", workspace, values[0])
		return 0
	case "show":
		mappings, err := projectmap.List(repoRoot)
		if err != nil {
			fmt.Fprintf(stderr, "project show failed: %v\n", err)
			return 1
		}
		for _, mapping := range mappings {
			fmt.Fprintf(stdout, "%s  %s\n", mapping.Project, mapping.Path)
		}
		return 0
	case "clear":
		if err := projectmap.Clear(repoRoot, workspace); err != nil {
			fmt.Fprintf(stderr, "project clear failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Cleared project mapping for %s\n", workspace)
		return 0
	default:
		fmt.Fprintln(stderr, "usage: knowblazer project <set|show|clear> [name] [--path <dir>] [--repo <path>]")
		return 2
	}
}

func runSync(args []string, stdout io.Writer, stderr io.Writer) int {
	command := "status"
	repoArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command = args[0]
		repoArgs = args[1:]
	}
	repoRoot, err := repoForArgs(repoArgs)
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	var result ksync.Result
	switch command {
	case "status":
		result, err = ksync.Status(repoRoot)
	case "commit":
		message := flagValue(repoArgs, "--message")
		if message == "" {
			message = flagValue(repoArgs, "-m")
		}
		result, err = ksync.Commit(repoRoot, message)
	case "push":
		result, err = ksync.Push(repoRoot)
	case "pull":
		result, err = ksync.Pull(repoRoot)
	default:
		fmt.Fprintln(stderr, "usage: knowblazer sync [status|commit|push|pull] [--repo <path>] [--message <text>]")
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "sync failed: %v\n", err)
		if result.Output != "" {
			fmt.Fprint(stderr, result.Output)
		}
		return 1
	}
	fmt.Fprint(stdout, result.Output)
	return 0
}

func runDoctor(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 0 && len(args) != 2 {
		fmt.Fprintln(stderr, "usage: knowblazer doctor [--repo <path>]")
		return 2
	}
	repoRoot, ok := valueForFlag(args, "--repo")
	if len(args) == 2 && (!ok || repoRoot == "") {
		fmt.Fprintln(stderr, "usage: knowblazer doctor [--repo <path>]")
		return 2
	}
	if !ok || repoRoot == "" {
		var err error
		repoRoot, err = discoverRepo()
		if err != nil {
			fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
			return 2
		}
	}

	result := doctor.Run(repoRoot)
	for _, check := range result.Checks {
		fmt.Fprintf(stdout, "%s  %s  %s\n", doctorStatusString(check.Status), check.Name, check.Message)
	}
	if result.HasFailures() {
		return 1
	}
	return 0
}

func runRecall(args []string, stdout io.Writer, stderr io.Writer) int {
	task, ok := valueForFlag(args, "--task")
	if !ok || task == "" {
		task = strings.TrimSpace(strings.Join(nonFlagArgs(args), " "))
	}
	if task == "" {
		fmt.Fprintln(stderr, "usage: knowblazer recall <task> [--repo <path>] [--project <name>] [--output <file>]")
		return 2
	}
	repoRoot, ok := valueForFlag(args, "--repo")
	if !ok || repoRoot == "" {
		var err error
		repoRoot, err = discoverRepo()
		if err != nil {
			fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
			return 2
		}
	}
	if err := repo.MustBeRepo(repoRoot); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}
	project, _ := valueForFlag(args, "--project")
	if project == "" {
		if cwd, err := os.Getwd(); err == nil {
			if mapped, ok, err := projectmap.Resolve(repoRoot, cwd); err == nil && ok {
				project = mapped
			}
		}
	}
	output, _ := valueForFlag(args, "--output")

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

func runPromote(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer promote <file> --to <target> --repo <path>")
		return 2
	}

	source := args[0]
	target, ok := valueForFlag(args[1:], "--to")
	if !ok || target == "" {
		fmt.Fprintln(stderr, "promote requires --to <target>")
		return 2
	}
	repoRoot, ok := valueForFlag(args[1:], "--repo")
	if !ok || repoRoot == "" {
		var err error
		repoRoot, err = discoverRepo()
		if err != nil {
			fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
			return 2
		}
	}
	if err := repo.MustBeRepo(repoRoot); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}

	result, err := promote.File(repoRoot, source, target)
	if err != nil {
		fmt.Fprintf(stderr, "promote failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Promoted to: %s\n", result.Path)
	return 0
}

func runCapture(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer capture <file> --repo <path>")
		return 2
	}

	source := args[0]
	repoRoot, ok := valueForFlag(args[1:], "--repo")
	if !ok || repoRoot == "" {
		var err error
		repoRoot, err = discoverRepo()
		if err != nil {
			fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
			return 2
		}
	}
	if err := repo.MustBeRepo(repoRoot); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}

	result, err := capture.Markdown(repoRoot, source)
	if err != nil {
		fmt.Fprintf(stderr, "capture failed: %v\n", err)
		return 1
	}

	if result.ScanLevel == scan.High {
		fmt.Fprintln(stdout, "Sensitive content detected.")
		fmt.Fprintf(stdout, "Moved to quarantine: %s\n", result.Path)
		fmt.Fprintln(stdout, "Review and sanitize before promoting or committing.")
		return 1
	}

	fmt.Fprintf(stdout, "Captured to inbox: %s\n", result.Path)
	fmt.Fprintf(stdout, "Scan result: %s\n", result.ScanLevel.String())
	return 0
}

func runScan(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 && len(args) != 3 {
		fmt.Fprintln(stderr, "usage: knowblazer scan <path> [--repo <path>]")
		return 2
	}
	if len(args) == 3 {
		repoRoot, ok := valueForFlag(args[1:], "--repo")
		if !ok || repoRoot == "" {
			fmt.Fprintln(stderr, "usage: knowblazer scan <path> [--repo <path>]")
			return 2
		}
	}

	result, err := scan.Path(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "scan failed: %v\n", err)
		return 2
	}

	highCount := 0
	for _, finding := range result.Findings {
		if finding.Level == scan.High {
			highCount++
		}
		fmt.Fprintf(
			stdout,
			"%s  %s:%d  %s  %s\n",
			levelString(finding.Level),
			filepath.Clean(finding.File),
			finding.Line,
			finding.Rule,
			finding.Snippet,
		)
	}
	fmt.Fprintf(stdout, "Summary: %s, %d findings, %d high-risk\n", result.Level.String(), len(result.Findings), highCount)

	if result.Level == scan.Clean {
		return 0
	}
	return 1
}

func runInit(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 1 {
		fmt.Fprintln(stderr, "usage: knowblazer init [path]")
		return 2
	}

	root := filepath.Join(os.Getenv("HOME"), "knowblazer-notes")
	if len(args) == 1 {
		root = args[0]
	}
	if err := repo.Init(root); err != nil {
		fmt.Fprintf(stderr, "init failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Initialized Knowblazer memory repo: %s\n", root)
	fmt.Fprintln(stdout, "Next steps:")
	fmt.Fprintf(stdout, "  1. Capture fresh memory with: knowblazer remember \"<lesson>\" --repo %s\n", root)
	fmt.Fprintf(stdout, "  2. Consolidate fresh memory with: knowblazer consolidate --repo %s\n", root)
	fmt.Fprintf(stdout, "  3. Generate dynamic context with: knowblazer recall --task \"<task>\" --repo %s\n", root)
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: knowblazer <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Core workflow:")
	fmt.Fprintln(w, "  start [--repo <path>] [--path <dir>] [--skip-mcp]")
	fmt.Fprintln(w, "  remember <file|text> [--repo <path>] [--daily]")
	fmt.Fprintln(w, "  recall <task> [--repo <path>] [--project <name>] [--output <file>]")
	fmt.Fprintln(w, "  consolidate [--repo <path>]")
	fmt.Fprintln(w, "  status [--repo <path>] [--path <dir>]")
	fmt.Fprintln(w, "  sync [status|commit|push|pull] [--repo <path>] [--message <text>]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Review and maintenance:")
	fmt.Fprintln(w, "  review <list|promote|reject> [file] [--to <target>] [--repo <path>]")
	fmt.Fprintln(w, "  doctor [--repo <path>]")
	fmt.Fprintln(w, "  backup <create|restore> [--repo <path>] [--output <file>] [--input <file>] [--target <path>] [--passphrase <text>]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Advanced commands:")
	fmt.Fprintln(w, "  setup claude [--repo <path>] [--project <name>] [--path <dir>]")
	fmt.Fprintln(w, "  scan <path> [--repo <path>]")
	fmt.Fprintln(w, "  capture <file> --repo <path>")
	fmt.Fprintln(w, "  promote <file> --to <target> --repo <path>")
	fmt.Fprintln(w, "  daily <add|show> [text] [--repo <path>] [--date YYYY-MM-DD]")
	fmt.Fprintln(w, "  project <set|show|clear> [name] [--path <dir>] [--repo <path>]")
	fmt.Fprintln(w, "  import specstory <path> [--repo <path>]")
	fmt.Fprintln(w, "  adapter <claude|codex|gemini|cursor> [--repo <path>] [--output <path>]")
	fmt.Fprintln(w, "  index <build|search> [query] [--repo <path>]")
	fmt.Fprintln(w, "  mcp serve [--repo <path>]")
	fmt.Fprintln(w, "  dream [--repo <path>]")
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
	}
	return false
}

func repoForStart(args []string) (string, error) {
	repoRoot := flagValue(args, "--repo")
	if repoRoot == "" {
		repoRoot = os.Getenv("KNOWBLAZER_REPO")
	}
	if repoRoot == "" {
		repoRoot = filepath.Join(os.Getenv("HOME"), "knowblazer-notes")
	}
	if err := repo.Init(repoRoot); err != nil {
		return "", err
	}
	if err := repo.MustBeRepo(repoRoot); err != nil {
		return "", err
	}
	return repoRoot, nil
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
	case "--daily", "--skip-mcp":
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
