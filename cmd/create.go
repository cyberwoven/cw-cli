/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new site from boilerplate",
	Run: func(cmd *cobra.Command, args []string) {

		siteTypePrompt := promptui.Select{
			Label: "Site type",
			Items: []string{"Drupal", "Netlify", "Wordpress"},
		}

		_, siteType, err := siteTypePrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		domainValidate := func(domain string) error {
			RegExp := regexp.MustCompile(`^([0-9a-zA-Z\-\.]+)$`)

			if !RegExp.MatchString(domain) {
				return errors.New("domain must contain only letters, numbers, dots, and dashes")
			}

			domain = strings.ToLower(domain)
			if _, err := os.Stat(ctx.SITES_DIR + "/" + domain); err == nil {
				return errors.New("specified domain already exists: " + ctx.SITES_DIR + "/" + domain)
			}

			return nil
		}

		domainPrompt := promptui.Prompt{
			Label:    "Domain name",
			Validate: domainValidate,
		}

		domain, err := domainPrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if siteType == "Netlify" {
			createNetlify(domain)
			return
		}

		databaseValidate := func(domain string) error {
			RegExp := regexp.MustCompile(`^([0-9a-zA-Z]+)$`)

			if !RegExp.MatchString(domain) {
				return errors.New("database must contain only letters and numbers")
			}

			return nil
		}

		databasePrompt := promptui.Prompt{
			Label:    "Database name",
			Validate: databaseValidate,
		}

		database, err := databasePrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		themeValidate := func(domain string) error {
			RegExp := regexp.MustCompile(`^([0-9a-zA-Z]+)$`)

			if !RegExp.MatchString(domain) {
				return errors.New("theme must contain only letters and numbers")
			}

			return nil
		}

		themePrompt := promptui.Prompt{
			Label:    "Theme machine name",
			Validate: themeValidate,
		}

		theme, err := themePrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		domain = strings.ToLower(domain)
		database = strings.ToLower(database)
		theme = strings.ToLower(theme)

		if siteType == "Drupal" {
			createDrupal(domain, database, theme)
		} else if siteType == "Wordpress" {
			createWordpress(domain, database, theme)
		}

	},
}

func createDrupal(domain string, database string, theme string) {
	fmt.Print("Creating a Drupal site, this will take a minute or two...")

	composerCmd := exec.Command(
		"composer",
		"create-project",
		"cyberwoven/drupal",
		ctx.SITES_DIR+"/"+domain,
	)

	composerCmd.Env = os.Environ()
	composerCmd.Env = append(composerCmd.Env, "DRUPAL_DATABASE_NAME="+database)
	composerCmd.Env = append(composerCmd.Env, "DRUPAL_THEME_NAME="+theme)

	stdout, _ := composerCmd.StdoutPipe()
	stderr, _ := composerCmd.StderrPipe()

	_ = composerCmd.Start()

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	_ = composerCmd.Wait()
}

func createNetlify(domain string) {
	fmt.Print("Creating a Netlify site. (Not implemented yet)")
}

func createWordpress(domain string, database string, theme string) {
	fmt.Print("Creating a Wordpress site. (Not implemented yet)")
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
