/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// wordpressCmd represents the wordpress command
var wordpressCmd = &cobra.Command{
	Use:   "wordpress",
	Short: "Create a wordpress site",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wordpress called")
	},
}

func init() {
	createCmd.AddCommand(wordpressCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wordpressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wordpressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
