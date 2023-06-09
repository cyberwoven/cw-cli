/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	cwutils "cw-cli/utils"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Pretty print the relevant vars of the current site",
	Run: func(cmd *cobra.Command, args []string) {
		cwutils.PrettyPrint(cwutils.GetContext())
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
