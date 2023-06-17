package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	Aliases:          []string{"t"},
	PersistentPreRun: performPersistentPreRunActions,
	Run: func(cmd *cobra.Command, args []string) {
		taskGetCmd.Run(cmd, args)
	},
}

func performPersistentPreRunActions(cmd *cobra.Command, args []string) {
	// Check if git repo
	if !ctx.IS_GIT_REPO {
		log.Fatal("Operation failed: Project is not a git repository.")
	}

	// Create .git/hooks dir if not exists
	err := createGitHooksDirInProjectIfNotExists()
	if err != nil {
		log.Fatal("Operation failed: ", err)
	}

	// Create .git/hooks dir if not exists
	err = createTaskIdFileIfNotExists()
	if err != nil {
		log.Fatal("Operation failed: ", err)
	}

	err = createGitIgnoreIfNotExists()
	if err != nil {
		log.Fatal("Operation failed: ", err)
	}

	err = addToGitIgnoreIfNotExists()
	if err != nil {
		log.Fatal("Operation failed: ", err)
	}

	// Create symlink
	err = createHookSymlinkIfNotExists()
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
			err := fmt.Errorf("unable to create git hook symlink.  A NON-symlink (regular file) prepare-commit-msg hook already exists at %s.  The file in place is not affiliated with the cw-cli.  The file must be manually deleted before replacing it with a symlink is possible", filepath.Join(ctx.PROJECT_ROOT, ".git", "hooks", PREPARE_COMMIT_MSG_HOOK_NAME))
			log.Fatal("Operation failed: ", err)
		}

		return nil
	}

	// Create the symbolic link if it doesn't exist
	fmt.Printf("Linking %s to %s\n", customHookPath, PREPARE_COMMIT_MSG_HOOK_NAME)
	err := os.Symlink(customHookPath, linkPath)
	if err != nil {
		return err
	}
	fmt.Printf("Symlinked %s to %s\n", customHookPath, PREPARE_COMMIT_MSG_HOOK_NAME)

	return nil
}
func createGitHooksDirInProjectIfNotExists() error {
	dirPath := filepath.Join(ctx.PROJECT_ROOT, ".git", "hooks")

	// Check if the directory already exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Create the directory
		err := os.Mkdir(dirPath, 0755)
		if err != nil {
			fmt.Println("Operation failed:", err)
			return err
		}

		fmt.Println("Created .git/hooks directory.")
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
		fmt.Println("Created .taskId file")
		defer f.Close()
		return nil
	}
	// some other error occurred
	return err
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
