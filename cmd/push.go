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
	"strings"

	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the current site's database and/or files to the test server",
	Run: func(cmd *cobra.Command, args []string) {
		isNew := createTestSite()

		if ctx.HAS_DATABASE {
			pushDbCmd.Run(cmd, []string{})
			pushFilesCmd.Run(cmd, []string{})
		} else {
			fmt.Println("This site has no database, so nothing to push.")
		}

		if isNew && ctx.SITE_TYPE == "drupal" {
			url := uliGenerateTestLink()
			uliOpenLink(url)
		}
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func createTestSite() bool {
	// Check to see if the site exists on the Test server.
	// If it does, then we'll return, and simply push the DB and files
	// If it does not, then we'll see if the git remote needs to be created
	remoteSiteExistsCmd := fmt.Sprintf(`[[ -d /var/www/vhosts/%s ]] && echo "YES" || echo "NO"`, ctx.SITE_NAME)
	remoteHost := fmt.Sprintf("%s@%s", ctx.SSH_TEST_USER, ctx.SSH_TEST_HOST)
	output, err := exec.Command("ssh", remoteHost, remoteSiteExistsCmd).Output()
	if err != nil {
		panic(err)
	}

	siteExists := strings.TrimSpace(string(output))
	if siteExists == "YES" {
		return false
	}

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
	fmt.Printf(" - Checking Bitbucket for repo existence [%s/%s]\n", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	status, _ := bbApi("GET", url, "")
	if status != 404 {
		fmt.Println("ABORT: Unable to push new site, remote repo already exists.")
		os.Exit(1)
	}

	// create repo
	fmt.Printf(" - Creating repo [%s/%s]\n", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	url = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	payload := fmt.Sprintf(`{
		"scm": "git",
		"is_private": true,
		"project": {
			"key": "%s"
		}
	}`, "PROJ")
	status, _ = bbApi("POST", url, payload)
	if status != 200 {
		fmt.Printf("ABORT: Unable to push new site, could not create new remote repo [STATUS: %d].\n", status)
		os.Exit(1)
	}

	// add test deploy key to repo
	// fmt.Println(" - Adding deploy key")
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
	// fmt.Println(" - Adding webhook")
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
	if status != 201 {
		fmt.Printf("ABORT: Unable to push new site, could not add autopilot webhook to new remote repo [STATUS:%d].\n", status)
		os.Exit(1)
	}

	// fmt.Println(" - Adding origin remote")
	originUrl := fmt.Sprintf("%s@%s:%s/%s", ctx.GIT_DEFAULT_USER, ctx.GIT_DEFAULT_DOMAIN, ctx.GIT_DEFAULT_WORKSPACE, ctx.SITE_NAME)
	exec.Command("git", "remote", "add", "origin", originUrl).Run()

	fmt.Println(" - Pushing repo to Bitbucket")
	exec.Command("git", "push", "-f", "--set-upstream", "origin", "master").Run()

	fmt.Println(" - Creating test site")
	remoteLoadCmd := fmt.Sprintf("sudo forest create %s", ctx.SITE_NAME)
	exec.Command("ssh", remoteHost, remoteLoadCmd).Run()

	return true
}

func bbApi(method string, url string, payload string) (int, string) {
	body := bytes.NewBuffer([]byte(payload))
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(ctx.BITBUCKET_USERNAME, ctx.BITBUCKET_APP_PASSWORD)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)

	// fmt.Printf("***\n%s\n***\n", resBody)
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
