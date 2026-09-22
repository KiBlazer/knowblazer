package textutil

import (
	"reflect"
	"sort"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "Deploy smoke tests",
			want:  []string{"deploy", "smoke", "tests"},
		},
		{
			input: "Plane部署",
			want:  []string{"plane", "部署", "部", "署"},
		},
		{
			input: "排查Docker问题",
			want:  []string{"docker", "排查", "排", "查", "问题", "问", "题"},
		},
	}

	for _, tt := range tests {
		got := Tokenize(tt.input)
		sort.Strings(got)
		sort.Strings(tt.want)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Tokenize(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestTermFrequencies(t *testing.T) {
	counts, total := TermFrequencies("deploy deploy smoke")
	if counts["deploy"] != 2 {
		t.Errorf("counts['deploy'] = %d, want 2", counts["deploy"])
	}
	if counts["smoke"] != 1 {
		t.Errorf("counts['smoke'] = %d, want 1", counts["smoke"])
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
}

