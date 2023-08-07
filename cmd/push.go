/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the current site's database and/or files to the test server",
	Run: func(cmd *cobra.Command, args []string) {
		isFlaggedNew, _ := cmd.Flags().GetBool("new")

		if isFlaggedNew {
			newSite()
		}

		// pushDbCmd.Run(cmd, []string{})
		// pushFilesCmd.Run(cmd, []string{})
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().BoolP("new", "n", false, "Creating the git repo and test site before pushing.")
}

func newSite() {
	if ctx.BITBUCKET_USERNAME == "" || ctx.BITBUCKET_APP_PASSWORD == "" {
		fmt.Println("ABORT: Unable to push new site, Bitbucket API creds are required.\nSee https://bitbucket.org/account/settings/app-passwords/")
		os.Exit(1)
	}

	if ctx.BITBUCKET_WEBHOOK_URL == "" {
		fmt.Println("ABORT: Unable to push, webhook URL is required when creating the remote git repo.")
		os.Exit(1)
	}

	if ctx.TEST_SERVER_DEPLOY_KEY == "" {
		fmt.Println("ABORT: Unable to push, test server deploy key is required when creating the remote git repo.")
		os.Exit(1)
	}

	if !ctx.IS_SITE {
		fmt.Println("ABORT: Unable to push new site, you must be in a site directory.")
		os.Exit(1)
	}

	if !ctx.IS_GIT_REPO {
		fmt.Println("ABORT: Unable to push new site, current site has no local git repo.")
		os.Exit(1)
	}

	if hasGitRemote(ctx.PROJECT_ROOT) {
		fmt.Println("ABORT: Unable to push new site, git repo already has a remote specified.")
		os.Exit(1)
	}

	if ctx.IS_PANTHEON {
		fmt.Println("ABORT: Unable to push new site, Pantheon is not supported.")
		os.Exit(1)
	}

	// url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	// getRepoResponse := bitbucketApiRequest("GET", url)

	// 1. ensure we have the necessary configs (BB app key, git host, git user, deploykey, webhook url, etc)
	// 2. check to see if remote repo already exists
	// 3. create repo, set deploy key, set webhook url
	// 4. set remote url on local git repo
	// 5. git push
	// 6. forest create-site <domain> over ssh to default test server
}

func hasGitRemote(dir string) bool {
	gitOutput, err := exec.Command("git", "remote").Output()
	if err != nil {
		fmt.Printf("ABORT: Unable to determine git remote(s) - %s\n%s", err.Error(), gitOutput)
		os.Exit(1)
	}

	return string(gitOutput) != ""
}
