package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure cw cli",

	Run: func(cmd *cobra.Command, args []string) {

		defaultConfigPath := ctx.HOME_DIR + "/.cw/config"

		// check if there is something to read on STDIN
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {

			input, err := godotenv.Parse(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}

			output, err := godotenv.Marshal(input)
			if err != nil {
				log.Fatal(err)
			}

			err = os.WriteFile(defaultConfigPath, []byte(output), 0644)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// no text coming in from STDIN, so prompt for each config var interactively

			var configVars map[string]string

			configVars, err := godotenv.Read(defaultConfigPath)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Specify new configuration values. Leave blank to keep existing values.")
			fmt.Println("")

			fmt.Println("\n\033[1m──────────────────────\033[0m")
			fmt.Println("\033[1mBitbucket App Password (1 of 15)\033[0m")
			fmt.Println(" • Used to create new git repositories in your Bitbucket account.")
			fmt.Println(" • See https://bitbucket.org/account/settings/app-passwords/")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["BITBUCKET_APP_PASSWORD"])
			fmt.Print("       New » ")
			var newBitbucketAppPassword string
			fmt.Scanf("%s", &newBitbucketAppPassword)
			if newBitbucketAppPassword != "" {
				configVars["BITBUCKET_APP_PASSWORD"] = newBitbucketAppPassword
			}

			fmt.Println("\n\033[1m────────────────────────\033[0m")
			fmt.Println("\033[1mBitbucket Drupal Project (2 of 15)\033[0m")
			fmt.Println(" • Default project key for new Drupal sites.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["BITBUCKET_PROJECT_DRUPAL"])
			fmt.Print("       New » ")
			var newBitbucketProjectDrupal string
			fmt.Scanf("%s", &newBitbucketProjectDrupal)
			if newBitbucketProjectDrupal != "" {
				configVars["BITBUCKET_PROJECT_DRUPAL"] = newBitbucketProjectDrupal
			}

			fmt.Println("\n\033[1m────────────────────────\033[0m")
			fmt.Println("\033[1mBitbucket Static Project (3 of 15)\033[0m")
			fmt.Println(" • Default project key for new static sites.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["BITBUCKET_PROJECT_STATIC"])
			fmt.Print("       New » ")
			var newBitbucketProjectStatic string
			fmt.Scanf("%s", &newBitbucketProjectStatic)
			if newBitbucketProjectStatic != "" {
				configVars["BITBUCKET_PROJECT_STATIC"] = newBitbucketProjectStatic
			}

			fmt.Println("\n\033[1m───────────────────────────\033[0m")
			fmt.Println("\033[1mBitbucket Wordpress Project (4 of 15)\033[0m")
			fmt.Println(" • Default project key for new Wordpress sites.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["BITBUCKET_PROJECT_WORDPRESS"])
			fmt.Print("       New » ")
			var newBitbucketProjectWordpress string
			fmt.Scanf("%s", &newBitbucketProjectWordpress)
			if newBitbucketProjectWordpress != "" {
				configVars["BITBUCKET_PROJECT_WORDPRESS"] = newBitbucketProjectWordpress
			}

			fmt.Println("\n\033[1m──────────────────\033[0m")
			fmt.Println("\033[1mBitbucket Username (5 of 15)\033[0m")
			fmt.Println(" • Default Bitbucket username used for creating repos.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["BITBUCKET_USERNAME"])
			fmt.Print("       New » ")
			var newBitbucketUsername string
			fmt.Scanf("%s", &newBitbucketUsername)
			if newBitbucketUsername != "" {
				configVars["BITBUCKET_USERNAME"] = newBitbucketUsername
			}

			fmt.Println("\n\033[1m─────────────────────\033[0m")
			fmt.Println("\033[1mBitbucket Webhook URL (6 of 15)\033[0m")
			fmt.Println(" • Default Autopilot webhook URL for new repos.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["BITBUCKET_WEBHOOK_URL"])
			fmt.Print("       New » ")
			var newBitbucketWebhookUrl string
			fmt.Scanf("%s", &newBitbucketWebhookUrl)
			if newBitbucketWebhookUrl != "" {
				configVars["BITBUCKET_WEBHOOK_URL"] = newBitbucketWebhookUrl
			}

			fmt.Println("\n\033[1m───────────────────\033[0m")
			fmt.Println("\033[1mBitbucket Workspace (7 of 15)\033[0m")
			fmt.Println(" • Default Workspace to use when creating repos.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["BITBUCKET_WORKSPACE"])
			fmt.Print("       New » ")
			var newBitbucketWorkspace string
			fmt.Scanf("%s", &newBitbucketWorkspace)
			if newBitbucketWorkspace != "" {
				configVars["BITBUCKET_WORKSPACE"] = newBitbucketWorkspace
			}

			fmt.Println("\n\033[1m───────────────────\033[0m")
			fmt.Println("\033[1mDatabase Import Dir (8 of 15)\033[0m")
			fmt.Println(" • Local directory for storing database dumps.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["DATABASE_IMPORT_DIR"])
			fmt.Print("       New » ")
			var newDatabaseImportDir string
			fmt.Scanf("%s", &newDatabaseImportDir)
			if newDatabaseImportDir != "" {
				configVars["DATABASE_IMPORT_DIR"] = newDatabaseImportDir
			}

			fmt.Println("\n\033[1m───────────────────\033[0m")
			fmt.Println("\033[1mGit Domain (9 of 15)\033[0m")
			fmt.Println(" • Default domain to use when cloning remote git repositories.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["GIT_DOMAIN"])
			fmt.Print("       New » ")
			var newGitDomain string
			fmt.Scanf("%s", &newGitDomain)
			if newGitDomain != "" {
				configVars["GIT_DOMAIN"] = newGitDomain
			}

			fmt.Println("\n\033[1m────────\033[0m")
			fmt.Println("\033[1mGit User (10 of 15)\033[0m")
			fmt.Println(" • Default user to use when cloning remote git repositories.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["GIT_USER"])
			fmt.Print("       New » ")
			var newGitUser string
			fmt.Scanf("%s", &newGitUser)
			if newGitUser != "" {
				configVars["GIT_USER"] = newGitUser
			}

			fmt.Println("\n\033[1m───────────────\033[0m")
			fmt.Println("\033[1mSites directory (11 of 15)\033[0m")
			fmt.Println(" • Local directory that sites are cloned into.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["SITES_DIR"])
			fmt.Print("       New » ")
			var newSitesDir string
			fmt.Scanf("%s", &newSitesDir)
			if newSitesDir != "" {
				configVars["SITES_DIR"] = newSitesDir
			}

			fmt.Println("\n\033[1m───────────────\033[0m")
			fmt.Println("\033[1mSSH Test Server (12 of 15)\033[0m")
			fmt.Println(" • Hostname of default test server that databases and files are pulled from.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["SSH_TEST_SERVER"])
			fmt.Print("       New » ")
			var newSshTestServer string
			fmt.Scanf("%s", &newSshTestServer)
			if newSshTestServer != "" {
				configVars["SSH_TEST_SERVER"] = newSshTestServer
			}

			fmt.Println("\n\033[1m────────\033[0m")
			fmt.Println("\033[1mSSH User (13 of 15)\033[0m")
			fmt.Println(" • Default SSH user for pulling site databases and files from test server.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["SSH_USER"])
			fmt.Print("       New » ")
			var newSshUser string
			fmt.Scanf("%s", &newSshUser)
			if newSshUser != "" {
				configVars["SSH_USER"] = newSshUser
			}

			fmt.Println("\n\033[1m────────\033[0m")
			fmt.Println("\033[1mTest server deploy key (14 of 15)\033[0m")
			fmt.Println(" • SSH public key used by the test server for pulling code from git.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["TEST_SERVER_DEPLOY_KEY"])
			fmt.Print("       New » ")
			var newTestServerDeployKey string
			fmt.Scanf("%s", &newTestServerDeployKey)
			if newTestServerDeployKey != "" {
				configVars["TEST_SERVER_DEPLOY_KEY"] = newTestServerDeployKey
			}

			fmt.Println("\n\033[1m────────\033[0m")
			fmt.Println("\033[1mTask URL Prefix (15 of 15)\033[0m")
			fmt.Println(" • URL prefix for task detail pages.")
			fmt.Println("")
			fmt.Println("   Current » " + configVars["TASK_URL_PREFIX"])
			fmt.Print("       New » ")
			var newTaskUrlPrefix string
			fmt.Scanf("%s", &newTaskUrlPrefix)
			if newTaskUrlPrefix != "" {
				configVars["TASK_URL_PREFIX"] = newTaskUrlPrefix
			}

			// for k, v := range configVars {
			// 	fmt.Printf("%s: %s\n", k, v)
			// }

			envContents, err := godotenv.Marshal(configVars)
			if err != nil {
				log.Fatal(err)
			}

			err = os.WriteFile(defaultConfigPath, []byte(envContents), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
