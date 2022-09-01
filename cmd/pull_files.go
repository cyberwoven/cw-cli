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

// pullFilesCmd represents the pullFiles command
var pullFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Pull files from test to sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		var vars = cwutils.GetProjectVars()

		if !vars.Is_pantheon {
			rsyncCmd := exec.Command("rsync",
				"-vcrtzP",
				"forest-web:"+vars.DEFAULT_DIR_FOREST+"/files",
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

			fmt.Println("[Running Command]: " + rsyncCmd.String())

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
			fmt.Println("[PANTHEON]: creating files tarball... this may take a minute...")
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
			fmt.Println("[PANTHEON]: downloading files tarball...")
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
			fmt.Println("[PANTHEON]: extracting files tarball...")
			mkdirCmd := exec.Command("mkdir", "/tmp/files_"+vars.Drupal_dbname)
			_ = mkdirCmd.Run()
			tarCmd := exec.Command("tar", "--directory=/tmp/files_"+vars.Drupal_dbname, "-xzvf", "/tmp/files_"+vars.Drupal_dbname+".tar.gz")
			_ = tarCmd.Run()
			// COPY EXTRACTED FILES ==================================================
			fmt.Println("[PANTHEON]: copying files into 'sites/default/files/' directory...")
			copyExtractCmd := exec.Command("rsync", "-vcrP", "--stats", "/tmp/files_"+vars.Drupal_dbname+"/files_dev/", vars.Drupal_root+"/sites/default/files/")
			_ = copyExtractCmd.Run()
			fmt.Println("[PANTHEON]: Finished pulling down files to local environment!")
		}
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
