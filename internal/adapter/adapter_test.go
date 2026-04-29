package adapter

import (
	"strings"
	"testing"
)

func TestGenerateSupportedAdapters(t *testing.T) {
	for _, tool := range Tools() {
		content, err := Generate(tool, "/tmp/memory")
		if err != nil {
			t.Fatalf("Generate(%s) error = %v", tool, err)
		}
		text := string(content)
		if !strings.Contains(text, "knowblazer recall") || !strings.Contains(text, "/tmp/memory") {
			t.Fatalf("adapter %s content invalid:\n%s", tool, text)
		}
	}
}

func TestGenerateRejectsUnknownAdapter(t *testing.T) {
	if _, err := Generate("unknown", "/tmp/memory"); err == nil {
		t.Fatal("Generate() error = nil, want unsupported adapter error")
	}
}
