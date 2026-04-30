package recall

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestGenerateIncludesRelevantMemoryAndExcludesInboxAndQuarantine(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	write := func(rel string, content string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	write("profile/preferences.md", "# Preferences\n\nPrefer explicit deployment checks.\n")
	write("projects/kiblazer.md", "# Kiblazer\n\nFrontend deploys require a smoke test.\n")
	write("experience/deployment/frontend-deploy.md", "# Frontend Deploy\n\nRun rollback checks after deploy.\n")
	write("inbox/2026-04-29/raw.md", "# Raw\n\nDo not include inbox.\n")
	write("quarantine/2026-04-29/secret.md", "# Secret\n\npassword=super-secret-password\n")

	pack, err := Generate(root, Options{Task: "deploy frontend", Project: "kiblazer"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	text := string(pack)
	for _, want := range []string{
		"# Knowblazer Recall Pack",
		"Task: deploy frontend",
		"Prefer explicit deployment checks.",
		"Frontend deploys require a smoke test.",
		"Run rollback checks after deploy.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("recall pack missing %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{"Do not include inbox.", "super-secret-password"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("recall pack included unwanted %q:\n%s", unwanted, text)
		}
	}
}

func TestGenerateLimitsTotalOutputSize(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	large := strings.Repeat("deploy ", 6000)
	profileDir := filepath.Join(root, "profile")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profileDir, "preferences.md"), []byte(large), 0o644); err != nil {
		t.Fatalf("write preferences: %v", err)
	}
	deployDir := filepath.Join(root, "experience", "deployment")
	if err := os.MkdirAll(deployDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}
	for i := 0; i < 5; i++ {
		path := filepath.Join(deployDir, fmt.Sprintf("deploy-%d.md", i))
		if err := os.WriteFile(path, []byte(large), 0o644); err != nil {
			t.Fatalf("write experience: %v", err)
		}
	}

	pack, err := Generate(root, Options{Task: "deploy"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(pack) > maxPackBytes {
		t.Fatalf("pack len = %d, want <= %d", len(pack), maxPackBytes)
	}
	if !strings.Contains(string(pack), "[truncated]") {
		t.Fatalf("pack missing truncation marker")
	}
}

func TestGeneratePrioritizesSynthesizedMemoryOverFreshAutoMemory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	write := func(rel string, content string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	for i := 0; i < 5; i++ {
		write(fmt.Sprintf("experience/auto/fresh-%d.md", i), fmt.Sprintf("# Fresh %d\n\ndeploy smoke auto memory %d\n", i, i))
	}
	write("experience/synthesized/deploy.md", "# Synthesized\n\ndeploy smoke synthesized memory\n")

	pack, err := Generate(root, Options{Task: "deploy smoke"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	text := string(pack)
	if !strings.Contains(text, "deploy smoke synthesized memory") {
		t.Fatalf("recall pack missing synthesized memory:\n%s", text)
	}
	if strings.Contains(text, "auto memory 4") {
		t.Fatalf("recall pack included auto memory before synthesized priority filled capacity:\n%s", text)
	}
	if !strings.Contains(text, "prioritizes synthesized memory") {
		t.Fatalf("recall pack missing synthesized caution:\n%s", text)
	}
}

func TestGenerateRequiresTask(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}

	if _, err := Generate(root, Options{}); err == nil {
		t.Fatal("Generate() error = nil, want missing task error")
	}
}

func TestGenerateSkipsExperienceReadmeFiles(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	readmeDir := filepath.Join(root, "experience", "deployment")
	if err := os.MkdirAll(readmeDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}
	readme := filepath.Join(readmeDir, "README.md")
	if err := os.WriteFile(readme, []byte("# Deployment Experience\n\nReusable deployment lessons go here.\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	pack, err := Generate(root, Options{Task: "deployment"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if strings.Contains(string(pack), "Reusable deployment lessons go here.") {
		t.Fatalf("recall pack included README content:\n%s", pack)
	}
}
