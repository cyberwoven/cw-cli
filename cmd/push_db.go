package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var pushDbCmd = &cobra.Command{
	Use:   "db",
	Short: "Push db from local up to test",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		if ctx.IS_PANTHEON {
			fmt.Print("ABORT - push db is not implemented for Pantheon sites!")
			os.Exit(1)
		}

		if ctx.DATABASE_NAME == "" {
			fmt.Print("ABORT - unable to push db, no database name found for this site!")
			os.Exit(1)
		}

		fmt.Println(" - Dump local database: " + ctx.DATABASE_NAME)
		databaseDumpFile := fmt.Sprintf("%s/database_dumps/%s-push.sql.gz", ctx.DEFAULT_CONFIG_DIR, ctx.DATABASE_NAME)
		err = exec.Command(
			"sh",
			"-c",
			fmt.Sprintf("mysqldump %s | gzip > %s", ctx.DATABASE_NAME, databaseDumpFile),
		).Run()
		if err != nil {
			fmt.Printf("Unable to dump data for database %s\n", ctx.DATABASE_NAME)
			log.Fatal(err)
		}

		//
		// Dump local DB
		//
		// fmt.Println(" - Dump local database [schema]: " + ctx.DATABASE_NAME)
		// databaseDumpDir := ctx.DEFAULT_CONFIG_DIR + "/database_dumps/" + ctx.DATABASE_NAME
		// err = exec.Command(
		// 	"mydumper",
		// 	"--no-data",
		// 	"--user", ctx.USERNAME,
		// 	"--database", ctx.DATABASE_NAME,
		// 	"--outputdir", databaseDumpDir,
		// ).Run()
		// if err != nil {
		// 	fmt.Printf("Unable to dump schema for database %s\n", ctx.DATABASE_NAME)
		// 	log.Fatal(err)
		// }

		// fmt.Println(" - Dump local database [data]: " + ctx.DATABASE_NAME)
		// err = exec.Command(
		// 	"mydumper",
		// 	"--no-schemas",
		// 	"--user", ctx.USERNAME,
		// 	"--database", ctx.DATABASE_NAME,
		// 	"--regex", "^(?!(${DB}.watchdog|${DB}.cache|${DB}.session))",
		// 	"--outputdir", databaseDumpDir,
		// ).Run()
		// if err != nil {
		// 	fmt.Printf("Unable to dump data for database %s\n", ctx.DATABASE_NAME)
		// 	log.Fatal(err)
		// }

		fmt.Println(" - rsync database files up to remote")
		rsyncDest := fmt.Sprintf("%s@%s:~/backups/transient", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST)
		err = exec.Command("rsync", databaseDumpFile, rsyncDest).Run()
		if err != nil {
			fmt.Printf("Unable to rsync database files up to remote host for database %s, local dump file: %s\n", ctx.DATABASE_NAME, databaseDumpFile)
			log.Fatal(err)
		}

		//
		// Load database on remote server
		//
		fmt.Println(" - Load remote database: " + ctx.DATABASE_NAME)
		remoteLoadCmd := fmt.Sprintf("bash ~/bin/database-load.sh %s", ctx.DATABASE_NAME)
		// remoteBackupFile := fmt.Sprintf("~/backups/transient/%s-push.sql.gz", ctx.DATABASE_NAME)
		// remoteLoadCmd := fmt.Sprintf("mysqladmin create %s; gunzip < %s | mysql %s", ctx.DATABASE_NAME, remoteBackupFile, ctx.DATABASE_NAME)
		remoteHost := fmt.Sprintf("%s@%s", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST)
		exec.Command("ssh", remoteHost, remoteLoadCmd).Run()

		//
		// Try to flush remote cache if this is a drupal site.
		//
		if ctx.SITE_TYPE == "drupal" {
			fmt.Println(" - Attempting remote drush cache rebuild: " + ctx.DRUPAL_DEFAULT_DIR_REMOTE)
			remoteLoadCmd := fmt.Sprintf("~/bin/drush cr --root=%s", ctx.DRUPAL_DEFAULT_DIR_REMOTE)
			exec.Command("ssh", remoteHost, remoteLoadCmd).Run()
		}
	},
}

func init() {
	pushCmd.AddCommand(pushDbCmd)
}
