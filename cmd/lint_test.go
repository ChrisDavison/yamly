package cmd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davison/yamly/cmd"
)

func makeLintVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// a.md: duplicate value in tags, empty value on "draft"
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("---\ntags:\n  - go\n  - go\n  - rust\ndraft:\n---\nBody\n"), 0644)
	// b.md: tags is a scalar — inconsistent with a.md's list
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("---\ntags: go\ntitle: Hello\n---\n"), 0644)
	// c.md: no frontmatter
	os.WriteFile(filepath.Join(dir, "c.md"), []byte("no frontmatter here\n"), 0644)
	// d.md: malformed YAML
	os.WriteFile(filepath.Join(dir, "d.md"), []byte("---\nbad: [unclosed\n---\n"), 0644)
	// e.md: clean
	os.WriteFile(filepath.Join(dir, "e.md"), []byte("---\nstatus: published\n---\n"), 0644)

	return dir
}

func runLintCmdDir(t *testing.T, dir string, extraArgs ...string) string {
	t.Helper()
	args := []string{"lint", "--dir", dir}
	args = append(args, extraArgs...)
	return execCmd(t, args...)
}

// execLintNoFail runs lint and returns output even if Execute returns an error.
// lint always exits 0, so a failure here is a real error.
func TestLintMalformedYAML(t *testing.T) {
	dir := makeLintVault(t)
	out := runLintCmdDir(t, dir)
	if !strings.Contains(out, "malformed-yaml") {
		t.Errorf("expected malformed-yaml finding, got:\n%s", out)
	}
	// d.md should be named in the finding
	if !strings.Contains(out, "d.md") {
		t.Errorf("expected d.md in output, got:\n%s", out)
	}
}

func TestLintDuplicateValue(t *testing.T) {
	dir := makeLintVault(t)
	out := runLintCmdDir(t, dir)
	if !strings.Contains(out, "duplicate-value") {
		t.Errorf("expected duplicate-value finding, got:\n%s", out)
	}
	if !strings.Contains(out, `"go"`) {
		t.Errorf("expected duplicate value 'go' mentioned, got:\n%s", out)
	}
}

func TestLintEmptyValue(t *testing.T) {
	dir := makeLintVault(t)
	out := runLintCmdDir(t, dir)
	if !strings.Contains(out, "empty-value") {
		t.Errorf("expected empty-value finding, got:\n%s", out)
	}
	if !strings.Contains(out, `"draft"`) {
		t.Errorf("expected key 'draft' mentioned in empty-value finding, got:\n%s", out)
	}
}

func TestLintInconsistentType(t *testing.T) {
	dir := makeLintVault(t)
	out := runLintCmdDir(t, dir)
	if !strings.Contains(out, "inconsistent-type") {
		t.Errorf("expected inconsistent-type finding, got:\n%s", out)
	}
	if !strings.Contains(out, "(across files)") {
		t.Errorf("expected '(across files)' prefix for cross-file finding, got:\n%s", out)
	}
}

func TestLintMissingFrontmatterOffByDefault(t *testing.T) {
	dir := makeLintVault(t)
	out := runLintCmdDir(t, dir)
	if strings.Contains(out, "missing-frontmatter") {
		t.Errorf("missing-frontmatter should not appear without --check-missing, got:\n%s", out)
	}
}

func TestLintCheckMissingFlag(t *testing.T) {
	dir := makeLintVault(t)
	out := runLintCmdDir(t, dir, "--check-missing")
	if !strings.Contains(out, "missing-frontmatter") {
		t.Errorf("expected missing-frontmatter with --check-missing, got:\n%s", out)
	}
	if !strings.Contains(out, "c.md") {
		t.Errorf("expected c.md in missing-frontmatter finding, got:\n%s", out)
	}
}

func TestLintEmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "empty.md"), []byte("---\n---\n"), 0644)

	// Without --check-missing: no finding
	out := runLintCmdDir(t, dir)
	if strings.Contains(out, "empty-frontmatter") {
		t.Errorf("empty-frontmatter should not appear without --check-missing, got:\n%s", out)
	}

	// With --check-missing: finding emitted
	out = runLintCmdDir(t, dir, "--check-missing")
	if !strings.Contains(out, "empty-frontmatter") {
		t.Errorf("expected empty-frontmatter with --check-missing, got:\n%s", out)
	}
}

func TestLintCleanVaultNoFindings(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.md"), []byte("---\nstatus: published\ntags:\n  - go\n  - rust\n---\n"), 0644)

	out := runLintCmdDir(t, dir)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected no output for clean vault, got:\n%s", out)
	}
}

func TestLintAsJSON(t *testing.T) {
	dir := makeLintVault(t)
	out := runLintCmdDir(t, dir, "--as-json")

	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Fatalf("expected JSON array, got:\n%s", out)
	}

	var findings []map[string]any
	if err := json.Unmarshal([]byte(out), &findings); err != nil {
		t.Fatalf("invalid JSON: %v\noutput:\n%s", err, out)
	}

	rules := make(map[string]bool)
	for _, f := range findings {
		if r, ok := f["rule"].(string); ok {
			rules[r] = true
		}
	}
	for _, want := range []string{"malformed-yaml", "duplicate-value", "empty-value", "inconsistent-type"} {
		if !rules[want] {
			t.Errorf("expected rule %q in JSON output; rules found: %v", want, rules)
		}
	}
}

func TestLintAsJSONEmptyIsArray(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.md"), []byte("---\nstatus: ok\n---\n"), 0644)
	cmd.SetArgs([]string{"lint", "--dir", dir, "--as-json"})
	out := runLintCmdDir(t, dir, "--as-json")
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected [] for clean vault JSON, got: %s", out)
	}
}
