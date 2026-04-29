package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/knowblazer/knowblazer/internal/capture"
	"github.com/knowblazer/knowblazer/internal/promote"
	"github.com/knowblazer/knowblazer/internal/recall"
	"github.com/knowblazer/knowblazer/internal/repo"
	"github.com/knowblazer/knowblazer/internal/scan"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "scan":
		return runScan(args[1:], stdout, stderr)
	case "capture":
		return runCapture(args[1:], stdout, stderr)
	case "promote":
		return runPromote(args[1:], stdout, stderr)
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

func runRecall(args []string, stdout io.Writer, stderr io.Writer) int {
	task, ok := valueForFlag(args, "--task")
	if !ok || task == "" {
		fmt.Fprintln(stderr, "recall requires --task <text>")
		return 2
	}
	repoRoot, ok := valueForFlag(args, "--repo")
	if !ok || repoRoot == "" {
		fmt.Fprintln(stderr, "recall requires --repo <path> in MVP")
		return 2
	}
	project, _ := valueForFlag(args, "--project")
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
		fmt.Fprintln(stderr, "promote requires --repo <path> in MVP")
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
		fmt.Fprintln(stderr, "capture requires --repo <path> in MVP")
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
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: knowblazer scan <path>")
		return 2
	}

	result, err := scan.Path(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "scan failed: %v\n", err)
		return 2
	}

	for _, finding := range result.Findings {
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

	if result.Level == scan.Clean {
		fmt.Fprintln(stdout, "clean")
		return 0
	}
	return 1
}

func runInit(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: knowblazer init <path>")
		return 2
	}

	root := args[0]
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
	fmt.Fprintln(w, "  init <path>    initialize a Knowblazer memory repo")
	fmt.Fprintln(w, "  scan <path>    scan a file or directory for sensitive content")
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
