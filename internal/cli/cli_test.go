package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelpIncludesDoctor(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("help code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "doctor [--repo <path>]") {
		t.Fatalf("stdout missing doctor usage: %s", stdout.String())
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
	if !strings.Contains(out, "WARN  git-repo") {
		t.Fatalf("stdout missing local-first git warning: %s", out)
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
	if !strings.Contains(out, "knowblazer capture <file>") {
		t.Fatalf("stdout missing next steps: %s", out)
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
	runTestGit(t, root, "init")
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
