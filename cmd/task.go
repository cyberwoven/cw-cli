package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NOTE: We are ASSUMING ~/.cw/git-hooks dir is in place.
// TODO: Add a way of hydrating and updating ~/.cw with this necessary dir and content

// Concatenate root with .taskId to get full .taskId filepath.
var TASK_ID_FILEPATH = ctx.PROJECT_ROOT + "/" + ".taskId"

// taskCmd represents the clone command
var taskCmd = &cobra.Command{
	Use:              "task",
	Short:            "Set, get, clear the task ID",
	PersistentPreRun: performPersistentPreRunActions,
	Run: func(cmd *cobra.Command, args []string) {

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

func performPersistentPreRunActions(cmd *cobra.Command, args []string) {
	// Check if git repo
	if !ctx.IS_GIT_REPO {
		log.Fatal("Operation failed: Project is not a git repository.")
	}

	// Create symlink
	err := createHookSymlinkIfNotExists()
	if err != nil {
		log.Fatal("Operation failed: ", err)
	}
}
func init() {
	rootCmd.AddCommand(taskCmd)
}

func createHookSymlinkIfNotExists() error {
	const PREPARE_COMMIT_MSG_HOOK_NAME = "prepare-commit-msg"

	// Define path where the link will go (inside project's .git/hooks/)
	linkPath := filepath.Join(ctx.PROJECT_ROOT, ".git", "hooks", PREPARE_COMMIT_MSG_HOOK_NAME)

	// TODO: Configure .cw/git-hooks dir creation in setup.go
	// TODO: Configure .cw self-update mechanism eventually
	// Get the absolute path of the custom hook file.  We are ASSUMING dir is in place.
	customHookPath := filepath.Join(ctx.DEFAULT_CONFIG_DIR, "git-hooks/task", PREPARE_COMMIT_MSG_HOOK_NAME)

	// Check if the linkPath already exists
	if _, err := os.Lstat(linkPath); err == nil {
		fileInfo, err := os.Lstat(linkPath)
		if err != nil {
			return err
		}

		// Check if it is a regular file
		if fileInfo.Mode().IsRegular() {
			err := fmt.Errorf("Unable to create git hook symlink.  A NON-symlink (regular file) prepare-commit-msg hook already exists at %s.  The file in place is not affiliated with the cw-cli.  The file must be manually deleted before replacing it with a symlink is possible.", filepath.Join(ctx.PROJECT_ROOT, ".git", "hooks", PREPARE_COMMIT_MSG_HOOK_NAME))
			log.Fatal("Operation failed: ", err)
		}

		return nil
	}

	// Create the symbolic link if it doesn't exist
	fmt.Printf("Linking %s to %s", customHookPath, PREPARE_COMMIT_MSG_HOOK_NAME)
	err := os.Symlink(customHookPath, linkPath)
	if err != nil {
		return err
	}
	fmt.Printf("Created git hook symlink from %s to %s\n", customHookPath, PREPARE_COMMIT_MSG_HOOK_NAME)

	return nil
}
