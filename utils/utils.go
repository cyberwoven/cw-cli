package cwutils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type DrushStatus struct {
	DbName        string `json:"db-name"`
	DrupalVersion string `json:"drupal-version"`
	Root          string `json:"root"`
}

type Context struct {
	USERNAME                  string
	HOME_DIR                  string
	DEFAULT_CONFIG_DIR        string
	PROJECT_CONFIG_FILE       string
	SITES_DIR                 string
	GIT_DEFAULT_USER          string
	GIT_DEFAULT_DOMAIN        string
	GIT_DEFAULT_WORKSPACE     string
	SSH_TEST_USER             string
	SSH_TEST_HOST             string
	IS_GIT_REPO               bool
	IS_SITE                   bool
	HAS_DATABASE              bool
	IS_DRUPAL7                bool
	IS_PANTHEON               bool
	GIT_BRANCH                string
	GIT_BRANCH_SLUG           string
	SITE_NAME                 string
	SITE_TYPE                 string
	SITE_DOCUMENT_ROOT        string
	PROJECT_ROOT              string
	DATABASE_IMPORT_DIR       string
	DATABASE_NAME             string
	DATABASE_BASENAME         string
	DRUPAL_CORE_VERSION       string
	DRUPAL_DEFAULT_DIR_LOCAL  string
	DRUPAL_DEFAULT_DIR_REMOTE string
	DRUPAL_PRIVATE_FILES_DIR  string
	DRUPAL_PUBLIC_FILES_DIR   string
	WP_UPLOADS_DIR_LOCAL      string
	WP_UPLOADS_DIR_REMOTE     string
	TASK_URL_PREFIX           string
}

