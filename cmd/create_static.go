/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// staticCmd represents the static command
var staticCmd = &cobra.Command{
	Use:   "static",
	Short: "Create a static site",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("static called")
	},
}

func init() {
	createCmd.AddCommand(staticCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// staticCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// staticCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
