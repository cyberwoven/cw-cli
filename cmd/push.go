/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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

	// check if repo exists (hopefully not)
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	status, _ := bbApi("GET", url, "")
	if status != 404 {
		fmt.Println("ABORT: Unable to push new site, remote repo already exists.")
		os.Exit(1)
	}

	// create repo
	url = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	payload := fmt.Sprintf(`{
		"scm": "git",
		"project": {
			"key": "%s"
		}
	}`, "DRUP")
	status, _ = bbApi("POST", url, payload)
	if status != 200 {
		fmt.Println("ABORT: Unable to push new site, could not create new remote repo.")
		os.Exit(1)
	}

	// add test deploy key to repo
	url = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/deploy-keys", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	payload = fmt.Sprintf(`{
		"key": "%s",
		"label": "Autopilot"
	}`, ctx.TEST_SERVER_DEPLOY_KEY)
	status, _ = bbApi("POST", url, payload)
	if status != 200 {
		fmt.Println("ABORT: Unable to push new site, could not add deploy key to new remote repo.")
		os.Exit(1)
	}

	// add test webhook to repo
	url = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/hooks", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	payload = fmt.Sprintf(`{
		"description": "Autopilot",
		"url": "%s",
		"active": true,
		"events": [
			"repo:push"
		]
	}`, ctx.BITBUCKET_WEBHOOK_URL)
	status, _ = bbApi("POST", url, payload)
	if status != 200 {
		fmt.Println("ABORT: Unable to push new site, could not add autopilot webhook to new remote repo.")
		os.Exit(1)
	}

	// 1. ensure we have the necessary configs (BB app key, git host, git user, deploykey, webhook url, etc)
	// 2. check to see if remote repo already exists
	// 3. create repo, set deploy key, set webhook url
	// 4. set remote url on local git repo
	// 5. git push
	// 6. forest create-site <domain> over ssh to default test server
}

func bbApi(method string, url string, payload string) (int, string) {
	body := bytes.NewBuffer([]byte(payload))
	req, err := http.NewRequest(method, url, body)
	req.SetBasicAuth(ctx.BITBUCKET_USERNAME, ctx.BITBUCKET_APP_PASSWORD)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)

	return res.StatusCode, string(resBody)
}

func hasGitRemote(dir string) bool {
	gitOutput, err := exec.Command("git", "remote").Output()
	if err != nil {
		fmt.Printf("ABORT: Unable to determine git remote(s) - %s\n%s", err.Error(), gitOutput)
		os.Exit(1)
	}

	return string(gitOutput) != ""
}
