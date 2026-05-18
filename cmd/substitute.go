package cmd

import (
	"fmt"
	"os"

	"github.com/davison/yamly/internal/frontmatter"
	"github.com/spf13/cobra"
)

var substituteCmd = &cobra.Command{
	Use:   "substitute <file> [<file>...]",
	Short: "Replace the value of a key in specific files",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSubstitute,
}

var subKey, subNewValue, subOldValue string
var subOverwrite, subDryRun, subJSON bool

func init() {
	rootCmd.AddCommand(substituteCmd)
	substituteCmd.Flags().StringVar(&subKey, "key", "", "frontmatter key (required)")
	substituteCmd.Flags().StringVar(&subNewValue, "new-value", "", "replacement value, parsed as YAML (required)")
	substituteCmd.Flags().StringVar(&subOldValue, "old-value", "", "only replace when current value matches this")
	substituteCmd.Flags().BoolVar(&subOverwrite, "overwrite", false, "replace if key exists; add key if absent")
	substituteCmd.Flags().BoolVar(&subDryRun, "dry-run", false, "print changes without writing")
	substituteCmd.Flags().BoolVar(&subJSON, "as-json", false, "output results as JSON")
	substituteCmd.MarkFlagRequired("key")
	substituteCmd.MarkFlagRequired("new-value")
}

func runSubstitute(cmd *cobra.Command, args []string) error {
	hasOldVal := cmd.Flags().Changed("old-value")
	isOverwrite := cmd.Flags().Changed("overwrite")
	isDryRun := cmd.Flags().Changed("dry-run")
	isJSON := cmd.Flags().Changed("as-json")

	if hasOldVal && isOverwrite {
		return fmt.Errorf("--old-value and --overwrite are mutually exclusive")
	}

	newVal, err := frontmatter.ParseValue(subNewValue)
	if err != nil {
		return fmt.Errorf("--new-value: %w", err)
	}

	var oldVal any
	if hasOldVal {
		oldVal, err = frontmatter.ParseValue(subOldValue)
		if err != nil {
			return fmt.Errorf("--old-value: %w", err)
		}
	}

	var results []actionResult
	for _, path := range args {
		f, err := frontmatter.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		fieldVal, exists := f.Data[subKey]

		switch {
		case !exists && isOverwrite:
			f.Data[subKey] = newVal
		case !exists:
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		case hasOldVal && !frontmatter.Matches(fieldVal, oldVal):
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		default:
			f.Data[subKey] = newVal
		}

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
