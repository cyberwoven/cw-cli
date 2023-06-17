package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var pushFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Push public files from local up to test",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		if ctx.IS_PANTHEON {
			fmt.Print("ABORT - push db is not implemented for Pantheon sites!")
			os.Exit(1)
		}

		if ctx.SITE_TYPE != "drupal" && ctx.SITE_TYPE != "wordpress" {
			fmt.Print("ABORT - file push is only supported for Drupal and Wordpress sites")
			os.Exit(1)
		}

		if ctx.SITE_TYPE == "drupal" {
			remoteHost := fmt.Sprintf("%s@%s", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST)
			remoteTestCmd := fmt.Sprintf("test -d %s", ctx.DRUPAL_DEFAULT_DIR_REMOTE)
			err = exec.Command("ssh", remoteHost, remoteTestCmd).Run()
			if err != nil {
				fmt.Printf("ABORT - remote directory does not exist: %s", ctx.DRUPAL_DEFAULT_DIR_REMOTE)
				os.Exit(1)
			}

			remoteAcquireOwnershipCmd := fmt.Sprintf("sudo acquire-site-file-ownership %s %s", ctx.SITE_NAME, ctx.GIT_BRANCH_SLUG)
			err = exec.Command("ssh", remoteHost, remoteAcquireOwnershipCmd).Run()
			if err != nil {
				fmt.Printf("ABORT - Unable to acquire site file ownership before rsync: %s %s", ctx.SITE_NAME, ctx.GIT_BRANCH_SLUG)
				os.Exit(1)
			}

			rsyncDest := fmt.Sprintf("%s@%s:%s", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST, ctx.DRUPAL_DEFAULT_DIR_REMOTE)
			err = exec.Command("rsync", "-avz", ctx.DRUPAL_PUBLIC_FILES_DIR, rsyncDest).Run()
			if err != nil {
				fmt.Printf("Unable to rsync database files up to remote host for database %s\n", ctx.DATABASE_NAME)
				log.Fatal(err)
			}

			remoteReturnOwnershipCmd := fmt.Sprintf("sudo return-site-file-ownership %s %s", ctx.SITE_NAME, ctx.GIT_BRANCH_SLUG)
			err = exec.Command("ssh", remoteHost, remoteReturnOwnershipCmd).Run()
			if err != nil {
				fmt.Printf("ABORT - Unable to return site file ownership before rsync: %s %s", ctx.SITE_NAME, ctx.GIT_BRANCH_SLUG)
				os.Exit(1)
			}

		} else if ctx.SITE_TYPE == "wordpress" {
			fmt.Print("push wp files")
		}

	},
}

func init() {
	pushCmd.AddCommand(pushFilesCmd)
}
