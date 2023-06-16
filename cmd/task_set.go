package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

			err := writeTaskID(TASK_ID_FILEPATH, args[0])
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
