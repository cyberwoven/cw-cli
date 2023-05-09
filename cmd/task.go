package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// taskCmd represents the clone command
var taskCmd = &cobra.Command{
	Use:   "task [action]",
	Short: "Set, get, unset the task Id",
	Run: func(cmd *cobra.Command, args []string) {

		// Get repository root.
		repoRoot := getRepoRoot()

		// Concatenate root with .taskId to get full .taskId filepath.
		taskIdFilePath := repoRoot + "/" + ".taskId"

		// Print taskId.
		if len(args) == 0 {
			taskID, err := readTaskID(taskIdFilePath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Task ID:", taskID)
			os.Exit(0)
		}

		// Clear taskId.
		if args[0] == "clear" {
			err := clearTaskId(taskIdFilePath)
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}

		// Write taskId.
		if isStringNumeric(args[0]) {

			err := createTaskIdFileIfNotExists(taskIdFilePath)
			if err != nil {
				log.Fatal(err)
			}

			err = createGitIgnoreIfNotExists(repoRoot)
			if err != nil {
				log.Fatal(err)
			}

			err = addToGitIgnoreIfNotExists("\n\n.taskId")
			if err != nil {
				log.Fatal(err)
			}

			err = createPrepareCommitMsgHook(repoRoot)
			if err != nil {
				log.Fatal(err)
			}

			err = writeTaskID(taskIdFilePath, args[0])
			if err != nil {
				log.Fatal(err)
			}

			os.Exit(0)

		}

	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
}

func isStringNumeric(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

func getRepoRoot() string {
	// execute the `git rev-parse --show-toplevel` command
	comm := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := comm.Output()
	if err != nil {
		panic(err)
	}

	// trim whitespace from the output and convert it to a string
	repoRoot := strings.TrimSpace(string(output))
	return repoRoot
}

func readTaskID(filepath string) (string, error) {
	// read the contents of the file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// convert the data to a string and return it
	taskID := string(data)
	return taskID, nil
}

func clearTaskId(filename string) error {
	// create an empty byte slice to overwrite the file contents
	emptyData := []byte("")

	// write the empty byte slice to the file, overwriting its contents
	err := ioutil.WriteFile(filename, emptyData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func writeTaskID(filename string, taskID string) error {
	// convert the task ID to a byte slice
	taskIDData := []byte(taskID)

	// write the byte slice to the file
	err := ioutil.WriteFile(filename, taskIDData, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Task ID %s written to %s\n", taskID, filename)
	return nil
}

func addToGitIgnoreIfNotExists(filename string) error {
	repoRoot := getRepoRoot()
	gitIgnoreFilePath := repoRoot + "/" + ".gitignore"
	gitIgnoreData, err := ioutil.ReadFile(gitIgnoreFilePath)
	if err != nil {
		return err
	}

	if !strings.Contains(string(gitIgnoreData), filename) {
		newGitIgnoreData := []byte(string(gitIgnoreData) + filename + "\n")
		err := ioutil.WriteFile(".gitignore", newGitIgnoreData, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTaskIdFileIfNotExists(filepath string) error {
	_, err := os.Stat(filepath)
	if err == nil {
		// file exists, no need to create it
		return nil
	}
	if os.IsNotExist(err) {
		// file doesn't exist, create it
		f, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer f.Close()
		return nil
	}
	// some other error occurred
	return err
}

func createGitIgnoreIfNotExists(repoRoot string) error {
	gitIgnorePath := repoRoot + "/" + ".gitignore"

	_, err := os.Stat(gitIgnorePath)
	if err == nil {
		// file exists, no need to create it
		return nil
	}
	if os.IsNotExist(err) {
		// file doesn't exist, create it
		f, err := os.Create(gitIgnorePath)
		if err != nil {
			return err
		}
		defer f.Close()
		return nil
	}
	// some other error occurred
	return err
}

func createPrepareCommitMsgHook(repoRoot string) error {

	gitHookPath := repoRoot + "/" + ".git/hooks/prepare-commit-msg"

	data := []byte(`
		#!/bin/bash

		# if the taskId file doesn't exist, exit
		[ ! -f .taskId ] && exit 0;
		# if the taskId is empty, exit
		[ ! -s .taskId ] && exit 0;
		
		# read taskID into a var
		TASK_ID=$( < .taskId )
		
		# read commit message into a var
		COMMIT_MSG=$( < $1 )
		
		# write the new commit message
		cat << EOT > $1
		$TASK_ID - $COMMIT_MSG
		
		https://cyberwoven.teamwork.com/app/tasks/$TASK_ID
	`)

	// Create or overwrite .git/hooks/prepare-commit-msg
	err := ioutil.WriteFile(gitHookPath, data, 0644)
	if err != nil {
		return err
	}

	err = os.Chmod(gitHookPath, 0755)
	if err != nil {
		return err
	}
	return nil
}
