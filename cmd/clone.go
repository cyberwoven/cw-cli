package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
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

		sitesDir := os.Getenv("CW_SITES_DIR")
		repo := args[0]
		repoDir := sitesDir + "/" + repo

		fmt.Println("Cloning " + repo + "...")
		cloneCmd := exec.Command("git", "clone", "git@bitbucket.org:cyberwoven/"+repo, repoDir)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		cloneCmd.Start()
		cloneCmd.Wait()

		// only bother with settings.local.php and syncing if it's a drupal site.
		// TODO: make this work for WP sites
		if _, err := os.Stat(repoDir + "/pub/sites/default"); !os.IsNotExist(err) {
			os.Chdir(repoDir + "/pub/sites/default/")

			settingsCopyCmd := exec.Command("cp", "default.settings.local.php", "settings.local.php")
			settingsCopyCmd.Run()

			fmt.Println("Syncing " + repo + "...")
			syncCmd := exec.Command("cw", "sync", "db")
			syncCmd.Stdout = os.Stdout
			syncCmd.Stderr = os.Stderr
			syncCmd.Start()
			syncCmd.Wait()
		}

	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
