package cmd

import (
	"fmt"
	"os"

	"github.com/davison/yamly/internal/frontmatter"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <file> [<file>...]",
	Short: "Rename a frontmatter key in specific files",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRename,
}

var renOldKey, renNewKey string
var renDryRun, renJSON bool

func init() {
	rootCmd.AddCommand(renameCmd)
	renameCmd.Flags().StringVar(&renOldKey, "old-key", "", "key to rename (required)")
	renameCmd.Flags().StringVar(&renNewKey, "new-key", "", "new key name (required)")
	renameCmd.Flags().BoolVar(&renDryRun, "dry-run", false, "print changes without writing")
	renameCmd.Flags().BoolVar(&renJSON, "as-json", false, "output results as JSON")
	renameCmd.MarkFlagRequired("old-key")
	renameCmd.MarkFlagRequired("new-key")
}

func runRename(cmd *cobra.Command, args []string) error {
	isDryRun := cmd.Flags().Changed("dry-run")
	isJSON := cmd.Flags().Changed("as-json")

	var results []actionResult
	for _, path := range args {
		f, err := frontmatter.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		val, exists := f.Data[renOldKey]
		if !exists {
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}
		if _, collision := f.Data[renNewKey]; collision {
			fmt.Fprintf(os.Stderr, "fail %s: key %q already exists\n", path, renNewKey)
			results = append(results, actionResult{File: path, Action: "failed"})
			continue
		}

		f.Data[renNewKey] = val
		delete(f.Data, renOldKey)

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
