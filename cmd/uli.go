package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var uliCmd = &cobra.Command{
	Use:   "uli",
	Short: "Output `drush uli` with a link that works in the desired environment",

	Run: func(cmd *cobra.Command, args []string) {
		isFlaggedTest, _ := cmd.Flags().GetBool("test")
		uid, _ := cmd.Flags().GetInt("uid")

		var url string

		if isFlaggedTest {
			url = uliGenerateTestLink()
		} else {
			url = uliGenerateLink(ctx.PROJECT_ROOT, ctx.SITE_NAME+".test", uid)
		}

		uliOpenLink(url)
	},
}

func init() {
	rootCmd.AddCommand(uliCmd)
	uliCmd.Flags().Int("uid", 1, "Specific Drupal user ID to login as")
	uliCmd.PersistentFlags().BoolP("test", "", false, "Run `drush uli` on the respective remote test environment.")
}

func uliOpenLink(url string) {
	url += "?destination=admin/reports/status"

	fmt.Println(url)

	err := exec.Command("open", url).Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func uliGenerateTestLink() string {
	remoteUliCmd := fmt.Sprintf("cd %s && ~/bin/uli", ctx.DRUPAL_DEFAULT_DIR_REMOTE)
	remoteHost := fmt.Sprintf("%s@%s", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST)
	drushCmd := exec.Command("ssh", remoteHost, remoteUliCmd)
	stdout, err := drushCmd.Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(stdout))
}

func uliGenerateLink(drupalRoot string, uri string, uids ...int) string {

	uid := 1
	if len(uids) > 0 {
		uid = uids[0]
	}

	var stdout []byte
	var err error
	var drushCmd *exec.Cmd

	os.Chdir(drupalRoot)

	// nuke all Drupal sessions so we don't get the access denied error
	err = exec.Command("drush", "sqlq", "TRUNCATE SESSIONS").Run()
	if err != nil {
		fmt.Println(err.Error())
	}

	drushCmd = exec.Command("drush", "uli", "--uid", strconv.Itoa(uid), "--uri", uri, "--no-browser")

	stdout, err = drushCmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	}

	return strings.TrimSpace(string(stdout))
}
