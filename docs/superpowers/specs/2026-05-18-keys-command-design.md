---
title: keys command design
date: 2026-05-18
---

# `keys` Command

## Purpose

Show all frontmatter keys found across the target directory, with a count of how many files contain each key.

## Usage

```
yamlsum keys [--dir <path>] [--as-json]
```

Flags:
- `--dir` — inherited from root, defaults to `.`
- `--as-json` — emit a JSON object `{"key": count}` instead of plain text

## Output

**Text (default):** one `key: N` line per key, sorted by count descending, then alphabetically.

```
title: 42
date: 40
tags: 31
draft: 12
```

**JSON (`--as-json`):** a single JSON object mapping key names to counts.

```json
{"date":40,"draft":12,"tags":31,"title":42}
```

## Implementation

Single file `cmd/keys.go`, mirroring the structure of `cmd/count.go`:

1. Walk files via `walk.Walk(dir)`
2. Parse frontmatter; skip files without it
3. For each key in `f.Data`, increment `counts[key]`
4. Sort entries by count desc, then key asc
5. Output text or JSON via `cmd.OutOrStdout()`

No new internal packages needed.
