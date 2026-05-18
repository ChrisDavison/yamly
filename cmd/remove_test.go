package cmd_test

import (
	"testing"

	"github.com/davison/yamlsum/internal/frontmatter"
)

func TestRemoveKey(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\nauthor: alice\n---\nBody\n")
	execCmd(t, "remove", path, "--key", "status")
	f, _ := frontmatter.Parse(path)
	if _, ok := f.Data["status"]; ok {
		t.Error("status key should have been removed")
	}
	if f.Data["author"] != "alice" {
		t.Error("author key should be unchanged")
	}
}

func TestRemoveKeyWithValue(t *testing.T) {
	dir := t.TempDir()
	pathA := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	pathB := writeMD(t, dir, "b.md", "---\nstatus: published\n---\nBody\n")
	execCmd(t, "remove", pathA, pathB, "--key", "status", "--value", "draft")
	fA, _ := frontmatter.Parse(pathA)
	fB, _ := frontmatter.Parse(pathB)
	if _, ok := fA.Data["status"]; ok {
		t.Error("a.md: status should have been removed (value matched)")
	}
	if fB.Data["status"] != "published" {
		t.Error("b.md: status should not have been removed (value didn't match)")
	}
}

func TestRemoveDryRun(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "remove", path, "--key", "status", "--dry-run")
	f, _ := frontmatter.Parse(path)
	if _, ok := f.Data["status"]; !ok {
		t.Error("--dry-run should not have written the file")
	}
}
