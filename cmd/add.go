package cmd

import (
	"fmt"
	"os"

	"github.com/davison/yamlsum/internal/frontmatter"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file> [<file>...]",
	Short: "Add a key/value pair to specific files",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

var (
	addKey, addValue                                         string
	addSkipIfExists, addFailIfExists, addOverwrite, addAppend bool
	addDryRun, addJSON                                       bool
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&addKey, "key", "", "frontmatter key (required)")
	addCmd.Flags().StringVar(&addValue, "value", "", "value to set, parsed as YAML (required)")
	addCmd.Flags().BoolVar(&addSkipIfExists, "skip-if-exists", false, "skip file if key already exists (default behaviour)")
	addCmd.Flags().BoolVar(&addFailIfExists, "fail-if-exists", false, "error if key already exists in any file")
	addCmd.Flags().BoolVar(&addOverwrite, "overwrite", false, "overwrite existing value")
	addCmd.Flags().BoolVar(&addAppend, "append", false, "append value to existing array; creates array if key absent")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "print changes without writing")
	addCmd.Flags().BoolVar(&addJSON, "as-json", false, "output results as JSON")
	addCmd.MarkFlagRequired("key")
	addCmd.MarkFlagRequired("value")
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Re-read booleans from Changed state to avoid bleed between test runs.
	addSkipIfExists = cmd.Flags().Changed("skip-if-exists")
	addFailIfExists = cmd.Flags().Changed("fail-if-exists")
	addOverwrite = cmd.Flags().Changed("overwrite")
	addAppend = cmd.Flags().Changed("append")
	addDryRun = cmd.Flags().Changed("dry-run")
	addJSON = cmd.Flags().Changed("as-json")

	exclusiveCount := 0
	for _, b := range []bool{addSkipIfExists, addFailIfExists, addOverwrite, addAppend} {
		if b {
			exclusiveCount++
		}
	}
	if exclusiveCount > 1 {
		return fmt.Errorf("--skip-if-exists, --fail-if-exists, --overwrite, and --append are mutually exclusive")
	}

	parsedVal, err := frontmatter.ParseValue(addValue)
	if err != nil {
		return fmt.Errorf("--value: %w", err)
	}

	var results []actionResult
	for _, path := range args {
		f, err := frontmatter.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		_, exists := f.Data[addKey]

		switch {
		case addAppend:
			if !exists {
				f.Data[addKey] = []any{parsedVal}
			} else {
				arr, ok := f.Data[addKey].([]any)
				if !ok {
					fmt.Fprintf(os.Stderr, "fail %s: --append requires existing array value for key %q\n", path, addKey)
					results = append(results, actionResult{File: path, Action: "failed"})
					continue
				}
				f.Data[addKey] = append(arr, parsedVal)
			}
		case exists && addFailIfExists:
			return fmt.Errorf("%s: key %q already exists", path, addKey)
		case exists && !addOverwrite:
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		default:
			f.Data[addKey] = parsedVal
		}

		if !addDryRun {
			if err := f.WriteFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "fail %s: %v\n", path, err)
				results = append(results, actionResult{File: path, Action: "failed"})
				continue
			}
		}
		results = append(results, actionResult{File: path, Action: "modified"})
	}
	return printResults(results, addJSON)
}