func GetContext() Context {
	ctx := Context{}

	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	ctx.USERNAME = user.Username

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

	USER_HOME_DIRECTORY, err := os.UserHomeDir()
	if err == nil {
		ctx.HOME_DIR = USER_HOME_DIRECTORY
	}

	var configVars map[string]string
	defaultConfigPath := ctx.HOME_DIR + "/.cw/config"
	projectConfigPath := ctx.PROJECT_ROOT + "/.cw/config"

	if _, err := os.Stat(projectConfigPath); err == nil {
		ctx.PROJECT_CONFIG_FILE = projectConfigPath
		configVars, err = godotenv.Read(defaultConfigPath, projectConfigPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		configVars, err = godotenv.Read(defaultConfigPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	sitesDir := configVars["SITES_DIR"]
	if strings.HasPrefix(sitesDir, "~/") {
		ctx.SITES_DIR = fmt.Sprintf("%s/%s", ctx.HOME_DIR, sitesDir[2:])
	} else if !strings.HasPrefix(sitesDir, "/") {
		ctx.SITES_DIR = fmt.Sprintf("%s/%s", ctx.HOME_DIR, sitesDir)
	} else {
		ctx.SITES_DIR = sitesDir
	}

	ctx.DEFAULT_CONFIG_DIR = filepath.Join(USER_HOME_DIRECTORY, ".cw")
	ctx.GIT_DEFAULT_USER = configVars["GIT_USER"]
	ctx.GIT_DEFAULT_DOMAIN = configVars["GIT_DOMAIN"]
	ctx.GIT_DEFAULT_WORKSPACE = configVars["GIT_WORKSPACE"]
	ctx.SSH_TEST_USER = configVars["SSH_USER"]
	ctx.SSH_TEST_HOST = configVars["SSH_TEST_SERVER"]
	ctx.TASK_URL_PREFIX = configVars["TASK_URL_PREFIX"]
	ctx.DATABASE_NAME = configVars["DATABASE_NAME"]
	if ctx.DATABASE_NAME != "" {
		ctx.HAS_DATABASE = true
	}

	if configVars["DATABASE_IMPORT_DIR"] == "" {
		ctx.DATABASE_IMPORT_DIR = ctx.HOME_DIR + "/.cw/database_dumps"
	} else {
		ctx.DATABASE_IMPORT_DIR = configVars["DATABASE_IMPORT_DIR"]

		if strings.HasPrefix(configVars["DATABASE_IMPORT_DIR"], "~/") {
			ctx.DATABASE_IMPORT_DIR = fmt.Sprintf("%s/%s", ctx.HOME_DIR, configVars["DATABASE_IMPORT_DIR"][2:])
		} else if !strings.HasPrefix(configVars["DATABASE_IMPORT_DIR"], "/") {
			ctx.DATABASE_IMPORT_DIR = fmt.Sprintf("%s/%s", ctx.HOME_DIR, configVars["DATABASE_IMPORT_DIR"])
		} else {
			ctx.DATABASE_IMPORT_DIR = configVars["DATABASE_IMPORT_DIR"]
		}
	}

	cwd, err := os.Getwd()
	if err == nil && cwd != ctx.SITES_DIR && strings.HasPrefix(cwd, ctx.SITES_DIR) {
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

	// dirty check for Pantheon
	if _, err := os.Stat(ctx.PROJECT_ROOT + "/pantheon.yml"); err == nil {
		ctx.IS_PANTHEON = true
	}

	// dirty check for wordpress
	if _, err := os.Stat(ctx.SITE_DOCUMENT_ROOT + "/wp-config.php"); err == nil {
		ctx.SITE_TYPE = "wordpress"
		ctx.WP_UPLOADS_DIR_LOCAL = fmt.Sprintf("%s/wp-content/uploads", ctx.SITE_DOCUMENT_ROOT)
		ctx.WP_UPLOADS_DIR_REMOTE = fmt.Sprintf("/var/www/vhosts/%s/%s/pub/wp-content/uploads", ctx.SITE_NAME, ctx.GIT_BRANCH_SLUG)

		wpDbCmd := exec.Command("wp", "config", "get", "DB_NAME")
		wpDbCmd.Dir = ctx.SITE_DOCUMENT_ROOT
		wpDbCmdOutput, err := wpDbCmd.Output()
		if err == nil {
			ctx.DATABASE_NAME = strings.TrimSpace(string(wpDbCmdOutput))
		}
	}

	// site has an index.php, but isn't WP? most likely it's Drupal, so
	// let's try to load the drupal vars via drush
	hasIndexPhp := false
	if _, err := os.Stat(ctx.SITE_DOCUMENT_ROOT + "/index.php"); err == nil {
		hasIndexPhp = true
	}

	var drushStatus DrushStatus
	if hasIndexPhp && ctx.SITE_TYPE != "wordpress" {
		drushCmd := exec.Command("drush", "status", "--format=json")
		drushCmdOutput, err := drushCmd.Output()
		if err == nil {
			json.Unmarshal([]byte(drushCmdOutput), &drushStatus)
		}
	}

	if drushStatus.DrupalVersion != "" {
		dbParts := strings.Split(drushStatus.DbName, "__")

		ctx.SITE_TYPE = "drupal"
		ctx.DRUPAL_CORE_VERSION = drushStatus.DrupalVersion
		ctx.DATABASE_NAME = drushStatus.DbName
		ctx.DATABASE_BASENAME = dbParts[0]

		if string(ctx.DRUPAL_CORE_VERSION[0]) == "7" {
			ctx.IS_DRUPAL7 = true
		}

		ctx.DRUPAL_DEFAULT_DIR_LOCAL = fmt.Sprintf("%s/sites/default", ctx.SITE_DOCUMENT_ROOT)
		ctx.DRUPAL_DEFAULT_DIR_REMOTE = fmt.Sprintf("/var/www/vhosts/%s/%s/pub/sites/default", ctx.SITE_NAME, ctx.GIT_BRANCH_SLUG)

		showPrivateFiles := "echo Drupal::service('file_system')->realpath('private://');"
		showPublicFiles := "echo Drupal::service('file_system')->realpath('public://');"

		drushPrivateFiles, err := exec.Command("drush", "php:eval", showPrivateFiles).Output()
		if err == nil {
			ctx.DRUPAL_PRIVATE_FILES_DIR = strings.TrimSpace(string(drushPrivateFiles))
		}

		ctx.DRUPAL_PUBLIC_FILES_DIR = ctx.DRUPAL_DEFAULT_DIR_LOCAL + "/files"
		drushPublicFiles, err := exec.Command("drush", "php:eval", showPublicFiles).Output()
		publicFiles := strings.TrimSpace(string(drushPublicFiles))
		if err == nil && publicFiles != "" {
			ctx.DRUPAL_PUBLIC_FILES_DIR = publicFiles
		}
	}

	if hasIndexPhp && ctx.SITE_TYPE == "" {
		ctx.SITE_TYPE = "custom"
	}

	return ctx
}

func ContextTest() {
	// create an empty Context struct
	ctx := GetContext()
	PrettyPrint(ctx)
}

func PrettyPrint(ctx Context) {
	ctxJson, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("\n%s\n", string(ctxJson))
}
