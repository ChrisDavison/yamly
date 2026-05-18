package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

type actionResult struct {
	File   string `json:"file"`
	Action string `json:"action"` // "modified", "skipped", "failed"
}

func printResults(results []actionResult, asJSON bool) error {
	if asJSON {
		return json.NewEncoder(os.Stdout).Encode(results)
	}
	for _, r := range results {
		fmt.Printf("%s: %s\n", r.Action, r.File)
	}
	return nil
}
