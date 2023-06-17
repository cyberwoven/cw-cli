/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure cw cli",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configuring...")
		fmt.Println(" ✔ set SITES directory,         save to ~/.cw/config if modified")
		fmt.Println(" ✔ set DEFAULT_TEST_HOSTNAME,   save to ~/.cw/config if modified")
		fmt.Println(" ✔ set DEFAULT_SSH_CA_HOSTNAME, save to ~/.cw/config if modified")
		fmt.Println(" ✔ ensure SAN cert exists")
		fmt.Println(" ✔ ensure MariaDB creds/privs are in place")
		fmt.Println(" ✔ configure & restart apache")
		fmt.Println(" ✔ configure & restart php-fpm")
		fmt.Println(" ✔ configure & restart dnsmasq")
		fmt.Println(" ✔ download latest php tools: composer, drush, drush8, terminus, psysh")

		// check if there is somethinig to read on STDIN
		// stat, _ := os.Stdin.Stat()
		// if (stat.Mode() & os.ModeCharDevice) == 0 {
		// 	var stdin []byte
		// 	scanner := bufio.NewScanner(os.Stdin)
		// 	for scanner.Scan() {
		// 		stdin = append(stdin, scanner.Bytes()...)
		// 	}
		// 	if err := scanner.Err(); err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	fmt.Printf("stdin = %s\n", stdin)
		// } else {
		// 	fmt.Println("Enter your name")

		// 	var name string
		// 	fmt.Scanf("%s", &name)
		// 	fmt.Printf("name = %s\n", name)
		// }
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
