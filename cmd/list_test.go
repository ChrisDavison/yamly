package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davison/yamlsum/cmd"
)

// execCmd runs yamlsum with the given args and captures stdout.
func execCmd(t *testing.T, args ...string) string {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(%v): %v", args, err)
	}
	return buf.String()
}

func makeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("---\nstatus: draft\ntags:\n  - go\n  - cli\n---\nBody\n"), 0644)
	os.WriteFile(filepath.Join(sub, "b.md"), []byte("---\nstatus: published\ntags:\n  - go\n---\nBody\n"), 0644)
	os.WriteFile(filepath.Join(dir, "c.md"), []byte("no frontmatter"), 0644)
	return dir
}

func TestListKey(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "list", "--dir", dir, "--key", "status")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 results, got %d: %v", len(lines), lines)
	}
}

func TestListKeyValue(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "list", "--dir", dir, "--key", "status", "--value", "draft")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 result, got %d: %v", len(lines), lines)
	}
}

func TestListArrayMembership(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "list", "--dir", dir, "--key", "tags", "--value", "go")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 results (both have 'go' tag), got %d: %v", len(lines), lines)
	}
}

func TestListAsJSON(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "list", "--dir", dir, "--key", "status", "--as-json")
	if !strings.HasPrefix(out, "[") {
		t.Errorf("expected JSON array, got: %s", out)
	}
}

func TestCountKey(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "count", "--dir", dir, "--key", "status")
	if !strings.Contains(out, "draft: 1") || !strings.Contains(out, "published: 1") {
		t.Errorf("unexpected count output: %s", out)
	}
}

func TestCountAsJSON(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "count", "--dir", dir, "--key", "status", "--as-json")
	if !strings.HasPrefix(out, "{") {
		t.Errorf("expected JSON object, got: %s", out)
	}
}

func TestKeysText(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "keys", "--dir", dir)
	if !strings.Contains(out, "status: 2") {
		t.Errorf("expected 'status: 2' in output, got: %s", out)
	}
	if !strings.Contains(out, "tags: 2") {
		t.Errorf("expected 'tags: 2' in output, got: %s", out)
	}
}

func TestKeysAsJSON(t *testing.T) {
	dir := makeFixture(t)
	out := execCmd(t, "keys", "--dir", dir, "--as-json")
	if !strings.HasPrefix(out, "{") {
		t.Errorf("expected JSON object, got: %s", out)
	}
	if !strings.Contains(out, `"status":2`) {
		t.Errorf("expected status:2 in JSON, got: %s", out)
	}
}
