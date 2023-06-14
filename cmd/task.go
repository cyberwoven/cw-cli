package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// Concatenate root with .taskId to get full .taskId filepath.
var TASK_ID_FILEPATH = ctx.PROJECT_ROOT + "/" + ".taskId"

// taskCmd represents the clone command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Set, get, clear the task ID",
	Run: func(cmd *cobra.Command, args []string) {
		if !ctx.IS_GIT_REPO {
			log.Fatal("Operation failed: Project is not a git repository.")
		}
		// Perform task reading by default
		if len(args) == 0 {
			taskID, err := readTaskId()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Task ID:", taskID)
			os.Exit(0)
		}

	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
