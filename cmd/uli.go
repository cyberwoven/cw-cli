package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	cwutils "cw-cli/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// uliCmd represents the uli command
var uliCmd = &cobra.Command{
	Use:   "uli",
	Short: "Output `drush uli` with a link that works in the desired environment",

	Run: func(cmd *cobra.Command, args []string) {
		isFlaggedTest, _ := cmd.Flags().GetBool("test")

		cwutils.InitViperConfigEnv()

		var vars cwutils.CwVars = cwutils.GetProjectVars()
		cwutils.CheckLocalConfigOverrides(vars.Project_root)

		uid, _ := cmd.Flags().GetInt("uid")

		var LOGIN_URL string
		var stdout []byte
		var err error
		var drushCmd *exec.Cmd

		if isFlaggedTest {
			var SSH_TEST_SERVER string = viper.GetString("CWCLI_SSH_TEST_SERVER")
			var SSH_USER string = viper.GetString("CWCLI_SSH_USER")

			remoteUliCmd := fmt.Sprintf("cd %s && ~/bin/uli", vars.DEFAULT_DIR_FOREST)
			drushCmd = exec.Command("ssh", SSH_USER+"@"+SSH_TEST_SERVER, remoteUliCmd)
		} else {
			// local uli. let's nuke all sessions so we don't get the access denied error
			err = exec.Command("drush", "sqlq", "TRUNCATE SESSIONS").Run()
			if err != nil {
				fmt.Println(err.Error())
			}

			drushCmd = exec.Command("drush", "uli", "--uid", strconv.Itoa(uid), "--uri="+vars.Drupal_site_name+".test", "--no-browser")
		}

		stdout, err = drushCmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		LOGIN_URL = strings.TrimSpace(string(stdout))
		LOGIN_ARRAY := []string{LOGIN_URL, "?destination=admin/reports/status"}
		LOGIN_URL_WITH_REDIRECT := strings.Join(LOGIN_ARRAY, "")

		fmt.Println(LOGIN_URL_WITH_REDIRECT)

		openCmd := exec.Command("open", LOGIN_URL_WITH_REDIRECT)

		err = openCmd.Run()
		if err != nil {
			fmt.Println(err.Error())
		}

	},
}

func init() {
	rootCmd.AddCommand(uliCmd)
	uliCmd.Flags().Int("uid", 1, "Specific Drupal user ID to login as")
	uliCmd.PersistentFlags().BoolP("test", "", false, "Run `drush uli` on the respective remote test environment.")
}
