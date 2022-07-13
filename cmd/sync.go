package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync database and/or files from remote test server into sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		sitesDir := os.Getenv("CW_SITES_DIR")

		cwd, _ := os.Getwd()

		if !strings.HasPrefix(cwd, sitesDir) || cwd == sitesDir {
			fmt.Println("It doesn't look like you're in a Drupal site.")
			os.Exit(1)
		}

		pathSuffix := strings.Trim(cwd[len(sitesDir):], "/")

		slashIndex := strings.Index(pathSuffix, "/")

		domain := pathSuffix
		if slashIndex != -1 {
			domain = pathSuffix[:strings.Index(pathSuffix, "/")]
		}

		fmt.Println("Syncing database and/or files for", domain)

		if args[0] == "db" {
			cloneCmd := exec.Command("bash", sitesDir+"/cw/bin/cw.sh", "sync", "db")
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			cloneCmd.Start()
			cloneCmd.Wait()
		} else if args[0] == "files" {
			cloneCmd := exec.Command("bash", sitesDir+"/cw/bin/cw.sh", "sync", "files")
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			cloneCmd.Start()
			cloneCmd.Wait()
		} else {
			cloneCmd := exec.Command("bash", sitesDir+"/cw/bin/cw.sh", "sync")
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			cloneCmd.Start()
			cloneCmd.Wait()
		}

		// 1. parse drush status to get dbname, root dir, drupal version
		// 2. determine if the site is pantheon
		// 3. determine if we're syncing db, files, or both
		// 4. call db sync func (if necessary)
		// 5. call file sync func (if necessary)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
