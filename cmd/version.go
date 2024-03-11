package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type versionOpts struct {
	*rootOpts
}

type versionCmd struct {
	*versionOpts
	cmd *cobra.Command
}

func newVersionCmd(opts *rootOpts) *versionCmd {
	c := &versionCmd{
		versionOpts: &versionOpts{rootOpts: opts},
	}
	c.cmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			color.Blue(c.version)
		},
	}
	return c
}
