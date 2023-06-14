package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// taskCmd represents the clone command
var taskClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the current task ID",
	Run: func(cmd *cobra.Command, args []string) {

		if !ctx.IS_GIT_REPO {
			log.Fatal("Operation failed: Project is not a git repository.")
		}
		// Clear taskId.
		err := clearTaskId()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	},
}

func init() {
	taskCmd.AddCommand(taskClearCmd)
}

func clearTaskId() error {
	// create an empty byte slice to overwrite the file contents
	emptyData := []byte("")

	// write the empty byte slice to the file, overwriting its contents
	err := os.WriteFile(TASK_ID_FILEPATH, emptyData, 0644)
	if err != nil {
		return err
	}

	fmt.Println(".taskId successfully cleared.")

	return nil
}
