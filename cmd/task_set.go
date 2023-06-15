package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// taskCmd represents the clone command
var taskSetCmd = &cobra.Command{
	Use:   "set [taskId]",
	Short: "Set the task ID",
	Run: func(cmd *cobra.Command, args []string) {

		// Handle argument count errors
		if len(args) == 0 {
			log.Fatal("Operation failed: No task ID provided.  Usage: cw task set [taskId]")
		}
		if len(args) > 1 {
			log.Fatalf("Operation failed: Expected 1 argument (taskId) and received %d.", len(args))
		}

		// Write taskId.
		if isStringNumeric(args[0]) {

			err := createTaskIdFileIfNotExists()
			if err != nil {
				log.Fatal(err)
			}

			err = createGitIgnoreIfNotExists()
			if err != nil {
				log.Fatal(err)
			}

			err = addToGitIgnoreIfNotExists()
			if err != nil {
				log.Fatal(err)
			}

			err = writeTaskID(TASK_ID_FILEPATH, args[0])
			if err != nil {
				log.Fatal(err)
			}

			os.Exit(0)
		} else {
			log.Fatal("Operation failed: Non-numeric task ID provided.")
		}

	},
}

func init() {
	taskCmd.AddCommand(taskSetCmd)
}

func isStringNumeric(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

func writeTaskID(filename string, taskID string) error {
	// convert the task ID to a byte slice
	taskIDData := []byte(taskID)

	// write the byte slice to the file
	err := os.WriteFile(filename, taskIDData, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Task ID %s written to %s\n", taskID, filename)
	return nil
}

func addToGitIgnoreIfNotExists() error {
	const TASK_ID_GITIGNORE_VALUE = ".taskId"

	gitIgnoreFilePath := ctx.PROJECT_ROOT + "/" + ".gitignore"
	gitIgnoreData, err := os.ReadFile(gitIgnoreFilePath)
	if err != nil {
		return err
	}

	if !strings.Contains(string(gitIgnoreData), TASK_ID_GITIGNORE_VALUE) {
		newGitIgnoreData := []byte(string(gitIgnoreData) + "\n\n" + TASK_ID_GITIGNORE_VALUE + "\n")
		err := os.WriteFile(".gitignore", newGitIgnoreData, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTaskIdFileIfNotExists() error {
	_, err := os.Stat(TASK_ID_FILEPATH)
	if err == nil {
		// file exists, no need to create it
		return nil
	}
	if os.IsNotExist(err) {
		// file doesn't exist, create it
		f, err := os.Create(TASK_ID_FILEPATH)
		if err != nil {
			return err
		}
		defer f.Close()
		return nil
	}
	// some other error occurred
	return err
}

func createGitIgnoreIfNotExists() error {
	gitIgnorePath := ctx.PROJECT_ROOT + "/" + ".gitignore"

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
