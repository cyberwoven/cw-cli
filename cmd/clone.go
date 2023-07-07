package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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

			settingsCopyCmd := exec.Command("cp", "default.settings.local.php", "settings.local.php")
			settingsCopyCmd.Run()

			// pull db
			pullDbCmd.Run(cmd, []string{})

			// logout all users first, so we avoid the access denied error if we're already logged in
			truncateSessionsCmd := exec.Command("drush", "sqlq", "TRUNCATE sessions")
			truncateSessionsCmd.Run()

			LOCAL_URI := fmt.Sprintf("%s.test", REPO_NAME)

			drushUliCmd := exec.Command("drush", "uli", fmt.Sprintf("--uri=%s", LOCAL_URI), "--no-browser")
			LOGIN_URL, err := drushUliCmd.Output()
			if err != nil {
				fmt.Println(string(err.Error()))
				os.Exit(1)
			}

			// open browser and login automagically
			openCmd := exec.Command("open", fmt.Sprintf("%s?destination=admin/reports/status", strings.TrimSpace(string(LOGIN_URL))))
			openCmd.Run()
		}
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
