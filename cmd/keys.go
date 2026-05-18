package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/davison/yamlsum/internal/frontmatter"
	"github.com/davison/yamlsum/internal/walk"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List all frontmatter keys and the number of files containing each",
	RunE:  runKeys,
}

var keysJSON bool

func init() {
	rootCmd.AddCommand(keysCmd)
	keysCmd.Flags().BoolVar(&keysJSON, "as-json", false, "output as JSON object {key: count}")
}

func runKeys(cmd *cobra.Command, args []string) error {
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
		for k := range f.Data {
			counts[k]++
		}
	}

	type entry struct {
		key   string
		count int
	}
	entries := make([]entry, 0, len(counts))
	for k, c := range counts {
		entries = append(entries, entry{k, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].key < entries[j].key
	})

	out := cmd.OutOrStdout()
	if keysJSON {
		m := make(map[string]int, len(entries))
		for _, e := range entries {
			m[e.key] = e.count
		}
		return json.NewEncoder(out).Encode(m)
	}
	for _, e := range entries {
		fmt.Fprintf(out, "%s: %d\n", e.key, e.count)
	}
	return nil
}
