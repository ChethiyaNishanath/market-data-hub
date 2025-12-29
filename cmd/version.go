package cmd

import (
	"fmt"

	"github.com/ChethiyaNishanath/market-data-hub/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nCommit: %s\nBuilt: %s\n",
			version.Version,
			version.Commit,
			version.Date,
		)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
