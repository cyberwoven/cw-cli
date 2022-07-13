/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// phpCmd represents the php command
var phpCmd = &cobra.Command{
	Use:   "php",
	Short: "Switch to a specific version of php-cli",

	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			fmt.Println("No new PHP version specified, sticking with current version.")
			os.Exit(0)
		}

		newVersion := args[0]

		currentVersionCmd := exec.Command("php", "-v")
		versionOutput, err := currentVersionCmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		currentVersionOutput := string(versionOutput)
		currentVersion := currentVersionOutput[4:7]

		if newVersion == currentVersion {
			fmt.Println("You're already using PHP", newVersion)
			os.Exit(0)
		}

		brewPrefixCmd := exec.Command("brew", "--prefix")
		prefixOutput, err := brewPrefixCmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		brewPrefixPath := strings.TrimSpace(string(prefixOutput))

		phpVersions, err := os.ReadDir(brewPrefixPath + "/etc/php")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		validVersion := false
		for _, dir := range phpVersions {
			if newVersion == dir.Name() {
				validVersion = true
			}
		}

		if validVersion {
			fmt.Println("Switching to PHP verion", newVersion)

			unlinkCmd := exec.Command("brew", "unlink", "php@"+currentVersion)
			unlinkCmd.Run()

			linkCmd := exec.Command("brew", "link", "php@"+newVersion, "--force", "--overwrite")
			linkCmd.Run()

		} else {
			fmt.Println("PHP version", newVersion, "is not available.")
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(phpCmd)
}
