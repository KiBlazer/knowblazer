package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRunHelpIncludesCoreWorkflow(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("help code = %d, stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"Common commands:", "start [--repo <path>]", "remember <file|text>", "recall <task>", "status [--repo <path>]", "sync [status|commit|push|pull]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q: %s", want, out)
		}
	}
	for _, hidden := range []string{"consolidate [--repo <path>]", "Advanced commands:", "setup claude", "mcp serve", "adapter <claude|codex|gemini|cursor>"} {
		if strings.Contains(out, hidden) {
			t.Fatalf("stdout should not show %q by default: %s", hidden, out)
		}
	}
}

func TestRunStartAutoConfiguresInstalledTools(t *testing.T) {
	home := t.TempDir()
	workspace := filepath.Join(t.TempDir(), "kiblazer")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	t.Setenv("HOME", home)
	oldLookPath := setupLookPath
	setupLookPath = func(tool string) bool {
		return tool == "claude" || tool == "codex"
	}
	t.Cleanup(func() {
		setupLookPath = oldLookPath
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"start", "--path", workspace, "--skip-mcp"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("start code = %d, stderr = %s", code, stderr.String())
	}
	root := filepath.Join(home, "knowblazer-notes")
	if _, err := os.Stat(filepath.Join(root, ".knowblazer", "config.json")); err != nil {
		t.Fatalf("expected default repo config: %v", err)
	}
	projectPath := filepath.Join(root, "projects", "kiblazer.md")
	if _, err := os.Stat(projectPath); err != nil {
		t.Fatalf("expected inferred project file: %v", err)
	}
	projectContent, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatalf("read project file: %v", err)
	}
	if !strings.Contains(string(projectContent), "## Workspace") || !strings.Contains(string(projectContent), workspace) {
		t.Fatalf("project file missing scaffold:\n%s", projectContent)
	}
	claudeMD := filepath.Join(workspace, "CLAUDE.md")
	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	for _, want := range []string{"knowblazer_context", "knowblazer_remember", "end-of-task memory candidate review", workspace} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("CLAUDE.md missing %q:\n%s", want, content)
		}
	}
	agentsMD := filepath.Join(workspace, "AGENTS.md")
	agentsContent, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	for _, want := range []string{"KNOWBLAZER-CODEX-SETUP:START", "knowblazer_context", "knowblazer_remember", "end-of-task memory candidate review", workspace} {
		if !strings.Contains(string(agentsContent), want) {
			t.Fatalf("AGENTS.md missing %q:\n%s", want, agentsContent)
		}
	}
	if !strings.Contains(stdout.String(), "Knowblazer is ready for Claude Code and Codex in this project.") || !strings.Contains(stdout.String(), "Next: run `claude` or `codex` from this project.") {
		t.Fatalf("start output missing auto setup details: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"status", "--repo", root, "--path", workspace}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("status code = %d, stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Project: kiblazer") || !strings.Contains(out, "Project memory: scaffold") || !strings.Contains(out, "Claude instructions:") || !strings.Contains(out, "Codex instructions:") {
		t.Fatalf("status output missing setup details: %s", out)
	}
}

