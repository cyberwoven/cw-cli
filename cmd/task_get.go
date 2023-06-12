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
	Short: "Get the task Id",
	Run: func(cmd *cobra.Command, args []string) {

		taskID, err := readTaskID()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Task ID:", taskID)
		os.Exit(0)

	},
}

func init() {
	taskCmd.AddCommand(taskGetCmd)
}

func readTaskID() (string, error) {
	// read the contents of the file
	data, err := os.ReadFile(TASK_ID_FILEPATH)
	if err != nil {
		return "", err
	}

	// convert the data to a string and return it
	taskID := string(data)
	return taskID, nil
}
