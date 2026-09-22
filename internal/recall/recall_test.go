package recall

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

func TestGenerateRecallWithChineseTask(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "deployment")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}
	content := "# Plane 部署经验\n\nPlane 平台使用 Docker Compose 部署在 /data/infra/plane，HTTP 端口 9103。\n"
	if err := os.WriteFile(filepath.Join(expDir, "plane.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write plane.md: %v", err)
	}

	pack, err := Generate(root, Options{Task: "排查Plane部署问题"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	text := string(pack)
	if !strings.Contains(text, "Plane 平台使用 Docker Compose 部署") {
		t.Fatalf("recall pack failed to match Chinese query:\n%s", text)
	}
}

func TestGeneratePrioritizesShortFocusedNoteOverLongDilutedNote(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "networking")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}

	// Focused short note: 2 hits in a concise 20-word note
	focused := "# gRPC Timeout\n\nFix grpc deadline exceeded by increasing client timeout to 10s.\n"
	if err := os.WriteFile(filepath.Join(expDir, "focused.md"), []byte(focused), 0o644); err != nil {
		t.Fatalf("write focused.md: %v", err)
	}

	// Diluted long note: 6 hits scattered across a 3000-word note
	diluted := "# Network Overview Log\n\n" + strings.Repeat("unrelated gateway routing traffic information packet header\n", 500)
	for i := 0; i < 6; i++ {
		diluted += "casually mentioned grpc deadline exceeded in background worker\n" + strings.Repeat("irrelevant log lines metrics counters\n", 50)
	}
	if err := os.WriteFile(filepath.Join(expDir, "diluted.md"), []byte(diluted), 0o644); err != nil {
		t.Fatalf("write diluted.md: %v", err)
	}

	pack, err := Generate(root, Options{Task: "fix grpc deadline exceeded timeout"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	text := string(pack)
	focusedIdx := strings.Index(text, "Fix grpc deadline exceeded")
	dilutedIdx := strings.Index(text, "Network Overview Log")
	if focusedIdx == -1 {
		t.Fatalf("focused note not found in pack:\n%s", text)
	}
	if dilutedIdx != -1 && focusedIdx > dilutedIdx {
		t.Fatalf("expected focused note (idx %d) to appear BEFORE diluted note (idx %d)", focusedIdx, dilutedIdx)
	}
}

func TestGenerateFiltersOutIncidentalLowConfidenceMatch(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "general")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}

	// 5 notes with common word "build" so "build" has high document frequency
	for i := 0; i < 5; i++ {
		content := fmt.Sprintf("# Note %d\n\nAlways build before running tests in CI %d.\n", i, i)
		if err := os.WriteFile(filepath.Join(expDir, fmt.Sprintf("note_%d.md", i)), []byte(content), 0o644); err != nil {
			t.Fatalf("write note: %v", err)
		}
	}

	// An incidental note that only mentions "build" once casually in an unrelated topic
	incidental := "# Cooking Recipe\n\nBuild sandwich with bread and butter.\n"
	if err := os.WriteFile(filepath.Join(expDir, "recipe.md"), []byte(incidental), 0o644); err != nil {
		t.Fatalf("write recipe: %v", err)
	}

	// Query has specific domain terms "kubernetes deployment rollout" plus common "build"
	// All notes contain "build" (df=N), so its score is miniscule (~0.07) and should be cut off by the quality gate.
	pack, err := Generate(root, Options{Task: "kubernetes pod crashloopbackoff build"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	text := string(pack)
	expSection := ""
	if idx := strings.Index(text, "Relevant Experience\n\n"); idx != -1 {
		rest := text[idx+len("Relevant Experience\n\n"):]
		if endIdx := strings.Index(rest, "\n## "); endIdx != -1 {
			expSection = rest[:endIdx]
		}
	}
	if !strings.Contains(expSection, "_No matching memory found._") {
		t.Fatalf("expected '_No matching memory found._' in Relevant Experience section, got:\n%s", expSection)
	}
}

func TestLongContextReorder(t *testing.T) {
	// <= 2 items unchanged
	if got := LongContextReorder([]string{"a"}); !reflect.DeepEqual(got, []string{"a"}) {
		t.Errorf("got %v, want [a]", got)
	}
	if got := LongContextReorder([]string{"a", "b"}); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("got %v, want [a, b]", got)
	}

	// 3 items: [1, 2, 3] -> [1, 3, 2]
	if got := LongContextReorder([]string{"1", "2", "3"}); !reflect.DeepEqual(got, []string{"1", "3", "2"}) {
		t.Errorf("3 items: got %v, want [1, 3, 2]", got)
	}

	// 4 items: [1, 2, 3, 4] -> [1, 3, 4, 2]
	if got := LongContextReorder([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(got, []string{"1", "3", "4", "2"}) {
		t.Errorf("4 items: got %v, want [1, 3, 4, 2]", got)
	}

	// 5 items: [1, 2, 3, 4, 5] -> [1, 3, 5, 4, 2]
	if got := LongContextReorder([]string{"1", "2", "3", "4", "5"}); !reflect.DeepEqual(got, []string{"1", "3", "5", "4", "2"}) {
		t.Errorf("5 items: got %v, want [1, 3, 5, 4, 2]", got)
	}
}

func TestGenerateAppliesLongContextReorderToPack(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "networking")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}

	// 3 notes with differing relevance:
	// Rank 1 (highest score: hits all 3 query terms "database connection timeout")
	writeNote := func(name, title, content string) {
		if err := os.WriteFile(filepath.Join(expDir, name), []byte(fmt.Sprintf("# %s\n\n%s\n", title, content)), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	writeNote("rank1.md", "Rank 1 Note", "database connection timeout in postgres pool")
	writeNote("rank2.md", "Rank 2 Note", "database connection pool settings max idle")
	writeNote("rank3.md", "Rank 3 Note", "database timeout retry logic query metrics logging info")

	pack, err := Generate(root, Options{Task: "database connection timeout"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	text := string(pack)

	idx1 := strings.Index(text, "Rank 1 Note")
	idx2 := strings.Index(text, "Rank 2 Note")
	idx3 := strings.Index(text, "Rank 3 Note")

	if idx1 == -1 || idx2 == -1 || idx3 == -1 {
		t.Fatalf("missing one or more notes in pack: idx1=%d, idx2=%d, idx3=%d\n%s", idx1, idx2, idx3, text)
	}

	// Reorder for 3 items [Rank1, Rank2, Rank3] -> [Rank1, Rank3, Rank2]
	if !(idx1 < idx3 && idx3 < idx2) {
		t.Fatalf("expected order [Rank1, Rank3, Rank2], got idx1=%d, idx3=%d, idx2=%d", idx1, idx3, idx2)
	}
}

func TestGenerateExtractsMatchingSectionWithBreadcrumbs(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "infra")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}

	multiSection := `# Infrastructure Guide

Overview of infrastructure setup.

## Database Setup

Postgres connection pool max connections should be set to 50.
Always configure SSL mode to require.

## Redis Cache

Redis cache eviction policy volatile-lru with 4GB memory limit.

## Email Gateway

SMTP server port 587 with TLS authentication.
`
	if err := os.WriteFile(filepath.Join(expDir, "infra.md"), []byte(multiSection), 0o644); err != nil {
		t.Fatalf("write infra.md: %v", err)
	}

	pack, err := Generate(root, Options{Task: "configure postgres connection pool"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	text := string(pack)

	if !strings.Contains(text, "Database Setup") || !strings.Contains(text, "Postgres connection pool") {
		t.Fatalf("pack missing Database Setup section:\n%s", text)
	}
	if !strings.Contains(text, "Infrastructure Guide > Database Setup") {
		t.Fatalf("pack missing breadcrumb 'Infrastructure Guide > Database Setup':\n%s", text)
	}
	// Sections that do not match should NOT be dumped into the pack
	if strings.Contains(text, "Redis cache eviction policy") {
		t.Fatalf("pack unnecessarily dumped unrelated Redis section:\n%s", text)
	}
	if strings.Contains(text, "SMTP server port 587") {
		t.Fatalf("pack unnecessarily dumped unrelated Email section:\n%s", text)
	}
}

func TestGenerateDualTierSemanticPromptPackaging(t *testing.T) {
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

	write("profile/preferences.md", "# Preferences\n\nPrefer existing patterns.\n")
	write("projects/kiblazer.md", "# Kiblazer\n\nMust pass linting before commit.\n")
	write("experience/deploy.md", "# Deploy\n\nPast lesson: restart worker before migrating.\n")

	pack, err := Generate(root, Options{Task: "deploy", Project: "kiblazer"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	text := string(pack)

	for _, want := range []string{
		"## Tier 1: Mandatory Constraints & Governance",
		"## Tier 2: Empirical Reference & Past Cases",
		"Tier 1 contains mandatory engineering constraints that must be preserved.",
		"Tier 2 provides historical lessons and reference heuristics to consult, not rigid commands.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("recall pack missing dual-tier element %q:\n%s", want, text)
		}
	}

	tier1Idx := strings.Index(text, "## Tier 1: Mandatory Constraints & Governance")
	tier2Idx := strings.Index(text, "## Tier 2: Empirical Reference & Past Cases")
	prefIdx := strings.Index(text, "Developer Preferences")
	projIdx := strings.Index(text, "Project Context")
	expIdx := strings.Index(text, "Relevant Experience")

	if !(tier1Idx < prefIdx && prefIdx < tier2Idx) {
		t.Fatalf("expected Developer Preferences inside Tier 1")
	}
	if !(tier1Idx < projIdx && projIdx < tier2Idx) {
		t.Fatalf("expected Project Context inside Tier 1")
	}
	if !(tier2Idx < expIdx) {
		t.Fatalf("expected Relevant Experience inside Tier 2")
	}
}

func TestGenerateWithExplainDiagnostics(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	expDir := filepath.Join(root, "experience", "deploy")
	if err := os.MkdirAll(expDir, 0o755); err != nil {
		t.Fatalf("mkdir experience: %v", err)
	}
	if err := os.WriteFile(filepath.Join(expDir, "deploy.md"), []byte("# Deploy Notes\n\nRun smoke test after deploy.\n"), 0o644); err != nil {
		t.Fatalf("write deploy.md: %v", err)
	}

	// With Explain: true
	packExplain, err := Generate(root, Options{Task: "deploy smoke", Explain: true})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	textExplain := string(packExplain)
	if !strings.Contains(textExplain, "## Recall Diagnostics") {
		t.Fatalf("expected '## Recall Diagnostics' in explain mode:\n%s", textExplain)
	}
	if !strings.Contains(textExplain, "Extracted Terms:") || !strings.Contains(textExplain, "BM25") {
		t.Fatalf("expected diagnostic details in explain mode:\n%s", textExplain)
	}

	// With Explain: false
	packNormal, err := Generate(root, Options{Task: "deploy smoke", Explain: false})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	textNormal := string(packNormal)
	if strings.Contains(textNormal, "## Recall Diagnostics") {
		t.Fatalf("did not expect '## Recall Diagnostics' in normal mode:\n%s", textNormal)
	}
}









