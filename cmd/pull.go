package cmd

import (
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:     "pull",
	Aliases: []string{"sync"},
	Short:   "Pull the current site's database and/or files",
	Run: func(cmd *cobra.Command, args []string) {
		pullDbCmd.Run(cmd, []string{})
		pullFilesCmd.Run(cmd, []string{})
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
