package cmd

import (
	"bufio"
	cwutils "cw-cli/utils"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone [repo]",
	Short: "Clone a git repository from Cyberwoven's Bitbucket workspace",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a repository to clone.")
			os.Exit(1)
		}

		USER_HOME_DIRECTORY, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(string(err.Error()))
			os.Exit(1)
		}

		cwutils.InitViperConfigEnv()
		REPO_NAME := args[0]
		SITES_DIRECTORY := fmt.Sprintf("%s/%s", USER_HOME_DIRECTORY, viper.GetString("CWCLI_SITES_DIR"))
		REPO_DIRECTORY := SITES_DIRECTORY + "/" + REPO_NAME
		GIT_DOMAIN := viper.GetString("CWCLI_GIT_DOMAIN")
		GIT_USER := viper.GetString("CWCLI_GIT_USER")

		// fmt.Println(SITES_DIRECTORY)

		fmt.Printf("[%s] Cloning site...", REPO_NAME)
		cloneCmd := exec.Command("git", "clone", fmt.Sprintf("git@%s:%s/%s", GIT_DOMAIN, GIT_USER, REPO_NAME), REPO_DIRECTORY)
		stdout, _ := cloneCmd.StdoutPipe()
		stderr, _ := cloneCmd.StderrPipe()

		_ = cloneCmd.Start()
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
		_ = cloneCmd.Wait()

		// only bother with settings.local.php and syncing if it's a drupal site.
		// TODO: make this work for WP sites
		// TODO: make this work for Pantheon sites
		if _, err := os.Stat(REPO_DIRECTORY + "/pub/sites/default"); !os.IsNotExist(err) {
			os.Chdir(REPO_DIRECTORY + "/pub/sites/default/")

			settingsCopyCmd := exec.Command("cp", "default.settings.local.php", "settings.local.php")
			settingsCopyCmd.Run()

			// pull db
			pullDbCmd.Run(cmd, []string{})
		}
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
