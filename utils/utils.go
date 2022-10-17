/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cwutils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/viper"
)

type DrushStatus struct {
	DbName        string `json:"db-name"`
	DrupalVersion string `json:"drupal-version"`
	Root          string `json:"root"`
}

type CwVars struct {
	Is_pantheon        bool
	Drupal_dbname      string
	Drupal_root        string
	Drupal_version     string
	Drupal_site_name   string
	Project_root       string
	Domain_local       string
	Domain_forest      string
	Branch_name        string
	DEFAULT_DIR_FOREST string
	DEFAULT_DIR_LOCAL  string
}

func GetProjectVars() CwVars {
	filesCmd := exec.Command("drush", "status", "--format=json")
	filesCmdOutput, err := filesCmd.Output()
	if err != nil {
		fmt.Println("drush error")
		log.Fatal(err)
	}

	var drushStatus DrushStatus
	json.Unmarshal([]byte(filesCmdOutput), &drushStatus)

	// strArr := strings.Split(string(filesCmdOutput), "\n")
	// drupal_dbname := strArr[0]
	// drupal_root := strArr[1]
	// drupal_version := strArr[2]

	is_pantheon := false
	drupal_site_name := path.Base(path.Dir(drushStatus.Root))
	domain_local := drupal_site_name + ".test"
	project_root := path.Dir(drushStatus.Root)

	forestCmd, err := exec.Command("/usr/bin/git", "-C", drushStatus.Root, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		fmt.Println("[error] get forest domain failed")
		log.Fatal(err)
	}

	domain_forest := strings.TrimSpace(string(forestCmd))
	domain_forest = path.Base(domain_forest)
	domain_forest = strings.Replace(domain_forest, ".git", "", 1)

	branchCmd, err := exec.Command("/usr/bin/git", "-C", drushStatus.Root, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		fmt.Println("[error] get branch name failed")
		log.Fatal(err)
	}

	branch_name := strings.TrimSpace(string(branchCmd))
	branch_name = strings.ReplaceAll(branch_name, ".", "-")
	branch_name = strings.ReplaceAll(branch_name, "/", "-")
	branch_name = strings.ReplaceAll(branch_name, "_", "-")
	DEFAULT_DIR_LOCAL := fmt.Sprintf("%s/pub/sites/default", project_root)
	DEFAULT_DIR_FOREST := fmt.Sprintf("/var/www/vhosts/%s/%s/pub/sites/default", domain_forest, branch_name)

	// fmt.Println(DEFAULT_DIR_LOCAL)
	// fmt.Println(DEFAULT_DIR_FOREST)
	// fmt.Println("[drupal_site_name]: ", drupal_site_name)
	// fmt.Println("[project_root]: ", project_root)
	// fmt.Println("[drupal_root]: ", drushStatus.Root)
	// fmt.Println("[drupal_dbname]: ", drushStatus.DbName)
	// fmt.Println("[drupal_version]: ", drushStatus.DrupalVersion)
	// os.Exit(1)
	// fmt.Println("[domain_local]: ", domain_local)
	// fmt.Println("[domain_forest]: ", domain_forest)
	// fmt.Println("[branch_name]: ", branch_name)
	// fmt.Println("[DEFAULT_DIR_FOREST]: ", DEFAULT_DIR_FOREST)
	// fmt.Println("[DEFAULT_DIR_LOCAL]: ", DEFAULT_DIR_LOCAL)

	if strings.Contains(drushStatus.Root, "/web") {
		is_pantheon = true
	}

	var vars = CwVars{
		Is_pantheon:        is_pantheon,
		Drupal_dbname:      drushStatus.DbName,
		Drupal_root:        drushStatus.Root,
		Drupal_version:     drushStatus.DrupalVersion,
		Drupal_site_name:   drupal_site_name,
		Project_root:       project_root,
		Domain_local:       domain_local,
		Domain_forest:      domain_forest,
		Branch_name:        branch_name,
		DEFAULT_DIR_FOREST: DEFAULT_DIR_FOREST,
		DEFAULT_DIR_LOCAL:  DEFAULT_DIR_LOCAL,
	}

	return vars
}

// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func InitViperConfigEnv() {
	USER_HOME_DIR, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	viper.SetEnvPrefix("CWCWLI")
	viper.AutomaticEnv()
	viper.SetConfigName("default")
	viper.SetConfigType("env")
	viper.AddConfigPath(fmt.Sprintf("%s/.cw", USER_HOME_DIR))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found in ~/.cw")
			log.Fatal(err)
		} else {
			fmt.Println("Config file was found but another error was produced")
			log.Fatal(err)
		}
	}
}

func CheckLocalConfigOverrides(projectRoot string) {
	viper.SetConfigName("config")
	viper.AddConfigPath(projectRoot + "/.cw")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if viper.GetBool("verbose") {
				fmt.Println("Local Config file not found in project root, using default cw-cli config...")
			}
		} else {
			fmt.Println("Local Config file was found but another error was produced...")
			log.Fatal(err)
		}
	}
}
