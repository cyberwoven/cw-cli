package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone [repo]",
	Short: "Clone a git repository from default git workspace",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a repository to clone.")
			os.Exit(1)
		}

		REPO_NAME := args[0]

		fmt.Printf("[%s] Cloning site...\n", REPO_NAME)
		repo_url := fmt.Sprintf("%s@%s:%s/%s",
			ctx.GIT_DEFAULT_USER,
			ctx.GIT_DEFAULT_DOMAIN,
			ctx.GIT_DEFAULT_WORKSPACE,
			REPO_NAME)
		project_root := fmt.Sprintf("%s/%s", ctx.SITES_DIR, REPO_NAME)
		cloneCmd := exec.Command("git", "clone", repo_url, project_root, "--progress", "--verbose")
		stdout, _ := cloneCmd.StdoutPipe()
		stderr, _ := cloneCmd.StderrPipe()

		_ = cloneCmd.Start()
		scanner := bufio.NewScanner(io.MultiReader(stderr, stdout))
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
		_ = cloneCmd.Wait()

		// only bother with settings.local.php and syncing if it's a drupal site.
		// TODO: make this work for WP sites
		// TODO: make this work for Pantheon sites
		if _, err := os.Stat(project_root + "/pub/sites/default"); !os.IsNotExist(err) {
			os.Chdir(project_root + "/pub/sites/default/")

			settings_local := ""
			if _, err := os.Stat("default.settings.local.php"); !os.IsNotExist(err) {
				settings_local = "default.settings.local.php"
			} else if _, err := os.Stat("default.sandbox.settings.local.php"); !os.IsNotExist(err) {
				settings_local = "default.sandbox.settings.local.php"
			}

			settingsCopyCmd := exec.Command("cp", settings_local, "settings.local.php")
			settingsCopyCmd.Run()

			// pull db
			newSitePullDbCmd := exec.Command("cw", "pull", "db")
			err := newSitePullDbCmd.Run()
			if err != nil {
				fmt.Println(string(err.Error()))
				os.Exit(1)
			}

			exec.Command("cw", "uli").Run()
		}
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
