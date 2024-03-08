package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "mrd",
	Short:         "This is a command line tool for mored.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&dry, "dry-run", "", false, "dry run.")
	if err := rootCmd.Execute(); err != nil {
		IfErrExit(err)
	}
}
