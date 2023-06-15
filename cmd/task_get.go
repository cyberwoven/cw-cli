package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// taskCmd represents the clone command
var taskGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the task ID",
	Run: func(cmd *cobra.Command, args []string) {

		taskId, err := readTaskId()
		if err != nil {
			log.Fatal(err)
		}
		if taskId != "" {
			fmt.Println("Task ID:", taskId)
		} else {
			fmt.Println("No Task ID set.")
		}
		os.Exit(0)

	},
}

func init() {
	taskCmd.AddCommand(taskGetCmd)
}

func readTaskId() (string, error) {
	// read the contents of the file
	data, err := os.ReadFile(TASK_ID_FILEPATH)
	if err != nil {
		return "", err
	}

	// convert the data to a string and return it
	taskID := string(data)
	return taskID, nil
}
