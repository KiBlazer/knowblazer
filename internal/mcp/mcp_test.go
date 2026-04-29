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
