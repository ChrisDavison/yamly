package cmd

import (
	"fmt"
	"os"

	"github.com/davison/yamly/internal/frontmatter"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge <file> [<file>...]",
	Short: "Merge the value of a source key into a destination key",
	Long: `Merge combines the value of a source key into a destination key,
turning the destination into a list if it wasn't already. The source
key is deleted afterwards unless --keep-source is specified.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMerge,
}

var mergeSrcKey, mergeDstKey string
var mergeKeepSource, mergeDryRun, mergeJSON bool

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringVarP(&mergeSrcKey, "source", "s", "", "source key to merge from (required)")
	mergeCmd.Flags().StringVarP(&mergeDstKey, "dest", "d", "", "destination key to merge into (required)")
	mergeCmd.Flags().BoolVar(&mergeKeepSource, "keep-source", false, "keep the source key after merging")
	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "print changes without writing")
	mergeCmd.Flags().BoolVar(&mergeJSON, "as-json", false, "output results as JSON")
	mergeCmd.MarkFlagRequired("source")
	mergeCmd.MarkFlagRequired("dest")
}

func toList(v any) []any {
	if list, ok := v.([]any); ok {
		return list
	}
	return []any{v}
}

func runMerge(cmd *cobra.Command, args []string) error {
	keepSource := cmd.Flags().Changed("keep-source")
	dryRun := cmd.Flags().Changed("dry-run")
	asJSON := cmd.Flags().Changed("as-json")

	if mergeSrcKey == mergeDstKey {
		return fmt.Errorf("source and destination keys must be different")
	}

	var results []actionResult
	for _, path := range args {
		f, err := frontmatter.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		srcVal, srcExists := f.Data[mergeSrcKey]
		if !srcExists {
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		dstVal, dstExists := f.Data[mergeDstKey]

		var merged []any
		if dstExists {
			merged = toList(dstVal)
		}
		merged = append(merged, toList(srcVal)...)

		f.Data[mergeDstKey] = merged
		if !keepSource {
			delete(f.Data, mergeSrcKey)
		}

		if !dryRun {
			if err := f.WriteFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "fail %s: %v\n", path, err)
				results = append(results, actionResult{File: path, Action: "failed"})
				continue
			}
		}
		results = append(results, actionResult{File: path, Action: "modified"})
	}
	return printResults(results, asJSON)
}
