# yamly — Design Spec

**Date:** 2026-05-18

## Purpose

A Go CLI tool for inspecting and bulk-editing YAML frontmatter across nested directories of markdown files. Intended for knowledge-base and notes management workflows where frontmatter keys need to be queried, audited, and migrated across many files.

---

## CLI Shape

Global default: search/resolve from current working directory. Override with `--dir <path>` on any subcommand.

All `--value` / `--new-value` / `--old-value` arguments are parsed as YAML scalars, so `--value true` matches a boolean `true`, `--value 42` matches an integer, and `--value "[a, b]"` matches a sequence.

**Array membership:** when a frontmatter field holds a YAML sequence, `--value <v>` matches if `v` is a member of the sequence (not an exact match on the whole array).

---

## Subcommands

### `list`

Find files that contain a key, optionally filtered by value.

```
yamly list --key <key> [--value <value>] [--dir <path>] [--as-json]
```

- Walks `--dir` recursively for `*.md` files.
- With `--value`: matches if the field equals the value, or (for array fields) contains it as a member.
- Output: one filepath per line. `--as-json`: JSON array of strings.

---

### `count`

Count occurrences of each distinct value for a key across all markdown files.

```
yamly count --key <key> [--dir <path>] [--as-json]
```

- Files without the key are ignored.
- Output: `value: N` lines sorted by count descending. `--as-json`: `{"value": N, ...}`.

---

### `add`

Add a key/value pair to specific files.

```
yamly add <file> [<file>...] --key <key> --value <value> [--as-json] [--dry-run]
           [--skip-if-exists | --fail-if-exists | --overwrite | --append]
```

- Requires at least one file argument; errors if none provided.
- `--skip-if-exists`: silently skip files where the key already exists (default if no flag given).
- `--fail-if-exists`: abort with a non-zero exit code if any file already has the key.
- `--overwrite`: replace the existing value if the key is already present.
- `--append`: append `--value` to an existing sequence field; if the key is absent, create it as a single-element array; if the key exists but is not a sequence, error.
- `--skip-if-exists`, `--fail-if-exists`, `--overwrite`, `--append` are mutually exclusive.
- `--dry-run`: print what would be changed without writing.

---

### `substitute`

Replace the value of a key in specific files.

```
yamly substitute <file> [<file>...] --key <key> --new-value <value>
                   [--old-value <value> | --overwrite] [--as-json] [--dry-run]
```

- Requires at least one file argument; errors if none provided.
- Default (no flag): replaces the value in files where the key exists, skips files where it is absent.
- `--old-value <v>`: only replace in files where the current value equals `v` (or, for arrays, contains `v`). Skips all others.
- `--overwrite`: replace if the key exists; **add** the key if it is absent.
- `--old-value` and `--overwrite` are mutually exclusive.
- `--dry-run`: print what would be changed without writing.

---

### `remove`

Remove a key from specific files.

```
yamly remove <file> [<file>...] --key <key> [--value <value>] [--dry-run]
```

- Requires at least one file argument; errors if none provided.
- `--value`: only remove the key if its current value matches (or, for arrays, contains) `v`.
- `--dry-run`: print what would be changed without writing.

---

### `rename`

Rename a key in specific files.

```
yamly rename <file> [<file>...] --old-key <key> --new-key <key> [--dry-run]
```

- Requires at least one file argument; errors if none provided.
- Files where `--old-key` is absent are skipped.
- Errors if `--new-key` already exists in a file (no silent overwrite).
- `--dry-run`: print what would be changed without writing.

---

## Architecture

```
yamly/
├── main.go
├── cmd/
│   ├── root.go        — global --dir flag, cobra root
│   ├── list.go
│   ├── count.go
│   ├── add.go
│   ├── substitute.go
│   ├── remove.go
│   └── rename.go
└── internal/
    ├── frontmatter/
    │   ├── parse.go   — extract YAML block from between --- delimiters
    │   ├── write.go   — serialise modified frontmatter back into file
    │   └── match.go   — value equality + array membership logic
    └── walk/
        └── walk.go    — recursive *.md walker respecting --dir
```

**Dependencies:** `github.com/spf13/cobra`, `gopkg.in/yaml.v3`.

Frontmatter is identified as the YAML block between the opening `---` and closing `---` at the top of the file. Files without frontmatter are skipped silently for read commands; for mutating commands (`add`, `substitute`, `remove`, `rename`) they are skipped with a warning unless `--fail-if-exists` or equivalent makes them an error.

yaml.v3 is used to round-trip frontmatter: parse into `map[string]interface{}`, mutate, re-serialise. Key ordering and comments within the frontmatter block may not be preserved (yaml.v3 limitation).

**`--as-json` output for mutating commands** (`add`, `substitute`, `remove`, `rename`): a JSON array of per-file result objects:
```json
[{"file": "notes/a.md", "action": "modified"}, {"file": "notes/b.md", "action": "skipped"}]
```
Possible `action` values: `modified`, `skipped`, `failed`. Without `--as-json`, one human-readable line per file is printed to stdout.

---

## Verification

```bash
# Build
go build ./...

# Smoke test against a fixture directory
mkdir -p /tmp/yamltest/sub
cat > /tmp/yamltest/a.md <<'EOF'
---
status: draft
tags: [go, cli]
---
Body
EOF
cat > /tmp/yamltest/sub/b.md <<'EOF'
---
status: published
tags: [go]
---
Body
EOF

yamly list --key status --dir /tmp/yamltest
yamly list --key tags --value go --dir /tmp/yamltest
yamly count --key status --dir /tmp/yamltest
yamly add /tmp/yamltest/a.md --key author --value alice
yamly substitute /tmp/yamltest/a.md --key status --new-value published --old-value draft
yamly remove /tmp/yamltest/a.md --key author
yamly rename /tmp/yamltest/a.md --old-key status --new-key state
yamly add /tmp/yamltest/a.md --key tags --value rust --append
yamly substitute /tmp/yamltest/sub/b.md --key missing --new-value x --overwrite  # should add key
yamly add /tmp/yamltest/a.md --key status --value x --dry-run
```
