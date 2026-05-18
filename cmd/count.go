package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/davison/yamly/internal/frontmatter"
	"github.com/davison/yamly/internal/walk"
	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Count occurrences of each value for a frontmatter key",
	RunE:  runCount,
}

var countKey string
var countJSON bool

func init() {
	rootCmd.AddCommand(countCmd)
	countCmd.Flags().StringVar(&countKey, "key", "", "frontmatter key to count values for (required)")
	countCmd.Flags().BoolVar(&countJSON, "as-json", false, "output as JSON object {value: count}")
	countCmd.MarkFlagRequired("key")
}

func runCount(cmd *cobra.Command, args []string) error {
	files, err := walk.Walk(dir)
	if err != nil {
		return err
	}

	counts := make(map[string]int)
	for _, path := range files {
		f, err := frontmatter.Parse(path)
		if err != nil {
			continue
		}
		fieldVal, ok := f.Data[countKey]
		if !ok {
			continue
		}
		counts[fmt.Sprintf("%v", fieldVal)]++
	}

	type entry struct {
		value string
		count int
	}
	entries := make([]entry, 0, len(counts))
	for v, c := range counts {
		entries = append(entries, entry{v, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].value < entries[j].value
	})

	out := cmd.OutOrStdout()
	if countJSON {
		m := make(map[string]int, len(entries))
		for _, e := range entries {
			m[e.value] = e.count
		}
		return json.NewEncoder(out).Encode(m)
	}
	for _, e := range entries {
		fmt.Fprintf(out, "%s: %d\n", e.value, e.count)
	}
	return nil
}
