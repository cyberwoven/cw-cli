/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "cw",
	Short:   "Cyberwoven local web site development tool",
	Version: "1.2.8",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cw-cli.yaml)")

	rootCmd.PersistentFlags().BoolP("verbose", "V", false, "Enables verbose logging for developer curiosity.")
	rootCmd.PersistentFlags().BoolP("force", "f", false, "Force stuff to happen.")
	rootCmd.PersistentFlags().BoolP("fast", "F", false, "Use experimental fast versions of commands, where available.")
	rootCmd.PersistentFlags().BoolP("slow", "S", false, "Use slower (more stable) pull comand.")
}
