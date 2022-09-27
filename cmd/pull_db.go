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
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/sfreiberg/simplessh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		cwutils.InitViperConfigEnv()
		cwutils.CheckLocalConfigOverrides(vars.Project_root)
		var SSH_TEST_SERVER string = viper.GetString("CWCLI_SSH_TEST_SERVER")
		var SSH_USER string = viper.GetString("CWCLI_SSH_USER")
		SSH_AGENT_PID, HAS_AGENT_PID := os.LookupEnv("SSH_AGENT_PID")

		// if HAS_AGENT_PID {
		// 	fmt.Println("YES PID")
		// 	fmt.Println("'" + SSH_AGENT_PID + "'")
		// } else {
		// 	fmt.Println("NO PID")
		// 	fmt.Println("'" + SSH_AGENT_PID + "'")
		// }
		// os.Exit(420)

		if !vars.Is_pantheon {
			fmt.Printf("[%s] Pulling down database, this could take awhile...\n", vars.Drupal_site_name)

			var client *simplessh.Client
			var err error

			if HAS_AGENT_PID {
				// fmt.Println("YES PID")
				fmt.Printf("SSH AGENT PID: %s\n", SSH_AGENT_PID)
				if client, err = simplessh.ConnectWithAgent(SSH_TEST_SERVER, SSH_USER); err != nil {
					log.Fatal(err)
				}
			} else {
				// fmt.Println("NO PID")
				// fmt.Println("'" + SSH_AGENT_PID + "'")
				if client, err = simplessh.ConnectWithKeyFile(SSH_TEST_SERVER, SSH_USER, ""); err != nil {
					log.Fatal(err)
				}
			}

			defer client.Close()

			res, err := client.Exec(createBackupString)
			if err != nil {
				log.Fatal(err)
			}

			localFile, err := os.Create(tempFilePath)
			if err != nil {
				log.Fatal(err)
			}

			defer localFile.Close()

			_, err = io.Copy(localFile, bytes.NewReader(res))
			if err != nil {
				fmt.Printf("[%s] Something went wrong when copying temp db gzip.\n", vars.Drupal_site_name)
				log.Fatal(err)
			}

			_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
			if err != nil {
				fmt.Printf("[%s] Database \"%s\" does not exist. Creating...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_ = exec.Command("bash", "-c", fmt.Sprintf("mysqladmin create %s", vars.Drupal_dbname)).Run()

				fmt.Printf("[%s] Restoring database \"%s\". This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
				if err != nil {
					fmt.Printf("[%s] Something went wrong when restoring database.\n", vars.Drupal_site_name)
					log.Fatal(err)
				}
			}
		} else {

			// CREATE BACKUP =========================================================
			fmt.Printf("[%s] creating database backup for \"%s\". This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)
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
			fmt.Printf("[%s] downloading database backup...\n", vars.Drupal_site_name)
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
			fmt.Printf("[%s] extracting database backup...\n", vars.Drupal_site_name)

			_, err := exec.Command("bash", "-c", gunzipCmdString).Output()
			if err != nil {
				fmt.Printf("[%s] Database \"%s\" does not exist. Creating...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_ = exec.Command("bash", "-c", fmt.Sprintf("mysqladmin create %s", vars.Drupal_dbname)).Run()

				fmt.Printf("[%s] Restoring database \"%s\". This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
				if err != nil {
					fmt.Printf("[%s] Something went wrong when restoring database.\n", vars.Drupal_site_name)
					log.Fatal(err)
				}
			}
		}

		fmt.Printf("[%s] Cleaning up temp files...\n", vars.Drupal_site_name)
		err := os.Remove(tempFilePath)
		if err != nil {
			fmt.Printf("[%s] Something went wrong when cleaning up temp files.\n", vars.Drupal_site_name)
			log.Fatal(err)
		}
		fmt.Printf("[%s] Finished pulling down database!\n\n", vars.Drupal_site_name)
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
