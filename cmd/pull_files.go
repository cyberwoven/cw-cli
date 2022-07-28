/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	// "strings"

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

		} else {
			fmt.Println("PANTHEON: creating files tarball ...")
			// terminusCmd := exec.Command("drush", "status", "--fields=db-name,root,drupal-version", "--format=list")
			// out, err := terminusCmd.Output()
			// if err != nil {
			// 	log.Fatal(err)
			// 	fmt.Println("is_pantheon error")
			// } else {
			// 	fmt.Println(out)
			// }
		}

		// strings.Index(drupal_root, "")

		// filesCmd.Stdout = os.Stdout
		// filesCmd.Stderr = os.Stderr
		// filesCmd.Start()
		// filesCmd.Wait()
		// thing := fmt.Sprint(filesCmd.Stdout)
		// stuff := strings.Split(thing, "\n")
		// fmt.Printf("%s", stuff)

		// dbname, root dir, drupal version
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
