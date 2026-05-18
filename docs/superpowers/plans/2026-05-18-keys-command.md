# keys Command Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `keys` subcommand that lists all frontmatter keys found in a directory along with the count of files containing each key.

**Architecture:** Single new file `cmd/keys.go` mirroring `cmd/count.go`. Walk all markdown files, accumulate a `map[string]int` of key→file-count, sort by count desc then key asc, output as text or JSON. Tests go in the existing `cmd/list_test.go` file alongside the other command tests (the file is a shared cmd_test package fixture).

**Tech Stack:** Go, cobra, gopkg.in/yaml.v3 (via internal/frontmatter), standard library

---

### Task 1: Write failing tests for `keys`

**Files:**
- Modify: `cmd/list_test.go`

The existing `makeFixture` helper creates:
- `a.md` — keys: `status`, `tags`
- `sub/b.md` — keys: `status`, `tags`
- `c.md` — no frontmatter

So `status` appears in 2 files, `tags` appears in 2 files.

- [ ] **Step 1: Add tests to `cmd/list_test.go`**

Append these functions to the file:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/davison/code/yamly && go test ./cmd/... -run TestKeys -v
```

Expected: FAIL — `unknown command "keys"`

---

### Task 2: Implement `cmd/keys.go`

**Files:**
- Create: `cmd/keys.go`

- [ ] **Step 1: Create the file**

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/davison/yamly/internal/frontmatter"
	"github.com/davison/yamly/internal/walk"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List all frontmatter keys and the number of files containing each",
	RunE:  runKeys,
}

var keysJSON bool

func init() {
	rootCmd.AddCommand(keysCmd)
	keysCmd.Flags().BoolVar(&keysJSON, "as-json", false, "output as JSON object {key: count}")
}

func runKeys(cmd *cobra.Command, args []string) error {
	files, err := walk.Walk(dir)
	if err != nil {
		return err
	}

	counts := make(map[string]int)
	for _, path := range files {
		f, err := frontmatter.Parse(path)
		if err != nil {
			continue
		}
		for k := range f.Data {
			counts[k]++
		}
	}

	type entry struct {
		key   string
		count int
	}
	entries := make([]entry, 0, len(counts))
	for k, c := range counts {
		entries = append(entries, entry{k, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].key < entries[j].key
	})

	out := cmd.OutOrStdout()
	if keysJSON {
		m := make(map[string]int, len(entries))
		for _, e := range entries {
			m[e.key] = e.count
		}
		return json.NewEncoder(out).Encode(m)
	}
	for _, e := range entries {
		fmt.Fprintf(out, "%s: %d\n", e.key, e.count)
	}
	return nil
}
```

- [ ] **Step 2: Run tests to verify they pass**

```bash
cd /home/davison/code/yamly && go test ./cmd/... -run TestKeys -v
```

Expected: PASS for `TestKeysText` and `TestKeysAsJSON`

- [ ] **Step 3: Run the full test suite**

```bash
cd /home/davison/code/yamly && go test ./...
```

Expected: all tests pass

- [ ] **Step 4: Commit**

```bash
git add cmd/keys.go cmd/list_test.go
git commit -m "feat: add keys subcommand"
```
