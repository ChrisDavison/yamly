package frontmatter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davison/yamly/internal/frontmatter"
)

func TestParseBytesScalar(t *testing.T) {
	content := []byte("---\nstatus: draft\n---\nBody text\n")
	f, err := frontmatter.ParseBytes(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Data["status"] != "draft" {
		t.Errorf("status = %v, want draft", f.Data["status"])
	}
}

func TestParseBytesArray(t *testing.T) {
	content := []byte("---\ntags:\n  - go\n  - cli\n---\nBody\n")
	f, err := frontmatter.ParseBytes(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags, ok := f.Data["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Fatalf("tags = %v, want [go cli]", f.Data["tags"])
	}
	if tags[0] != "go" || tags[1] != "cli" {
		t.Errorf("tags = %v, want [go cli]", tags)
	}
}

func TestParseBytesNoFrontmatter(t *testing.T) {
	content := []byte("Just body text\n")
	_, err := frontmatter.ParseBytes(content)
	if err != frontmatter.ErrNoFrontmatter {
		t.Errorf("err = %v, want ErrNoFrontmatter", err)
	}
}

func TestParseBytesEmptyFrontmatter(t *testing.T) {
	content := []byte("---\n---\nBody\n")
	f, err := frontmatter.ParseBytes(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Data) != 0 {
		t.Errorf("expected empty data, got %v", f.Data)
	}
}

func TestWriteFileRoundTrip(t *testing.T) {
	content := []byte("---\nstatus: draft\n---\nBody text\n")
	f, err := frontmatter.ParseBytes(content)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	f.Data["status"] = "published"

	tmp := filepath.Join(t.TempDir(), "test.md")
	if err := f.WriteFile(tmp); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	f2, err := frontmatter.Parse(tmp)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if f2.Data["status"] != "published" {
		t.Errorf("status = %v, want published", f2.Data["status"])
	}
}

func TestWriteFilePreservesBody(t *testing.T) {
	content := []byte("---\nstatus: draft\n---\nHello world\n")
	f, err := frontmatter.ParseBytes(content)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tmp := filepath.Join(t.TempDir(), "test.md")
	if err := f.WriteFile(tmp); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	written, _ := os.ReadFile(tmp)
	if string(written[len(written)-12:]) != "Hello world\n" {
		t.Errorf("body not preserved, got: %q", string(written))
	}
}
