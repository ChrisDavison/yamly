package cmd_test

import (
	"testing"

	"github.com/davison/yamlsum/internal/frontmatter"
)

func TestRenameKey(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "rename", path, "--old-key", "status", "--new-key", "state")
	f, _ := frontmatter.Parse(path)
	if _, ok := f.Data["status"]; ok {
		t.Error("old key 'status' should be gone")
	}
	if f.Data["state"] != "draft" {
		t.Errorf("state = %v, want draft", f.Data["state"])
	}
}

func TestRenameSkipsMissingOldKey(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nauthor: alice\n---\nBody\n")
	execCmd(t, "rename", path, "--old-key", "status", "--new-key", "state")
	f, _ := frontmatter.Parse(path)
	if _, ok := f.Data["state"]; ok {
		t.Error("state should not have been created when old-key was absent")
	}
}

func TestRenameErrorsIfNewKeyExists(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\nstate: old\n---\nBody\n")
	// collision → failed result, but Execute itself does not error
	execCmd(t, "rename", path, "--old-key", "status", "--new-key", "state")
	f, _ := frontmatter.Parse(path)
	if f.Data["status"] != "draft" {
		t.Error("status should be unchanged when rename fails due to new-key collision")
	}
}

func TestRenameDryRun(t *testing.T) {
	dir := t.TempDir()
	path := writeMD(t, dir, "a.md", "---\nstatus: draft\n---\nBody\n")
	execCmd(t, "rename", path, "--old-key", "status", "--new-key", "state", "--dry-run")
	f, _ := frontmatter.Parse(path)
	if _, ok := f.Data["status"]; !ok {
		t.Error("--dry-run should not have written the file")
	}
}
