package frontmatter_test

import (
	"testing"

	"github.com/davison/yamly/internal/frontmatter"
)

func TestParseValue(t *testing.T) {
	v, err := frontmatter.ParseValue("true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != true {
		t.Errorf("got %v (%T), want true (bool)", v, v)
	}

	v, err = frontmatter.ParseValue("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Errorf("got %v (%T), want 42 (int)", v, v)
	}

	v, err = frontmatter.ParseValue("draft")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "draft" {
		t.Errorf("got %v, want draft", v)
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name  string
		field any
		query any
		want  bool
	}{
		{"scalar string match", "draft", "draft", true},
		{"scalar string no match", "draft", "published", false},
		{"scalar bool match", true, true, true},
		{"scalar bool no match", true, false, false},
		{"scalar int match", 42, 42, true},
		{"array contains value", []any{"go", "cli"}, "go", true},
		{"array missing value", []any{"go", "cli"}, "rust", false},
		{"array exact no match", []any{"go"}, []any{"go"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frontmatter.Matches(tt.field, tt.query)
			if got != tt.want {
				t.Errorf("Matches(%v, %v) = %v, want %v", tt.field, tt.query, got, tt.want)
			}
		})
	}
}
