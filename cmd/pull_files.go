/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	cwutils "cw-cli/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// pullFilesCmd represents the pullFiles command
var pullFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Pull files from test to sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
			log.Fatal(err)
		}
		var vars = cwutils.GetProjectVars()
		cwutils.InitViperConfigEnv()
		cwutils.CheckLocalConfigOverrides(vars.Project_root)
		var sshUsername string = viper.GetString("CWCLI_SSH_USER")
		var sshServerUrl string = viper.GetString("CWCLI_SSH_TEST_SERVER")
		var rsyncRemote string = fmt.Sprintf("%s@%s:%s/files", sshUsername, sshServerUrl, vars.DEFAULT_DIR_FOREST)

		if !vars.Is_pantheon {
			rsyncCmd := exec.Command("rsync",
				"-vcrtzP",
				rsyncRemote,
				vars.DEFAULT_DIR_LOCAL,
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

			// fmt.Println("[Running Command]: " + rsyncCmd.String())
			fmt.Printf("[%s] Pulling down files...\n", vars.Drupal_site_name)

			_ = rsyncCmd.Start()

			scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				m := scanner.Text()
				fmt.Println(m)
			}
			_ = rsyncCmd.Wait()

		} else {

			// CREATE BACKUP =========================================================
			fmt.Printf("[%s]: creating files tarball... this may take a minute...\n", vars.Drupal_site_name)
			terminusBackupCreateCmd := exec.Command("terminus", "backup:create", vars.Drupal_site_name+".dev", "--element=files")
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
			fmt.Printf("[%s]: downloading files tarball...\n", vars.Drupal_site_name)
			terminusBackupGetCmd := exec.Command("terminus", "backup:get", vars.Drupal_site_name+".dev", "--element=files")
			// fmt.Println("[Running Command]: " + terminusBackupGetCmd.String())
			backupGetStdout, _ := terminusBackupGetCmd.Output()
			// fmt.Println("[PANTHEON]: backup url - " + string(backupGetStdout))
			wgetCmd := exec.Command("wget", "--quiet", "--show-progress", strings.TrimSpace(string(backupGetStdout)), "-O", "/tmp/files_"+vars.Drupal_dbname+".tar.gz")
			wgetStdout, _ := wgetCmd.StdoutPipe()
			wgetStderr, _ := wgetCmd.StderrPipe()
			// fmt.Println("[Running Command]: " + wgetCmd.String())
			_ = wgetCmd.Start()
			wgetScanner := bufio.NewScanner(io.MultiReader(wgetStderr, wgetStdout))
			wgetScanner.Split(bufio.ScanLines)
			for wgetScanner.Scan() {
				m := wgetScanner.Text()
				fmt.Println(m)
			}
			_ = wgetCmd.Wait()

			// EXTRACT BACKUP ========================================================
			if viper.GetBool("verbose") {
				fmt.Printf("[%s]: extracting files tarball...\n", vars.Drupal_site_name)
			}
			mkdirCmd := exec.Command("mkdir", "/tmp/files_"+vars.Drupal_dbname)
			_ = mkdirCmd.Run()
			tarCmd := exec.Command("tar", "--directory=/tmp/files_"+vars.Drupal_dbname, "-xzvf", "/tmp/files_"+vars.Drupal_dbname+".tar.gz")
			_ = tarCmd.Run()
			// COPY EXTRACTED FILES ==================================================
			if viper.GetBool("verbose") {
				fmt.Printf("[%s]: copying files into 'sites/default/files/' directory...\n", vars.Drupal_site_name)
			}
			copyExtractCmd := exec.Command("rsync", "-vcrP", "--stats", "/tmp/files_"+vars.Drupal_dbname+"/files_dev/", vars.Drupal_root+"/sites/default/files/")
			_ = copyExtractCmd.Run()
		}

		fmt.Printf("[%s]: Finished pulling down files to local environment!\n\n", vars.Drupal_site_name)
	},
}

func init() {
	pullCmd.AddCommand(pullFilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullFilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullFilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-exists
// func exists(path string) (bool, error) {
// 	_, err := os.Stat(path)
// 	if err == nil {
// 		return true, nil
// 	}
// 	if os.IsNotExist(err) {
// 		return false, nil
// 	}
// 	return false, err
// }
