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

		explicitDatabaseName, _ := cmd.PersistentFlags().GetString("name")

		user, err := user.Current()
		if err != nil {
			log.Fatalf(err.Error())
		}

		username := user.Username

		databaseDumpParentDir := viper.GetString("CWCLI_DATABASE_IMPORT_DIR")
		if databaseDumpParentDir == "" {
			homeDir, _ := os.UserHomeDir()
			databaseDumpParentDir = homeDir + "/.cw/database_dumps"
		}

		err = os.MkdirAll(databaseDumpParentDir, os.ModePerm)
		if err != nil {
			fmt.Printf("UNABLE TO MKDIR [%s]: %s", databaseDumpParentDir, err.Error())
			os.Exit(1)
		}

		var vars cwutils.CwVars
		var tempFilePath string
		var gunzipCmdString string

		cwutils.InitViperConfigEnv()

		var SSH_TEST_SERVER string = viper.GetString("CWCLI_SSH_TEST_SERVER")
		var SSH_USER string = viper.GetString("CWCLI_SSH_USER")

		databaseName := ""
		if explicitDatabaseName != "" {
			databaseName = explicitDatabaseName
		} else {
			vars = cwutils.GetProjectVars()
			cwutils.CheckLocalConfigOverrides(vars.Project_root)
			tempFilePath = fmt.Sprintf("%s/%s.sql.gz", databaseDumpParentDir, vars.Drupal_dbname)
			gunzipCmdString = fmt.Sprintf("gunzip < %s | mysql %s", tempFilePath, vars.Drupal_dbname)
			databaseName = vars.Drupal_dbname
		}

		siteName := "non-drupal-site"
		if vars.Drupal_site_name != "" {
			siteName = vars.Drupal_site_name
		}

		fmt.Printf("[%s] Starting database pull for '%s'.\n", siteName, databaseName)

		/**
		 * --force? then we drop the database...
		 */
		if isFlaggedForce {
			fmt.Printf("[%s] Dropping database '%s' for a full import (FORCE)...\n", siteName, databaseName)
			exec.Command("mysqladmin", "drop", databaseName, "-f").Run()
		}

		err = exec.Command("mysqladmin", "create", databaseName).Run()
		if err == nil {
			fmt.Printf("[%s] Database '%s' was just created, proceeding with a fresh import...\n", siteName, databaseName)
		}

		if !vars.Is_pantheon {

			var err error

			if !isFlaggedSlow {

				databaseDumpDir := fmt.Sprintf("%s/%s", databaseDumpParentDir, databaseName)

				err = os.MkdirAll(databaseDumpDir, os.ModePerm)
				if err != nil {
					fmt.Printf("UNABLE TO MKDIR [%s]: %s", databaseDumpDir, err.Error())
					os.Exit(1)
				}

				fmt.Printf("[%s] Dumping remote database '%s'...\n", siteName, databaseName)
				remoteMysqlDumpCmd := fmt.Sprintf("~/bin/database-dump.sh %s", databaseName)
				err = exec.Command("ssh", SSH_USER+"@"+SSH_TEST_SERVER, remoteMysqlDumpCmd).Run()

				if err != nil {
					fmt.Printf("[%s] ABORT: Unable to backup remote database '%s'. SSH_USER=%s, SSH_HOST=%s\n", siteName, databaseName, SSH_USER, SSH_TEST_SERVER)
					os.Exit(1)
				}

				fmt.Printf("[%s] Downloading remote database '%s'...\n", siteName, databaseName)
				rsyncSrc := fmt.Sprintf("%s@%s:~/backups/transient/%s", SSH_USER, SSH_TEST_SERVER, databaseName)
				rsyncOutput, _ := exec.Command("rsync", "-avz", rsyncSrc, databaseDumpParentDir, "--delete").CombinedOutput()

				if isFlaggedVerbose {
					fmt.Printf("%s\n", rsyncOutput)
				}

				fmt.Printf("[%s] Dumping local session table %s\n", siteName, databaseDumpDir)

				mydumperArgs := []string{
					"--user", username,
					"--database", databaseName,
					"--regex", fmt.Sprintf("%s.sessions", databaseName),
					"--outputdir", databaseDumpDir,
				}

				mydumperOutput, err := exec.Command("mydumper", mydumperArgs...).CombinedOutput()
				if err != nil {
					fmt.Printf("MYDUMPER ERROR: %s\n%s", err.Error(), mydumperOutput)
					os.Exit(1)
				}

				fmt.Printf("[%s] Importing database files from %s\n", siteName, databaseDumpDir)

				myloaderArgs := []string{
					"--user", username,
					"--database", databaseName,
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

				// halt here, since a db name was provided. no need to clear caches, etc
				if explicitDatabaseName != "" {
					os.Exit(0)
				}

			} else {
				remoteDatabaseDumpFilename := vars.Drupal_dbname + "-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
				remoteMysqlDumpCmd := fmt.Sprintf("mysqldump %s |gzip> ~/backups/transient/%s.sql.gz", vars.Drupal_dbname, remoteDatabaseDumpFilename)

				fmt.Printf("[%s] Dumping remote database '%s' (mysqldump)\n", vars.Drupal_site_name, vars.Drupal_dbname)
				sshCmd := exec.Command("ssh", SSH_USER+"@"+SSH_TEST_SERVER, remoteMysqlDumpCmd)
				sshCmd.Start()
				sshCmd.Wait()

				fmt.Printf("[%s] Downloading remote database file ~/backups/transient/%s.sql.gz ...\n", vars.Drupal_site_name, remoteDatabaseDumpFilename)
				scpSrc := fmt.Sprintf("%s@%s:~/backups/transient/%s.sql.gz", SSH_USER, SSH_TEST_SERVER, remoteDatabaseDumpFilename)
				scpCmd := exec.Command("scp", scpSrc, tempFilePath)
				scpCmd.Start()
				scpCmd.Wait()

				fmt.Printf("[%s] Importing database '%s' into sandbox.\n", vars.Drupal_site_name, vars.Drupal_dbname)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()

				if err != nil {
					fmt.Printf("[%s] Something went wrong when restoring database.\n", vars.Drupal_site_name)
					log.Fatal(err)
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
				fmt.Printf("[%s] Something went wrong when restoring database.\n", vars.Drupal_site_name)
				log.Fatal(err)
			}
		}

		if isFlaggedVerbose {
			fmt.Printf("[%s] Cleaning up temp files...\n", vars.Drupal_site_name)
		}

		if isFlaggedSlow {
			err := os.Remove(tempFilePath)
			if err != nil {
				fmt.Printf("[%s] Something went wrong when cleaning up temp files in %s\n", vars.Drupal_site_name, tempFilePath)
				log.Fatal(err)
			}
		}

		fmt.Printf("[%s] Clearing Drupal cache...\n", vars.Drupal_site_name)

		var dv string = vars.Drupal_version[0:1]
		drupal_version, _ := strconv.Atoi(dv)

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

	pullDbCmd.PersistentFlags().String("name", "", "Pull a specific database by name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullDbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullDbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
