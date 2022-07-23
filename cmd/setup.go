/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup local sandbox",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Setting up...")
		fmt.Println(" ✔ set SITES directory,         save to ~/.cw/config if modified")
		fmt.Println(" ✔ set DEFAULT_TEST_HOSTNAME,   save to ~/.cw/config if modified")
		fmt.Println(" ✔ set DEFAULT_SSH_CA_HOSTNAME, save to ~/.cw/config if modified")
		fmt.Println(" ✔ ensure SAN cert exists")
		fmt.Println(" ✔ ensure MariaDB creds/privs are in place")
		fmt.Println(" ✔ configure & restart apache")
		fmt.Println(" ✔ configure & restart php-fpm")
		fmt.Println(" ✔ configure & restart dnsmasq")
		fmt.Println(" ✔ download latest php tools: composer, drush, drush8, terminus, psysh")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
