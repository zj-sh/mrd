package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		color.Blue("v1.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
