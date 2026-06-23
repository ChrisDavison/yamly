package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/davison/yamly/internal/frontmatter"
	"github.com/davison/yamly/internal/walk"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Check frontmatter for common problems",
	Long: `lint scans all markdown files under --dir and reports findings for:

  malformed-yaml      YAML that fails to parse
  duplicate-value     a list key containing repeated elements
  empty-value         a key whose value is null, empty string, or empty list
  inconsistent-type   a key that is a scalar in some files and a list in others

With --check-missing also reports:

  missing-frontmatter  file has no --- block
  empty-frontmatter    frontmatter block contains no keys`,
	RunE: runLint,
}

var lintCheckMissing, lintJSON bool

func init() {
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().BoolVar(&lintCheckMissing, "check-missing", false, "also report files with missing or empty frontmatter")
	lintCmd.Flags().BoolVar(&lintJSON, "as-json", false, "output findings as a JSON array")
}

// lintFinding describes a single lint problem. File is empty for cross-file findings.
type lintFinding struct {
	File    string `json:"file"`
	Rule    string `json:"rule"`
	Key     string `json:"key,omitempty"`
	Message string `json:"message"`
}

func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	if arr, ok := v.([]any); ok {
		return len(arr) == 0
	}
	return false
}

func yamlKind(v any) string {
	switch v.(type) {
	case []any:
		return "list"
	case map[string]any:
		return "map"
	default:
		return "scalar"
	}
}

func runLint(cmd *cobra.Command, args []string) error {
	checkMissing := cmd.Flags().Changed("check-missing")
	asJSON := cmd.Flags().Changed("as-json")

	files, err := walk.Walk(dir, excludes)
	if err != nil {
		return err
	}

	var findings []lintFinding
	// keyKinds tracks which yamlKinds have been seen per key, with one example file per kind.
	keyKinds := make(map[string]map[string]string) // key → kind → first file

	for _, path := range files {
		f, err := frontmatter.Parse(path)
		if err != nil {
			if errors.Is(err, frontmatter.ErrNoFrontmatter) {
				if checkMissing {
					findings = append(findings, lintFinding{
						File:    path,
						Rule:    "missing-frontmatter",
						Message: "no frontmatter block found",
					})
				}
			} else {
				findings = append(findings, lintFinding{
					File:    path,
					Rule:    "malformed-yaml",
					Message: err.Error(),
				})
			}
			continue
		}

		if len(f.Data) == 0 {
			if checkMissing {
				findings = append(findings, lintFinding{
					File:    path,
					Rule:    "empty-frontmatter",
					Message: "frontmatter block contains no keys",
				})
			}
			continue
		}

		for k, v := range f.Data {
			if isEmptyValue(v) {
				findings = append(findings, lintFinding{
					File:    path,
					Rule:    "empty-value",
					Key:     k,
					Message: fmt.Sprintf("key %q has an empty or null value", k),
				})
			}

			if arr, ok := v.([]any); ok {
				seen := make(map[string]int)
				for _, item := range arr {
					seen[fmt.Sprintf("%v", item)]++
				}
				// Emit one finding per distinct duplicate, sorted for determinism.
				var dupes []string
				for val, count := range seen {
					if count > 1 {
						dupes = append(dupes, fmt.Sprintf("%q (×%d)", val, count))
					}
				}
				sort.Strings(dupes)
				for _, d := range dupes {
					findings = append(findings, lintFinding{
						File:    path,
						Rule:    "duplicate-value",
						Key:     k,
						Message: fmt.Sprintf("key %q contains duplicate value %s", k, d),
					})
				}
			}

			// Record kind for cross-file consistency check.
			kind := yamlKind(v)
			if keyKinds[k] == nil {
				keyKinds[k] = make(map[string]string)
			}
			if _, exists := keyKinds[k][kind]; !exists {
				keyKinds[k][kind] = path
			}
		}
	}

	// Emit one inconsistent-type finding per key that has mixed kinds across files.
	type mixedKey struct {
		key   string
		kinds map[string]string
	}
	var mixed []mixedKey
	for k, kinds := range keyKinds {
		if len(kinds) > 1 {
			mixed = append(mixed, mixedKey{k, kinds})
		}
	}
	sort.Slice(mixed, func(i, j int) bool { return mixed[i].key < mixed[j].key })

	for _, m := range mixed {
		parts := make([]string, 0, len(m.kinds))
		for kind, file := range m.kinds {
			parts = append(parts, fmt.Sprintf("%s in %s", kind, file))
		}
		sort.Strings(parts)
		findings = append(findings, lintFinding{
			File:    "",
			Rule:    "inconsistent-type",
			Key:     m.key,
			Message: fmt.Sprintf("key %q has mixed types: %s", m.key, strings.Join(parts, ", ")),
		})
	}

	// Sort: per-file findings first (by file, rule, key); cross-file last (by rule, key).
	sort.SliceStable(findings, func(i, j int) bool {
		fi, fj := findings[i], findings[j]
		if fi.File == "" && fj.File != "" {
			return false
		}
		if fi.File != "" && fj.File == "" {
			return true
		}
		if fi.File != fj.File {
			return fi.File < fj.File
		}
		if fi.Rule != fj.Rule {
			return fi.Rule < fj.Rule
		}
		return fi.Key < fj.Key
	})

	out := cmd.OutOrStdout()
	if asJSON {
		if findings == nil {
			findings = []lintFinding{}
		}
		return json.NewEncoder(out).Encode(findings)
	}
	for _, f := range findings {
		file := f.File
		if file == "" {
			file = "(across files)"
		}
		fmt.Fprintf(out, "%s: [%s] %s\n", file, f.Rule, f.Message)
	}
	return nil
}
