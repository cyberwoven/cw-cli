/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

// pullFilesCmd represents the pullFiles command
var pullFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Pull files from test to sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		filesCmd := exec.Command("drush", "status", "--fields=db-name,root,drupal-version", "--format=list")
		filesCmdOutput, err := filesCmd.Output()
		if err != nil {
			log.Fatal(err)
			fmt.Println("drush error")
		}

		is_pantheon := false
		strArr := strings.Split(string(filesCmdOutput), "\n")
		drupal_dbname := strArr[0]
		drupal_root := strArr[1]
		drupal_version := strArr[2]
		drupal_site_name := path.Base(path.Dir(drupal_root))
		domain_local := drupal_site_name + ".test"
		project_root := path.Dir(drupal_root)

		forestCmd, err := exec.Command("/usr/bin/git", "-C", drupal_root, "config", "--get", "remote.origin.url").Output()
		if err != nil {
			log.Fatal(err)
			fmt.Println("[error] get forest domain failed")
		}
		domain_forest := strings.TrimSpace(string(forestCmd))
		domain_forest = path.Base(domain_forest)
		domain_forest = strings.Replace(domain_forest, ".git", "", 1)

		branchCmd, err := exec.Command("/usr/bin/git", "-C", drupal_root, "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err != nil {
			log.Fatal(err)
			fmt.Println("[error] get branch name failed")
		}
		branch_name := strings.TrimSpace(string(branchCmd))
		branch_name = strings.ReplaceAll(branch_name, ".", "-")
		branch_name = strings.ReplaceAll(branch_name, "/", "-")
		branch_name = strings.ReplaceAll(branch_name, "_", "-")

		// brewCmd, err := exec.Command("brew", "--prefix").Output()
		// if err != nil {
		// 	log.Fatal(err)
		// 	fmt.Println("[error] get branch name failed")
		// }
		// brew := strings.TrimSpace(string(brewCmd))

		// DEFAULT_DIR_LOCAL := brew + `/var/www/vhosts/` + domain_local + `var`
		// DEFAULT_DIR_FOREST := `/var/www/vhosts/` + `/$BRANCH/pub/sites/default`
		DEFAULT_DIR_LOCAL := fmt.Sprintf("%s/pub/sites/default", project_root)
		DEFAULT_DIR_FOREST := fmt.Sprintf("/var/www/vhosts/%s/%s/pub/sites/default", domain_forest, branch_name)
		fmt.Println(DEFAULT_DIR_LOCAL)
		fmt.Println(DEFAULT_DIR_FOREST)

		fmt.Println("[drupal_site_name]: ", drupal_site_name)
		fmt.Println("[project_root]: ", project_root)
		fmt.Println("[drupal_root]: ", drupal_root)
		fmt.Println("[drupal_dbname]: ", drupal_dbname)
		fmt.Println("[drupal_version]: ", drupal_version)
		fmt.Println("[domain_local]: ", domain_local)
		fmt.Println("[domain_forest]: ", domain_forest)
		fmt.Println("[branch_name]: ", branch_name)
		fmt.Println("[DEFAULT_DIR_FOREST]: ", DEFAULT_DIR_FOREST)
		fmt.Println("[DEFAULT_DIR_LOCAL]: ", DEFAULT_DIR_LOCAL)

		if strings.Contains(drupal_root, "/web") {
			is_pantheon = true
		}

		isDirLocalExist, err := exists(DEFAULT_DIR_LOCAL)
		if err != nil {
			log.Fatal(err)
			fmt.Println("[error] get branch name failed")
		}

		// fmt.Println(DEFAULT_DIR_LOCAL)
		// fmt.Println(isDirLocalExist)

		if !isDirLocalExist {
			log.Fatal("ABORTING! Local default dir does not exist: " + DEFAULT_DIR_LOCAL)
			os.Exit(1)
		}

		if !is_pantheon {
			rsyncCmd := exec.Command("rsync",
				"-vcrtzP",
				"forest-web:"+DEFAULT_DIR_FOREST+"/files",
				DEFAULT_DIR_LOCAL,
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
			terminusBackupCreateCmd := exec.Command("terminus", "backup:create", drupal_site_name+".dev", "--element=files")
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
			terminusBackupGetCmd := exec.Command("terminus", "backup:get", drupal_site_name+".dev", "--element=files")
			// fmt.Println("[Running Command]: " + terminusBackupGetCmd.String())
			backupGetStdout, _ := terminusBackupGetCmd.Output()
			// fmt.Println("[PANTHEON]: backup url - " + string(backupGetStdout))
			wgetCmd := exec.Command("wget", "--quiet", "--show-progress", strings.TrimSpace(string(backupGetStdout)), "-O", "/tmp/files_"+drupal_dbname+".tar.gz")
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

			// EXTRACT BACKUP ========================================================
			fmt.Println("[PANTHEON]: extracting files tarball...")
			mkdirCmd := exec.Command("mkdir", "/tmp/files_"+drupal_dbname)
			_ = mkdirCmd.Run()
			tarCmd := exec.Command("tar", "--directory=/tmp/files_"+drupal_dbname, "-xzvf", "/tmp/files_"+drupal_dbname+".tar.gz")
			_ = tarCmd.Run()
			// COPY EXTRACTED FILES ==================================================
			fmt.Println("[PANTHEON]: copying files into 'sites/default/files/' directory...")
			copyExtractCmd := exec.Command("rsync", "-vcrP", "--stats", "/tmp/files_"+drupal_dbname+"/files_dev/", drupal_root+"/sites/default/files/")
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
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
