package cmd_test

import (
	"testing"

	"github.com/davison/yamlsum/internal/frontmatter"
)

func TestSubstituteDefaultReplaces(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "substitute", path, "--key", "status", "--new-value", "published")
	f, _ := frontmatter.Parse(path)
	if f.Data["status"] != "published" {
		t.Errorf("status = %v, want published", f.Data["status"])
	}
}

func TestSubstituteDefaultSkipsMissing(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nauthor: alice\n---\nBody\n")
	execCmd(t, "substitute", path, "--key", "status", "--new-value", "published")
	f, _ := frontmatter.Parse(path)
	if _, ok := f.Data["status"]; ok {
		t.Error("status should not have been added when key is absent and no --overwrite")
	}
}

func TestSubstituteOldValue(t *testing.T) {
	dir := t.TempDir()
	pathA := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	pathB := writeMD(t, dir, "b.md", "---\nstatus: published\n---\nBody\n")
	execCmd(t, "substitute", pathA, pathB, "--key", "status", "--new-value", "archived", "--old-value", "draft")
	fA, _ := frontmatter.Parse(pathA)
	fB, _ := frontmatter.Parse(pathB)
	if fA.Data["status"] != "archived" {
		t.Errorf("a.md status = %v, want archived", fA.Data["status"])
	}
	if fB.Data["status"] != "published" {
		t.Errorf("b.md status = %v, want published (should be skipped)", fB.Data["status"])
	}
}

func TestSubstituteOverwriteAddsKey(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nauthor: alice\n---\nBody\n")
	execCmd(t, "substitute", path, "--key", "status", "--new-value", "published", "--overwrite")
	f, _ := frontmatter.Parse(path)
	if f.Data["status"] != "published" {
		t.Errorf("status = %v, want published", f.Data["status"])
	}
}

func TestSubstituteDryRun(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "substitute", path, "--key", "status", "--new-value", "published", "--dry-run")
	f, _ := frontmatter.Parse(path)
	if f.Data["status"] != "draft" {
		t.Error("--dry-run should not have written the file")
	}
}
