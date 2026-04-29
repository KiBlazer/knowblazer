package mcp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knowblazer/knowblazer/internal/repo"
)

func TestServeClaudeFriendlyTools(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	workspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "projects", "workspace.md"), []byte("# Workspace\n\nUse smoke tests.\n"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".knowblazer"), 0o755); err != nil {
		t.Fatalf("mkdir .knowblazer: %v", err)
	}
	projectJSON := `{"mappings":[{"path":"` + filepath.ToSlash(workspace) + `","project":"workspace"}]}`
	if err := os.WriteFile(filepath.Join(root, ".knowblazer", "projects.json"), []byte(projectJSON), 0o644); err != nil {
		t.Fatalf("write project map: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "CLAUDE.md"), []byte("<!-- KNOWBLAZER-CLAUDE-SETUP:START -->\n"), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	input := strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
			`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n" +
			`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n" +
			`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"knowblazer_context","arguments":{"task":"deploy","workspace":"` + filepath.ToSlash(workspace) + `"}}}` + "\n" +
			`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"knowblazer_remember","arguments":{"text":"Remember smoke tests"}}}` + "\n" +
			`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"knowblazer_status","arguments":{"workspace":"` + filepath.ToSlash(workspace) + `"}}}` + "\n",
	)
	var output bytes.Buffer
	if err := Serve(root, input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d responses: %s", len(lines), output.String())
	}
	if !strings.Contains(lines[0], "serverInfo") || !strings.Contains(lines[1], "knowblazer_context") || !strings.Contains(lines[2], "Use smoke tests.") || !strings.Contains(lines[3], "auto_promoted") || !strings.Contains(lines[4], "claude_configured") {
		t.Fatalf("unexpected responses:\n%s", output.String())
	}
}

func TestServeRecallRequest(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory")
	if err := repo.Init(root); err != nil {
		t.Fatalf("repo.Init() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "profile", "preferences.md"), []byte("# Preferences\n\nPrefer tests.\n"), 0o644); err != nil {
		t.Fatalf("write preferences: %v", err)
	}
	input := strings.NewReader(`{"id":1,"method":"knowblazer_recall","params":{"task":"test"}}` + "\n")
	var output bytes.Buffer
	if err := Serve(root, input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	var resp Response
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result, _ := resp.Result.(string)
	if resp.Error != "" || !strings.Contains(result, "Prefer tests.") {
		t.Fatalf("response = %#v", resp)
	}
}
