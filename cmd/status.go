/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	cwutils "cw-cli/utils"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Pretty print the relevant vars of the current site",
	Run: func(cmd *cobra.Command, args []string) {

		cwVarsJson, err := json.MarshalIndent(cwutils.GetProjectVars(), "", "  ")
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Printf("ProjectVars %s\n", string(cwVarsJson))
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
