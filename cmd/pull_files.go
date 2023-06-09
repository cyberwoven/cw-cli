/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	cwutils "cw-cli/utils"

	"github.com/spf13/cobra"
)

var pullFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Pull files from test to sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		isFlaggedVerbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
		isFlaggedSlow, _ := rootCmd.PersistentFlags().GetBool("slow")

		ctx := cwutils.GetContext()
		
		var rsyncRemote string = fmt.Sprintf("%s@%s:%s/files", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST, ctx.DRUPAL_DEFAULT_DIR_REMOTE)

		fmt.Printf("[%s] Starting file pull from branch '%s'.\n", ctx.SITE_NAME, ctx.GIT_BRANCH)

		if !ctx.IS_PANTHEON {
			rsyncCmd := exec.Command("rsync",
				"-vcrtzP",
				rsyncRemote,
				ctx.DRUPAL_PUBLIC_FILES_DIR,
				"--stats",
				// "--dry-run",
				"--exclude=advagg_css",
				"--exclude=advagg_js",
				"--exclude=css",
				"--exclude=ctools",
				"--exclude=js",
				"--exclude=php",
				"--exclude=styles",
				"--exclude=tmp")

			stdout, _ := rsyncCmd.StdoutPipe()
			stderr, _ := rsyncCmd.StderrPipe()

			fmt.Printf("[%s] Pulling down files...\n", ctx.SITE_NAME)

			_ = rsyncCmd.Start()

			if isFlaggedVerbose {
				scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
				scanner.Split(bufio.ScanLines)
				for scanner.Scan() {
					m := scanner.Text()
					fmt.Println(m)
				}
			}
			_ = rsyncCmd.Wait()

		} else {

			pantheonEnv := "dev"
			if ctx.GIT_BRANCH_SLUG != "master" {
				pantheonEnv = ctx.GIT_BRANCH_SLUG
			}

			if !isFlaggedSlow {
				/**
				 * We assume the terminus rsync plugin is installed
				 */
				terminusRsyncArgs := []string{
					"rsync",
					fmt.Sprintf("%s.%s:files", ctx.SITE_NAME, pantheonEnv),
					ctx.DRUPAL_DEFAULT_DIR_LOCAL,
				}

				rsyncCmd := exec.Command("terminus", terminusRsyncArgs...)
				rsyncStdout, _ := rsyncCmd.StdoutPipe()
				rsyncStderr, _ := rsyncCmd.StderrPipe()
				_ = rsyncCmd.Start()

				if isFlaggedVerbose {
					fmt.Printf("[rsyncCmd] %s\n", rsyncCmd)
					
					rsyncScanner := bufio.NewScanner(io.MultiReader(rsyncStderr, rsyncStdout))
					rsyncScanner.Split(bufio.ScanLines)
					for rsyncScanner.Scan() {
						m := rsyncScanner.Text()
						fmt.Println(m)
					}
				}
				_ = rsyncCmd.Wait()

			} else {

				// CREATE BACKUP =========================================================
				fmt.Printf("[%s] creating remote files archive...\n", ctx.SITE_NAME)
				terminusBackupCreateCmd := exec.Command("terminus", "backup:create", ctx.SITE_NAME+"."+pantheonEnv, "--element=files")
				backupCreateStdout, _ := terminusBackupCreateCmd.StdoutPipe()
				backupCreateStderr, _ := terminusBackupCreateCmd.StderrPipe()
				// fmt.Println("[Running Command]: " + terminusBackupCreateCmd.String())
				_ = terminusBackupCreateCmd.Start()

				if isFlaggedVerbose {
					backupCreateScanner := bufio.NewScanner(io.MultiReader(backupCreateStdout, backupCreateStderr))
					backupCreateScanner.Split(bufio.ScanLines)
					for backupCreateScanner.Scan() {
						m := backupCreateScanner.Text()
						fmt.Println(m)
					}
				}
				_ = terminusBackupCreateCmd.Wait()

				// GET BACKUP ============================================================
				fmt.Printf("[%s] downloading files tarball...\n", ctx.SITE_NAME)
				terminusBackupGetCmd := exec.Command("terminus", "backup:get", ctx.SITE_NAME+"."+pantheonEnv, "--element=files")
				// fmt.Println("[Running Command]: " + terminusBackupGetCmd.String())
				backupGetStdout, _ := terminusBackupGetCmd.Output()
				// fmt.Println("[PANTHEON]: backup url - " + string(backupGetStdout))
				wgetCmd := exec.Command("wget", "--quiet", "--show-progress", strings.TrimSpace(string(backupGetStdout)), "-O", "/tmp/files_"+ctx.SITE_NAME+".tar.gz")
				wgetStdout, _ := wgetCmd.StdoutPipe()
				wgetStderr, _ := wgetCmd.StderrPipe()
				// fmt.Println("[Running Command]: " + wgetCmd.String())
				_ = wgetCmd.Start()

				if isFlaggedVerbose {
					wgetScanner := bufio.NewScanner(io.MultiReader(wgetStderr, wgetStdout))
					wgetScanner.Split(bufio.ScanLines)
					for wgetScanner.Scan() {
						m := wgetScanner.Text()
						fmt.Println(m)
					}
				}

				_ = wgetCmd.Wait()

				// EXTRACT BACKUP ========================================================
				if isFlaggedVerbose {
					fmt.Printf("[%s] extracting files tarball...\n", ctx.SITE_NAME)
				}
				mkdirCmd := exec.Command("mkdir", "/tmp/files_"+ctx.DATABASE_NAME)
				_ = mkdirCmd.Run()
				tarCmd := exec.Command("tar", "--directory=/tmp/files_"+ctx.DATABASE_NAME, "-xzvf", "/tmp/files_"+ctx.DATABASE_NAME+".tar.gz")
				_ = tarCmd.Run()
				// COPY EXTRACTED FILES ==================================================
				if isFlaggedVerbose {
					fmt.Printf("[%s] copying files into 'sites/default/files/' directory...\n", ctx.SITE_NAME)
				}
				copyExtractCmd := exec.Command("rsync", "-vcrP", "--stats", "/tmp/files_"+ctx.DATABASE_NAME+"/files_dev/", ctx.DRUPAL_PUBLIC_FILES_DIR)
				_ = copyExtractCmd.Run()

			}
		}

		fmt.Printf("[%s] File pull complete.\n\n", ctx.SITE_NAME)
	},
}

func init() {
	pullCmd.AddCommand(pullFilesCmd)
}
