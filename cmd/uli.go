package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// uliCmd represents the uli command
var uliCmd = &cobra.Command{
	Use:   "uli",
	Short: "Output `drush uli` with a link that works in sandbox",

	Run: func(cmd *cobra.Command, args []string) {
		uid, _ := cmd.Flags().GetInt("uid")
		sitesDir := os.Getenv("CW_SITES_DIR")

		cwd, _ := os.Getwd()

		if !strings.HasPrefix(cwd, sitesDir) {
			fmt.Println("It doesn't look like you're in a Drupal site.")
			os.Exit(1)
		}

		pathSuffix := strings.Trim(cwd[len(sitesDir):], "/")
		slashIndex := strings.Index(pathSuffix, "/")
		domain := pathSuffix
		if slashIndex != -1 {
			domain = pathSuffix[:strings.Index(pathSuffix, "/")]
		}

		drushCmd := exec.Command("drush", "uli", "--uid", strconv.Itoa(uid), "--no-browser")
		stdout, err := drushCmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			return
		} else {
			LOGIN_URL := strings.TrimSpace(strings.Replace(string(stdout), "default", domain+".test", 1))
			LOGIN_ARRAY := []string{LOGIN_URL, "?destination=admin/reports/status"}
			LOGIN_URL_WITH_REDIRECT := strings.Join(LOGIN_ARRAY, "")

			fmt.Println(LOGIN_URL_WITH_REDIRECT)

			truncateCmd := exec.Command("drush", "sqlq", "\"TRUNCATE SESSIONS\"")
			openCmd := exec.Command("open", LOGIN_URL_WITH_REDIRECT)

			err = truncateCmd.Run()
			if err != nil {
				fmt.Println(err.Error())
			}
			err = openCmd.Run()
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(uliCmd)
	uliCmd.Flags().Int("uid", 1, "Specific Drupal user ID to login as")
}
