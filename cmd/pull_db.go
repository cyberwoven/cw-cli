package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var pullDbCmd = &cobra.Command{
	Use:   "db",
	Short: "Pull database from test down to sandbox",
	Run: func(cmd *cobra.Command, args []string) {

		isFlaggedVerbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
		isFlaggedSlow, _ := rootCmd.PersistentFlags().GetBool("slow")
		isFlaggedForce, _ := rootCmd.PersistentFlags().GetBool("force")
		explicitDatabaseName, _ := cmd.PersistentFlags().GetString("name")

		err := os.MkdirAll(ctx.DATABASE_IMPORT_DIR, os.ModePerm)
		if err != nil {
			fmt.Printf("UNABLE TO MKDIR [%s]: %s", ctx.DATABASE_IMPORT_DIR, err.Error())
			os.Exit(1)
		}

		var tempFilePath string
		var gunzipCmdString string

		databaseName := ""
		if explicitDatabaseName != "" {
			databaseName = explicitDatabaseName
		} else {
			tempFilePath = fmt.Sprintf("%s/%s.sql.gz", ctx.DATABASE_IMPORT_DIR, ctx.DATABASE_NAME)
			gunzipCmdString = fmt.Sprintf("gunzip < %s | mysql %s", tempFilePath, ctx.DATABASE_NAME)
			databaseName = ctx.DATABASE_NAME
		}

		if databaseName == "" {
			fmt.Print("ABORT: database name is empty")
			os.Exit(1)
		}

		siteName := "non-drupal-site"
		if ctx.SITE_NAME != "" {
			siteName = ctx.SITE_NAME
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

		if !ctx.IS_PANTHEON {

			var err error

			if !isFlaggedSlow {

				databaseDumpDir := fmt.Sprintf("%s/%s", ctx.DATABASE_IMPORT_DIR, databaseName)

				err = os.MkdirAll(databaseDumpDir, os.ModePerm)
				if err != nil {
					fmt.Printf("UNABLE TO MKDIR [%s]: %s", databaseDumpDir, err.Error())
					os.Exit(1)
				}

				fmt.Printf("[%s] Dumping remote database '%s'...\n", siteName, databaseName)
				remoteMysqlDumpCmd := fmt.Sprintf("~/bin/database-dump.sh %s", databaseName)
				remoteHost := fmt.Sprintf("%s@%s", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST)
				err = exec.Command("ssh", remoteHost, remoteMysqlDumpCmd).Run()

				if err != nil {
					fmt.Printf("[%s] ABORT: Unable to backup remote database '%s'. SSH_USER=%s, SSH_HOST=%s\n", siteName, databaseName, ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST)
					os.Exit(1)
				}

				fmt.Printf("[%s] Downloading remote database '%s'...\n", siteName, databaseName)
				rsyncSrc := fmt.Sprintf("%s@%s:~/backups/transient/%s", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST, databaseName)
				rsyncOutput, _ := exec.Command("rsync", "-avz", rsyncSrc, ctx.DATABASE_IMPORT_DIR, "--delete").CombinedOutput()

				if isFlaggedVerbose {
					fmt.Printf("%s\n", rsyncOutput)
				}

				if ctx.SITE_TYPE == "drupal" {
					fmt.Printf("[%s] Dumping local session table %s\n", siteName, databaseDumpDir)

					mydumperArgs := []string{
						"--user", ctx.USERNAME,
						"--database", databaseName,
						"--regex", fmt.Sprintf("%s.sessions", databaseName),
						"--outputdir", databaseDumpDir,
					}

					mydumperOutput, err := exec.Command("mydumper", mydumperArgs...).CombinedOutput()
					if err != nil {
						fmt.Printf("MYDUMPER ERROR: %s\n%s", err.Error(), mydumperOutput)
						os.Exit(1)
					}
				}

				fmt.Printf("[%s] Importing database files from %s\n", siteName, databaseDumpDir)

				myloaderArgs := []string{
					"--user", ctx.USERNAME,
					"--database", databaseName,
					"--directory", databaseDumpDir,
					"--overwrite-tables",
					"--purge-mode", "TRUNCATE",
				}

				myloaderOutput, err := exec.Command("myloader", myloaderArgs...).CombinedOutput()
				if isFlaggedVerbose {
					fmt.Printf("%s", myloaderOutput)
				}

				if err != nil {
					fmt.Printf("MYLOADER ERROR: %s\n\n", err.Error())
					fmt.Print("Try running the command again with the --force flag to drop and recreate the database.\n")
					os.Exit(1)
				}

				// halt here, since a db name was provided. no need to clear caches, etc
				if explicitDatabaseName != "" {
					os.Exit(0)
				}

			} else {
				remoteDatabaseDumpFilename := ctx.DATABASE_NAME + "-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
				remoteMysqlDumpCmd := fmt.Sprintf("mysqldump %s |gzip> ~/backups/transient/%s.sql.gz", ctx.DATABASE_NAME, remoteDatabaseDumpFilename)

				fmt.Printf("[%s] Dumping remote database '%s' (mysqldump)\n", ctx.SITE_NAME, ctx.DATABASE_NAME)
				sshCmd := exec.Command("ssh", ctx.SSH_TEST_USER+"@"+ctx.SSH_TEST_HOST, remoteMysqlDumpCmd)
				sshCmd.Start()
				sshCmd.Wait()

				fmt.Printf("[%s] Downloading remote database file ~/backups/transient/%s.sql.gz ...\n", ctx.SITE_NAME, remoteDatabaseDumpFilename)
				scpSrc := fmt.Sprintf("%s@%s:~/backups/transient/%s.sql.gz", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST, remoteDatabaseDumpFilename)
				scpCmd := exec.Command("scp", scpSrc, tempFilePath)
				scpCmd.Start()
				scpCmd.Wait()

				fmt.Printf("[%s] Importing database '%s' into sandbox.\n", ctx.SITE_NAME, ctx.DATABASE_NAME)
				_, err = exec.Command("bash", "-c", gunzipCmdString).Output()

				if err != nil {
					fmt.Printf("[%s] Something went wrong when restoring database.\n", ctx.SITE_NAME)
					log.Fatal(err)
				}

			}
		} else {

			pantheonEnv := "dev"
			if ctx.GIT_BRANCH_SLUG != "master" {
				pantheonEnv = ctx.GIT_BRANCH_SLUG
			}

			// CREATE BACKUP =========================================================
			fmt.Printf("[%s] Creating remote database backup for '%s'...\n", ctx.SITE_NAME, ctx.DATABASE_NAME)
			terminusBackupCreateCmd := exec.Command("terminus", "backup:create", ctx.SITE_NAME+"."+pantheonEnv, "--element=db")
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
			fmt.Printf("[%s] Downloading remote database backup...\n", ctx.SITE_NAME)
			terminusBackupGetCmd := exec.Command("terminus", "backup:get", ctx.SITE_NAME+"."+pantheonEnv, "--element=db")
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
			fmt.Printf("[%s] Importing database '%s' into sandbox...\n", ctx.SITE_NAME, ctx.DATABASE_NAME)

			_, err := exec.Command("bash", "-c", gunzipCmdString).Output()
			if err != nil {
				fmt.Printf("[%s] Something went wrong when restoring database.\n", ctx.SITE_NAME)
				log.Fatal(err)
			}
		}

		if isFlaggedVerbose {
			fmt.Printf("[%s] Cleaning up temp files...\n", ctx.SITE_NAME)
		}

		if isFlaggedSlow {
			err := os.Remove(tempFilePath)
			if err != nil {
				fmt.Printf("[%s] Something went wrong when cleaning up temp files in %s\n", ctx.SITE_NAME, tempFilePath)
				log.Fatal(err)
			}
		}

		if ctx.SITE_TYPE == "drupal" {
			fmt.Printf("[%s] Clearing Drupal cache...\n", ctx.SITE_NAME)

			drushArgs := []string{"cr"}
			if ctx.IS_DRUPAL7 {
				drushArgs = []string{"cc", "all"}
			}
			drushCmd := exec.Command("drush", drushArgs...)
			_ = drushCmd.Run()
		}

		fmt.Printf("[%s] Database pull complete.\n\n", ctx.SITE_NAME)
	},
}

func init() {
	pullCmd.AddCommand(pullDbCmd)

	pullDbCmd.PersistentFlags().String("name", "", "Pull a specific database by name")
}
