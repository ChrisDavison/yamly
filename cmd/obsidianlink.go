package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/davison/yamly/internal/frontmatter"
	"github.com/spf13/cobra"
)

var obsidianLinkCmd = &cobra.Command{
	Use:   "obsidianlink <file> [<file>...]",
	Short: "Wrap frontmatter values as Obsidian links",
	Long: `obsidianlink wraps string values at a given key as Obsidian wiki-links ([[value]]).
Values already wrapped in [[ ]] are left unchanged. If the key holds a list,
each string element is wrapped individually.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runObsidianLink,
}

var obsKey string
var obsDryRun, obsJSON bool

func init() {
	rootCmd.AddCommand(obsidianLinkCmd)
	obsidianLinkCmd.Flags().StringVar(&obsKey, "key", "", "frontmatter key to transform (required)")
	obsidianLinkCmd.Flags().BoolVar(&obsDryRun, "dry-run", false, "print changes without writing")
	obsidianLinkCmd.Flags().BoolVar(&obsJSON, "as-json", false, "output results as JSON")
	obsidianLinkCmd.MarkFlagRequired("key")
}

func wrapLink(s string) string {
	if strings.HasPrefix(s, "[[") && strings.HasSuffix(s, "]]") {
		return s
	}
	return "[[" + s + "]]"
}

func runObsidianLink(cmd *cobra.Command, args []string) error {
	dryRun := cmd.Flags().Changed("dry-run")
	asJSON := cmd.Flags().Changed("as-json")

	var results []actionResult
	for _, path := range args {
		f, err := frontmatter.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		val, exists := f.Data[obsKey]
		if !exists {
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
		}

		switch v := val.(type) {
		case string:
			f.Data[obsKey] = wrapLink(v)
		case []any:
			wrapped := make([]any, len(v))
			for i, elem := range v {
				if s, ok := elem.(string); ok {
					wrapped[i] = wrapLink(s)
				} else {
					wrapped[i] = elem
				}
			}
			f.Data[obsKey] = wrapped
		default:
			results = append(results, actionResult{File: path, Action: "skipped"})
			continue
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
