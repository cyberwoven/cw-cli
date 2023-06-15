/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the current site's database and/or files to the test server",

	Run: func(cmd *cobra.Command, args []string) {
		pushDbCmd.Run(cmd, []string{})
		// pushFilesCmd.Run(cmd, []string{})
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
