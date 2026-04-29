package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	if strings.Contains(out, "super-secret-password") {
		t.Fatalf("stdout leaked secret: %s", out)
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
