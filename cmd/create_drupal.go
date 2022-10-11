/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// drupalCmd represents the drupal command
var drupalCmd = &cobra.Command{
	Use:   "drupal",
	Short: "Create a Drupal 9 site",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("drupal called!")
		// 1. clone drupal-starter repo
		for i, s := range args {
			fmt.Println(i, s)
		}
		// git clone git@bitbucket.org:cyberwoven/drupal-project-starter
	},
}

func init() {
	createCmd.AddCommand(drupalCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// drupalCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// drupalCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
