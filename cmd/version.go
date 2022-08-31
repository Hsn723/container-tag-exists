package cmd

import (
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version",
		Args:  cobra.NoArgs,
		Run:   runVersion,
	}
	version string
	commit  string
	date    string
	builtBy string
)

func runVersion(cmd *cobra.Command, _ []string) {
	cmd.Printf("container-tag-exists version: %s, commit: %s, date: %s, builtBy: %s\n", version, commit, date, builtBy)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
