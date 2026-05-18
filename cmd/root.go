package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

var dir string

var rootCmd = &cobra.Command{
	Use:   "yamlsum",
	Short: "Inspect and edit YAML frontmatter in markdown files",
}

func Execute() error {
	return rootCmd.Execute()
}

func SetOut(w io.Writer) {
	rootCmd.SetOut(w)
}

func SetArgs(args []string) {
	rootCmd.SetArgs(args)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dir, "dir", ".", "root directory to search (default: current directory)")
}
