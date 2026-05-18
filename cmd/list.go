package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/davison/yamlsum/internal/frontmatter"
	"github.com/davison/yamlsum/internal/walk"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Find files containing a frontmatter key",
	RunE:  runList,
}

var listKey, listValue string
var listJSON bool

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listKey, "key", "", "frontmatter key to search for (required)")
	listCmd.Flags().StringVar(&listValue, "value", "", "filter by value (YAML-parsed)")
	listCmd.Flags().BoolVar(&listJSON, "as-json", false, "output as JSON array")
	listCmd.MarkFlagRequired("key")
}

func runList(cmd *cobra.Command, args []string) error {
	var (
		queryVal any
		hasVal   = cmd.Flags().Changed("value")
	)
	if hasVal {
		v, err := frontmatter.ParseValue(listValue)
		if err != nil {
			return fmt.Errorf("--value: %w", err)
		}
		queryVal = v
	}

	files, err := walk.Walk(dir)
	if err != nil {
		return err
	}

	var matches []string
	for _, path := range files {
		f, err := frontmatter.Parse(path)
		if err != nil {
			continue
		}
		fieldVal, ok := f.Data[listKey]
		if !ok {
			continue
		}
		if hasVal && !frontmatter.Matches(fieldVal, queryVal) {
			continue
		}
		matches = append(matches, path)
	}

	if listJSON {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(matches)
	}
	for _, m := range matches {
		fmt.Fprintln(cmd.OutOrStdout(), m)
	}
	return nil
}
