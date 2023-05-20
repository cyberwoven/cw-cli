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

type Context struct {
	Git  *Git
	Site *Site
	Ssh  *Ssh
}

type Ssh struct {
	Username string // "cyberwoven"
	Hostname string // "exmp.test.cyberwoven.net"
}

type Git struct {
	Branch    string // "dev/new-homepage"
	Domain    string // "bitbucket.org"
	Username  string // "git"
	Workspace string // "cyberwoven"
}

type Site struct {
	IsPantheon   bool   // false
	IsDrupal     bool   // true
	Domain       string // "www.example.com"
	ProjectRoot  string // "/home/username/Sites/www.example.com"
	DocumentRoot string // "/home/username/Sites/www.example.com/pub"
	DatabaseName string // "example"
	Framework    *Framework
}

type Framework struct {
	Name    string // "drupal"
	Version string // "9.5.3"
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
	if projectRoot == "" {
		return
	}

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

type FlatContext struct {
	SITES_DIR					string
	GIT_DEFAULT_DOMAIN			string
	GIT_DEFAULT_USER			string
	SSH_TEST_USER				string
	SSH_TEST_HOST				string
	IS_GIT_REPO					bool
	IS_SITE						bool
	HAS_DATABASE				bool
	IS_DRUPAL7					bool
	IS_PANTHEON					bool
	GIT_BRANCH					string
	GIT_BRANCH_SLUG				string
	SITE_NAME					string
	SITE_TYPE					string
	SITE_DOCUMENT_ROOT			string
	PROJECT_ROOT				string
	DATABASE_IMPORT_DIR			string
	DATABASE_NAME				string
	DATABASE_BASENAME			string
	DRUPAL_CORE_VERSION			string
	DRUPAL_DEFAULT_DIR_LOCAL	string
	DRUPAL_DEFAULT_DIR_REMOTE	string
	DRUPAL_PRIVATE_FILES_DIR	string
	DRUPAL_PUBLIC_FILES_DIR		string
	WP_UPLOADS_DIR_LOCAL		string
	WP_UPLOADS_DIR_REMOTE		string
}

func LoadContext() FlatContext {

	ctx := FlatContext{}

	isGitRepoCmd, err := exec.Command("/usr/bin/git", "rev-parse", "--is-inside-work-tree").Output()
	if err == nil {
		ctx.IS_GIT_REPO = strings.TrimSpace(string(isGitRepoCmd)) == "true"
	}

	if ctx.IS_GIT_REPO {
		gitBranchCmd, err := exec.Command("/usr/bin/git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err == nil {
			branchName := strings.TrimSpace(string(gitBranchCmd))
			branchSlug := strings.ReplaceAll(branchName, ".", "-")
			branchSlug = strings.ReplaceAll(branchSlug, "/", "-")
			branchSlug = strings.ReplaceAll(branchSlug, "_", "-")
			ctx.GIT_BRANCH = branchName
			ctx.GIT_BRANCH_SLUG = branchSlug

		}

		gitRootCmd, err := exec.Command("/usr/bin/git", "rev-parse", "--show-toplevel").Output()
		if err == nil {
			ctx.PROJECT_ROOT = strings.TrimSpace(string(gitRootCmd))
		}
	}

	InitViperConfigEnv()
	CheckLocalConfigOverrides(ctx.PROJECT_ROOT)

	USER_HOME_DIRECTORY, err := os.UserHomeDir()
	if err == nil {
		ctx.SITES_DIR = fmt.Sprintf("%s/%s", USER_HOME_DIRECTORY, viper.GetString("CWCLI_SITES_DIR"))
	}
	
	ctx.GIT_DEFAULT_DOMAIN = viper.GetString("CWCLI_GIT_DOMAIN")
	ctx.GIT_DEFAULT_USER = viper.GetString("CWCLI_GIT_USER")
	ctx.SSH_TEST_USER = viper.GetString("CWCLI_SSH_USER")
	ctx.SSH_TEST_HOST = viper.GetString("CWCLI_SSH_TEST_SERVER")

	cwd, err := os.Getwd()
    if err == nil {
		if strings.HasPrefix(cwd, ctx.SITES_DIR) {
			childDir := strings.TrimPrefix(cwd, ctx.SITES_DIR)
			
			/**
			 * if childDir == "/www.example.com/pub/sites/default"
			 *
			 *  dirParts[0] == ""
			 *  dirParts[1] == "www.example.com"
			 */
			dirParts := strings.Split(childDir, "/")
			siteName := dirParts[1]
			projectRoot := fmt.Sprintf("%s/%s", ctx.SITES_DIR, siteName)
			documentRoot := fmt.Sprintf("%s/%s", projectRoot, "pub")

			if _, err := os.Stat(documentRoot); err == nil {
				ctx.IS_SITE = true
				ctx.SITE_NAME = siteName
				ctx.PROJECT_ROOT = projectRoot
				ctx.SITE_DOCUMENT_ROOT = documentRoot
			}
		}
    }
    
	// dirty check for Pantheon
	if _, err := os.Stat(ctx.PROJECT_ROOT + "/pantheon.yml"); err == nil {
		ctx.IS_PANTHEON = true
	} 

	// dirty check for D7
	if _, err := os.Stat(ctx.SITE_DOCUMENT_ROOT + "/misc/drupal.js"); err == nil {
		ctx.IS_DRUPAL7 = true
	}

	// dirty check for wordpress
	if _, err := os.Stat(ctx.SITE_DOCUMENT_ROOT + "/wp-config.php"); err == nil {
		ctx.SITE_TYPE = "wordpress"
		ctx.WP_UPLOADS_DIR_LOCAL = fmt.Sprintf("%s/wp-content/uploads", ctx.SITE_DOCUMENT_ROOT)
		ctx.WP_UPLOADS_DIR_REMOTE = fmt.Sprintf("/var/www/vhosts/%s/%s/wp-content/uploads", ctx.SITE_NAME, ctx.GIT_BRANCH_SLUG)
	}

	/**
	 * Try detecting Drupal stuff now...only if we know:
	 *  - it's NOT wordpress
	 *  - there's a pub/index.php file present
	 */
	hasIndexPhp := false
	if _, err := os.Stat(ctx.SITE_DOCUMENT_ROOT + "/index.php"); err == nil {
		hasIndexPhp = true
	}

	if hasIndexPhp && ctx.SITE_TYPE != "wordpress" {

		drushCmd := exec.Command("drush", "status", "--format=json")
		drushCmdOutput, err := drushCmd.Output()
		if err == nil {
			var drushStatus DrushStatus
			json.Unmarshal([]byte(drushCmdOutput), &drushStatus)
	
			dbParts := strings.Split(drushStatus.DbName, "__")

			ctx.SITE_TYPE = "drupal"
			ctx.DRUPAL_CORE_VERSION = drushStatus.DrupalVersion
			ctx.DATABASE_NAME = drushStatus.DbName
			ctx.DATABASE_BASENAME = dbParts[0]
			
			//drush php:eval "echo \Drupal::service('file_system')->realpath('private://');"

			showPrivateFiles := "echo Drupal::service('file_system')->realpath('private://');"
			showPublicFiles  := "echo Drupal::service('file_system')->realpath('public://');"

			drushPrivateFiles, err := exec.Command("drush", "php:eval", showPrivateFiles).Output()
			if err == nil {
				ctx.DRUPAL_PRIVATE_FILES_DIR = strings.TrimSpace(string(drushPrivateFiles))
			}

			drushPublicFiles, err := exec.Command("drush", "php:eval", showPublicFiles).Output()
			if err == nil {
				ctx.DRUPAL_PUBLIC_FILES_DIR = strings.TrimSpace(string(drushPublicFiles))
			}


		}
		
	}
	
	
	return ctx
}

func FlatContextTest() {
	// create an empty Context struct
	ctx := LoadContext()
	prettyPrint(ctx)
}

// func ContextTest() {
// 	// create an empty Context struct
// 	ctx := Context{}
// 	prettyPrint(ctx)

// 	// create a Site struct in the Context
// 	// we use &Site{} here b/c the struct has *Site (a pointer to a Site struct).
// 	//
// 	// why *Site? b/c the Site might be null. Maybe we run `cw` in a dir
// 	// where there's no detectable site, so we dont' want a full Site struct
// 	// full of empty string variables. We just want the ctx.Site to end up being nil
// 	ctx.Site = &Site{}
// 	prettyPrint(ctx)

// 	// we can directly assign the Domain b/c it's just a normal string
// 	ctx.Site.Domain = "www.example.com"

// 	// but we have to assign Database to a var, then assign the
// 	// address of that var to the struct's database member,
// 	// since it's a pointer to a string.
// 	//
// 	// why make it a pointer in the struct? so it can be be nil,
// 	// like for a static site where there's no db at all
// 	var database = "exampledb"
// 	ctx.Site.DatabaseName = database

// 	// same thing with *Git. Maybe there's no git repo, but it's a legitimate site
// 	// that only exists on the local machine.

// 	prettyPrint(ctx)
// }

func prettyPrint(ctx FlatContext) {
	ctxJson, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("\n%s\n", string(ctxJson))
}
