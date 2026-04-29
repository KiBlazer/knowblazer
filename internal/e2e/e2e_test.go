package e2e

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/cli"
)

func TestMVPFlow(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := filepath.Join(tmp, "memory")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := cli.Run([]string{"init", repoRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, stderr.String())
	}

	lesson := filepath.Join(tmp, "frontend-deploy.md")
	if err := os.WriteFile(lesson, []byte("# Frontend Deploy\n\nRun smoke tests after frontend deploy.\n"), 0o644); err != nil {
		t.Fatalf("write lesson: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if code := cli.Run([]string{"capture", lesson, "--repo", repoRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("capture code = %d, stderr = %s", code, stderr.String())
	}
	capturedPath := strings.TrimSpace(strings.TrimPrefix(strings.Split(stdout.String(), "\n")[0], "Captured to inbox:"))
	if capturedPath == "" {
		t.Fatalf("could not parse captured path from %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := cli.Run([]string{"promote", capturedPath, "--to", "experience/deployment", "--repo", repoRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("promote code = %d, stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := cli.Run([]string{"recall", "--task", "frontend deploy", "--repo", repoRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("recall code = %d, stderr = %s", code, stderr.String())
	}
	pack := stdout.String()
	if !strings.Contains(pack, "# Knowblazer Recall Pack") {
		t.Fatalf("recall pack missing title:\n%s", pack)
	}
	if !strings.Contains(pack, "Run smoke tests after frontend deploy.") {
		t.Fatalf("recall pack missing promoted lesson:\n%s", pack)
	}
	if strings.Contains(pack, "Raw candidate memory goes here.") {
		t.Fatalf("recall pack included inbox README:\n%s", pack)
	}
}

func TestMVPFlowQuarantinesSecrets(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := filepath.Join(tmp, "memory")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := cli.Run([]string{"init", repoRoot}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d, stderr = %s", code, stderr.String())
	}

	lesson := filepath.Join(tmp, "secret.md")
	if err := os.WriteFile(lesson, []byte("# Secret\n\nAuthorization: Bearer abcdefghijklmnopqrstuvwxyz\n"), 0o644); err != nil {
		t.Fatalf("write lesson: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if code := cli.Run([]string{"capture", lesson, "--repo", repoRoot}, &stdout, &stderr); code != 1 {
		t.Fatalf("capture code = %d, want 1; stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Moved to quarantine:") {
		t.Fatalf("stdout missing quarantine message: %s", stdout.String())
	}
}
