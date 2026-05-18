package cmd

import (
	"io"

	flag "github.com/spf13/pflag"
	"github.com/spf13/cobra"
)

var dir string
var excludes []string

var rootCmd = &cobra.Command{
	Use:   "yamly",
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
	// Reset "changed" state on all subcommand flags so repeated Execute()
	// calls in tests don't bleed flag values across runs.
	for _, sub := range rootCmd.Commands() {
		sub.Flags().VisitAll(func(f *flag.Flag) { f.Changed = false })
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dir, "dir", ".", "root directory to search (default: current directory)")
	rootCmd.PersistentFlags().StringArrayVar(&excludes, "exclude", nil, "directory name to exclude (repeatable)")
}
