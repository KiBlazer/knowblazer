package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/knowblazer/knowblazer/internal/adapter"
	"github.com/knowblazer/knowblazer/internal/backup"
	"github.com/knowblazer/knowblazer/internal/capture"
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
	case "review":
		return runReview(args[1:], stdout, stderr)
	case "scan":
		return runScan(args[1:], stdout, stderr)
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
	fmt.Fprintf(stdout, "Dream suggestions written to: %s\n", path)
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
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: knowblazer sync <status|commit|push|pull> [--repo <path>] [--message <text>]")
		return 2
	}
	repoRoot, err := repoForArgs(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, "Knowblazer repo not found. Run `knowblazer init <path>`, pass --repo, or set KNOWBLAZER_REPO.")
		return 2
	}
	var result ksync.Result
	switch args[0] {
	case "status":
		result, err = ksync.Status(repoRoot)
	case "commit":
		message := flagValue(args[1:], "--message")
		if message == "" {
			message = flagValue(args[1:], "-m")
		}
		result, err = ksync.Commit(repoRoot, message)
	case "push":
		result, err = ksync.Push(repoRoot)
	case "pull":
		result, err = ksync.Pull(repoRoot)
	default:
		fmt.Fprintln(stderr, "usage: knowblazer sync <status|commit|push|pull> [--repo <path>] [--message <text>]")
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
		fmt.Fprintln(stderr, "recall requires --task <text>")
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
	fmt.Fprintln(stdout, "  1. Edit AI-SETUP.md")
	fmt.Fprintf(stdout, "  2. Capture a lesson with: knowblazer capture <file> --repo %s\n", root)
	fmt.Fprintf(stdout, "  3. Generate recall with: knowblazer recall --task \"<task>\" --repo %s\n", root)
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: knowblazer <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  init [path]    initialize a Knowblazer memory repo")
	fmt.Fprintln(w, "  scan <path> [--repo <path>]    scan a file or directory for sensitive content")
	fmt.Fprintln(w, "  adapter <claude|codex|gemini|cursor> [--repo <path>] [--output <path>]")
	fmt.Fprintln(w, "  backup <create|restore> [--repo <path>] [--output <file>] [--input <file>] [--target <path>] [--passphrase <text>]")
	fmt.Fprintln(w, "  dream [--repo <path>]")
	fmt.Fprintln(w, "  import specstory <path> [--repo <path>]")
	fmt.Fprintln(w, "  index <build|search> [query] [--repo <path>]")
	fmt.Fprintln(w, "  mcp serve [--repo <path>]")
	fmt.Fprintln(w, "  review <list|promote|reject> [file] [--to <target>] [--repo <path>]")
	fmt.Fprintln(w, "  daily <add|show> [text] [--repo <path>] [--date YYYY-MM-DD]")
	fmt.Fprintln(w, "  doctor [--repo <path>]")
	fmt.Fprintln(w, "  project <set|show|clear> [name] [--path <dir>] [--repo <path>]")
	fmt.Fprintln(w, "  sync <status|commit|push|pull> [--repo <path>] [--message <text>]")
	fmt.Fprintln(w, "  capture <file> --repo <path>")
	fmt.Fprintln(w, "  promote <file> --to <target> --repo <path>")
	fmt.Fprintln(w, "  recall --task <text> --repo <path> [--project <name>] [--output <file>]")
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
		if strings.HasPrefix(args[i], "-") {
			i++
			continue
		}
		out = append(out, args[i])
	}
	return out
}

func discoverRepo() (string, error) {
	return repo.Discover(repo.Options{
		EnvRepo:     os.Getenv("KNOWBLAZER_REPO"),
		DefaultRepo: filepath.Join(os.Getenv("HOME"), "knowblazer-notes"),
	})
}
