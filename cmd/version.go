package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "dev"
	date    = "dev"
)

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version, commit, and build date",
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "mkcr version %s\ncommit: %s\ndate: %s\n", version, commit, date)
	},
}
