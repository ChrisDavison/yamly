package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davison/yamly/cmd"
	"github.com/davison/yamly/internal/frontmatter"
)

func writeMD(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestAddNewKey(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "add", path, "--key", "author", "--value", "alice")
	f, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if f.Data["author"] != "alice" {
		t.Errorf("author = %v, want alice", f.Data["author"])
	}
}

func TestAddSkipIfExists(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "add", path, "--key", "status", "--value", "published", "--skip-if-exists")
	f, _ := frontmatter.Parse(path)
	if f.Data["status"] != "draft" {
		t.Errorf("status changed despite --skip-if-exists, got %v", f.Data["status"])
	}
}

func TestAddOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "add", path, "--key", "status", "--value", "published", "--overwrite")
	f, _ := frontmatter.Parse(path)
	if f.Data["status"] != "published" {
		t.Errorf("status = %v, want published", f.Data["status"])
	}
}

func TestAddAppend(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\ntags:\n  - go\n---\nBody\n")
	execCmd(t, "add", path, "--key", "tags", "--value", "rust", "--append")
	f, _ := frontmatter.Parse(path)
	tags, ok := f.Data["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Fatalf("tags = %v, want [go rust]", f.Data["tags"])
	}
	if tags[1] != "rust" {
		t.Errorf("tags[1] = %v, want rust", tags[1])
	}
}

func TestAddAppendCreatesArray(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "add", path, "--key", "tags", "--value", "go", "--append")
	f, _ := frontmatter.Parse(path)
	tags, ok := f.Data["tags"].([]any)
	if !ok || len(tags) != 1 || tags[0] != "go" {
		t.Errorf("tags = %v, want [go]", f.Data["tags"])
	}
}

func TestAddDryRun(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "add", path, "--key", "author", "--value", "alice", "--dry-run")
	f, _ := frontmatter.Parse(path)
	if _, ok := f.Data["author"]; ok {
		t.Error("--dry-run should not have written the file")
	}
}

func TestAddNoFiles(t *testing.T) {
	cmd.SetArgs([]string{"add", "--key", "status", "--value", "draft"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no files provided")
	}
}
