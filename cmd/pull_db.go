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
	"path"
	"path/filepath"
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
		if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
			log.Fatal(err)
		}

		if err := viper.BindPFlag("fast", rootCmd.PersistentFlags().Lookup("fast")); err != nil {
			log.Fatal(err)
		}

		if err := viper.BindPFlag("force", rootCmd.PersistentFlags().Lookup("force")); err != nil {
			log.Fatal(err)

		}
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

		if !vars.Is_pantheon {

			var err error

			if viper.GetBool("fast") {
				databaseDumpParentDir := viper.GetString("CWCLI_DATABASE_IMPORT_DIR")
				if databaseDumpParentDir == "" {
					databaseDumpParentDir = "/tmp/database_dumps"
				}

				databaseDumpDir := fmt.Sprintf("%s/%s", databaseDumpParentDir, vars.Drupal_dbname)
				databaseImportDir := fmt.Sprintf("%s/%s", databaseDumpDir, "import")

				err = os.MkdirAll(databaseDumpDir, os.ModePerm)
				if err != nil {
					fmt.Printf("UNABLE TO MKDIR [%s]: %s", databaseDumpDir, err.Error())
					os.Exit(1)
				}

				/**
				 * --force? then we drop the database...
				 */
				if viper.GetBool("force") {
					fmt.Printf("[%s] Dropping database '%s' for a full import (FORCE)...\n", vars.Drupal_site_name, vars.Drupal_dbname)
					exec.Command("mysqladmin", "drop", vars.Drupal_dbname, "-f").Run()
				}

				databaseExists := true
				err = exec.Command("mysqladmin", "create", vars.Drupal_dbname).Run()
				if err == nil {
					databaseExists = false
					fmt.Printf("[%s] Database '%s' was just created, gotta do a full import...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				}

				fmt.Printf("[%s] Dumping remote database '%s'...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				remoteMysqlDumpCmd := fmt.Sprintf("~/bin/database-dump.sh %s", vars.Drupal_dbname)
				exec.Command("ssh", SSH_USER+"@"+SSH_TEST_SERVER, remoteMysqlDumpCmd).Run()

				fmt.Printf("[%s] Downloading remote database '%s'...\n", vars.Drupal_site_name, vars.Drupal_dbname)
				rsyncSrc := fmt.Sprintf("%s@%s:~/backups/transient/%s", SSH_USER, SSH_TEST_SERVER, vars.Drupal_dbname)
				rsyncOutput, _ := exec.Command("rsync", "-vcrz", rsyncSrc, databaseDumpParentDir, "--delete").CombinedOutput()

				/**
				 * Database already exists? Let's import only the tables that have changed
				 */
				if databaseExists {
					fmt.Printf("[%s] Optimizing import files...\n", vars.Drupal_site_name)

					/**
					 * ./import/ will contain symlinks to the files we actually want to import
					 * so nuke the current ./import/ dir if it exists, then recreate it
					 */
					err = os.RemoveAll(databaseImportDir)
					if err != nil {
						log.Fatal(" UNABLE TO DELETE import dir: " + err.Error())
					}

					err = os.MkdirAll(databaseImportDir, os.ModePerm)
					if err != nil {
						log.Fatal(" UNABLE TO MKDIR import: " + err.Error())
					}

					/**
					 * Create symlinks for each file that came down in the rsync
					 */
					scanner := bufio.NewScanner(strings.NewReader(string(rsyncOutput)))
					for scanner.Scan() {
						line := scanner.Text()

						if _, err := os.Stat(databaseDumpParentDir + "/" + line); err == nil {
							filename := path.Base(line)
							if filename != "." {
								targetFile := fmt.Sprintf("../%s", filename)
								linkFile := fmt.Sprintf("%s/%s", databaseImportDir, filename)
								if viper.GetBool("verbose") {
									fmt.Printf("Create symlink: [%s => %s]\n", linkFile, targetFile)
								}
								err2 := os.Symlink(targetFile, linkFile)
								if err2 != nil {
									fmt.Printf(" %s\n", err2.Error())
								}
							}
						}
					}

					/**
					 * Also create symlinks for any SQL schemas (CREATE statements) that are needed.
					 *
					 * I.e., for any data file, we need to also symlink its schema, so that the table
					 * will get re-created and rows re-inserted
					 */
					dataFiles, err := filepath.Glob(databaseImportDir + "/*0*.sql")
					if err != nil {
						log.Fatal(err.Error())
					}
					for _, dataFile := range dataFiles {
						dataFilename := path.Base(dataFile)
						dataFileParts := strings.Split(dataFilename, ".")
						schemaFilename := strings.Join(dataFileParts[:len(dataFileParts)-2], ".") + "-schema.sql"

						targetFile := fmt.Sprintf("../%s", schemaFilename)
						linkFile := fmt.Sprintf("%s/%s", databaseImportDir, schemaFilename)
						if viper.GetBool("verbose") {
							fmt.Printf("Create symlink schema file: [%s => %s]\n", linkFile, targetFile)
						}

						if _, err := os.Lstat(linkFile); err == nil {
							os.Remove(linkFile)
						}

						err := os.Symlink(targetFile, linkFile)
						if err != nil {
							log.Fatal(err.Error())
						}
					}

				} else {
					databaseImportDir = databaseDumpDir
				}

				fmt.Printf("[%s] Importing database files from %s\n", vars.Drupal_site_name, databaseImportDir)

				myloaderArgs := []string{"--database", vars.Drupal_dbname, "--directory", databaseImportDir, "--overwrite-tables"}
				myloaderOutput, err := exec.Command("myloader", myloaderArgs...).CombinedOutput()
				//err = exec.Command("myloader", "--database", vars.Drupal_dbname, "--directory", databaseImportDir, "--overwrite-tables").Run()
				if err != nil {
					log.Fatal("MYLOADER ERROR: " + err.Error())
				}

				if viper.GetBool("verbose") {
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
			fmt.Printf("[%s] creating database backup for '%s'. This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)
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
			fmt.Printf("[%s] downloading database backup...\n", vars.Drupal_site_name)
			terminusBackupGetCmd := exec.Command("terminus", "backup:get", vars.Drupal_site_name+"."+pantheonEnv, "--element=db")
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
			fmt.Printf("[%s] Restoring database '%s'. This could take awhile...\n", vars.Drupal_site_name, vars.Drupal_dbname)

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

		if viper.GetBool("verbose") {
			fmt.Printf("[%s] Cleaning up temp files...\n", vars.Drupal_site_name)
		}

		if !viper.GetBool("fast") {
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
