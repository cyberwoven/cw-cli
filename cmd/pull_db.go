/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"bytes"
	cwutils "cw-cli/utils"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sfreiberg/simplessh"
	"github.com/spf13/cobra"
)

// pullDbCmd represents the pullDb command
var pullDbCmd = &cobra.Command{
	Use:   "db",
	Short: "Pull database from test down to sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		var vars cwutils.CwVars = cwutils.GetProjectVars()
		var tempFilePath string = fmt.Sprintf("/tmp/db_%s.sql.gz", vars.Drupal_dbname)
		var createBackupString string = fmt.Sprintf("mysqldump %s | gzip", vars.Drupal_dbname)
		var gunzipCmdString = fmt.Sprintf("gunzip < %s | mysql %s", tempFilePath, vars.Drupal_dbname)

		if !vars.Is_pantheon {
			fmt.Printf("Pulling down database for %s. This could take awhile...\n", vars.Drupal_site_name)

			var client *simplessh.Client
			var err error

			if client, err = simplessh.ConnectWithKeyFile("forest-db.test.cyberwoven.net", "cyberwoven", ""); err != nil {
				fmt.Println(string(err.Error()))
				os.Exit(1)
			}

			defer client.Close()

			res, err := client.Exec(createBackupString)
			if err != nil {
				fmt.Println(string(err.Error()))
				os.Exit(1)
			}

			localFile, err := os.Create(tempFilePath)
			if err != nil {
				fmt.Println(string(err.Error()))
				os.Exit(1)
			}

			defer localFile.Close()

			_, err = io.Copy(localFile, bytes.NewReader(res))
			if err != nil {
				fmt.Println("Something went wrong when copying temp db gzip.")
				os.Exit(1)
			}

			_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
			if err != nil {
				fmt.Printf("Database \"%s\" does not exist. Creating...\n", vars.Drupal_dbname)
				_ = exec.Command("bash", "-c", fmt.Sprintf("mysqladmin create %s", vars.Drupal_dbname)).Run()

				fmt.Printf("Restoring database \"%s\". This could take awhile...\n", vars.Drupal_dbname)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
				if err != nil {
					fmt.Println("Something went wrong when restoring database.")
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}
		} else {

			// CREATE BACKUP =========================================================
			fmt.Printf("[PANTHEON]: creating database backup for \"%s\". This could take awhile...\n", vars.Drupal_dbname)
			terminusBackupCreateCmd := exec.Command("terminus", "backup:create", vars.Drupal_site_name+".dev", "--element=db")
			backupCreateStdout, _ := terminusBackupCreateCmd.StdoutPipe()
			backupCreateStderr, _ := terminusBackupCreateCmd.StderrPipe()
			// fmt.Println("[Running Command]: " + terminusBackupCreateCmd.String())
			_ = terminusBackupCreateCmd.Start()
			backupCreateScanner := bufio.NewScanner(io.MultiReader(backupCreateStdout, backupCreateStderr))
			backupCreateScanner.Split(bufio.ScanLines)
			for backupCreateScanner.Scan() {
				m := backupCreateScanner.Text()
				fmt.Println(m)
			}
			_ = terminusBackupCreateCmd.Wait()

			// GET BACKUP ============================================================
			fmt.Println("[PANTHEON]: downloading database backup...")
			terminusBackupGetCmd := exec.Command("terminus", "backup:get", vars.Drupal_site_name+".dev", "--element=db")
			backupGetStdout, _ := terminusBackupGetCmd.Output()

			wgetCmd := exec.Command("wget", "--quiet", "--show-progress", strings.TrimSpace(string(backupGetStdout)), "-O", tempFilePath)
			wgetStdout, _ := wgetCmd.StdoutPipe()
			wgetStderr, _ := wgetCmd.StderrPipe()

			_ = wgetCmd.Start()
			wgetScanner := bufio.NewScanner(io.MultiReader(wgetStderr, wgetStdout))
			wgetScanner.Split(bufio.ScanLines)
			for wgetScanner.Scan() {
				m := wgetScanner.Text()
				fmt.Println(m)
			}
			_ = wgetCmd.Wait()

			// EXTRACT BACKUP ========================================================
			fmt.Println("[PANTHEON]: extracting database backup...")

			_, err := exec.Command("bash", "-c", gunzipCmdString).Output()
			if err != nil {
				fmt.Printf("Database \"%s\" does not exist. Creating...\n", vars.Drupal_dbname)
				_ = exec.Command("bash", "-c", fmt.Sprintf("mysqladmin create %s", vars.Drupal_dbname)).Run()

				fmt.Printf("Restoring database \"%s\". This could take awhile...\n", vars.Drupal_dbname)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
				if err != nil {
					fmt.Println("Something went wrong when restoring database.")
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}
		}

		fmt.Println("Cleaning up temp files...")
		err := os.Remove(tempFilePath)
		if err != nil {
			fmt.Println("Something went wrong when cleaning up temp files.")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("Finished!")
	},
}

func init() {
	pullCmd.AddCommand(pullDbCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullDbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullDbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