func TestRunStartAutoSkipMCPWritesInstructionsWithoutInstalledTools(t *testing.T) {
	home := t.TempDir()
	workspace := filepath.Join(t.TempDir(), "kiblazer")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	t.Setenv("HOME", home)
	oldLookPath := setupLookPath
	setupLookPath = func(tool string) bool {
		return false
	}
	t.Cleanup(func() {
		setupLookPath = oldLookPath
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"start", "--path", workspace, "--skip-mcp"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("start code = %d, stderr = %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(workspace, "CLAUDE.md")); err != nil {
		t.Fatalf("expected CLAUDE.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "AGENTS.md")); err != nil {
		t.Fatalf("expected AGENTS.md: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Claude Code MCP skipped") || !strings.Contains(out, "Codex MCP skipped") {
		t.Fatalf("start output missing skipped MCP details: %s", out)
	}
}

func TestRunStartAutoContinuesWhenOneToolMCPFails(t *testing.T) {
	home := t.TempDir()
	workspace := filepath.Join(t.TempDir(), "kiblazer")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	t.Setenv("HOME", home)
	bin := t.TempDir()
	writeFakeCLI(t, bin, "claude", "add failed", 1)
	writeFakeCLI(t, bin, "codex", "added", 0)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Cleanup(func() {
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"start", "--path", workspace}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("start code = %d, stderr = %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(workspace, "CLAUDE.md")); err != nil {
		t.Fatalf("expected CLAUDE.md despite Claude MCP failure: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "AGENTS.md")); err != nil {
		t.Fatalf("expected AGENTS.md after Codex setup: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Claude Code MCP failed:") || !strings.Contains(out, "Codex MCP configured: knowblazer") {
		t.Fatalf("start output missing partial result details: %s", out)
	}
}

func writeFakeCLI(t *testing.T, dir string, name string, addOutput string, addCode int) {
	t.Helper()
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"mcp\" ] && [ \"$2\" = \"get\" ]; then exit 1; fi\n" +
		"if [ \"$1\" = \"mcp\" ] && [ \"$2\" = \"add\" ]; then echo \"" + addOutput + "\"; exit " + strconv.Itoa(addCode) + "; fi\n" +
		"exit 0\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake %s: %v", name, err)
	}
}

func TestRunStartCanTargetCodex(t *testing.T) {
	home := t.TempDir()
	workspace := filepath.Join(t.TempDir(), "kiblazer")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	t.Setenv("HOME", home)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"start", "--tool", "codex", "--path", workspace, "--skip-mcp"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("start code = %d, stderr = %s", code, stderr.String())
	}
	root := filepath.Join(home, "knowblazer-notes")
	if _, err := os.Stat(filepath.Join(root, "projects", "kiblazer.md")); err != nil {
		t.Fatalf("expected inferred project file: %v", err)
	}
	agentsMD := filepath.Join(workspace, "AGENTS.md")
	content, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "KNOWBLAZER-CODEX-SETUP:START") || !strings.Contains(text, "knowblazer_context") || !strings.Contains(text, root) {
		t.Fatalf("AGENTS.md missing Codex Knowblazer instructions:\n%s", text)
	}
	out := stdout.String()
	if !strings.Contains(out, "Knowblazer is ready for Codex in this project.") || !strings.Contains(out, "Codex instructions:") || !strings.Contains(out, "Next: run `codex`") {
		t.Fatalf("start output missing Codex details: %s", out)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"status", "--repo", root, "--path", workspace}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("status code = %d, stderr = %s", code, stderr.String())
	}
	out = stdout.String()
	if !strings.Contains(out, "Project: kiblazer") || !strings.Contains(out, "Codex instructions:") {
		t.Fatalf("status output missing Codex setup details: %s", out)
	}
}

func TestRunDoctorUsesRepoFlag(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"doctor", "--repo", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("doctor code = %d, stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "OK  config") {
		t.Fatalf("stdout missing config check: %s", out)
	}
	if !strings.Contains(out, "WARN  git-remote") {
		t.Fatalf("stdout missing local-first remote warning: %s", out)
	}
}

func TestRunDoctorDiscoversRepoFromEnvironment(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	t.Setenv("KNOWBLAZER_REPO", root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"doctor"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("doctor code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "OK  config") {
		t.Fatalf("stdout missing config check: %s", stdout.String())
	}
}

func TestRunDoctorReturnsOneForFailures(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	if err := os.Remove(filepath.Join(root, ".knowblazer", "config.json")); err != nil {
		t.Fatalf("remove config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"doctor", "--repo", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("doctor code = %d, want 1; stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "FAIL  config") {
		t.Fatalf("stdout missing config failure: %s", stdout.String())
	}
}

func TestRunInitUsesDefaultHomePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"init"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	root := filepath.Join(home, "knowblazer-notes")
	if _, err := os.Stat(filepath.Join(root, ".knowblazer", "config.json")); err != nil {
		t.Fatalf("expected default config to exist: %v", err)
	}
	if !strings.Contains(stdout.String(), root) {
		t.Fatalf("stdout missing default path: %s", stdout.String())
	}
}

func TestRunInitCreatesRepoAndPrintsNextSteps(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"init", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}

	if _, err := os.Stat(filepath.Join(root, ".knowblazer", "config.json")); err != nil {
		t.Fatalf("expected config to exist: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Initialized Knowblazer memory repo:") {
		t.Fatalf("stdout missing init message: %s", out)
	}
	if !strings.Contains(out, "knowblazer remember \"<lesson>\"") || !strings.Contains(out, "knowblazer recall \"<task>\"") || !strings.Contains(out, "knowblazer sync status") {
		t.Fatalf("stdout missing dynamic memory next steps: %s", out)
	}
}

func TestRunUnknownCommandReturnsError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"unknown"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("Run() code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr missing unknown command: %s", stderr.String())
	}
}

func TestRunScanPrintsRedactedFinding(t *testing.T) {
	file := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(file, []byte("password=super-secret-password\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"scan", file}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1; stderr = %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "HIGH") {
		t.Fatalf("stdout missing HIGH: %s", out)
	}
	if !strings.Contains(out, "password=****") {
		t.Fatalf("stdout missing redacted snippet: %s", out)
	}
	if !strings.Contains(out, "Summary: high, 1 findings, 1 high-risk") {
		t.Fatalf("stdout missing scan summary: %s", out)
	}
	if strings.Contains(out, "super-secret-password") {
		t.Fatalf("stdout leaked secret: %s", out)
	}
}

func TestRunScanPrintsCleanSummary(t *testing.T) {
	file := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(file, []byte("# Lesson\n\nNo secrets here.\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"scan", file}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Summary: clean, 0 findings, 0 high-risk") {
		t.Fatalf("stdout missing clean summary: %s", stdout.String())
	}
}

func TestRunScanAcceptsRepoFlagForSpecCompatibility(t *testing.T) {
	file := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(file, []byte("# Lesson\n\nNo secrets here.\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"scan", file, "--repo", t.TempDir()}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Summary: clean, 0 findings, 0 high-risk") {
		t.Fatalf("stdout missing clean summary: %s", stdout.String())
	}
}

func TestRunCaptureUsesRepoFlag(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	source := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(source, []byte("# Deploy Lesson\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"capture", source, "--repo", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("capture code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Captured to inbox:") {
		t.Fatalf("stdout missing capture message: %s", stdout.String())
	}
}

func TestRunCaptureDiscoversRepoFromEnvironment(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	t.Setenv("KNOWBLAZER_REPO", root)

	source := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(source, []byte("# Deploy Lesson\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"capture", source}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("capture code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Captured to inbox:") {
		t.Fatalf("stdout missing capture message: %s", stdout.String())
	}
}

func TestRunPromoteUsesRepoAndTargetFlags(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	source := filepath.Join(t.TempDir(), "lesson.md")
	if err := os.WriteFile(source, []byte("# Deploy Lesson\n\nNo secrets.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	var captureOut bytes.Buffer
	var captureErr bytes.Buffer
	if code := Run([]string{"capture", source, "--repo", root}, &captureOut, &captureErr); code != 0 {
		t.Fatalf("capture code = %d, stderr = %s", code, captureErr.String())
	}
	capturedPath := strings.TrimSpace(strings.TrimPrefix(strings.Split(captureOut.String(), "\n")[0], "Captured to inbox:"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"promote", capturedPath, "--to", "experience/deployment", "--repo", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("promote code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Promoted to:") {
		t.Fatalf("stdout missing promote message: %s", stdout.String())
	}
}

func TestRunAdapterImportAndIndex(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"adapter", "claude", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("adapter code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "knowblazer recall") {
		t.Fatalf("adapter output missing recall instructions: %s", stdout.String())
	}

	source := filepath.Join(t.TempDir(), "history.md")
	if err := os.WriteFile(source, []byte("# Deploy History\n\nRun smoke tests after deploy.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"import", "specstory", source, "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("import code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "inbox") {
		t.Fatalf("import output missing inbox path: %s", stdout.String())
	}

	experienceDir := filepath.Join(root, "experience", "deployment")
	if err := os.MkdirAll(experienceDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}
	experience := filepath.Join(experienceDir, "deploy.md")
	if err := os.WriteFile(experience, []byte("# Deploy\n\nRun smoke tests after deploy.\n"), 0o644); err != nil {
		t.Fatalf("write experience: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"index", "build", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("index build code = %d, stderr = %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"index", "search", "deploy", "smoke", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("index search code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "experience/deployment/deploy.md") {
		t.Fatalf("index search missing hit: %s", stdout.String())
	}
}

func TestRunRememberTextAndPositionalRecall(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"remember", "Deploys", "need", "smoke", "tests", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("remember code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Remembered to fresh memory:") {
		t.Fatalf("remember output missing fresh memory path: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), filepath.Join("experience", "auto")) {
		t.Fatalf("remember output missing experience auto path: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Auto-consolidated 1 fresh memories into:") {
		t.Fatalf("remember output missing automatic consolidation: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"recall", "deploy", "smoke", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("recall code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Knowblazer Recall Pack") {
		t.Fatalf("recall output missing pack: %s", stdout.String())
	}
}

func TestRunExplicitConsolidateIsNoopAfterAutomaticConsolidation(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"remember", "Deploys", "need", "smoke", "tests", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("remember code = %d, stderr = %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"status", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("status code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Fresh auto memories: 0") || !strings.Contains(stdout.String(), "Synthesized memories: 1") {
		t.Fatalf("status missing automatic consolidation counts: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"consolidate", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("consolidate code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No fresh auto memories to consolidate.") {
		t.Fatalf("consolidate should be noop after automatic consolidation: %s", stdout.String())
	}
}

func TestRunRememberDaily(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"remember", "finished", "deploy", "--daily", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("remember daily code = %d, stderr = %s", code, stderr.String())
	}
	content, err := os.ReadFile(filepath.Join(root, "daily", dailyFileName()))
	if err != nil {
		t.Fatalf("read daily note: %v", err)
	}
	if strings.Contains(string(content), root) {
		t.Fatalf("daily text included --repo value: %s", content)
	}
}

func dailyFileName() string {
	return time.Now().Format("2006-01-02") + ".md"
}

func TestRunDailyAddAndShow(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"daily", "add", "finished", "deploy", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("daily add code = %d, stderr = %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"daily", "show", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("daily show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "finished deploy") {
		t.Fatalf("daily show missing entry: %s", stdout.String())
	}
}

func TestRunSetupClaudeUpdatesProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	workspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"setup", "claude", "--repo", root, "--path", workspace, "--project", "kiblazer", "--skip-mcp"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("setup code = %d, stderr = %s", code, stderr.String())
	}
	claudeMD := filepath.Join(workspace, "CLAUDE.md")
	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "KNOWBLAZER-CLAUDE-SETUP:START") || !strings.Contains(text, root) || !strings.Contains(text, workspace) || !strings.Contains(text, "end-of-task memory candidate review") {
		t.Fatalf("CLAUDE.md missing Knowblazer instructions:\n%s", text)
	}
	if !strings.Contains(stdout.String(), "Claude Code integration ready.") {
		t.Fatalf("setup output missing ready message: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"project", "show", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("project show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "kiblazer") || !strings.Contains(stdout.String(), workspace) {
		t.Fatalf("project mapping missing: %s", stdout.String())
	}
}

func TestRunSetupUpgradesDefaultProjectStub(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	workspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	projectPath := filepath.Join(root, "projects", "kiblazer.md")
	if err := os.WriteFile(projectPath, []byte("# kiblazer\n\nAdd durable project context here.\n"), 0o644); err != nil {
		t.Fatalf("write project stub: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"setup", "claude", "--repo", root, "--path", workspace, "--project", "kiblazer", "--skip-mcp"}, &stdout, &stderr); code != 0 {
		t.Fatalf("setup code = %d, stderr = %s", code, stderr.String())
	}
	content, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatalf("read project: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "## Workspace") || !strings.Contains(text, workspace) {
		t.Fatalf("project stub not upgraded:\n%s", text)
	}
}

func TestRunSetupKeepsCustomProjectMemory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	workspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	projectPath := filepath.Join(root, "projects", "kiblazer.md")
	custom := "# kiblazer\n\nCustom durable context.\n"
	if err := os.WriteFile(projectPath, []byte(custom), 0o644); err != nil {
		t.Fatalf("write project custom: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"setup", "claude", "--repo", root, "--path", workspace, "--project", "kiblazer", "--skip-mcp"}, &stdout, &stderr); code != 0 {
		t.Fatalf("setup code = %d, stderr = %s", code, stderr.String())
	}
	content, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatalf("read project: %v", err)
	}
	if string(content) != custom {
		t.Fatalf("custom project memory was overwritten:\n%s", content)
	}
}

func TestRunSetupCodexUpdatesProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	workspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"setup", "codex", "--repo", root, "--path", workspace, "--project", "kiblazer", "--skip-mcp"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("setup code = %d, stderr = %s", code, stderr.String())
	}
	agentsMD := filepath.Join(workspace, "AGENTS.md")
	content, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "KNOWBLAZER-CODEX-SETUP:START") || !strings.Contains(text, root) || !strings.Contains(text, workspace) || !strings.Contains(text, "end-of-task memory candidate review") {
		t.Fatalf("AGENTS.md missing Knowblazer instructions:\n%s", text)
	}
	if !strings.Contains(stdout.String(), "Codex integration ready.") {
		t.Fatalf("setup output missing ready message: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"project", "show", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("project show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "kiblazer") || !strings.Contains(stdout.String(), workspace) {
		t.Fatalf("project mapping missing: %s", stdout.String())
	}
}

func TestRunProjectSetShowAndRecallMapping(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	workspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	project := filepath.Join(root, "projects", "kiblazer.md")
	if err := os.WriteFile(project, []byte("# Kiblazer\n\nMapped project context.\n"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"project", "set", "kiblazer", "--path", workspace, "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("project set code = %d, stderr = %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"project", "show", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("project show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "kiblazer") {
		t.Fatalf("project show missing mapping: %s", stdout.String())
	}
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir workspace: %v", err)
	}
	defer os.Chdir(oldwd)
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"recall", "--task", "deploy", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("recall code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Mapped project context.") {
		t.Fatalf("recall missing mapped project context: %s", stdout.String())
	}
}

func TestRunSyncStatus(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	if err := os.WriteFile(filepath.Join(root, "projects", "kiblazer.md"), []byte("# Kiblazer\n"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"sync", "status", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("sync status code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "projects/") {
		t.Fatalf("sync status missing projects directory: %s", stdout.String())
	}
}

func TestRunSyncDefaultsToStatus(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	if err := os.WriteFile(filepath.Join(root, "projects", "kiblazer.md"), []byte("# Kiblazer\n"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"sync", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("sync default code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "projects/") {
		t.Fatalf("sync default missing projects directory: %s", stdout.String())
	}
}

func TestRunBackupDreamAndReview(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}

	if err := os.WriteFile(filepath.Join(root, "daily", "2026-04-29.md"), []byte("# 2026-04-29\n\n- Fixed deploy issue.\n"), 0o644); err != nil {
		t.Fatalf("write daily: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"dream", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("dream code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Dynamic memory suggestions written to:") {
		t.Fatalf("dream output missing path: %s", stdout.String())
	}

	candidate := filepath.Join(root, "inbox", "note.md")
	if err := os.WriteFile(candidate, []byte("# Note\n\nRemember smoke tests.\n"), 0o644); err != nil {
		t.Fatalf("write candidate: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"review", "list", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("review list code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), candidate) {
		t.Fatalf("review list missing candidate: %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"review", "reject", candidate, "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("review reject code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Rejected to:") {
		t.Fatalf("review reject missing message: %s", stdout.String())
	}

	promoteSource := filepath.Join(root, "inbox", "promote.md")
	if err := os.WriteFile(promoteSource, []byte("# Promote\n\nReviewed lesson.\n"), 0o644); err != nil {
		t.Fatalf("write promote source: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"review", "promote", promoteSource, "--to", "experience/deployment", "--repo", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("review promote code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Promoted to:") {
		t.Fatalf("review promote missing message: %s", stdout.String())
	}

	backupPath := filepath.Join(t.TempDir(), "backup.tgz")
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"backup", "create", "--repo", root, "--output", backupPath, "--passphrase", "secret"}, &stdout, &stderr); code != 0 {
		t.Fatalf("backup create code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Backup written to:") {
		t.Fatalf("backup create missing message: %s", stdout.String())
	}
	restorePath := filepath.Join(t.TempDir(), "restore")
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"backup", "restore", "--input", backupPath, "--target", restorePath, "--passphrase", "secret"}, &stdout, &stderr); code != 0 {
		t.Fatalf("backup restore code = %d, stderr = %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(restorePath, ".knowblazer", "config.json")); err != nil {
		t.Fatalf("restored config missing: %v", err)
	}
}

func runTestGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func TestRunRecallWritesOutputFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	output := filepath.Join(t.TempDir(), "recall.md")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"recall", "--task", "deploy frontend", "--repo", root, "--output", output}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("recall code = %d, stderr = %s", code, stderr.String())
	}
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(content), "# Knowblazer Recall Pack") {
		t.Fatalf("output missing recall pack: %s", content)
	}
	if !strings.Contains(stdout.String(), "Recall pack written to:") {
		t.Fatalf("stdout missing output message: %s", stdout.String())
	}
}

func TestRunRecallPrintsPack(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", root}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, initErr.String())
	}
	project := filepath.Join(root, "projects", "kiblazer.md")
	if err := os.WriteFile(project, []byte("# Kiblazer\n\nFrontend deploys require smoke tests.\n"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"recall", "--task", "deploy frontend", "--project", "kiblazer", "--repo", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("recall code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Knowblazer Recall Pack") {
		t.Fatalf("stdout missing recall pack: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Frontend deploys require smoke tests.") {
		t.Fatalf("stdout missing project context: %s", stdout.String())
	}
}
