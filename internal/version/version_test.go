package version

import (
	"strings"
	"testing"
)

func TestGetDefault(t *testing.T) {
	v := Get()
	if v == "" {
		t.Fatalf("Get() returned empty version")
	}
}

func TestFullString(t *testing.T) {
	s := FullString()
	if !strings.HasPrefix(s, "knowblazer v") {
		t.Fatalf("FullString() = %q, want prefix 'knowblazer v'", s)
	}
}
