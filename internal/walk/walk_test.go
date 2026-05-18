package walk_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/davison/yamly/internal/walk"
)

func TestWalkFindsMarkdown(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)

	os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(sub, "b.md"), []byte("# B"), 0644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("ignore"), 0644)

	got, err := walk.Walk(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Strings(got)
	if len(got) != 2 {
		t.Fatalf("got %d files, want 2: %v", len(got), got)
	}
}

func TestWalkEmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := walk.Walk(dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestWalkExcludesDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "skip")
	os.MkdirAll(sub, 0755)

	os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(sub, "b.md"), []byte("# B"), 0644)

	got, err := walk.Walk(dir, []string{"skip"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1: %v", len(got), got)
	}
}

func TestWalkNonexistentDir(t *testing.T) {
	_, err := walk.Walk("/nonexistent/path/abc123", nil)
	if err == nil {
		t.Error("expected error for nonexistent dir, got nil")
	}
}
