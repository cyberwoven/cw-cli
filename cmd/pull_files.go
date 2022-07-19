/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// pullFilesCmd represents the pullFiles command
var pullFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Pull files from test to sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pullFiles called")
	},
}

func init() {
	pullCmd.AddCommand(pullFilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullFilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullFilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
