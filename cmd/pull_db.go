/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	cwutils "cw-cli/utils"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// pullDbCmd represents the pullDb command
var pullDbCmd = &cobra.Command{
	Use:   "db",
	Short: "Pull database from test down to sandbox",
	Run: func(cmd *cobra.Command, args []string) {

		isFlaggedVerbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
		isFlaggedSlow, _ := rootCmd.PersistentFlags().GetBool("slow")
		isFlaggedForce, _ := rootCmd.PersistentFlags().GetBool("force")

		user, err := user.Current()
		if err != nil {
			log.Fatalf(err.Error())
		}

		username := user.Username

		var vars cwutils.CwVars = cwutils.GetProjectVars()
		var tempFilePath string = fmt.Sprintf("/tmp/db_%s.sql.gz", vars.Drupal_dbname)
		var gunzipCmdString = fmt.Sprintf("gunzip < %s | mysql %s", tempFilePath, vars.Drupal_dbname)
		cwutils.InitViperConfigEnv()
		cwutils.CheckLocalConfigOverrides(vars.Project_root)
		var SSH_TEST_SERVER string = viper.GetString("CWCLI_SSH_TEST_SERVER")
		var SSH_USER string = viper.GetString("CWCLI_SSH_USER")

		if len(vars.Drupal_version) == 0 {
			log.Fatal("Drupal_version is empty!")
		}

		var dv string = vars.Drupal_version[0:1]
		drupal_version, _ := strconv.Atoi(dv)

		fmt.Printf("[%s] Starting database pull for '%s'.\n", vars.Drupal_site_name, vars.Drupal_dbname)

		if !vars.Is_pantheon {

			var err error

			if !isFlaggedSlow {
				databaseDumpParentDir := viper.GetString("CWCLI_DATABASE_IMPORT_DIR")
				if databaseDumpParentDir == "" {
					databaseDumpParentDir = "/tmp/database_dumps"
				}

				databaseDumpDir := fmt.Sprintf("%s/%s", databaseDumpParentDir, vars.Drupal_dbname)
				//databaseImportDir := fmt.Sprintf("%s/%s", databaseDumpDir, "import")

				err = os.MkdirAll(databaseDumpDir, os.ModePerm)
				if err != nil {
					fmt.Printf("UNABLE TO MKDIR [%s]: %s", databaseDumpDir, err.Error())
					os.Exit(1)
				}

				/**
				 * --force? then we drop the database...
				 */
				if isFlaggedForce {
					fmt.Printf("[%s] Dropping database '%s' for a full import (FORCE)...\n", vars.Drupal_site_name, vars.Drupal_dbname)
					exec.Command("mysqladmin", "drop", vars.Drupal_dbname, "-f").Run()
				}

				//databaseExists := true
				err = exec.Command("mysqladmin", "create", vars.Drupal_dbname).Run()
				if err == nil {
					//databaseExists = false
					fmt.Printf("[%s] Database '%s' was just created, gotta do a full import...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				}

				fmt.Printf("[%s] Dumping remote database '%s'...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				remoteMysqlDumpCmd := fmt.Sprintf("~/bin/database-dump.sh %s", vars.Drupal_dbname)
				exec.Command("ssh", SSH_USER+"@"+SSH_TEST_SERVER, remoteMysqlDumpCmd).Run()

				fmt.Printf("[%s] Downloading remote database '%s'...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				rsyncSrc := fmt.Sprintf("%s@%s:~/backups/transient/%s", SSH_USER, SSH_TEST_SERVER, vars.Drupal_dbname)
				rsyncOutput, _ := exec.Command("rsync", "-avz", rsyncSrc, databaseDumpParentDir, "--delete").CombinedOutput()

				if isFlaggedVerbose {
					fmt.Printf("%s\n", rsyncOutput)
				}

				fmt.Printf("[%s] Dumping local session table %s\n", vars.Drupal_site_name, databaseDumpDir)

				mydumperArgs := []string{
					"--user", username,
					"--database", vars.Drupal_dbname,
					"--regex", fmt.Sprintf("%s.sessions", vars.Drupal_dbname),
					"--outputdir", databaseDumpDir,
				}

				mydumperOutput, err := exec.Command("mydumper", mydumperArgs...).CombinedOutput()
				if err != nil {
					fmt.Printf("MYDUMPER ERROR: %s\n%s", err.Error(), mydumperOutput)
					os.Exit(1)
				}

				fmt.Printf("[%s] Importing database files from %s\n", vars.Drupal_site_name, databaseDumpDir)

				myloaderArgs := []string{
					"--user", username,
					"--database", vars.Drupal_dbname,
					"--directory", databaseDumpDir,
					"--overwrite-tables",
					"--purge-mode", "TRUNCATE",
				}

				myloaderOutput, err := exec.Command("myloader", myloaderArgs...).CombinedOutput()
				if err != nil {
					log.Fatal("MYLOADER ERROR: " + err.Error())
				}

				if isFlaggedVerbose {
					fmt.Printf("%s", myloaderOutput)
				}

			} else {
				remoteDatabaseDumpFilename := vars.Drupal_dbname + "-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
				remoteMysqlDumpCmd := fmt.Sprintf("mysqldump %s |gzip> /tmp/database_dumps/%s.sql.gz", vars.Drupal_dbname, remoteDatabaseDumpFilename)

				fmt.Printf("[%s] Dumping remote database '%s'...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				sshCmd := exec.Command("ssh", SSH_USER+"@"+SSH_TEST_SERVER, remoteMysqlDumpCmd)
				sshCmd.Start()
				sshCmd.Wait()

				fmt.Printf("[%s] Downloading remote database '%s'...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				scpSrc := fmt.Sprintf("%s@%s:/tmp/database_dumps/%s.sql.gz", SSH_USER, SSH_TEST_SERVER, remoteDatabaseDumpFilename)
				scpCmd := exec.Command("scp", scpSrc, tempFilePath)
				scpCmd.Start()
				scpCmd.Wait()

				fmt.Printf("[%s] Restoring database '%s'. This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
				if err != nil {
					fmt.Printf("[%s] Database '%s' does not exist. Creating...\n", vars.Drupal_site_name, vars.Drupal_dbname)
					_ = exec.Command("bash", "-c", fmt.Sprintf("mysqladmin create %s", vars.Drupal_dbname)).Run()

					fmt.Printf("[%s] Attempting to restore database '%s' again. This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)
					_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
					if err != nil {
						fmt.Printf("[%s] Something went wrong when restoring database.\n", vars.Drupal_site_name)
						log.Fatal(err)
					}
				}
			}
		} else {

			pantheonEnv := "dev"
			if vars.Branch_name != "master" {
				pantheonEnv = vars.Branch_name
			}

			// CREATE BACKUP =========================================================
			fmt.Printf("[%s] Creating remote database backup for '%s'...\n", vars.Drupal_site_name, vars.Drupal_dbname)
			terminusBackupCreateCmd := exec.Command("terminus", "backup:create", vars.Drupal_site_name+"."+pantheonEnv, "--element=db")
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
			fmt.Printf("[%s] Downloading remote database backup...\n", vars.Drupal_site_name)
			terminusBackupGetCmd := exec.Command("terminus", "backup:get", vars.Drupal_site_name+"."+pantheonEnv, "--element=db")
			backupGetStdout, _ := terminusBackupGetCmd.Output()

			wgetCmd := exec.Command("wget", "--quiet", "--show-progress", strings.TrimSpace(string(backupGetStdout)), "-O", tempFilePath)
			wgetStdout, _ := wgetCmd.StdoutPipe()
			wgetStderr, _ := wgetCmd.StderrPipe()

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
			fmt.Printf("[%s] Importing database '%s' into sandbox...\n", vars.Drupal_site_name, vars.Drupal_dbname)

			_, err := exec.Command("bash", "-c", gunzipCmdString).Output()
			if err != nil {
				fmt.Printf("[%s] Database '%s' does not exist. Creating...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_ = exec.Command("bash", "-c", fmt.Sprintf("mysqladmin create %s", vars.Drupal_dbname)).Run()

				fmt.Printf("[%s] Attempting to restore database '%s' again. This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()
				if err != nil {
					fmt.Printf("[%s] Something went wrong when restoring database.\n", vars.Drupal_site_name)
					log.Fatal(err)
				}
			}
		}

		if isFlaggedVerbose {
			fmt.Printf("[%s] Cleaning up temp files...\n", vars.Drupal_site_name)
		}

		if isFlaggedSlow {
			err := os.Remove(tempFilePath)
			if err != nil {
				fmt.Printf("[%s] Something went wrong when cleaning up temp files.\n", vars.Drupal_site_name)
				log.Fatal(err)
			}
		}

		fmt.Printf("[%s] Clearing drupal cache...\n", vars.Drupal_site_name)

		drushArgs := []string{"cr"}
		if drupal_version == 7 {
			drushArgs = []string{"cc", "all"}
		}
		drushCmd := exec.Command("drush", drushArgs...)
		_ = drushCmd.Run()

		fmt.Printf("[%s] Database pull complete.\n\n", vars.Drupal_site_name)
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
