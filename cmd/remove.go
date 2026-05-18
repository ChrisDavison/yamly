package cmd

import (
	"fmt"
	"os"

	"github.com/davison/yamlsum/internal/frontmatter"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <file> [<file>...]",
	Short: "Remove a key from specific files",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

var rmKey, rmValue string
var rmDryRun, rmJSON bool

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringVar(&rmKey, "key", "", "frontmatter key to remove (required)")
	removeCmd.Flags().StringVar(&rmValue, "value", "", "only remove if current value matches this")
	removeCmd.Flags().BoolVar(&rmDryRun, "dry-run", false, "print changes without writing")
	removeCmd.Flags().BoolVar(&rmJSON, "as-json", false, "output results as JSON")
	removeCmd.MarkFlagRequired("key")
}

func runRemove(cmd *cobra.Command, args []string) error {
	hasVal := cmd.Flags().Changed("value")
	isDryRun := cmd.Flags().Changed("dry-run")
	isJSON := cmd.Flags().Changed("as-json")

	var queryVal any
	if hasVal {
		v, err := frontmatter.ParseValue(rmValue)
		if err != nil {
			return fmt.Errorf("--value: %w", err)
		}
		queryVal = v
	}

	var results []actionResult
	for _, path := range args {
		f, err := frontmatter.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		fieldVal, exists := f.Data[rmKey]
		if !exists || (hasVal && !frontmatter.Matches(fieldVal, queryVal)) {
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		delete(f.Data, rmKey)

		if !isDryRun {
			if err := f.WriteFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "fail %s: %v\n", path, err)
				results = append(results, actionResult{File: path, Action: "failed"})
				continue
			}
		}
		results = append(results, actionResult{File: path, Action: "modified"})
	}
	return printResults(results, isJSON)
}
